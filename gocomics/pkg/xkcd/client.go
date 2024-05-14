package xkcd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/database"

	"github.com/spf13/viper"
)

type Client struct {
	client http.Client
}

func New() *Client {
	return &Client{client: http.Client{}}
}

func (c Client) GetCurrentMaxComicNum(sourceURL string) (int, error) {
	resp, err := c.client.Get("https://xkcd.com/info.0.json")
	if err != nil {
		return 0, fmt.Errorf("error fetching the latest comic: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("received non-200 response status: %d", resp.StatusCode)
	}

	var comic database.Comic
	if err := json.NewDecoder(resp.Body).Decode(&comic); err != nil {
		return 0, fmt.Errorf("error decoding comic: %v", err)
	}
	return comic.Num, nil
}

func (c Client) FetchComic(id int, sourceURL string) (*database.Comic, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", viper.GetString("source_url"), id)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching comic: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("comic %d not found", id)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response status: %d", resp.StatusCode)
	}

	var comic database.Comic
	if err := json.NewDecoder(resp.Body).Decode(&comic); err != nil {
		return nil, fmt.Errorf("error decoding comic: %v", err)
	}

	return &comic, nil
}
