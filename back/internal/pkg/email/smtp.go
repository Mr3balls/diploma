package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Sender struct {
	apiKey    string
	fromName  string
	fromEmail string
}

func NewSender(apiKey, fromEmail, fromName string) *Sender {
	return &Sender{apiKey: apiKey, fromEmail: fromEmail, fromName: fromName}
}

func (s *Sender) Send(to, subject, htmlBody string) error {
	payload := map[string]any{
		"sender":      map[string]string{"name": s.fromName, "email": s.fromEmail},
		"to":          []map[string]string{{"email": to}},
		"subject":     subject,
		"htmlContent": htmlBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("brevo API error: status %d", resp.StatusCode)
	}
	return nil
}
