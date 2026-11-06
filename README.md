# Sonatype Nexus IQ MCP Server

A Model Context Protocol (MCP) server implemented in Go 1.27+ that exposes read-only tools for interacting with the Sonatype Nexus IQ REST API. The server enables LLM clients to query component metadata, discover version histories, and calculate policy-compliant upgrade paths.

## Features & Architecture

- **Read-Only Execution:** All requests to Nexus IQ are strictly limited to `GET` and read-centric `POST` operations.
- **PURL-First Design:** Ecosystem-agnostic Package URLs (PURLs) are used exclusively to identify software components, with proactive validation before upstream calls.
- **Smart Identifier Resolution:** Tools accept human-readable application names, `publicId` values, or internal UUIDs and transparently resolve them to internal Nexus IQ application UUIDs.
- **Stateless Design:** Relies entirely on the upstream Nexus IQ instance as the source of truth. Component details and remediation replies pass through unchanged.

## Building

```bash
make                # lint, then build for the current platform into dist/
make help           # list the other targets
```

The binary has two commands: `nexus-iq-mcp serve` runs the MCP server over stdio, and `nexus-iq-mcp version [--verbose]` prints the build details.

## Environment Configuration

The server requires the following environment variables:

| Variable | Description | Example |
| :--- | :--- | :--- |
| `NEXUS_IQ_URL` | Base URL of the Nexus IQ instance | `http://localhost:8070` |
| `NEXUS_IQ_USERNAME` | API service account username | `admin` |
| `NEXUS_IQ_PASSWORD` | API service account password or access token | `admin123` |

## MCP Tools

1. **`get_component_details`**
   - Retrieves metadata, security vulnerabilities (`securityData`), license data (`licenseData`), and popularity for a specific dependency version.
   - **Parameters:** `purl` (required)
   - **Upstream Endpoint:** `POST /api/v2/components/details`

2. **`get_component_versions`**
   - Retrieves the release history for a dependency to determine available versions.
   - **Parameters:** `purl` (required)
   - **Upstream Endpoint:** `POST /api/v2/components/versions`

3. **`get_component_remediation`**
   - Suggests the optimal upgraded version of a dependency that resolves policy violations for a specific application and lifecycle stage.
   - **Parameters:**
     - `purl` (required): Package URL of the violating component.
     - `application_identifier` (required): Exact application `name`, `publicId`, or internal UUID.
     - `stage_id` (optional): Lifecycle stage: `develop`, `source`, `build`, `stage-release`, `release` or `operate` (defaults to `build`).
   - **Upstream Endpoint:** `POST /api/v2/components/remediation/application/{resolved_internal_id}?stageId={stage_id}`

## Known Vulnerabilities

`make vuln` lists one accepted finding (`ACCEPTED_VULNS` in `custom.mk`) and passes; any other finding fails it.

- **[GO-2026-6911](https://pkg.go.dev/vuln/GO-2026-6911)** (`github.com/mark3labs/mcp-go`, no fix available): the Streamable HTTP client in `client/transport` follows redirects and sends its credentials to the new host. This server uses only the `mcp` and `server` packages, so the vulnerable package is not in the binary. Remove this note when the advisory is fixed.

## Client Configuration & Installation

Always use **absolute paths** to your Go project or compiled binary when configuring MCP clients.

### Option 1: Claude Code (CLI)

```bash
claude mcp add \
  --env NEXUS_IQ_URL="http://localhost:8070" \
  --env NEXUS_IQ_USERNAME="admin" \
  --env NEXUS_IQ_PASSWORD="admin123" \
  --transport stdio \
  nexus-iq \
  -- go run /absolute/path/to/nexus-iq-mcp/cmd/nexus-iq-mcp serve
```

### Option 2: Cline (VS Code / Zed)

Add the `nexus-iq` server block to your `cline_mcp_settings.json`:

```json
{
  "mcpServers": {
    "nexus-iq": {
      "command": "go",
      "args": ["run", "/absolute/path/to/nexus-iq-mcp/cmd/nexus-iq-mcp", "serve"],
      "env": {
        "NEXUS_IQ_URL": "http://localhost:8070",
        "NEXUS_IQ_USERNAME": "admin",
        "NEXUS_IQ_PASSWORD": "admin123"
      }
    }
  }
}
```
