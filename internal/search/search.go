package search

import (
	"fmt"
	"log"
	"os"

	"github.com/Eduard-Bodreev/Yadro/gocomics/internal/database"
	"github.com/Eduard-Bodreev/Yadro/gocomics/internal/words"
)

func HandleSearchQuery(dbFile, indexFile, query string) {
	index, err := words.LoadIndex(indexFile)
	if err != nil {
		log.Fatalf("Failed to load index: %v", err)
	}

	comics, err := database.LoadAllComics(dbFile)
	if err != nil {
		log.Fatalf("Failed to load comics: %v", err)
	}

	results := words.SearchIndex(query, index)
	for i, id := range results {
		if i >= 10 {
			break
		}
		comic, ok := comics[id]
		if !ok {
			log.Printf("Failed to get comic %d: comic not found", id)
			continue
		}
		fmt.Printf("Comic ID: %d, URL: %s, Page URL: https://xkcd.com/%d\n", comic.Num, comic.Img, comic.Num)
	}
	os.Exit(0)
}
