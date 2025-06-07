package utils

import (
	"fmt"
	"math/rand/v2"

	"github.com/ascii-arcade/farkle/language"
)

func GenerateName(lang *language.Language) string {
AGAIN:
	a := lang.UsernameFirstWords[rand.IntN(len(lang.UsernameFirstWords))]
	b := lang.UsernameSecondWords[rand.IntN(len(lang.UsernameSecondWords))]
	n := fmt.Sprintf("%s %s", a, b)
	if len(n) > 15 {
		goto AGAIN
	}

	return n
}

func GenerateCode() string {
	for {
		b := make([]byte, 7)
		letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		for i := range b {
			if i == 3 {
				b[i] = '-'
				continue
			}
			b[i] = letters[rand.IntN(len(letters))]
		}
		code := string(b)
		return code
	}
}

func ToPointer[T any](v T) *T {
	return &v
}
