package nexusiq

import (
	"context"
	json "encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidatePURL(t *testing.T) {
	valid := []string{
		"pkg:golang/github.com/gin-gonic/gin@v1.9.1",
		"pkg:maven/org.apache.commons/commons-lang3@3.12.0",
		"pkg:npm/%40angular/core@16.0.0",
		"pkg:pypi/requests",
	}
	for _, purl := range valid {
		if err := ValidatePURL(purl); err != nil {
			t.Errorf("expected valid PURL %q, got error: %v", purl, err)
		}
	}

	invalid := []string{
		"",
		"   ",
		"golang/github.com/gin-gonic/gin@v1.9.1",
		"http://example.com/pkg",
		"pkg:",
		"pkg:golang",
		"pkg:golang/",
		"pkg:/gin@v1.0.0",
		"pkg:golang/gin@",
		"pkg:golang/gin @v1.0.0",
	}
	for _, purl := range invalid {
		if err := ValidatePURL(purl); !errors.Is(err, ErrInvalidPURL) {
			t.Errorf("expected ErrInvalidPURL for %q, got: %v", purl, err)
		}
	}
}

func TestClientAndEndpoints(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v2/components/details", func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var req ComponentDetailsRequest
		if err := json.UnmarshalRead(r.Body, &req); err != nil || len(req.Components) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp := ComponentDetailsResponse{
			ComponentDetails: []ComponentDetail{
				{
					Component:  ComponentInfo{PackageURL: req.Components[0].PackageURL},
					MatchState: "exact",
					SecurityData: SecurityData{
						SecurityIssues: []SecurityIssue{
							{Source: "cve", Reference: "CVE-2023-0001", Severity: 7.5},
						},
					},
					LicenseData: LicenseData{
						DeclaredLicenses: []License{{LicenseId: "MIT", LicenseName: "MIT License"}},
					},
				},
			},
		}
		_ = json.MarshalWrite(w, resp)
	})

	mux.HandleFunc("POST /api/v2/components/versions", func(w http.ResponseWriter, r *http.Request) {
		var req PurlRequest
		if err := json.UnmarshalRead(r.Body, &req); err != nil || req.PackageURL == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.MarshalWrite(w, []string{"v1.9.0", "v1.9.1", "v1.10.0"})
	})

	mux.HandleFunc("GET /api/v2/applications", func(w http.ResponseWriter, r *http.Request) {
		pubID := r.URL.Query().Get("publicId")
		if pubID == "my-public-app" {
			_ = json.MarshalWrite(w, ApplicationsResponse{
				Applications: []Application{
					{ID: "4537e6fe-8265-43f1-8c0d-9f88e4ef1234", PublicID: "my-public-app", Name: "My Public App"},
				},
			})
			return
		}
		if pubID != "" {
			_ = json.MarshalWrite(w, ApplicationsResponse{Applications: []Application{}})
			return
		}
		_ = json.MarshalWrite(w, ApplicationsResponse{
			Applications: []Application{
				{ID: "9999e6fe-8265-43f1-8c0d-9f88e4ef5678", PublicID: "other-app", Name: "My Named Application"},
			},
		})
	})

	mux.HandleFunc("POST /api/v2/components/remediation/application/{appId}", func(w http.ResponseWriter, r *http.Request) {
		stageID := r.URL.Query().Get("stageId")
		if stageID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.MarshalWrite(w, map[string]any{
			"remediation": map[string]any{
				"versionChanges": []VersionChange{
					{
						Type: "next-no-violations",
						Data: VersionChangeData{
							Component: ComponentInfo{PackageURL: "pkg:golang/github.com/gin-gonic/gin@v1.10.0"},
						},
					},
				},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewClient(srv.URL, "admin", "secret")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()

	// 1. Component Details
	details, err := client.GetComponentDetails(ctx, "pkg:golang/github.com/gin-gonic/gin@v1.9.1")
	if err != nil {
		t.Fatalf("GetComponentDetails failed: %v", err)
	}
	if details.MatchState != "exact" || len(details.SecurityData.SecurityIssues) != 1 {
		t.Fatalf("unexpected details: %+v", details)
	}

	// 2. Component Versions
	versions, err := client.GetComponentVersions(ctx, "pkg:golang/github.com/gin-gonic/gin@v1.9.1")
	if err != nil {
		t.Fatalf("GetComponentVersions failed: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}

	// 3. Resolve Application ID by UUID directly
	uuidStr := "11111111-2222-3333-4444-555555555555"
	resolved, err := client.ResolveApplicationID(ctx, uuidStr)
	if err != nil || resolved != uuidStr {
		t.Fatalf("expected direct UUID resolution %s, got %s (err=%v)", uuidStr, resolved, err)
	}

	// 4. Resolve Application ID by publicId
	resolved, err = client.ResolveApplicationID(ctx, "my-public-app")
	if err != nil || resolved != "4537e6fe-8265-43f1-8c0d-9f88e4ef1234" {
		t.Fatalf("unexpected publicId resolution: %s (err=%v)", resolved, err)
	}

	// 5. Resolve Application ID by Name fallback
	resolved, err = client.ResolveApplicationID(ctx, "My Named Application")
	if err != nil || resolved != "9999e6fe-8265-43f1-8c0d-9f88e4ef5678" {
		t.Fatalf("unexpected name resolution: %s (err=%v)", resolved, err)
	}

	// 6. Resolve Application ID Not Found
	_, err = client.ResolveApplicationID(ctx, "nonexistent-app")
	if !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf("expected ErrApplicationNotFound, got: %v", err)
	}

	// 7. Remediation
	rem, err := client.GetComponentRemediation(ctx, "pkg:golang/github.com/gin-gonic/gin@v1.9.1", resolved, "build")
	if err != nil {
		t.Fatalf("GetComponentRemediation failed: %v", err)
	}
	if len(rem.VersionChanges) != 1 || rem.VersionChanges[0].Type != "next-no-violations" {
		t.Fatalf("unexpected remediation response: %+v", rem)
	}

	// 8. Unauthorized client
	badClient, _ := NewClient(srv.URL, "admin", "wrong")
	_, err = badClient.GetComponentDetails(ctx, "pkg:golang/github.com/gin-gonic/gin@v1.9.1")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}
