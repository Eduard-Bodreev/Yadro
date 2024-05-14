package words

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
)

var re = regexp.MustCompile(`[\p{L}-]+`)

type Index map[string][]int

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

func LoadIndex(indexFile string) (Index, error) {
	file, err := os.Open(indexFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file: %v", err)
	}
	defer file.Close()

	var index Index
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to decode index: %v", err)
	}

	return index, nil
}

func SearchIndex(query string, index Index) []int {
	words := NormalizeInput(query)
	results := make(map[int]int)
	for _, word := range words {
		if ids, ok := index[word]; ok {
			for _, id := range ids {
				results[id]++
			}
		}
	}

	type kv struct {
		Key   int
		Value int
	}
	var sortedResults []kv
	for k, v := range results {
		sortedResults = append(sortedResults, kv{k, v})
	}
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].Value > sortedResults[j].Value
	})

	var finalResults []int
	for _, kv := range sortedResults {
		finalResults = append(finalResults, kv.Key)
	}

	return finalResults
}
