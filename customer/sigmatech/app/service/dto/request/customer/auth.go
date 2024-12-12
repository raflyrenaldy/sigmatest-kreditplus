package user

import (
	"customer/sigmatech/app/service/util"
	"errors"
	"fmt"
)

type SignInRequest struct {
	Email    string `json:"email"`
	User     string `json:"user"`
	Password string `json:"password" binding:"required"`
}

func (s *SignInRequest) Validate() error {
	if s.Password == "" {
		return errors.New("password is required")
	}

	if s.Email == "" {
		if !util.IsValidEmail(s.Email) {
			return fmt.Errorf("email is not valid")
		}

		return fmt.Errorf("email can't be empty")
	}

	return nil
}

type UpdatePassword struct {
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	OldPassword          string `json:"old_password"`
}

func (s *UpdatePassword) Validate() error {
	if s.Password == "" {
		return errors.New("password is required")
	}
	if s.PasswordConfirmation == "" {
		return errors.New("password_confirmation is required")
	}
	if s.OldPassword == "" {
		return errors.New("old_password is required")
	}

	if s.PasswordConfirmation != s.Password {
		return errors.New("password and password confirmation must match")
	}

	return nil
}
