package githubproc

import (
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(
		&db.Agent{}, &db.Project{}, &db.ProjectMember{}, &db.Post{}, &db.Comment{},
		&db.Webhook{}, &db.GitHubWebhook{}, &db.Notification{},
	); err != nil {
		t.Fatal(err)
	}
	return gdb
}

func TestGitHubPostMentionsValidated(t *testing.T) {
	gdb := testDB(t)
	now := time.Now().UTC()
	pid := domain.NewEntityID()
	botID := domain.NewEntityID()
	aliceID := domain.NewEntityID()
	if err := gdb.Create(&db.Agent{ID: botID, Name: "GitHubBot", APIKey: "k1", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&db.Agent{ID: aliceID, Name: "Alice", APIKey: "k2", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&db.Project{ID: pid, Name: "P", Description: "", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	cfg := db.GitHubWebhook{
		ID: domain.NewEntityID(), ProjectID: pid, Secret: "s", Active: true, CreatedAt: now,
	}
	cfg.SetEvents([]string{"pull_request"})
	cfg.SetLabels(nil)
	if err := gdb.Create(&cfg).Error; err != nil {
		t.Fatal(err)
	}
	ref := "https://github.com/o/r/pull/99"
	payload := map[string]any{
		"action": "opened",
		"pull_request": map[string]any{
			"number":   float64(99),
			"title":    "T",
			"html_url": ref,
			"body":     "cc @Alice @GhostWhoDoesNotExist",
			"user":     map[string]any{"login": "dev"},
		},
		"repository": map[string]any{"full_name": "o/r"},
	}
	var bot db.Agent
	if err := gdb.First(&bot, "id = ?", botID).Error; err != nil {
		t.Fatal(err)
	}
	allMu := sync.Mutex{}
	var result map[string]any
	if err := gdb.Transaction(func(tx *gorm.DB) error {
		result = ProcessGitHubEvent(tx, &cfg, "pull_request", payload, &bot, map[string]time.Time{}, &allMu)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if result == nil || result["post_id"] == nil {
		t.Fatal("expected post_created")
	}
	postID, _ := result["post_id"].(string)
	var post db.Post
	if err := gdb.First(&post, "id = ?", postID).Error; err != nil {
		t.Fatal(err)
	}
	got := post.Mentions()
	if len(got) != 1 || got[0] != "Alice" {
		t.Fatalf("mentions = %#v, want [Alice]", got)
	}
}

func TestGitHubAllStrippedWhenDisallowed(t *testing.T) {
	gdb := testDB(t)
	now := time.Now().UTC()
	pid := domain.NewEntityID()
	botID := domain.NewEntityID()
	if err := gdb.Create(&db.Agent{ID: botID, Name: "GitHubBot", APIKey: "k1", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&db.Project{ID: pid, Name: "P", Description: "", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	cfg := db.GitHubWebhook{
		ID: domain.NewEntityID(), ProjectID: pid, Secret: "s", Active: true, CreatedAt: now,
	}
	cfg.SetEvents([]string{"pull_request"})
	cfg.SetLabels(nil)
	if err := gdb.Create(&cfg).Error; err != nil {
		t.Fatal(err)
	}
	ref := "https://github.com/o/r/pull/100"
	payload := map[string]any{
		"action": "opened",
		"pull_request": map[string]any{
			"number":   float64(100),
			"title":    "All test",
			"html_url": ref,
			"body":     "hey @all team",
			"user":     map[string]any{"login": "dev"},
		},
		"repository": map[string]any{"full_name": "o/r"},
	}
	var bot db.Agent
	if err := gdb.First(&bot, "id = ?", botID).Error; err != nil {
		t.Fatal(err)
	}
	allMu := sync.Mutex{}
	var result map[string]any
	if err := gdb.Transaction(func(tx *gorm.DB) error {
		result = ProcessGitHubEvent(tx, &cfg, "pull_request", payload, &bot, map[string]time.Time{}, &allMu)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	postID, _ := result["post_id"].(string)
	var post db.Post
	if err := gdb.First(&post, "id = ?", postID).Error; err != nil {
		t.Fatal(err)
	}
	for _, m := range post.Mentions() {
		if m == "all" {
			t.Fatal("did not expect @all in stored mentions when bot cannot use @all")
		}
	}
}

func TestGitHubSyncCommentReplyToPostAuthor(t *testing.T) {
	gdb := testDB(t)
	now := time.Now().UTC()
	pid := domain.NewEntityID()
	botID := domain.NewEntityID()
	aliceID := domain.NewEntityID()
	if err := gdb.Create(&db.Agent{ID: botID, Name: "GitHubBot", APIKey: "k1", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&db.Agent{ID: aliceID, Name: "Alice", APIKey: "k2", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&db.Project{ID: pid, Name: "P", Description: "", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	cfg := db.GitHubWebhook{
		ID: domain.NewEntityID(), ProjectID: pid, Secret: "s", Active: true, CreatedAt: now,
	}
	cfg.SetEvents([]string{"pull_request"})
	cfg.SetLabels(nil)
	if err := gdb.Create(&cfg).Error; err != nil {
		t.Fatal(err)
	}
	ref := "https://github.com/o/r/pull/101"
	postID := domain.NewEntityID()
	githubRef := ref
	post := db.Post{
		ID: postID, ProjectID: pid, AuthorID: aliceID, Title: "PR", Content: "x",
		Type: "review", Status: "open", CreatedAt: now, UpdatedAt: now,
	}
	post.GithubRef = &githubRef
	if err := gdb.Create(&post).Error; err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{
		"action": "synchronize",
		"pull_request": map[string]any{
			"number":   float64(101),
			"title":    "PR",
			"html_url": ref,
			"body":     "",
			"user":     map[string]any{"login": "dev"},
		},
		"repository": map[string]any{"full_name": "o/r"},
	}
	var bot db.Agent
	if err := gdb.First(&bot, "id = ?", botID).Error; err != nil {
		t.Fatal(err)
	}
	allMu := sync.Mutex{}
	if err := gdb.Transaction(func(tx *gorm.DB) error {
		ProcessGitHubEvent(tx, &cfg, "pull_request", payload, &bot, map[string]time.Time{}, &allMu)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := gdb.Model(&db.Notification{}).Where("agent_id = ? AND type = ?", aliceID, "reply").Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("reply notifications = %d, want 1", n)
	}
}
