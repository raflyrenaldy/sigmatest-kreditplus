package user

import (
	"errors"
)

type CheckDevice struct {
	DeviceID          string  `json:"device_id"`
	DeviceCode        string  `json:"device_code"`
	OS                *string `json:"os"`
	DeviceInformation *string `json:"device_information"`
	VersionAPK        *string `json:"version_apk"`
}

func (s *CheckDevice) Validate() error {
	if s.DeviceID == "" {
		return errors.New("device_id is required")
	}
	if s.DeviceCode == "" {
		return errors.New("device_code is required")
	}

	return nil
}

type SignInWithDeviceAccess struct {
	DeviceID          string  `json:"device_id"`
	DeviceCode        string  `json:"device_code"`
	UserID            int     `json:"user_id"`
	PIN               string  `json:"pin"`
	OS                *string `json:"os"`
	VersionAPK        *string `json:"version_apk"`
	DeviceInformation *string `json:"device_information"`
}

func (s *SignInWithDeviceAccess) Validate() error {
	if s.DeviceID == "" {
		return errors.New("device_id is required")
	}
	if s.DeviceCode == "" {
		return errors.New("device_code is required")
	}
	if s.UserID == 0 {
		return errors.New("user_id is required")
	}
	if s.PIN == "" {
		return errors.New("pin is required")
	}

	return nil
}
