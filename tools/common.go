package tools

import "math/rand"

var (
	alphaNumericRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func RandomAlphanumericKey(length int) string {
	b := make([]rune, length)

	for i := range b {
		b[i] = alphaNumericRunes[rand.Intn(len(alphaNumericRunes))]
	}
	return string(b)
}
