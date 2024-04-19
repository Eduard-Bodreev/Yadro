package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func SaveComicData(comic Comic, dbFile string) error {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	ComicBuffer = append(ComicBuffer, ComicKeywords{
		Num:      comic.Num,
		Img:      comic.Img,
		Keywords: words.NormalizeInput(comic.Transcript + " " + comic.Alt),
	})

	if len(ComicBuffer) >= BufferSize {
		return FlushComicData(dbFile)
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

func GetLastComicNum(dbFile string) int {
	file, err := os.Open(dbFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		log.Fatalf("Failed to open database.json: %v", err)
	}
	defer file.Close()

	var comics []ComicKeywords
	if err := json.NewDecoder(file).Decode(&comics); err != nil {
		log.Fatalf("Failed to decode database.json: %v", err)
	}

	maxNum := 0
	for _, comic := range comics {
		if comic.Num > maxNum {
			maxNum = comic.Num
		}
	}
	return maxNum
}
