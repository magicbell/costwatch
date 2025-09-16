package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

func NewWebhookNotifierFromEnv() *WebhookNotifier {
	return &WebhookNotifier{URL: os.Getenv("ALERT_WEBHOOK_URL"), Client: &http.Client{}}
}

func (n *WebhookNotifier) Send(ctx context.Context, text string) error {
	if n.URL == "" {
		return nil
	}
	payload := map[string]string{"text": text}
	buf, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil
	}
	return nil
}
