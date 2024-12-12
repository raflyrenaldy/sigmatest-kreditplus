package transaction_installments

import (
	"github.com/google/uuid"
	"time"
)

const (
	TABLE_NAME              = "transaction_installments"
	COLUM_UUID              = "uuid"
	COLUMN_TRANSACTION_UUID = "transaction_uuid"
	COLUMN_METHOD_PAYMENT   = "method_payment"
	COLUMN_TERM             = "term"
	COLUMN_DUE_DATE         = "due_date"
	COLUMN_PAYMENT_AT       = "payment_at"
	COLUMN_AMOUNT           = "amount"
	COLUMN_AMOUNT_PAID      = "amount_paid"
	COLUMN_CREATED_AT       = "created_at"
	COLUMN_UPDATED_AT       = "updated_at"
)

type TransactionInstallment struct {
	Uuid            uuid.UUID  `json:"uuid"`
	TransactionUuid uuid.UUID  `json:"transaction_uuid"`
	MethodPayment   *string    `json:"payment_method"`
	Term            int        `json:"term"`
	DueDate         *time.Time `json:"due_date"`
	PaymentAt       *time.Time `json:"payment_at"`
	Amount          float64    `json:"amount"`
	AmountPaid      float64    `json:"amount_paid"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (u *TransactionInstallment) Validate() error {

	return nil
}
