package ipgeolocation

// URL constants.
const (
	// Base URL
	BaseURL = "https://api.ipgeolocation.io"

	// Endpoints
	GetLocationInfoEndpoint     = "/ipgeo"
	GetLocationInfoBulkEndpoint = "/ipgeo-bulk"
	GetTimezoneInfoEndpoint     = "/timezone"
	GetUserAgentInfoEndpoint    = "/user-agent"
	GetAstronomyInfoEndpoint    = "/astronomy"
)

// Parameter constants.
const (
	// Global parameters
	APIKeyParam = "apiKey"

	// GetLocationInfo parameters
	IPAddressParam = "ip"
	LanguageParam  = "lang"
	FieldsParam    = "fields"
	ExcludesParam  = "excludes"
	IncludeParam   = "include"
)
