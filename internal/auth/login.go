package auth

import "forms-app/internal/api"

// RequestVerificationCode sends phone to backend
func RequestVerificationCode(phone string) error {
	_, err := api.PostJSON("auth/login", map[string]string{
		"phone": phone,
	})
	return err
}
