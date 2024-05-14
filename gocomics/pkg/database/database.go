package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/words"
)

type ComicKeywords struct {
	Num      int      `json:"num"`
	Img      string   `json:"img"`
	Keywords []string `json:"keywords"`
}

func SaveComicData(comic Comic, dbFile string) error {
	file, err := os.OpenFile(dbFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error opening %s for writing: %v", dbFile, err)
	}
	defer file.Close()

	fullText := comic.Transcript + " " + comic.Alt
	keywords := words.NormalizeInput(fullText)

	newData := map[int]ComicKeywords{
		comic.Num: {
			Num:      comic.Num,
			Img:      comic.Img,
			Keywords: keywords,
		},
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(newData); err != nil {
		return fmt.Errorf("error encoding JSON to %s: %v", dbFile, err)
	}

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

	data := make(map[int]Comic)
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		log.Fatalf("Failed to decode database.json: %v", err)
	}

	maxNum := 0
	for num := range data {
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum
}
