package user

import (
	"errors"
	"fmt"
	"strings"
)

type SignInRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	User     string `json:"user"`
	Password string `json:"password" binding:"required"`
}

func (s *SignInRequest) Validate() error {
	if s.Password == "" {
		return errors.New("password is required")
	}

	count := 0
	if s.Email != "" {
		count++
	}
	if s.Username != "" {
		count++
	}
	if s.Phone != "" {
		count++
	}

	// Ensure only one of email, username, phone, or user is provided
	if count > 1 {
		return errors.New("exactly one of email, username, phone, or user should be provided")
	}

	// If 'User' is provided, set other fields to its value
	if s.User != "" {
		s.Email = s.User
		s.Username = s.User
		s.Phone = s.User
	}

	return nil
}

type SignUpRequest struct {
	Name            string `json:"name"`
	Address         string `json:"address"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirmation"`
}

func (s *SignUpRequest) Validate() error {
	// create a new empty error list
	var errReturn []error

	if s.Name == "" {
		errReturn = append(errReturn, fmt.Errorf("name can't be empty"))
	}

	if s.Email == "" {
		errReturn = append(errReturn, fmt.Errorf("email can't be empty"))
	}

	if s.Phone == "" {
		errReturn = append(errReturn, fmt.Errorf("phone can't be empty"))
	}

	if s.Password == "" {
		errReturn = append(errReturn, fmt.Errorf("password can't be empty"))
	}

	if s.PasswordConfirm == "" {
		errReturn = append(errReturn, fmt.Errorf("password_confirmation can't be empty"))
	}

	if s.Password != s.PasswordConfirm {
		errReturn = append(errReturn, fmt.Errorf("Password and confirmation do not match."))
	}

	if len(errReturn) > 0 {
		var errMessages []string
		for _, v := range errReturn {
			errMessages = append(errMessages, v.Error())
		}

		return fmt.Errorf("%s", strings.Join(errMessages, ", "))
	}

	return nil
}
