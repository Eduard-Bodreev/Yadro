package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/Eduard-Bodreev/Yadro/gocomics/config"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/database"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/words"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/xkcd"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	if err := words.LoadStopWords(""); err != nil {
		log.Fatalf("Failed to load stop words: %v", err)
	}
	config.InitConfig()
}

func main() {

	client := xkcd.New()

	outputFlag := flag.Bool("o", false, "Output the comics in stdout")
	numComics := flag.Int("n", 1, "Number of comics to fetch")
	flag.Parse()

	sourceURL := viper.GetString("source_url")
	dbFile := viper.GetString("db_file")

	lastComicNum := database.GetLastComicNum(dbFile)
	startNum := lastComicNum + 1

	maxComicNum, err := client.GetCurrentMaxComicNum(sourceURL)
	if err != nil {
		log.Fatalf("Error fetching the maximum comic number: %v", err)
	}

	comicsToFetch := min(*numComics, maxComicNum-startNum+1)

	var comics []database.Comic

	for i := 0; i < comicsToFetch; i++ {
		comicNum := startNum + i
		comic, err := client.FetchComic(comicNum, sourceURL)
		if err != nil {
			log.Printf("Error fetching comic %d: %v", comicNum, err)
			continue
		}

		err = database.SaveComicData(*comic, dbFile)
		if err != nil {
			log.Printf("Error saving comic %d: %v", comicNum, err)
			break
		}

		comics = append(comics, *comic)
		if *outputFlag {
			fmt.Printf("Comic #%d: %s\n", comic.Num, comic.Img)
		}
	}

	if *outputFlag && len(comics) > 0 {
		jsonData, err := json.MarshalIndent(comics, "", "    ")
		if err != nil {
			log.Fatalf("Failed to marshal comics: %v", err)
		}
		fmt.Println(string(jsonData))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
