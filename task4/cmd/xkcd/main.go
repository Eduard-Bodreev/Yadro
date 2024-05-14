package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/Eduard-Bodreev/Yadro/gocomics/config"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/database"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/words"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/xkcd"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	dbFile      string
	indexFile   string
	searchQuery = flag.String("s", "", "Search query for comics")
)

func init() {
	configPath := flag.String("c", "./config/config.yaml", "Path to config file")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	viper.SetConfigFile(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	if err := words.LoadStopWords(""); err != nil {
		log.Fatalf("Failed to load stop words: %v", err)
	}
}

var ErrNotFound = errors.New("comic not found")

func main() {
	config := config.InitConfig()
	client := xkcd.New()
	sourceURL := viper.GetString("source_url")
	dbFile = viper.GetString("db_file")
	indexFile = viper.GetString("index_file")
	downloadWorkers := config.Parallel
	processWorkers := 2

	if *searchQuery != "" {
		database.HandleSearchQuery(dbFile, indexFile, *searchQuery)
		return
	}

	lastComicNum, existingComics := database.GetLastComicNum(dbFile)
	var comicNum int64 = int64(lastComicNum)

	comicsChan := make(chan *database.Comic, downloadWorkers)
	processedChan := make(chan *database.Comic, processWorkers)
	errsChan := make(chan error)
	downloadWg := sync.WaitGroup{}
	processWg := sync.WaitGroup{}

	var errorCount int32 = 0
	const maxErrors = 2

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	var shouldShutdown int32

	go func() {
		<-sigChan
		fmt.Println("Received shutdown signal, exiting...")
		atomic.StoreInt32(&shouldShutdown, 1)
		close(comicsChan)
	}()

	for i := 1; i <= lastComicNum; i++ {
		if !existingComics[i] {
			comic, err := client.FetchComic(i, sourceURL)
			if err != nil {
				log.Printf("Failed to fetch missing comic %d: %v", i, err)
				continue
			}
			if err := database.SaveComicData(*comic, dbFile); err != nil {
				log.Printf("Error saving comic %d: %v", comic.Num, err)
			}
			if err := database.BuildIndex(dbFile, config.IndexFile); err != nil {
				log.Printf("Error building index: %v", err)
			}
		}
	}

	for i := 0; i < downloadWorkers; i++ {
		downloadWg.Add(1)
		go func() {
			defer downloadWg.Done()
			for {
				if atomic.LoadInt32(&shouldShutdown) == 1 {
					return
				}

				num := atomic.AddInt64(&comicNum, 1)
				comic, err := client.FetchComic(int(num), sourceURL)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						continue
					}
					errsChan <- err
					if atomic.AddInt32(&errorCount, 1) >= maxErrors {
						log.Printf("Max error count reached, shutting down...")
						atomic.StoreInt32(&shouldShutdown, 1)
						close(comicsChan)
					}
					continue
				}
				comicsChan <- comic
			}
		}()
	}

	for i := 0; i < processWorkers; i++ {
		processWg.Add(1)
		go func() {
			defer processWg.Done()
			for comic := range processedChan {
				if err := database.SaveComicData(*comic, dbFile); err != nil {
					log.Printf("Error saving comic %d: %v", comic.Num, err)
				}
			}
		}()
	}

	go func() {
		downloadWg.Wait()
		close(comicsChan)
	}()

	go func() {
		processWg.Wait()
		close(processedChan)
	}()

	for i := 0; i < processWorkers; i++ {
		go func() {
			for err := range errsChan {
				log.Printf("Error fetching comic: %v", err)
			}
		}()
	}

	for comic := range comicsChan {
		processedChan <- comic
	}

	close(errsChan)
	fmt.Println("All comics fetched and saved.")
}
