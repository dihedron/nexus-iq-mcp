package tools

import (
	"context"
	"encoding/json/jsontext"
	json "encoding/json/v2" // Using Go 1.27's json/v2
	"fmt"
	"strings"

	"github.com/dihedron/nexus-iq-mcp/internal/nexusiq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Register attaches the Nexus IQ tools to the MCP server instance
func Register(s *server.MCPServer, client *nexusiq.Client) {

	// Tool 1: Component Details
	detailsTool := mcp.NewTool("get_component_details",
		mcp.WithDescription("Retrieves metadata, security vulnerabilities, license data, and popularity for a specific dependency version."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("purl", mcp.Required(), mcp.Description("The Package URL of the component (e.g., pkg:golang/github.com/gin-gonic/gin@v1.9.1)")),
	)

	s.AddTool(detailsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		purl, err := request.RequireString("purl")
		purl = strings.TrimSpace(purl)
		if err != nil || purl == "" {
			return mcp.NewToolResultError("purl argument is missing or invalid"), nil
		}
		if err := nexusiq.ValidatePURL(purl); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		details, err := client.GetComponentDetails(ctx, purl)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		responseBytes, err := json.Marshal(details, jsontext.WithIndent("  "))
		if err != nil {
			return mcp.NewToolResultError("failed to serialize Nexus IQ response"), nil
		}

		return mcp.NewToolResultText(string(responseBytes)), nil
	})

	// Tool 2: Component Versions
	versionsTool := mcp.NewTool("get_component_versions",
		mcp.WithDescription("Retrieves the list of all known versions for a given dependency to determine the latest available version."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("purl", mcp.Required(), mcp.Description("The Package URL of the component (e.g., pkg:golang/github.com/gin-gonic/gin@v1.9.1)")),
	)

	s.AddTool(versionsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		purl, err := request.RequireString("purl")
		purl = strings.TrimSpace(purl)
		if err != nil || purl == "" {
			return mcp.NewToolResultError("purl argument is missing or invalid"), nil
		}
		if err := nexusiq.ValidatePURL(purl); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		versions, err := client.GetComponentVersions(ctx, purl)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		responseBytes, err := json.Marshal(versions, jsontext.WithIndent("  "))
		if err != nil {
			return mcp.NewToolResultError("failed to serialize Nexus IQ response"), nil
		}

		return mcp.NewToolResultText(string(responseBytes)), nil
	})

	// Tool 3: Component Remediation
	remediationTool := mcp.NewTool("get_component_remediation",
		mcp.WithDescription("Suggests the optimal upgraded version of a dependency that resolves policy violations for a specific application context."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("purl", mcp.Required(), mcp.Description("The Package URL of the violating component.")),
		mcp.WithString("application_identifier", mcp.Required(), mcp.Description("The exact name, publicId, or internal UUID of the application in Nexus IQ.")),
		mcp.WithString("stage_id", mcp.Description("The lifecycle stage (e.g., build, release, operate). Defaults to 'build'.")),
	)

	s.AddTool(remediationTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		purl, err := request.RequireString("purl")
		purl = strings.TrimSpace(purl)
		if err != nil || purl == "" {
			return mcp.NewToolResultError("purl argument is missing or invalid"), nil
		}
		if err := nexusiq.ValidatePURL(purl); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		appIdentifier, err := request.RequireString("application_identifier")
		if err != nil || strings.TrimSpace(appIdentifier) == "" {
			return mcp.NewToolResultError("application_identifier argument is missing or invalid"), nil
		}

		stageID := strings.TrimSpace(request.GetString("stage_id", "build"))
		if stageID == "" {
			stageID = "build"
		}

		internalAppID, err := client.ResolveApplicationID(ctx, appIdentifier)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Application Resolution Error: %v", err)), nil
		}

		remediation, err := client.GetComponentRemediation(ctx, purl, internalAppID, stageID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		responseBytes, err := json.Marshal(remediation, jsontext.WithIndent("  "))
		if err != nil {
			return mcp.NewToolResultError("failed to serialize Nexus IQ remediation response"), nil
		}

		return mcp.NewToolResultText(string(responseBytes)), nil
	})
}
