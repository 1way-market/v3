package model

// Language represents the supported languages
type Language int

const (
	LangRussian Language = 1
	LangEnglish Language = 2
	LangTurkish Language = 3
)

// MultiLangText represents a text in multiple languages
type MultiLangText struct {
	Lang Language `json:"lang"`
	Text string   `json:"text"`
}

// GetText returns the text for the specified language, falling back to English if not found
func GetTextForLang(texts []MultiLangText, lang Language) string {
	// First try to find exact match
	for _, t := range texts {
		if t.Lang == lang {
			return t.Text
		}
	}

	// Fallback to English if available
	for _, t := range texts {
		if t.Lang == LangEnglish {
			return t.Text
		}
	}

	// If no English, return the first available text
	if len(texts) > 0 {
		return texts[0].Text
	}

	return ""
}
