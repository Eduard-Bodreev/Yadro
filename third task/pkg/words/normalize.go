package words

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
)

var re = regexp.MustCompile(`[\p{L}-]+`)

func NormalizeInput(input string) []string {
	if altIndex := strings.Index(input, "{{Alt:"); altIndex != -1 {
		input = input[:altIndex]
	}

	tokens := re.FindAllString(input, -1)

	var normalizedWords []string
	for _, token := range tokens {
		if _, err := strconv.Atoi(token); err == nil {
			continue
		}

		cleanedToken := strings.Map(func(r rune) rune {
			if unicode.IsPunct(r) {
				return -1
			}
			return r
		}, token)

		if IsStopWord(cleanedToken) {
			continue
		}

		stemmedToken := english.Stem(cleanedToken, false)
		normalizedWords = append(normalizedWords, strings.ToLower(stemmedToken))
	}

	return normalizedWords
}
