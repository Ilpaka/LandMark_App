package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// Stub always succeeds with a synthetic reservation id.
type Stub struct{}

func (Stub) ReserveNickname(ctx context.Context, nickname string) (ports.ProfileReservation, error) {
	_ = ctx
	_ = nickname
	return ports.ProfileReservation{ReservationID: uuid.New()}, nil
}

func (Stub) ReleaseReservation(ctx context.Context, reservationID uuid.UUID) error {
	_ = ctx
	_ = reservationID
	return nil
}

func (Stub) ConfirmReservation(ctx context.Context, reservationID uuid.UUID, userID uuid.UUID) error {
	_ = ctx
	_ = reservationID
	_ = userID
	return nil
}

// HTTPClient calls Profile service when BaseURL is non-empty.
type HTTPClient struct {
	BaseURL string
	HTTP    *http.Client
}

func (c *HTTPClient) client() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 5 * time.Second}
}

func (c *HTTPClient) ReserveNickname(ctx context.Context, nickname string) (ports.ProfileReservation, error) {
	body, _ := json.Marshal(map[string]string{"nickname": nickname})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/internal/v1/nicknames/reserve", bytes.NewReader(body))
	if err != nil {
		return ports.ProfileReservation{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return ports.ProfileReservation{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return ports.ProfileReservation{}, domain.ErrNicknameUnavailable
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return ports.ProfileReservation{}, fmt.Errorf("profile reserve: status %d", resp.StatusCode)
	}
	var out struct {
		ReservationID string `json:"reservation_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ports.ProfileReservation{}, err
	}
	id, err := uuid.Parse(out.ReservationID)
	if err != nil {
		return ports.ProfileReservation{}, err
	}
	return ports.ProfileReservation{ReservationID: id}, nil
}

func (c *HTTPClient) ReleaseReservation(ctx context.Context, reservationID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/internal/v1/nicknames/reserve/%s", c.BaseURL, reservationID), nil)
	if err != nil {
		return err
	}
	resp, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("profile release: status %d", resp.StatusCode)
	}
	return nil
}

func (c *HTTPClient) ConfirmReservation(ctx context.Context, reservationID uuid.UUID, userID uuid.UUID) error {
	body, _ := json.Marshal(map[string]string{"user_id": userID.String()})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/internal/v1/nicknames/reserve/%s/confirm", c.BaseURL, reservationID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("profile confirm: status %d", resp.StatusCode)
	}
	return nil
}
