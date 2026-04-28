package placesclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type Client struct {
	BaseURL     string
	InternalKey string
	HTTP        *http.Client
}

func New() *Client {
	base := os.Getenv("PLACES_SERVICE_URL")
	if base == "" {
		base = "http://places-service:8080"
	}
	return &Client{
		BaseURL:     base,
		InternalKey: os.Getenv("INTERNAL_API_KEY"),
		HTTP:        &http.Client{},
	}
}

func (c *Client) ApprovePlace(ctx context.Context, placeID uuid.UUID) error {
	url := fmt.Sprintf("%s/v1/places/internal/%s/approve", c.BaseURL, placeID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	req.Header.Set("X-Internal-Key", c.InternalKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("places approve: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) RejectPlace(ctx context.Context, placeID uuid.UUID, reason string) error {
	url := fmt.Sprintf("%s/v1/places/internal/%s/reject", c.BaseURL, placeID)
	body, _ := json.Marshal(map[string]string{"reason": reason})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.InternalKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("places reject: status %d", resp.StatusCode)
	}
	return nil
}
