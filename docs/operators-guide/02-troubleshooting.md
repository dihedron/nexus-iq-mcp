# Troubleshooting

`NEXUS_IQ_URL, NEXUS_IQ_USERNAME, and NEXUS_IQ_PASSWORD environment variables are required`
:   The MCP client did not pass all three variables to the server. Check its configuration.

`authentication to Nexus IQ failed (401 Unauthorized)`
:   Nexus IQ rejected the credentials. Check `NEXUS_IQ_USERNAME` and `NEXUS_IQ_PASSWORD`.

`application not found in Nexus IQ`
:   No application has that name, `publicId` or UUID. Check the `application_identifier` argument.
