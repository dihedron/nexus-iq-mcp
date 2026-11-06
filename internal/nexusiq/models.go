package nexusiq

import "encoding/json/jsontext"

// Replies about components pass through as raw JSON, so that the client sees
// every field Nexus IQ returns; only what this server itself reads is modelled.

// --- Component Details API Request/Response Models ---

type ComponentDetailsRequest struct {
	Components []ComponentRequestItem `json:"components"`
}

type ComponentRequestItem struct {
	PackageURL string `json:"packageUrl,omitempty"`
}

type ComponentDetailsResponse struct {
	ComponentDetails []jsontext.Value `json:"componentDetails"`
}

// --- Component Versions & Remediation API Request Models ---

type PurlRequest struct {
	PackageURL string `json:"packageUrl"`
}

// --- Applications API Response Models ---

type ApplicationsResponse struct {
	Applications []Application `json:"applications"`
}

type Application struct {
	ID       string `json:"id"`
	PublicID string `json:"publicId"`
	Name     string `json:"name"`
}
