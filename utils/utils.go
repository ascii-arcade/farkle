package utils

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"math/big"
)

//go:embed animals.json adjectives.json
var words embed.FS

var (
	animals    []string
	adjectives []string
)

func init() {
	animalsBytes, err := words.ReadFile("animals.json")
	if err != nil {
		panic("Failed to load animals.json: " + err.Error())
	}

	adjectivesBytes, err := words.ReadFile("adjectives.json")
	if err != nil {
		panic("Failed to load adjectives.json: " + err.Error())
	}

	if err := json.Unmarshal(animalsBytes, &animals); err != nil {
		panic("Failed to unmarshal animals.json: " + err.Error())
	}

	if err := json.Unmarshal(adjectivesBytes, &adjectives); err != nil {
		panic("Failed to unmarshal adjectives.json: " + err.Error())
	}
}

func GenerateName() string {
	// Generate a random name using an adjective and an animal
	adjective := adjectives[randomInt(len(adjectives))]
	animal := animals[randomInt(len(animals))]
	return adjective + " " + animal
}

func randomInt(max int) int64 {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return num.Int64()
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
			b[i] = letters[randomInt(len(letters))]
		}
		code := string(b)
		return code
	}
}
