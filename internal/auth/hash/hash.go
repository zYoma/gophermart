package hash

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword принимает пароль в виде строки и возвращает хеш этого пароля
// или ошибку, если процесс хеширования не удался.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
