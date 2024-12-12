package customers

import (
	"customer/sigmatech/app/service/util"
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	TABLE_NAME        = "customers"
	COLUM_UUID        = "uuid"
	COLUMN_NAME       = "name"
	COLUMN_EMAIL      = "email"
	COLUMN_PASSWORD   = "password"
	COLUMN_IS_ACTIVE  = "is_active"
	COLUMN_CREATED_AT = "created_at"
	COLUMN_UPDATED_AT = "updated_at"
)

type Customer struct {
	Uuid      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *Customer) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	if u.Email == "" {
		if !util.IsValidEmail(u.Email) {
			return fmt.Errorf("email is not valid")
		}

		return fmt.Errorf("email can't be empty")
	}
	if u.Password == "" {
		return fmt.Errorf("password can't be empty")
	}

	hashedPassword, err := util.GenerateHash(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	return nil
}
