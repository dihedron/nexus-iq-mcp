package nexusiq

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// isUUID checks whether identifier is a canonical 36-character hyphenated UUID.
func isUUID(identifier string) bool {
	if len(identifier) != 36 {
		return false
	}
	return uuid.Validate(identifier) == nil
}

// ResolveApplicationID translates a generic identifier (Name, Public ID, or UUID) into the internal Application ID.
func (c *Client) ResolveApplicationID(ctx context.Context, identifier string) (string, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return "", fmt.Errorf("%w: application_identifier cannot be empty", ErrApplicationNotFound)
	}

	// 1. If application_identifier is a valid UUID, use it directly.
	if isUUID(identifier) {
		return identifier, nil
	}

	// 2. Query GET /api/v2/applications?publicId={application_identifier}. If a match is found, extract id.
	var pubIDResp ApplicationsResponse
	endpoint := fmt.Sprintf("/api/v2/applications?publicId=%s", url.QueryEscape(identifier))

	if err := c.DoRequest(ctx, http.MethodGet, endpoint, nil, &pubIDResp); err != nil {
		if errors.Is(err, ErrUnauthorized) {
			return "", err
		}
		// Other errors (such as 404 Not Found) fall through to full application list lookup
	} else if len(pubIDResp.Applications) > 0 {
		return pubIDResp.Applications[0].ID, nil
	}

	// 3. Fallback: Query GET /api/v2/applications, iterate, and find the application where name == application_identifier.
	var allAppsResp ApplicationsResponse
	if err := c.DoRequest(ctx, http.MethodGet, "/api/v2/applications", nil, &allAppsResp); err != nil {
		return "", fmt.Errorf("failed to fetch applications list for resolution: %w", err)
	}

	for _, app := range allAppsResp.Applications {
		if app.Name == identifier || app.PublicID == identifier || app.ID == identifier {
			return app.ID, nil
		}
	}

	// 4. If unresolved, return a structured error guiding the LLM that the application_identifier does not exist.
	return "", fmt.Errorf("%w: application identifier %q does not exist in the configured Nexus IQ instance (verify the application name, publicId, or internal UUID)", ErrApplicationNotFound, identifier)
}

// GetComponentRemediation fetches the suggested upgrade paths for a component in a specific application context.
func (c *Client) GetComponentRemediation(ctx context.Context, purl, appID, stageID string) (*RemediationResponse, error) {
	if err := ValidatePURL(purl); err != nil {
		return nil, err
	}
	if strings.TrimSpace(stageID) == "" {
		stageID = "build"
	}

	reqPayload := PurlRequest{
		PackageURL: purl,
	}

	var respPayload RemediationResponse

	endpoint := fmt.Sprintf("/api/v2/components/remediation/application/%s?stageId=%s",
		url.PathEscape(appID),
		url.QueryEscape(stageID),
	)

	err := c.DoRequest(ctx, http.MethodPost, endpoint, reqPayload, &respPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remediation paths: %w", err)
	}

	// Normalize nested remediation.versionChanges if returned by Nexus IQ v2 API
	if respPayload.Remediation != nil && len(respPayload.VersionChanges) == 0 {
		respPayload.VersionChanges = respPayload.Remediation.VersionChanges
		respPayload.Remediation = nil
	}

	return &respPayload, nil
}
