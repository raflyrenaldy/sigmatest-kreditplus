package customer_information_files

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	TABLE_NAME            = "customer_information_files"
	COLUM_UUID            = "uuid"
	COLUMN_CUSTOMER_UUID  = "customer_uuid"
	COLUMN_FULL_NAME      = "full_name"
	COLUMN_LEGAL_NAME     = "legal_name"
	COLUMN_CIF_NUMBER     = "cif_number"
	COLUMN_NIK            = "nik"
	COLUMN_PLACE_OF_BIRTH = "place_of_birth"
	COLUMN_DATE_OF_BIRTH  = "date_of_birth"
	COLUMN_GENDER         = "gender"
	COLUMN_SALARY         = "salary"
	COLUMN_CARD_PHOTO     = "card_photo"
	COLUMN_SELFIE_PHOTO   = "selfie_photo"
	COLUMN_CREATED_AT     = "created_at"
	COLUMN_UPDATED_AT     = "updated_at"
)

type CustomerInformationFile struct {
	Uuid         uuid.UUID  `json:"uuid"`
	CustomerUuid uuid.UUID  `json:"customer_uuid"`
	CifNumber    string     `json:"cif_number"`
	Nik          string     `json:"nik"`
	FullName     string     `json:"full_name"`
	LegalName    string     `json:"legal_name"`
	PlaceOfBirth string     `json:"place_of_birth"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	Gender       *string    `json:"gender"`
	Salary       float64    `json:"salary"`
	CardPhoto    string     `json:"card_photo"`
	SelfiePhoto  string     `json:"selfie_photo"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (u *CustomerInformationFile) Validate() error {
	if u.Nik == "" {
		return fmt.Errorf("nik can't be empty")
	}
	if u.LegalName == "" {
		return fmt.Errorf("legal name can't be empty")
	}
	if u.PlaceOfBirth == "" {
		return fmt.Errorf("place of birth can't be empty")
	}
	if u.DateOfBirth == nil {
		return fmt.Errorf("date of birth can't be empty")
	}
	if u.Salary == 0 {
		return fmt.Errorf("salary can't be empty")
	}
	if u.FullName == "" {
		return fmt.Errorf("full name can't be empty")
	}
	return nil
}
