package server

const (
	// General Routes
	HEALTH_CHECK = "/health-check"

	// User Routes
	USER = "user"

	// Customer Routes
	CUSTOMER = "customer"

	DETAIL   = "detail"
	PASSWORD = "password"

	// Authentication Routes
	SIGN_UP       = "/sign-up"
	SIGN_IN       = "/sign-in"
	REFRESH_TOKEN = "/refresh-token"

	// Account Routes (Reused for Customer and Vendor)
	ACCOUNT          = "/account"
	PROFILE          = "/profile"
	PROFILE_PASSWORD = "/profile/password"
)
