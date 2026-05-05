package moderation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Client — HTTP-клиент к moderation-service для постановки места в очередь
// модерации. Отдельный пакет вместо синхронного outbox-publisher'а — простое
// и достаточное решение для текущего объёма; outbox остаётся для async-аудита.
type Client struct {
	baseURL     string
	internalKey string
	http        *http.Client
}

func NewClient(baseURL, internalKey string) *Client {
	return &Client{
		baseURL:     baseURL,
		internalKey: internalKey,
		http:        &http.Client{Timeout: 5 * time.Second},
	}
}

// SubmitPlace ставит место в очередь модерации. Идемпотентно на стороне
// moderation-service (повторная отправка одного и того же target_id отдаёт
// существующий queue-item). Если URL не сконфигурирован — no-op (для тестов).
func (c *Client) SubmitPlace(ctx context.Context, placeID, submittedBy uuid.UUID) error {
	if c == nil || c.baseURL == "" {
		return nil
	}
	body := map[string]string{
		"target_type":  "place",
		"target_id":    placeID.String(),
		"submitted_by": submittedBy.String(),
	}
	buf, _ := json.Marshal(body)
	url := c.baseURL + "/v1/internal/moderation/submit"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.internalKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("moderation submit: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("moderation submit: status %d", resp.StatusCode)
	}
	return nil
}
