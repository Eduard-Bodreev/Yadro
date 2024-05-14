package words

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

var stopwords = make(map[string]bool)

func LoadStopWords(filePath string) error {
	if filePath == "" {
		var ok bool
		filePath, ok = os.LookupEnv("STOPWORDS_FILE")
		if !ok {
			return errors.New("file path wasn't read from env, be sure you are using flags")
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		stopwords[strings.ToLower(word)] = true
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func IsStopWord(word string) bool {
	return stopwords[strings.ToLower(word)]
}
