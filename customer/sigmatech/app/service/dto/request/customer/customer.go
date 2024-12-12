package user

import (
	"customer/sigmatech/app/service/util"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type SignUpReq struct {
	CustomerUUID uuid.UUID  `json:"customer_uuid"`
	CifUuid      uuid.UUID  `json:"cif_uuid"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	FullName     string     `json:"full_name"`
	LegalName    string     `json:"legal_name"`
	Nik          string     `json:"nik"`
	PlaceOfBirth string     `json:"place_of_birth"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	Gender       string     `json:"gender"`
	Salary       float64    `json:"salary"`
	CardPhoto    string     `json:"card_photo"`
	SelfiePhoto  string     `json:"selfie_photo"`
	Password     string     `json:"password,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (u *SignUpReq) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	if u.FullName == "" {
		return fmt.Errorf("full name can't be empty")
	}
	if u.LegalName == "" {
		return fmt.Errorf("legal name can't be empty")
	}
	if u.Nik == "" {
		return fmt.Errorf("nik can't be empty")
	}
	if len(u.Nik) != 16 {
		return fmt.Errorf("nik length must be 16")
	}
	if u.PlaceOfBirth == "" {
		return fmt.Errorf("place of birth can't be empty")
	}
	if u.DateOfBirth == nil {
		return fmt.Errorf("date of birth can't be empty")
	}
	if u.Salary <= 0 {
		return fmt.Errorf("salary can't be empty")
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

	if u.CardPhoto == "" {
		return fmt.Errorf("card photo can't be empty")
	}

	if u.SelfiePhoto == "" {
		return fmt.Errorf("selfie photo can't be empty")
	}

	// TODO validate password with character length is 8, need contains Uppercase, lowercase, Symbol, Numeric

	hashedPassword, err := util.GenerateHash(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword
	return nil

}
