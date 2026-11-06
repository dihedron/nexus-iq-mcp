package nexusiq

import (
	"context"
	"encoding/json/jsontext"
	"fmt"
	"net/http"
)

// GetComponentDetails queries Nexus IQ for metadata, security, and license info for a specific PURL.
func (c *Client) GetComponentDetails(ctx context.Context, purl string) (jsontext.Value, error) {
	// Construct the POST payload exactly as Nexus IQ expects it
	reqPayload := ComponentDetailsRequest{
		Components: []ComponentRequestItem{
			{PackageURL: purl},
		},
	}

	var respPayload ComponentDetailsResponse

	// Execute the HTTP request to the upstream API
	err := c.doRequest(ctx, http.MethodPost, "/api/v2/components/details", reqPayload, &respPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch component details from Nexus IQ: %w", err)
	}

	// The API returns an array; extract our specific component
	if len(respPayload.ComponentDetails) == 0 {
		return nil, fmt.Errorf("no details found for purl: %s", purl)
	}

	return respPayload.ComponentDetails[0], nil
}

// GetComponentVersions retrieves the complete list of known versions for a given dependency.
// It uses POST /api/v2/components/versions as required by the Nexus IQ API.
func (c *Client) GetComponentVersions(ctx context.Context, purl string) ([]string, error) {
	reqPayload := PurlRequest{
		PackageURL: purl,
	}

	// The API returns a direct JSON array of version strings
	var versions []string

	err := c.doRequest(ctx, http.MethodPost, "/api/v2/components/versions", reqPayload, &versions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch component versions from Nexus IQ: %w", err)
	}

	return versions, nil
}
