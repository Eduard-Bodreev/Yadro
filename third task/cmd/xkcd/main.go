package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
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

func main() {
	config := config.InitConfig()
	client := xkcd.New()
	sourceURL := viper.GetString("source_url")
	dbFile := viper.GetString("db_file")
	parallel := config.Parallel

	lastComicNum := database.GetLastComicNum(dbFile)
	var comicNum int64 = int64(lastComicNum + 1)

	comicsChan := make(chan *database.Comic)
	errsChan := make(chan error)
	wg := sync.WaitGroup{}

	var notFoundCount int32 = 0
	const maxNotFound = 5

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for comic := range comicsChan {
				if err := database.SaveComicData(*comic, dbFile); err != nil {
					log.Printf("Error saving comic %d: %v", comic.Num, err)
				}
			}
		}()
	}

	go func() {
		defer close(errsChan)
		for {
			if atomic.LoadInt32(&notFoundCount) >= maxNotFound {
				log.Println("Max not found count reached, flushing remaining data.")
				database.FlushComicData(dbFile)
				close(comicsChan)
				return
			}
			num := atomic.AddInt64(&comicNum, 1)
			comic, err := client.FetchComic(int(num), sourceURL)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					atomic.AddInt32(&notFoundCount, 1)
				} else {
					errsChan <- err
				}
				continue
			}
			atomic.StoreInt32(&notFoundCount, 0)
			comicsChan <- comic
		}
	}()

	go func() {
		wg.Wait()
		if len(database.ComicBuffer) > 0 {
			database.FlushComicData(dbFile)
		}
		log.Println("All comics fetched and saved.")
	}()

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, flushing data and closing channels...")
		close(comicsChan)
		wg.Wait()
		database.FlushComicData(dbFile)
		os.Exit(0)
	}()

	for err := range errsChan {
		log.Printf("Error fetching comic: %v", err)
	}
}
