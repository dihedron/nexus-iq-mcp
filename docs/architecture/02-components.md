# Components

![The MCP client, the server and Nexus IQ](overview.pdf)

- `cmd/nexus-iq-mcp` holds the entry point and the command line.
- `internal/command` implements the `serve` and `version` commands.
- `internal/tools` registers the MCP tools and validates their arguments.
- `internal/nexusiq` is the HTTP client for the Nexus IQ REST API.
- `pkg/metadata` holds the build metadata set at link time.
