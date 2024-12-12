package transactions

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	TABLE_NAME                 = "transactions"
	COLUM_UUID                 = "uuid"
	COLUMN_CUSTOMER_UUID       = "customer_uuid"
	COLUMN_CUSTOMER_LIMIT_UUID = "customer_limit_uuid"
	COLUMN_ASSET_NAME          = "asset_name"
	COLUMN_CONTRACT_NUMBER     = "contract_number"
	COLUMN_IS_DONE             = "is_done"
	COLUMN_OTR                 = "otr"
	COLUMN_ADMIN_FEE           = "admin_fee"
	COLUMN_TOTAL               = "total"
	COLUMN_INSTALLMENT_AMOUNT  = "installment_amount"
	COLUMN_INSTALLMENT_COUNT   = "installment_count"
	COLUMN_TOTAL_INTEREST      = "total_interest"
	COLUMN_CREATED_AT          = "created_at"
	COLUMN_UPDATED_AT          = "updated_at"
)

type Transaction struct {
	Uuid              uuid.UUID `json:"uuid"`
	CustomerUuid      uuid.UUID `json:"customer_uuid"`
	CustomerLimitUuid uuid.UUID `json:"customer_limit_uuid"`
	AssetName         string    `json:"asset_name"`
	ContractNumber    string    `json:"contract_number"`
	IsDone            *bool     `json:"is_done"`
	Otr               float64   `json:"otr"`
	AdminFee          float64   `json:"admin_fee"`
	Total             float64   `json:"total"`
	InstallmentAmount float64   `json:"installment_amount"`
	InstallmentCount  int       `json:"installment_count"`
	TotalInterest     float64   `json:"total_interest"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (u *Transaction) Validate() error {
	if u.CustomerLimitUuid == uuid.Nil {
		return fmt.Errorf("limit uuid can't be empty")
	}
	if u.AssetName == "" {
		return fmt.Errorf("asset name can't be empty")
	}
	if u.Otr == 0 {
		return fmt.Errorf("otr can't be empty")
	}

	return nil
}
