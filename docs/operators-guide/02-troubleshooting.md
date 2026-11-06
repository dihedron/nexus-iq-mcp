# Troubleshooting

`NEXUS_IQ_URL, NEXUS_IQ_USERNAME, and NEXUS_IQ_PASSWORD environment variables are required`
:   The MCP client did not pass all three variables to the server. Check its configuration.

`authentication to Nexus IQ failed (401 Unauthorized)`
:   Nexus IQ rejected the credentials. Check `NEXUS_IQ_USERNAME` and `NEXUS_IQ_PASSWORD`.

`application not found in Nexus IQ`
:   No application has that name, `publicId` or UUID. Check the `application_identifier` argument.

`ambiguous application identifier`
:   Several applications have that name. Use one of the `publicId` values the error lists.

`invalid base URL`
:   `NEXUS_IQ_URL` must be an `http` or `https` URL without credentials, query or fragment. Put the credentials in `NEXUS_IQ_USERNAME` and `NEXUS_IQ_PASSWORD`.

`redirected to ...: point NEXUS_IQ_URL at the final address`
:   Nexus IQ or a proxy answered with a redirect, which the server does not follow. Set `NEXUS_IQ_URL` to the address shown.
