package language

type LanguagePreference struct {
	Lang *Language
}

func (lp *LanguagePreference) SetLanguage(abbr string) {
	if lp == nil {
		lp = &LanguagePreference{}
	}

	lang, exists := Languages[abbr]
	if !exists {
		return
	}
	lp.Lang = lang
}
