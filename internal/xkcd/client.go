package xkcd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Eduard-Bodreev/Yadro/gocomics/internal/models"
)

type Client struct {
	client    http.Client
	sourceURL string
}

func New(sourceURL string) *Client {
	return &Client{
		client:    http.Client{},
		sourceURL: sourceURL,
	}
}

func (c *Client) FetchComic(id int) (*models.Comic, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.sourceURL, id)
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

	var comic models.Comic
	if err := json.NewDecoder(resp.Body).Decode(&comic); err != nil {
		return nil, fmt.Errorf("error decoding comic: %v", err)
	}

	return &comic, nil
}
