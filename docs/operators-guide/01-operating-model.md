# Operating model

The MCP client starts and stops `nexus-iq-mcp serve`, one process per client session. The server keeps no state, so there is nothing to back up.

Standard output carries the MCP protocol. Logs go to standard error, or to a file:

| Variable | Values |
| :--------------------- | :--------------------------------------------- |
| `NEXUS_IQ_MCP_LOG_LEVEL` | `debug`, `info`, `warn`, `error`, `off` (default: no logs) |
| `NEXUS_IQ_MCP_LOG_STREAM` | `stderr` (default), `file` |

`nexus-iq-mcp version --verbose` prints the version and build details of the binary.

At level `debug`, each request to Nexus IQ logs its method, endpoint, status and duration. Failed requests log at level `error`. Credentials and response bodies are never logged. A plain `http` URL to a host other than loopback logs a warning at level `warn`.
