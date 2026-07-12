package validator

import "net/mail"

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidPhone(phone string) bool {
	if len(phone) < 8 || len(phone) > 20 {
		return false
	}
	for _, c := range phone {
		if c < '+' || (c > '9' && c != ' ') {
			return false
		}
	}
	return true
}
