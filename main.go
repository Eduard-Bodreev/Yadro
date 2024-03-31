package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
)

func normalizeInput(input string) string {
	words := strings.FieldsFunc(input, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})

	var normalizedWords []string
	for _, word := range words {
		if !isStopWord(word) {
			normalizedWords = append(normalizedWords, strings.ToLower(word)+" ")
		}
	}

	res := strings.Join(normalizedWords, "")
	if len(res) > 0 {
		res = res[:len(res)-1]
	}
	return res
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	var stopWordsFilePath, input string

	flag.StringVar(&stopWordsFilePath, "stopwords", "", "Path to the stopwords file")
	flag.StringVar(&input, "s", "", "String to normalize")
	flag.Parse()

	if err := loadStopWords(stopWordsFilePath); err != nil {
		fmt.Printf("Failed to load stop words: %v\n", err)
		return
	}

	if input == "" {
		fmt.Println("No input provided")
		return
	}

	normalized := normalizeInput(input)
	fmt.Println(normalized)
}
