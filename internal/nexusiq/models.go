package nexusiq

// --- Component Details API Request/Response Models ---

type ComponentDetailsRequest struct {
	Components []ComponentRequestItem `json:"components"`
}

type ComponentRequestItem struct {
	PackageURL string `json:"packageUrl,omitempty"`
}

type ComponentDetailsResponse struct {
	ComponentDetails []ComponentDetail `json:"componentDetails"`
}

type ComponentDetail struct {
	Component          ComponentInfo `json:"component"`
	MatchState         string        `json:"matchState"`
	CatalogDate        string        `json:"catalogDate,omitempty"`
	RelativePopularity *float64      `json:"relativePopularity,omitempty"`
	HygieneRating      string        `json:"hygieneRating,omitempty"`
	SecurityData       SecurityData  `json:"securityData"`
	LicenseData        LicenseData   `json:"licenseData"`
}

type ComponentInfo struct {
	PackageURL          string `json:"packageUrl,omitempty"`
	Hash                string `json:"hash,omitempty"`
	ComponentIdentifier any    `json:"componentIdentifier,omitempty"`
}

type SecurityData struct {
	SecurityIssues []SecurityIssue `json:"securityIssues"`
}

type SecurityIssue struct {
	Source         string  `json:"source"`
	Reference      string  `json:"reference"`
	Severity       float64 `json:"severity"`
	URL            string  `json:"url,omitempty"`
	ThreatCategory string  `json:"threatCategory,omitempty"`
}

type LicenseData struct {
	DeclaredLicenses []License `json:"declaredLicenses"`
	ObservedLicenses []License `json:"observedLicenses"`
}

type License struct {
	LicenseId   string `json:"licenseId"`
	LicenseName string `json:"licenseName"`
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

// --- Remediation API Response Models ---

type RemediationData struct {
	VersionChanges []VersionChange `json:"versionChanges"`
}

type RemediationResponse struct {
	Remediation    *RemediationData `json:"remediation,omitempty"`
	VersionChanges []VersionChange  `json:"versionChanges"`
}

type VersionChange struct {
	Type string            `json:"type"` // e.g., "recommended-non-breaking-with-dependencies"
	Data VersionChangeData `json:"data"`
}

type VersionChangeData struct {
	Component ComponentInfo `json:"component"`
}
