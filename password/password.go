package password

import (
	"golang.org/x/crypto/bcrypt"
)

func Encrypt(pwd string) (string, error) {
	encrypted, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(encrypted), nil
}

func Verify(pwd string, encrypted string) error {
	return bcrypt.CompareHashAndPassword([]byte(encrypted), []byte(pwd))
}
