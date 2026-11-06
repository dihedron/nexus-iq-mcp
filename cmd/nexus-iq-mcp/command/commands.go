package command

import (
	"github.com/dihedron/nexus-iq-mcp/internal/command/serve"
	"github.com/dihedron/nexus-iq-mcp/internal/command/version"
)

// Commands is the set of root commands.
type Commands struct {
	// Version prints the program version information.
	Version version.Version `command:"version" alias:"v" description:"Print program version information."`
	// Serve runs the MCP server over stdio.
	Serve serve.Serve `command:"serve" description:"Run the MCP server over stdio."`
}
