# Installation and configuration

## Installing

Install the `deb` or `rpm` package from the release page, or unpack the `tar.gz` archive for your platform. The binary is `/usr/bin/nexus-iq-mcp` when installed from a package.

Check the signatures of a release before installing it: see "Verifying a Release" in the project's README.

Verify the installation:

```bash
nexus-iq-mcp version
```

## Configuring

`nexus-iq-mcp serve` reads its settings from the environment. All three variables are required.

| Variable | Description |
| :--------------------- | :--------------------------------------------- |
| `NEXUS_IQ_URL` | Base URL of the Nexus IQ instance |
| `NEXUS_IQ_USERNAME` | API service account username |
| `NEXUS_IQ_PASSWORD` | API service account password or access token |

## Registering with an MCP client

For Claude Code:

```bash
claude mcp add \
  --env NEXUS_IQ_URL="https://iq.example.com" \
  --env NEXUS_IQ_USERNAME="service-account" \
  --env NEXUS_IQ_PASSWORD="..." \
  --transport stdio \
  nexus-iq \
  -- /usr/bin/nexus-iq-mcp serve
```
