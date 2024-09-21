package utils

import (
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func BcryptHash(value string) (string, error) {
	costStr := os.Getenv("BCRYPT_ROUNDS")
	cost, _ := strconv.Atoi(costStr)
	hashedValue, err := bcrypt.GenerateFromPassword([]byte(value), cost)
	if err != nil {
		return "", err
	}
	return string(hashedValue), nil
}
