package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/words"
)

type ComicKeywords struct {
	Num      int      `json:"num"`
	Img      string   `json:"img"`
	Keywords []string `json:"keywords"`
}

var (
	ComicBuffer []ComicKeywords
	bufferMutex sync.Mutex
)

const BufferSize = 10

func HandleSearchQuery(dbFile, indexFile, query string) {
	index, err := words.LoadIndex(indexFile)
	if err != nil {
		log.Fatalf("Failed to load index: %v", err)
	}

	results := words.SearchIndex(query, index)
	for i, id := range results {
		if i >= 10 {
			break
		}
		comic, err := GetComicByID(dbFile, id)
		if err != nil {
			log.Printf("Failed to get comic %d: %v", id, err)
			continue
		}
		fmt.Printf("Comic ID: %d, URL: %s\n", comic.Num, comic.Img)
	}
	os.Exit(0)
}

func SaveComicData(comic Comic, dbFile string) error {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	ComicBuffer = append(ComicBuffer, ComicKeywords{
		Num:      comic.Num,
		Img:      comic.Img,
		Keywords: words.NormalizeInput(comic.Transcript + " " + comic.Alt),
	})

	if len(ComicBuffer) >= BufferSize {
		if err := FlushComicData(dbFile); err != nil {
			return err
		}
		indexFile := filepath.Join(filepath.Dir(dbFile), "index.json")
		if err := BuildIndex(dbFile, indexFile); err != nil {
			return err
		}
	}
	return nil
}

func MaybeFlushComicData(dbFile string) error {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	if len(ComicBuffer) > 0 {
		return FlushComicData(dbFile)
	}
	return nil
}

func FlushComicData(dbFile string) error {
	tempFile := dbFile + ".tmp"
	existingData, err := os.ReadFile(dbFile)
	if err == nil && len(existingData) > 0 {
		var existingComics []ComicKeywords
		if err := json.Unmarshal(existingData, &existingComics); err != nil {
			return fmt.Errorf("error decoding JSON from %s: %v", dbFile, err)
		}
		ComicBuffer = append(existingComics, ComicBuffer...)
	}

	newData, err := json.MarshalIndent(ComicBuffer, "", " ")
	if err != nil {
		return fmt.Errorf("error encoding JSON to %s: %v", tempFile, err)
	}

	if err := os.WriteFile(tempFile, newData, 0666); err != nil {
		return fmt.Errorf("error writing to %s: %v", tempFile, err)
	}

	if err := os.Rename(tempFile, dbFile); err != nil {
		return fmt.Errorf("error renaming %s to %s: %v", tempFile, dbFile, err)
	}

	ComicBuffer = nil
	return nil
}

func GetLastComicNum(dbFile string) (int, map[int]bool) {
	file, err := os.Open(dbFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		log.Fatalf("Failed to open database.json: %v", err)
	}
	defer file.Close()

	var comics []ComicKeywords
	if err := json.NewDecoder(file).Decode(&comics); err != nil {
		log.Fatalf("Failed to decode database.json: %v", err)
	}

	maxNum := 0
	existingNums := make(map[int]bool)
	for _, comic := range comics {
		if comic.Num > maxNum {
			maxNum = comic.Num
		}
		existingNums[comic.Num] = true
	}
	return maxNum, existingNums
}

func BuildIndex(dbFile string, indexFile string) error {
	file, err := os.Open(dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database file: %v", err)
	}
	defer file.Close()

	var comics []ComicKeywords
	if err := json.NewDecoder(file).Decode(&comics); err != nil {
		return fmt.Errorf("failed to decode database: %v", err)
	}

	index := make(map[string][]int)
	for _, comic := range comics {
		for _, keyword := range comic.Keywords {
			index[keyword] = append(index[keyword], comic.Num)
		}
	}

	indexData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode index: %v", err)
	}

	if err := os.WriteFile(indexFile, indexData, 0666); err != nil {
		return fmt.Errorf("failed to write index file: %v", err)
	}

	return nil
}

func GetComicByID(dbFile string, id int) (*ComicKeywords, error) {
	file, err := os.Open(dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %v", err)
	}
	defer file.Close()

	var comics []ComicKeywords
	if err := json.NewDecoder(file).Decode(&comics); err != nil {
		return nil, fmt.Errorf("failed to decode database: %v", err)
	}

	for _, comic := range comics {
		if comic.Num == id {
			return &comic, nil
		}
	}
	return nil, fmt.Errorf("comic not found")
}
