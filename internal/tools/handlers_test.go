package tools

import (
	"context"
	json "encoding/json/v2"
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
