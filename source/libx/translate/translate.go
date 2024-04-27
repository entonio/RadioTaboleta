package translate

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
)

type Translator struct {
	Languages    []string
	keyLanguage  string
	translations map[string]map[string]string
}

func NewTranslator(file string, keyLanguage string) Translator {
	keyLanguage = languageCode(keyLanguage)
	languages, translations, _ := read(file, keyLanguage)
	return Translator{Languages: languages, keyLanguage: keyLanguage, translations: translations}
}

func (self Translator) TranslateFormatted(language string, key string, parameters ...any) string {
	return fmt.Sprintf(self.Translate(language, key), parameters...)
}

func (self Translator) Translate(language string, key string) string {
	if len(language) > 0 && language != self.keyLanguage {
		if self.translations != nil {
			translations := self.translations[language]
			if translations != nil {
				return translations[key]
			}
		}
	}
	return key
}

func languageCode(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func read(path string, keyLanguage string) (languages []string, translations map[string]map[string]string, err error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Error reading %s: %s\n", path, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading %s: %s\n", path, err)
		return
	}
	if len(records) < 2 {
		log.Printf("Transalations file %s is empty\n", path)
		return
	}
	if len(records[0]) > 0 {
		bom := string([]byte{239, 187, 191})
		records[0][0] = strings.TrimPrefix(records[0][0], bom)
	}

	keyIndex := -1
	languages = records[0]
	for i, language := range languages {
		language = languageCode(language)
		fmt.Printf("[%+v][%+v][%+v][%+v][%+v]\n", keyLanguage, language, []byte(language), len(language),
			language == keyLanguage)
		if language == keyLanguage {
			keyIndex = i
		}
		languages[i] = language
	}

	if keyIndex < 0 {
		log.Printf("Transalations file %s doesn't contain the key language (%s)\n", path, keyLanguage)
		return
	}

	translations = make(map[string]map[string]string)
	for _, row := range records[1:] {
		if len(row) < keyIndex {
			continue
		}
		key := row[keyIndex]
		for i, translation := range row {
			if i == keyIndex {
				continue
			}
			if i > len(languages) {
				continue
			}
			language := languages[i]
			lt := translations[language]
			if lt == nil {
				lt = make(map[string]string)
			}
			lt[key] = translation
			translations[language] = lt
		}
	}
	return
}
