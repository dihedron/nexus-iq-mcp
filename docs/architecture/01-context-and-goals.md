# Context and goals

`nexus-iq-mcp` lets an LLM client query component metadata, version histories and policy-compliant upgrade paths from Sonatype Nexus IQ.

- **Read-only.** Requests are limited to `GET` and read-centric `POST` operations.
- **PURL-first.** Package URLs identify components, and are validated before any upstream call.
- **Stateless.** The upstream Nexus IQ instance is the only source of truth. Component details and remediation replies pass through unchanged, so no field is lost.
- **No redirects.** The client never follows a redirect, so the password cannot follow one to another port or to plain `http`.
