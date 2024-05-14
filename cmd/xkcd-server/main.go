package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"

	"github.com/Eduard-Bodreev/Yadro/gocomics/config"
	"github.com/Eduard-Bodreev/Yadro/gocomics/internal/database"
	"github.com/Eduard-Bodreev/Yadro/gocomics/internal/words"
	"github.com/Eduard-Bodreev/Yadro/gocomics/internal/xkcd"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
)

var cfg config.Config

func main() {
	var port string
	flag.StringVar(&port, "p", "", "Port to run the server on")
	flag.Parse()

	cfg = config.InitConfig()

	if port != "" {
		cfg.Port = port
	}

	log.Printf("Using DSN: %s", cfg.DSN)

	log.Printf("Connecting to database at %s", cfg.DSN)
	db, err := sql.Open("sqlite3", cfg.DSN)
	if err != nil {
		log.Fatalf("Cannot open database: %v", err)
	}
	defer db.Close()

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		log.Fatalf("Failed to create migrate driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file:///home/edik/Yadro/task5/migrations",
		"sqlite3", driver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	go ScheduleDailyUpdates()

	srv := &http.Server{
		Addr: ":" + cfg.Port,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the XKCD server!")
	})

	http.HandleFunc("/update", handleUpdate)
	http.HandleFunc("/pics", handlePics)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server is starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", cfg.Port, err)
		}
	}()

	<-stop

	log.Println("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func ScheduleDailyUpdates() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		xkcdClient := xkcd.New(cfg.SourceURL)
		if _, _, err := database.UpdateComics(cfg.DBFile, xkcdClient); err != nil {
			log.Printf("Error during scheduled update: %v", err)
		}
	}
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	xkcdClient := xkcd.New(cfg.SourceURL)
	newComics, totalComics, err := database.UpdateComics(cfg.DBFile, xkcdClient)
	if err != nil {
		log.Printf("Error updating comics: %v", err)
		http.Error(w, fmt.Sprintf("Error updating database: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Updated/Total comics: %d/%d", newComics, totalComics)

	response := map[string]int{"new": newComics, "total": totalComics}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handlePics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query().Get("search")
	if query == "" {
		http.Error(w, "Query parameter 'search' is required", http.StatusBadRequest)
		return
	}

	log.Printf("Search query received: %s", query)

	index, err := words.LoadIndex(cfg.IndexFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading index: %v", err), http.StatusInternalServerError)
		return
	}
	ids := words.SearchIndex(query, index)

	pages := make([]string, 0)
	for _, id := range ids {
		comic, err := database.GetComicByID(cfg.DBFile, id)
		if err != nil {
			log.Printf("Failed to get comic by ID %d: %v", id, err)
			continue
		}
		pageURL := fmt.Sprintf("https://xkcd.com/%d/", comic.Num)
		pages = append(pages, pageURL)
	}

	if len(pages) > 10 {
		pages = pages[:10]
	}

	log.Printf("Found %d comics matching query: %s", len(pages), query)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	formattedPages, err := json.MarshalIndent(pages, "", "    ")
	if err != nil {
		log.Printf("Error formatting JSON: %v", err)
		http.Error(w, "Error formatting response", http.StatusInternalServerError)
		return
	}

	w.Write(formattedPages)
}
