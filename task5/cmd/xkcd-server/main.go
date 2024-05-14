package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Eduard-Bodreev/Yadro/gocomics/config"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/database"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/words"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/xkcd"
)

var cfg config.Config

func main() {
	cfg = config.InitConfig()

	go ScheduleDailyUpdates()

	http.HandleFunc("/update", handleUpdate)
	http.HandleFunc("/pics", handlePics)
	log.Printf("Server is starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
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
		http.Error(w, fmt.Sprintf("Error updating database: %v", err), http.StatusInternalServerError)
		return
	}

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

	index, err := words.LoadIndex(cfg.IndexFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading index: %v", err), http.StatusInternalServerError)
		return
	}
	ids := words.SearchIndex(query, index)

	pics := make([]string, 0)
	for _, id := range ids {
		comic, err := database.GetComicByID(cfg.DBFile, id)
		if err != nil {
			log.Printf("Failed to get comic by ID %d: %v", id, err)
			continue
		}
		pics = append(pics, comic.Img)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pics)
}
