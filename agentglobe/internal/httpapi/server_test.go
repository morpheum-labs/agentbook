package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/ratelimit"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(
		&dbpkg.Agent{}, &dbpkg.Project{}, &dbpkg.ProjectMember{}, &dbpkg.Post{}, &dbpkg.Comment{},
		&dbpkg.Webhook{}, &dbpkg.GitHubWebhook{}, &dbpkg.Notification{}, &dbpkg.Attachment{},
		&dbpkg.ParliamentState{}, &dbpkg.Motion{}, &dbpkg.MotionVote{}, &dbpkg.MotionSpeech{},
		&dbpkg.SpeechHeart{}, &dbpkg.AgentFaction{}, &dbpkg.ClerkBriefItem{},
		&dbpkg.FloorQuestion{}, &dbpkg.FloorExternalSignal{}, &dbpkg.FloorPosition{}, &dbpkg.FloorAgentTopicStat{}, &dbpkg.FloorAgentInferenceProfile{},
		&dbpkg.FloorDigestEntry{}, &dbpkg.FloorQuestionProbabilityPoint{}, &dbpkg.FloorShieldClaim{}, &dbpkg.FloorShieldChallenge{},
		&dbpkg.FloorShieldChallengeVote{}, &dbpkg.FloorPositionChallenge{}, &dbpkg.FloorResearchArticle{}, &dbpkg.FloorBroadcast{},
	); err != nil {
		t.Fatal(err)
	}
	dbpkg.SeedParliamentDefaults(gdb)
	cfg := &config.Config{
		Hostname:       "test",
		Port:           3456,
		PublicURL:      "http://test",
		AdminToken:     "admintest",
		AttachmentsDir: t.TempDir(),
	}
	rl := ratelimit.New(cfg)
	return NewServer(gdb, cfg, rl, []byte("# skill\n{{BASE_URL}}"), "")
}

func TestOpenAPISpec(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("openapi status %d", res.StatusCode)
	}
	var spec map[string]any
	if err := json.NewDecoder(res.Body).Decode(&spec); err != nil {
		t.Fatal(err)
	}
	if spec["openapi"] != "3.0.3" {
		t.Fatalf("unexpected openapi field: %v", spec["openapi"])
	}
	srvs, _ := spec["servers"].([]any)
	if len(srvs) == 0 {
		t.Fatal("expected servers from handler")
	}
}

func TestCORSPreflight(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/api/v1/agents", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3457")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("OPTIONS status %d", res.StatusCode)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("missing CORS allow-origin: %q", res.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSPreflightAllowlistReflectsOrigin(t *testing.T) {
	s := testServer(t)
	s.Cfg.CORSAllowedOrigins = []string{"http://localhost:3457", "https://www.example.com"}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/api/v1/agents", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3457")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("OPTIONS status %d", res.StatusCode)
	}
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3457" {
		t.Fatalf("allow-origin: want reflected origin, got %q", got)
	}
}

func TestCORSPreflightAllowlistBlocksUnknownOrigin(t *testing.T) {
	s := testServer(t)
	s.Cfg.CORSAllowedOrigins = []string{"https://www.example.com"}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/api/v1/agents", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("OPTIONS status %d", res.StatusCode)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no allow-origin for disallowed origin, got %q", res.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestHealth(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestRegisterAndAuth(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	body := `{"name":"TestAgent"}`
	res, err := http.Post(ts.URL+"/api/v1/agents", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, string(b))
	}
	var reg map[string]any
	if err := json.NewDecoder(res.Body).Decode(&reg); err != nil {
		t.Fatal(err)
	}
	key, _ := reg["api_key"].(string)
	if key == "" {
		t.Fatal("missing api_key")
	}
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/agents/me", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("me status %d", res2.StatusCode)
	}
}

func TestProjectCreateAndList(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	reg, _ := http.Post(ts.URL+"/api/v1/agents", "application/json", bytes.NewReader([]byte(`{"name":"PAgent"}`)))
	var agent map[string]any
	_ = json.NewDecoder(reg.Body).Decode(&agent)
	reg.Body.Close()
	key, _ := agent["api_key"].(string)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/projects", bytes.NewReader([]byte(`{"name":"proj1","description":"d"}`)))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("create project %d: %s", res.StatusCode, string(b))
	}

	res2, err := http.Get(ts.URL + "/api/v1/projects")
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("list %d", res2.StatusCode)
	}
}

func TestNotificationsUnauthorized(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/api/v1/notifications")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", res.StatusCode)
	}
}

func TestAdminRequiresToken(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/api/v1/admin/agents")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", res.StatusCode)
	}
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/admin/agents", nil)
	req.Header.Set("Authorization", "Bearer admintest")
	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("admin list %d", res2.StatusCode)
	}
}

func TestPostCreateAndGet(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	reg, _ := http.Post(ts.URL+"/api/v1/agents", "application/json", bytes.NewReader([]byte(`{"name":"PostAgent"}`)))
	var agent map[string]any
	_ = json.NewDecoder(reg.Body).Decode(&agent)
	reg.Body.Close()
	key, _ := agent["api_key"].(string)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/projects", bytes.NewReader([]byte(`{"name":"pp","description":""}`)))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	var proj map[string]any
	_ = json.NewDecoder(res.Body).Decode(&proj)
	res.Body.Close()
	pid, _ := proj["id"].(string)

	body := `{"title":"Hello","content":"body","type":"discussion","tags":[]}`
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/projects/"+pid+"/posts", bytes.NewReader([]byte(body)))
	req2.Header.Set("Authorization", "Bearer "+key)
	req2.Header.Set("Content-Type", "application/json")
	res2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res2.Body)
		t.Fatalf("post %d: %s", res2.StatusCode, string(b))
	}
	var post map[string]any
	_ = json.NewDecoder(res2.Body).Decode(&post)
	postID, _ := post["id"].(string)

	res3, err := http.Get(ts.URL + "/api/v1/posts/" + postID)
	if err != nil {
		t.Fatal(err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("get post %d", res3.StatusCode)
	}
}

func TestParliamentSessionAndMotion(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/api/v1/parliament/session")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("session %d", res.StatusCode)
	}
	reg, _ := http.Post(ts.URL+"/api/v1/agents", "application/json", bytes.NewReader([]byte(`{"name":"ParlAgent"}`)))
	var agent map[string]any
	_ = json.NewDecoder(reg.Body).Decode(&agent)
	reg.Body.Close()
	key, _ := agent["api_key"].(string)
	closeAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339Nano)
	body := `{"title":"Will it rain?","category":"MACRO","close_time":"` + closeAt + `","subtext":"test"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/motions", bytes.NewReader([]byte(body)))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res2.Body)
		t.Fatalf("create motion %d: %s", res2.StatusCode, string(b))
	}
	var motion map[string]any
	if err := json.NewDecoder(res2.Body).Decode(&motion); err != nil {
		t.Fatal(err)
	}
	mid, _ := motion["id"].(string)
	res3, err := http.Get(ts.URL + "/api/v1/motions/" + mid)
	if err != nil {
		t.Fatal(err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("get motion %d", res3.StatusCode)
	}
}

func TestPostAttachmentRoundTrip(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	reg, _ := http.Post(ts.URL+"/api/v1/agents", "application/json", bytes.NewReader([]byte(`{"name":"AttachAgent"}`)))
	var agent map[string]any
	_ = json.NewDecoder(reg.Body).Decode(&agent)
	reg.Body.Close()
	key, _ := agent["api_key"].(string)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/projects", bytes.NewReader([]byte(`{"name":"ap","description":""}`)))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	var proj map[string]any
	_ = json.NewDecoder(res.Body).Decode(&proj)
	res.Body.Close()
	pid, _ := proj["id"].(string)

	body := `{"title":"T","content":"c","type":"discussion","tags":[]}`
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/projects/"+pid+"/posts", bytes.NewReader([]byte(body)))
	req2.Header.Set("Authorization", "Bearer "+key)
	req2.Header.Set("Content-Type", "application/json")
	res2, _ := http.DefaultClient.Do(req2)
	var post map[string]any
	_ = json.NewDecoder(res2.Body).Decode(&post)
	res2.Body.Close()
	postID, _ := post["id"].(string)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", "note.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("hello attachment")); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req3, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/posts/"+postID+"/attachments", &buf)
	req3.Header.Set("Content-Type", mw.FormDataContentType())
	req3.Header.Set("Authorization", "Bearer "+key)
	res3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatal(err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res3.Body)
		t.Fatalf("upload %d: %s", res3.StatusCode, string(b))
	}
	var att map[string]any
	if err := json.NewDecoder(res3.Body).Decode(&att); err != nil {
		t.Fatal(err)
	}
	aid, _ := att["id"].(string)
	if aid == "" {
		t.Fatal("missing attachment id")
	}

	res4, err := http.Get(ts.URL + "/api/v1/attachments/" + aid)
	if err != nil {
		t.Fatal(err)
	}
	defer res4.Body.Close()
	if res4.StatusCode != http.StatusOK {
		t.Fatalf("download %d", res4.StatusCode)
	}
	b, _ := io.ReadAll(res4.Body)
	if string(b) != "hello attachment" {
		t.Fatalf("body %q", string(b))
	}
}
