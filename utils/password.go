package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func CheckPassword(hashedPassword []byte, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(plainPassword))
	return err == nil // nil error indicates successful password verification
}
