package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
)

func normalizeInput(input string) string {
	words := strings.Fields(input)
	normalizedWords := make([]string, 0)

	for _, word := range words {
		if !isStopWord(word) {
			normalizedWords = append(normalizedWords, strings.ToLower(word)+" ")
		}
	}

	return strings.Join(normalizedWords, "")
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	stopWordsFilePath := flag.String("stopwords", "", "Path to the stopwords file")
	input := flag.String("s", "", "String to normalize")
	flag.Parse()

	if err := loadStopWords(*stopWordsFilePath); err != nil {
		fmt.Printf("Failed to load stop words: %v\n", err)
		return
	}

	if *input == "" {
		fmt.Println("No input provided")
		return
	}

	normalized := normalizeInput(*input)
	fmt.Println(normalized)
}
