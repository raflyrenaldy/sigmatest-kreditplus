package user

import (
	"fmt"
	"github.com/google/uuid"
)

type ApproveCustomerReq struct {
	CustomerUuid   uuid.UUID       `json:"customer_uuid"`
	CustomerLimits []CustomerLimit `json:"customer_limits"`
}

type CustomerLimit struct {
	Uuid   uuid.UUID `json:"uuid"`
	Amount float64   `json:"amount"`
}

func (u *ApproveCustomerReq) Validate() error {
	if u.CustomerUuid == uuid.Nil {
		return fmt.Errorf("customer uuid can't be empty")
	}

	if len(u.CustomerLimits) == 0 {
		return fmt.Errorf("customer limits can't be empty")
	}
	return nil

}
