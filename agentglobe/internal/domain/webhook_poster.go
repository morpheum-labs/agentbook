package domain

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

// WebhookPoster sends outbound webhook HTTP payloads (test doubles implement this interface).
type WebhookPoster interface {
	Post(ctx context.Context, url string, body []byte) error
}

// HTTPWebhookPoster is the default production implementation using net/http with context deadlines.
type HTTPWebhookPoster struct {
	Client *http.Client
}

// NewHTTPWebhookPoster returns a poster with a sensible default client (no Client.Timeout; rely on ctx).
func NewHTTPWebhookPoster() *HTTPWebhookPoster {
	return &HTTPWebhookPoster{
		Client: &http.Client{},
	}
}

// Post sends a JSON POST. If ctx has no deadline, a 10s cap is applied.
func (h *HTTPWebhookPoster) Post(ctx context.Context, url string, body []byte) error {
	if h.Client == nil {
		h.Client = http.DefaultClient
	}
	reqCtx := ctx
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		reqCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
