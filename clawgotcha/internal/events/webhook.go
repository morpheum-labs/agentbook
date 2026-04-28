package events

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"gorm.io/gorm"
)

const signatureHeader = "X-Clawgotcha-Signature"

// WebhookDispatcher POSTs signed payloads to runtime callback URLs for matching subscriptions.
type WebhookDispatcher struct {
	DB     *gorm.DB
	Client *http.Client
}

// Deliver loads enabled subscriptions (with runtime callback URLs) and POSTs the event.
func (d *WebhookDispatcher) Deliver(ev ChangeEvent) {
	if d == nil || d.DB == nil {
		return
	}
	body, err := MarshalJSONBytes(ev)
	if err != nil {
		return
	}
	var subs []db.SwarmWebhookSubscription
	q := d.DB.Preload("Runtime").Where("enabled = ?", true)
	if err := q.Find(&subs).Error; err != nil {
		return
	}
	client := d.Client
	if client == nil {
		client = http.DefaultClient
	}
	for i := range subs {
		s := &subs[i]
		if s.Runtime == nil || s.Runtime.CallbackURL == "" {
			continue
		}
		if len(s.EventTypes) > 0 && !MatchesSubscription(s.EventTypes, ev.EventType) {
			continue
		}
		mac := hmac.New(sha256.New, []byte(s.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req, err := http.NewRequest(http.MethodPost, s.Runtime.CallbackURL, bytes.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set(signatureHeader, "sha256="+sig)
		req.Header.Set("X-Clawgotcha-Event-Type", ev.EventType)
		ctx := req.Context()
		_ = ctx
		resp, err := client.Do(req)
		if resp != nil {
			_ = resp.Body.Close()
		}
		_ = err
	}
}

// DefaultHTTPClient returns a client with sane timeouts for webhook delivery.
func DefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}
