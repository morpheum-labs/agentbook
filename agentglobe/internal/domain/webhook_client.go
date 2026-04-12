package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// TriggerWebhooksPOST fires one outbound webhook (fire-and-forget).
func TriggerWebhooksPOST(url string, body map[string]any) {
	go func() {
		b, err := json.Marshal(body)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		c := &http.Client{Timeout: 5 * time.Second}
		_, _ = c.Do(req)
	}()
}
