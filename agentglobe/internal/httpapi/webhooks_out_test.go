package httpapi

import (
	"context"
	"sync"
	"testing"
	"time"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
)

var _ domain.WebhookPoster = (*fakeWebhookPoster)(nil)

type fakeWebhookPoster struct {
	mu    sync.Mutex
	urls  []string
	bodies [][]byte
}

func (f *fakeWebhookPoster) Post(ctx context.Context, url string, body []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.urls = append(f.urls, url)
	f.bodies = append(f.bodies, append([]byte(nil), body...))
	return nil
}

func TestFireWebhooksInvokesPoster(t *testing.T) {
	s := testServer(t)
	fake := &fakeWebhookPoster{}
	s.WebhookPoster = fake

	db := s.DB
	pid := domain.NewEntityID()
	if err := db.Create(&dbpkg.Project{
		ID: pid, Name: "hookproj", Description: "", CreatedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatal(err)
	}
	wh := dbpkg.Webhook{
		ID: "wh-test-1", ProjectID: pid, URL: "https://example.invalid/webhook",
		Active: true, CreatedAt: time.Now().UTC(),
	}
	wh.SetEvents([]string{"new_post"})
	if err := db.Create(&wh).Error; err != nil {
		t.Fatal(err)
	}

	s.fireWebhooks(db, pid, "new_post", map[string]any{"hello": "world"})

	deadline := time.Now().Add(2 * time.Second)
	for {
		fake.mu.Lock()
		n := len(fake.urls)
		fake.mu.Unlock()
		if n >= 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("expected WebhookPoster.Post to be called")
		}
		time.Sleep(5 * time.Millisecond)
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.urls) != 1 || fake.urls[0] != wh.URL {
		t.Fatalf("unexpected urls: %#v", fake.urls)
	}
}
