package utils

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

func GenerateUsername(prefixLength int) string {
	uuid := uuid.New().String()
	return fmt.Sprintf("%s-%s", uuid, _randomString(prefixLength))
}

func _randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")
	print()
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
