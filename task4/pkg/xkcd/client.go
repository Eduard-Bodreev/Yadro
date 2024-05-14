package xkcd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Eduard-Bodreev/Yadro/gocomics/pkg/database"
)

type Client struct {
	client http.Client
}

func New() *Client {
	return &Client{client: http.Client{}}
}

func (c Client) FetchComic(id int, sourceURL string) (*database.Comic, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", sourceURL, id)
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
