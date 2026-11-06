package tools

import (
	"context"
	json "encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dihedron/nexus-iq-mcp/internal/nexusiq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestMCPToolHandlers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v2/components/details", func(w http.ResponseWriter, r *http.Request) {
		_ = json.MarshalWrite(w, nexusiq.ComponentDetailsResponse{
			ComponentDetails: []nexusiq.ComponentDetail{
				{
					Component:  nexusiq.ComponentInfo{PackageURL: "pkg:golang/github.com/gin-gonic/gin@v1.9.1"},
					MatchState: "exact",
				},
			},
		})
	})
	mux.HandleFunc("POST /api/v2/components/versions", func(w http.ResponseWriter, r *http.Request) {
		_ = json.MarshalWrite(w, []string{"v1.9.0", "v1.9.1"})
	})
	mux.HandleFunc("POST /api/v2/components/remediation/application/{appId}", func(w http.ResponseWriter, r *http.Request) {
		_ = json.MarshalWrite(w, nexusiq.RemediationResponse{
			VersionChanges: []nexusiq.VersionChange{
				{Type: "next-no-violations"},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := nexusiq.NewClient(srv.URL, "admin", "admin123")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	mcpServer := server.NewMCPServer("Sonatype Nexus IQ", "1.0.0")
	Register(mcpServer, client)

	ctx := context.Background()

	// Test invalid PURL error handling via get_component_details
	invalidReq := mcp.CallToolRequest{}
	invalidReq.Params.Name = "get_component_details"
	invalidReq.Params.Arguments = map[string]any{"purl": "not-a-valid-purl"}
	tool := mcpServer.GetTool("get_component_details")
	if tool == nil {
		t.Fatal("get_component_details tool not registered")
	}
	res, err := tool.Handler(ctx, invalidReq)
	if err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for invalid PURL")
	}

	// Test valid get_component_details
	validReq := mcp.CallToolRequest{}
	validReq.Params.Name = "get_component_details"
	validReq.Params.Arguments = map[string]any{"purl": "pkg:golang/github.com/gin-gonic/gin@v1.9.1"}
	res, err = tool.Handler(ctx, validReq)
	if err != nil || res.IsError {
		t.Fatalf("expected valid details response, got err=%v, isError=%v", err, res.IsError)
	}

	// Test valid get_component_versions
	versionsTool := mcpServer.GetTool("get_component_versions")
	res, err = versionsTool.Handler(ctx, validReq)
	if err != nil || res.IsError {
		t.Fatalf("expected valid versions response, got err=%v, isError=%v", err, res.IsError)
	}

	// Test valid get_component_remediation with UUID
	remTool := mcpServer.GetTool("get_component_remediation")
	remReq := mcp.CallToolRequest{}
	remReq.Params.Name = "get_component_remediation"
	remReq.Params.Arguments = map[string]any{
		"purl":                   "pkg:golang/github.com/gin-gonic/gin@v1.9.1",
		"application_identifier": "4537e6fe-8265-43f1-8c0d-9f88e4ef1234",
	}
	res, err = remTool.Handler(ctx, remReq)
	if err != nil || res.IsError {
		t.Fatalf("expected valid remediation response, got err=%v, isError=%v", err, res.IsError)
	}
	textContent, ok := res.Content[0].(mcp.TextContent)
	if !ok || !strings.Contains(textContent.Text, "next-no-violations") {
		t.Fatalf("unexpected remediation content: %+v", res.Content)
	}
}

func TestToolsAreAnnotatedReadOnly(t *testing.T) {
	client, err := nexusiq.NewClient("http://localhost:8070", "admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	mcpServer := server.NewMCPServer("test", "0")
	Register(mcpServer, client)

	for _, name := range []string{"get_component_details", "get_component_versions", "get_component_remediation"} {
		a := mcpServer.GetTool(name).Tool.Annotations
		if a.ReadOnlyHint == nil || !*a.ReadOnlyHint {
			t.Errorf("%s: readOnlyHint is not true", name)
		}
		if a.DestructiveHint == nil || *a.DestructiveHint {
			t.Errorf("%s: destructiveHint is not false", name)
		}
		if a.IdempotentHint == nil || !*a.IdempotentHint {
			t.Errorf("%s: idempotentHint is not true", name)
		}
	}
}

// TestHandlersSendTrimmedPURL checks that what goes upstream is the trimmed
// value, not only what passes validation.
func TestHandlersSendTrimmedPURL(t *testing.T) {
	const purl = "pkg:maven/a/b@1"
	var got []string
	record := func(w http.ResponseWriter, r *http.Request) string {
		body, _ := io.ReadAll(r.Body)
		got = append(got, string(body))
		return string(body)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v2/components/details", func(w http.ResponseWriter, r *http.Request) {
		record(w, r)
		_, _ = w.Write([]byte(`{"componentDetails":[{"matchState":"exact"}]}`))
	})
	mux.HandleFunc("POST /api/v2/components/versions", func(w http.ResponseWriter, r *http.Request) {
		record(w, r)
		_, _ = w.Write([]byte(`["1"]`))
	})
	mux.HandleFunc("POST /api/v2/components/remediation/application/{appId}", func(w http.ResponseWriter, r *http.Request) {
		record(w, r)
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := nexusiq.NewClient(srv.URL, "admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	mcpServer := server.NewMCPServer("test", "0")
	Register(mcpServer, client)

	for _, name := range []string{"get_component_details", "get_component_versions", "get_component_remediation"} {
		req := mcp.CallToolRequest{}
		req.Params.Name = name
		req.Params.Arguments = map[string]any{
			"purl":                   "  " + purl + "\n",
			"application_identifier": "4537e6fe-8265-43f1-8c0d-9f88e4ef1234",
		}
		res, err := mcpServer.GetTool(name).Handler(context.Background(), req)
		if err != nil || res.IsError {
			t.Fatalf("%s: err=%v, result=%+v", name, err, res)
		}
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 upstream calls, got %d", len(got))
	}
	for _, body := range got {
		if !strings.Contains(body, `"`+purl+`"`) {
			t.Errorf("upstream body is not trimmed: %s", body)
		}
	}
}
