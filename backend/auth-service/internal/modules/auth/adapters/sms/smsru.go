package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SMSruSender sends SMS via https://sms.ru/ API.
type SMSruSender struct {
	APIID   string
	Client  *http.Client
	Timeout time.Duration
}

type smsruResponse struct {
	Status     string          `json:"status"`
	StatusCode int             `json:"status_code"`
	Sms        json.RawMessage `json:"sms"`
}

// SendOTP implements ports.SMSSender.
func (s *SMSruSender) SendOTP(ctx context.Context, phoneE164, code string) error {
	if s.APIID == "" {
		return fmt.Errorf("smsru: empty api id")
	}
	phoneDigits := strings.TrimPrefix(strings.TrimSpace(phoneE164), "+")
	q := url.Values{}
	q.Set("api_id", s.APIID)
	q.Set("to", phoneDigits)
	q.Set("msg", "Код подтверждения: "+code)
	q.Set("json", "1")
	u := "https://sms.ru/sms/send?" + q.Encode()
	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("smsru: http %d", resp.StatusCode)
	}
	var out smsruResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("smsru: json: %w", err)
	}
	// status "OK" and status_code 100 = enqueued; see SMS.ru docs.
	if out.Status != "OK" {
		if out.StatusCode != 0 {
			return fmt.Errorf("smsru: status %s code %d", out.Status, out.StatusCode)
		}
		return fmt.Errorf("smsru: status %s", out.Status)
	}
	if out.StatusCode != 0 && out.StatusCode != 100 {
		return fmt.Errorf("smsru: status_code %d", out.StatusCode)
	}
	return nil
}
