package constants

import "customer/sigmatech/config"

var Config *config.ServiceConfig

const (
	//Header constants
	AUTHORIZATION      = "Authorization"
	BEARER             = "Bearer "
	CTK_CLAIM_KEY      = CONTEXT_KEY("claims")
	CORRELATION_KEY_ID = CORRELATION_KEY("X-Correlation-ID")
	DEFAULT_ID         = 1
	STATUS_CODE        = "status_code"
)

type (
	ENVIRONMENT     string
	CONTEXT_KEY     string
	CORRELATION_KEY string
)

func (c CONTEXT_KEY) String() string {
	return string(c)
}

func (c CORRELATION_KEY) String() string {
	return string(c)
}

var (
	Local       ENVIRONMENT = "local"
	Development ENVIRONMENT = "development"
	Staging     ENVIRONMENT = "staging"
	Production  ENVIRONMENT = "production"
)

func (c ENVIRONMENT) String() string {
	return string(c)
}
