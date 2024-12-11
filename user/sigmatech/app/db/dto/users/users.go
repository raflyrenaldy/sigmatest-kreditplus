package users

import (
	"fmt"
	"github.com/google/uuid"
	"time"
	"user/sigmatech/app/service/util"
)

const (
	TABLE_NAME        = "users"
	COLUM_UUID        = "uuid"
	COLUMN_NAME       = "name"
	COLUMN_EMAIL      = "email"
	COLUMN_PASSWORD   = "password"
	COLUMN_CREATED_AT = "created_at"
	COLUMN_CREATED_BY = "created_by"
	COLUMN_UPDATED_AT = "updated_at"
	COLUMN_UPDATED_BY = "updated_by"
)

type User struct {
	Uuid      uuid.UUID  `json:"uuid"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Password  string     `json:"password,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy *uuid.UUID `json:"created_by"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdatedBy *uuid.UUID `json:"updated_by"`
}

func (u *User) ValidateUser() error {
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

func (u User) ValidateSignUpDetails() error {
	if u.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	if u.Email == "" {
		return fmt.Errorf("email can't be empty")
	}
	if u.Password == "" {
		return fmt.Errorf("password can't be empty")
	}
	return nil
}
