package auth

import "forms-app/internal/api"

// VerifyCode confirms OTP
func VerifyCode(phone, code string) (string, error) {
	res, err := api.PostJSON("auth/verify", map[string]string{
		"phone": phone,
		"code":  code,
	})
	if err != nil {
		return "", err
	}
	if token, ok := res["token"].(string); ok {
		return token, nil
	}
	return "", nil
}
