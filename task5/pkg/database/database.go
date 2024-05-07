package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/models"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/words"
)

type ComicKeywords struct {
	Num      int      `json:"num"`
	Img      string   `json:"img"`
	Keywords []string `json:"keywords"`
}

type ComicFetcher interface {
	FetchComic(num int) (*models.Comic, error)
}

var (
	ComicBuffer []ComicKeywords
	bufferMutex sync.Mutex
)

const BufferSize = 10

func SaveComicData(comic models.Comic, dbFile string) error {
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

func LoadAllComics(dbFile string) (map[int]*ComicKeywords, error) {
	file, err := os.Open(dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %v", err)
	}
	defer file.Close()

	var comics []ComicKeywords
	if err := json.NewDecoder(file).Decode(&comics); err != nil {
		return nil, fmt.Errorf("failed to decode database: %v", err)
	}

	comicsMap := make(map[int]*ComicKeywords)
	for i := range comics {
		comicsMap[comics[i].Num] = &comics[i]
	}
	return comicsMap, nil
}

func UpdateComics(dbFile string, fetcher ComicFetcher) (int, int, error) {
	lastComicNum, existingComics := GetLastComicNum(dbFile)
	newComicsCount := 0
	for i := lastComicNum + 1; ; i++ {
		comic, err := fetcher.FetchComic(i)
		if err != nil {
			break
		}
		if existingComics[comic.Num] {
			continue
		}
		err = SaveComicData(*comic, dbFile)
		if err != nil {
			log.Printf("Failed to save comic %d: %v", comic.Num, err)
			continue
		}
		newComicsCount++
	}
	totalComicsCount := len(existingComics) + newComicsCount
	return newComicsCount, totalComicsCount, nil
}
