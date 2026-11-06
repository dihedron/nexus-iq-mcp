package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/dihedron/nexus-iq-mcp/internal/nexusiq"
	"github.com/dihedron/nexus-iq-mcp/internal/tools"
	"github.com/dihedron/nexus-iq-mcp/pkg/metadata"
	"github.com/mark3labs/mcp-go/server"
)

// Serve is the command that runs the MCP server over stdio. The connection
// settings come only from the environment: a password on the command line
// would show in the process list.
type Serve struct{}

// Execute is the real implementation of the Serve command.
func (cmd *Serve) Execute(args []string) error {
	baseURL := os.Getenv("NEXUS_IQ_URL")
	username := os.Getenv("NEXUS_IQ_USERNAME")
	password := os.Getenv("NEXUS_IQ_PASSWORD")

	if baseURL == "" || username == "" || password == "" {
		return errors.New("NEXUS_IQ_URL, NEXUS_IQ_USERNAME, and NEXUS_IQ_PASSWORD environment variables are required")
	}

	client, err := nexusiq.NewClient(baseURL, username, password)
	if err != nil {
		return fmt.Errorf("failed to initialize Nexus IQ client: %w", err)
	}

	mcpServer := server.NewMCPServer("Sonatype Nexus IQ", metadata.Version)
	tools.Register(mcpServer, client)

	slog.Info("serving MCP over stdio", "nexus_iq_url", client.BaseURL)
	// SIGTERM and SIGINT cancel the server's context: that is a normal stop
	if err := server.ServeStdio(mcpServer); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
