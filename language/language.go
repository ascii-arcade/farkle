package language

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed en.json
var enJSON []byte

//go:embed es.json
var esJSON []byte

var Languages = map[string]*Language{
	"EN": LoadLanguage(enJSON),
	"ES": LoadLanguage(esJSON),
}

type Language struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Translations map[string]any `json:"translations"`

	UsernameFirstWords  []string `json:"username_first_words"`
	UsernameSecondWords []string `json:"username_second_words"`
}

var DefaultLanguage = Languages["EN"]

func (l *Language) Get(pathList ...string) string {
	if len(pathList) == 0 {
		return ""
	}
	var current any = l.Translations
	for i, key := range pathList {
		m, ok := current.(map[string]any)
		if !ok {
			return missingTranslationValue(pathList[0], key)
		}
		val, exists := m[key]
		if !exists {
			return missingTranslationValue(pathList[0], key)
		}
		if i == len(pathList)-1 {
			str, ok := val.(string)
			if !ok {
				return missingTranslationValue(pathList[0], key)
			}
			return str
		}
		current = val
	}
	return ""
}

func missingTranslationValue(section, key string) string {
	return "i18n-missing:'" + section + "." + key + "'"
}

func LoadLanguage(data []byte) *Language {
	var lang Language
	if err := json.Unmarshal(data, &lang); err != nil {
		log.Fatal("failed to decode language data: %w", err)
	}
	return &lang
}
