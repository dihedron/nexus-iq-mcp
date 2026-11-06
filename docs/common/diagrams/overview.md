```mermaid
flowchart LR
    client["MCP client (LLM host)"] -- "MCP over stdio" --> server["nexus-iq-mcp serve"]
    server -- "HTTP(S), basic auth" --> iq["Sonatype Nexus IQ REST API"]
```
