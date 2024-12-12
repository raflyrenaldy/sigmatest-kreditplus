package server

const (
	// General Routes
	HEALTH_CHECK = "/health-check"

	// Customer Routes
	CUSTOMER = "/customer"
	LIMIT    = "limit"

	PASSWORD = "password"

	// Transaction Routes
	TRANSACTION = "transaction"

	// Authentication Routes
	SIGN_UP       = "/sign-up"
	SIGN_IN       = "/sign-in"
	REFRESH_TOKEN = "/refresh-token"

	// Account Routes (Reused for Customer)
	ACCOUNT          = "/account"
	PROFILE          = "/profile"
	PROFILE_PASSWORD = "/profile/password"
)
