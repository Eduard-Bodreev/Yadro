package database

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/words"
)

var (
	comicsData  []ComicKeywords
	searchIndex map[string][]int
)

func loadComicsData() {
	data, err := os.ReadFile("database.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &comicsData); err != nil {
		panic(err)
	}
}

func loadIndex() {
	data, err := os.ReadFile("index.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &searchIndex); err != nil {
		panic(err)
	}
}

func init() {
	loadComicsData()
	loadIndex()
}

func BenchmarkSearchByIndex(b *testing.B) {
	query := "I'm following your questions"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = words.SearchIndex(query, searchIndex)
	}
}

func BenchmarkLinearSearch(b *testing.B) {
	query := "I'm following your questions"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = linearSearch(query, comicsData)
	}
}

func linearSearch(query string, data []ComicKeywords) []int {
	var results []int
	for _, comic := range data {
		if contains(comic.Keywords, query) {
			results = append(results, comic.Num)
		}
	}
	return results
}

func contains(keywords []string, query string) bool {
	for _, word := range keywords {
		if word == query {
			return true
		}
	}
	return false
}
