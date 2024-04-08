package words

import (
	"strings"
	"unicode"
)

func NormalizeInput(input string) []string {

	if altIndex := strings.Index(input, "{{Alt:"); altIndex != -1 {
		input = input[:altIndex]
	}

	wordsToNormalize := strings.FieldsFunc(input, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})

	var normalizedWords []string
	for _, word := range wordsToNormalize {
		cleanedWord := strings.Map(func(r rune) rune {
			if unicode.IsPunct(r) {
				return -1
			}
			return r
		}, word)

		if IsStopWord(cleanedWord) {
			normalizedWords = append(normalizedWords, strings.ToLower(cleanedWord))
		}
	}

	return normalizedWords
}
