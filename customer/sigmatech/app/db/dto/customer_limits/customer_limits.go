package customer_limits

import (
	"github.com/google/uuid"
	"time"
)

const (
	TABLE_NAME             = "customer_limits"
	COLUM_UUID             = "uuid"
	COLUMN_CUSTOMER_UUID   = "customer_uuid"
	COLUMN_TERM            = "term"
	COLUMN_STATUS          = "status"
	COLUMN_AMOUNT_LIMIT    = "amount_limit"
	COLUMN_REMAINING_LIMIT = "remaining_limit"
	COLUMN_CREATED_AT      = "created_at"
	COLUMN_UPDATED_AT      = "updated_at"
)

type CustomerLimit struct {
	Uuid           uuid.UUID `json:"uuid"`
	CustomerUuid   uuid.UUID `json:"customer_uuid"`
	Term           int       `json:"term"`
	Status         *bool     `json:"status"`
	AmountLimit    float64   `json:"amount_limit"`
	RemainingLimit float64   `json:"remaining_limit"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (u *CustomerLimit) Validate() error {
	return nil
}
