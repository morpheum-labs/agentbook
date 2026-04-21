package db

import (
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSanitizeDebatePlain_stripsHTML(t *testing.T) {
	in := `  <p>Hello</p><script>alert(1)</script><iframe src="x"></iframe>  `
	got := SanitizeDebatePlain(in, 1000)
	if strings.Contains(strings.ToLower(got), "<script") || strings.Contains(got, "<iframe") {
		t.Fatalf("expected scripts/iframes removed, got %q", got)
	}
	if !strings.Contains(got, "Hello") {
		t.Fatalf("expected plain text preserved, got %q", got)
	}
}

func TestSanitizeDebatePlain_truncatesRunes(t *testing.T) {
	in := strings.Repeat("é", 10) // 2-byte UTF-8 runes
	got := SanitizeDebatePlain(in, 5)
	if got == "" || len([]rune(got)) != 5 {
		t.Fatalf("want 5 runes, got %q len runes %d", got, len([]rune(got)))
	}
}

func TestSanitizeDebateToken(t *testing.T) {
	if g := SanitizeDebateToken("Spam-AD_v2!", 20); g != "spam-ad_v2" {
		t.Fatalf("got %q", g)
	}
}

func TestDebatePostBeforeSave_rejectsEmptyAfterStrip(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file:debate_sanitize?mode=memory&cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(&Agent{}, &DebateThread{}, &DebatePost{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	aid := uuid.NewString()
	if err := gdb.Create(&Agent{ID: aid, Name: "d1", APIKey: "k1", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	tid := uuid.NewString()
	if err := gdb.Create(&DebateThread{ID: tid, Title: "T", CreatedByAgentID: aid}).Error; err != nil {
		t.Fatal(err)
	}
	p := &DebatePost{
		ID:       uuid.NewString(),
		ThreadID: tid,
		AuthorID: aid,
		Content:  `<script>z</script>`,
	}
	if err := gdb.Create(p).Error; err == nil {
		t.Fatal("expected error for empty content after sanitization")
	}
	p.Content = "ok " + strings.Repeat("x", 50)
	if err := gdb.Create(p).Error; err != nil {
		t.Fatal(err)
	}
	var loaded DebatePost
	if err := gdb.First(&loaded, "id = ?", p.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(loaded.Content, "ok ") {
		t.Fatalf("content: %q", loaded.Content)
	}
}
