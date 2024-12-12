package variable_globals

import (
	"github.com/google/uuid"
	"time"
)

const (
	TABLE_NAME         = "variable_globals"
	COLUM_UUID         = "uuid"
	COLUMN_CODE        = "code"
	COLUMN_VALUE       = "value"
	COLUMN_DESCRIPTION = "description"
	COLUMN_CREATED_AT  = "created_at"
	COLUMN_UPDATED_AT  = "updated_at"
)

type VariableGlobal struct {
	Uuid        uuid.UUID `json:"uuid"`
	Code        string    `json:"code"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *VariableGlobal) Validate() error {
	return nil
}
