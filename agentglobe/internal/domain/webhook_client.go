package domain

import (
	"context"
	"encoding/json"
)

// TriggerWebhooksPOST fires one outbound webhook in a new goroutine (fire-and-forget).
// Prefer injecting [WebhookPoster] on the HTTP server for bounded concurrency; this remains for ad-hoc use.
func TriggerWebhooksPOST(url string, body map[string]any) {
	go func() {
		b, err := json.Marshal(body)
		if err != nil {
			return
		}
		_ = NewHTTPWebhookPoster().Post(context.Background(), url, b)
	}()
}
