package user

import (
	"fmt"
	users_DBModels "user/sigmatech/app/db/dto/users"
	"user/sigmatech/app/service/util"
)

type CreateUserReq struct {
	Name       string                `json:"name"`
	Username   string                `json:"username"`
	Email      string                `json:"email"`
	Phone      string                `json:"phone"`
	Password   string                `json:"password,omitempty"`
	UserRoleID int                   `json:"user_role_id"`
	Status     users_DBModels.Status `json:"status"`
}

func (u *CreateUserReq) ValidateUser() error {
	if u.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	if u.Username == "" {
		return fmt.Errorf("username can't be empty")
	}
	if u.Email == "" {
		if !util.IsValidEmail(u.Email) {
			return fmt.Errorf("email is not valid")
		}

		return fmt.Errorf("email can't be empty")
	}
	if u.Phone == "" {
		return fmt.Errorf("phone can't be empty")
	}
	if u.Password == "" {
		return fmt.Errorf("password can't be empty")
	}
	if u.Status == "" {
		u.Status = "Active"
	}

	hashedPassword, err := util.GenerateHash(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword
	return nil

}
