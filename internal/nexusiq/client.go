package nexusiq

import (
	"bytes"
	"context"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// ErrInvalidPURL indicates that the provided Package URL is malformed.
	ErrInvalidPURL = errors.New("invalid Package URL (PURL)")
	// ErrUnauthorized indicates an upstream 401/403 authentication or authorization failure.
	ErrUnauthorized = errors.New("authentication to Nexus IQ failed (401 Unauthorized): verify NEXUS_IQ_USERNAME and NEXUS_IQ_PASSWORD configuration")
	// ErrApplicationNotFound indicates the requested application identifier does not exist in Nexus IQ.
	ErrApplicationNotFound = errors.New("application not found in Nexus IQ")
	// ErrNotFound indicates an upstream 404 response.
	ErrNotFound = errors.New("not found")
	// ErrAmbiguousApplication indicates that the identifier matches the names of several applications.
	ErrAmbiguousApplication = errors.New("ambiguous application identifier")
)

// maxErrorBody is how much of an upstream error body reaches the client: a
// proxy's error page must not fill the LLM's context.
const maxErrorBody = 4096

// ValidatePURL proactively validates that a string is a well-formed Package URL (pkg:<type>/<name>@<version>).
func ValidatePURL(purl string) error {
	if purl == "" {
		return fmt.Errorf("%w: purl cannot be empty", ErrInvalidPURL)
	}
	if strings.ContainsAny(purl, " \t\r\n") {
		return fmt.Errorf("%w: %q must not contain whitespace", ErrInvalidPURL, purl)
	}
	if !strings.HasPrefix(purl, "pkg:") {
		return fmt.Errorf("%w: %q must start with the 'pkg:' scheme (e.g., pkg:golang/github.com/gin-gonic/gin@v1.9.1)", ErrInvalidPURL, purl)
	}

	remainder := strings.TrimPrefix(purl, "pkg:")
	remainder = strings.TrimLeft(remainder, "/")
	if remainder == "" {
		return fmt.Errorf("%w: %q is missing package type and name", ErrInvalidPURL, purl)
	}

	// Strip qualifiers (?) and subpath (#) for structural validation
	if idx := strings.IndexAny(remainder, "?#"); idx != -1 {
		remainder = remainder[:idx]
	}

	// Check optional @version
	var version string
	hasVersion := false
	if atIdx := strings.LastIndex(remainder, "@"); atIdx != -1 {
		hasVersion = true
		version = remainder[atIdx+1:]
		remainder = remainder[:atIdx]
		if version == "" {
			return fmt.Errorf("%w: %q has an empty version after '@'", ErrInvalidPURL, purl)
		}
	}
	_ = hasVersion

	slashIdx := strings.Index(remainder, "/")
	if slashIdx <= 0 || slashIdx == len(remainder)-1 {
		return fmt.Errorf("%w: %q must include both package type and name separated by '/' (e.g., pkg:golang/github.com/gin-gonic/gin@v1.9.1)", ErrInvalidPURL, purl)
	}

	pkgType := remainder[:slashIdx]
	pkgName := strings.Trim(remainder[slashIdx+1:], "/")
	if pkgType == "" || pkgName == "" {
		return fmt.Errorf("%w: %q must include a non-empty package type and name", ErrInvalidPURL, purl)
	}

	return nil
}

// Client is a stateless HTTP client for Sonatype Nexus IQ
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

// NewClient initializes a new Nexus IQ API client with basic auth credentials
func NewClient(baseURL, username, password string) (*Client, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		// a *url.Error quotes the whole URL, which may hold a password
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid base URL %q: scheme must be http or https", u.Redacted())
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid base URL %q: host is required", u.Redacted())
	}
	if u.User != nil {
		return nil, fmt.Errorf("invalid base URL %q: credentials belong in NEXUS_IQ_USERNAME and NEXUS_IQ_PASSWORD", u.Redacted())
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("invalid base URL: it must not have a query or a fragment")
	}
	if sendsPasswordInClear(u) {
		slog.Warn("NEXUS_IQ_URL uses plain http: the password travels unencrypted", "host", u.Host)
	}
	// Normalize base URL by removing trailing slashes
	u.Path = strings.TrimRight(u.Path, "/")

	return &Client{
		BaseURL:  u.String(),
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // Prevent hanging requests
			// Go keeps the password on a redirect to another port, to a
			// subdomain or from https to http: a redirect is an error instead
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// sendsPasswordInClear reports whether basic auth to u would cross a network
// unencrypted.
func sendsPasswordInClear(u *url.URL) bool {
	host := u.Hostname()
	return u.Scheme == "http" && host != "localhost" && !net.ParseIP(host).IsLoopback()
}

// DoRequest executes a read-only HTTP request against the Nexus IQ API
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, payload any, response any) error {
	if method != http.MethodGet && method != http.MethodPost {
		return fmt.Errorf("unsupported HTTP method %q: server is strictly read-only (only GET and POST allowed)", method)
	}

	var bodyReader io.Reader

	// Strictly marshal payload if present using json/v2
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	reqURL := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Inject Authentication and Content-Type
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		slog.Error("Nexus IQ request failed", "method", method, "endpoint", endpoint, "error", err)
		return fmt.Errorf("http request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() // read-only: a close error loses nothing

	slog.Debug("Nexus IQ request", "method", method, "endpoint", endpoint, "status", resp.StatusCode, "duration", time.Since(start))
	// 404 is a normal answer to a lookup; the error body may hold anything, so it is not logged
	if (resp.StatusCode < 200 || resp.StatusCode >= 300) && resp.StatusCode != http.StatusNotFound {
		slog.Error("Nexus IQ returned an error", "method", method, "endpoint", endpoint, "status", resp.StatusCode)
	}

	// Translate 401 Unauthorized into a descriptive configuration error
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	// Handle other non-2xx responses proactively
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if location := resp.Header.Get("Location"); location != "" {
			return fmt.Errorf("api error (status %d): redirected to %s: point NEXUS_IQ_URL at the final address", resp.StatusCode, location)
		}
		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		if err != nil {
			slog.Warn("failed to read the error body", "endpoint", endpoint, "error", err)
		}
		apiErr := fmt.Errorf("api error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("%w: %w", ErrNotFound, apiErr)
		}
		return apiErr
	}

	// Strictly unmarshal response if expected using json/v2
	if response != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if err := json.Unmarshal(bodyBytes, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
