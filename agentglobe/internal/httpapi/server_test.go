package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/ratelimit"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	memName := strings.ReplaceAll(strings.ReplaceAll(t.Name(), "/", "_"), " ", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", memName)
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(
		&dbpkg.Category{}, &dbpkg.Agent{}, &dbpkg.Project{}, &dbpkg.ProjectMember{}, &dbpkg.Post{}, &dbpkg.Comment{},
		&dbpkg.Webhook{}, &dbpkg.GitHubWebhook{}, &dbpkg.Notification{}, &dbpkg.Attachment{},
		&dbpkg.FloorQuestion{}, &dbpkg.FloorExternalSignal{}, &dbpkg.FloorPosition{}, &dbpkg.FloorAgentTopicStat{}, &dbpkg.FloorAgentInferenceProfile{},
		&dbpkg.FloorDigestEntry{}, &dbpkg.FloorQuestionProbabilityPoint{}, &dbpkg.FloorPositionChallenge{}, &dbpkg.FloorResearchArticle{}, &dbpkg.FloorTopicProposal{}, &dbpkg.FloorBroadcast{},
		&dbpkg.FloorIndexPageMeta{}, &dbpkg.FloorIndexEntry{},
		&dbpkg.DebateThread{}, &dbpkg.DebatePost{}, &dbpkg.DebatePostReport{}, &dbpkg.AgentSanction{},
		&dbpkg.CapabilityService{},
		&dbpkg.MCPMemory{},
	); err != nil {
		t.Fatal(err)
	}
	if err := dbpkg.MigrateCategoryReferences(gdb); err != nil {
		t.Fatal(err)
	}
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

func TestCapabilityServicesRegisterListHeartbeat(t *testing.T) {
	s := testServer(t)
	s.Cfg.ServiceRegistryToken = "regtest"
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	base := ts.URL + "/api/v1/capability-services"
	res, err := http.Get(base)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list: %d", res.StatusCode)
	}
	var list0 struct {
		Count int `json:"count"`
		Items any `json:"items"`
	}
	_ = json.NewDecoder(res.Body).Decode(&list0)
	if list0.Count != 0 {
		t.Fatalf("count want 0")
	}
	body := map[string]any{
		"name":    "testsvc",
		"version": "0.0.1",
		"base_url": "http://127.0.0.1:9",
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, base+"/register", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer regtest")
	req.Header.Set("Content-Type", "application/json")
	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(res2.Body)
		t.Fatalf("register: %d %s", res2.StatusCode, string(bb))
	}
	var regOut struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res2.Body).Decode(&regOut); err != nil {
		t.Fatal(err)
	}
	if regOut.ID == "" {
		t.Fatal("register response missing id")
	}
	hb, _ := json.Marshal(map[string]string{
		"name": "testsvc", "base_url": "http://127.0.0.1:9",
	})
	req3, _ := http.NewRequest(http.MethodPost, base+"/heartbeat", bytes.NewReader(hb))
	req3.Header.Set("Authorization", "Bearer regtest")
	req3.Header.Set("Content-Type", "application/json")
	res3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatal(err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(res3.Body)
		t.Fatalf("heartbeat: %d %s", res3.StatusCode, string(bb))
	}
	res4, err := http.Get(base)
	if err != nil {
		t.Fatal(err)
	}
	defer res4.Body.Close()
	if res4.StatusCode != http.StatusOK {
		t.Fatalf("list2: %d", res4.StatusCode)
	}
	var list1 struct {
		Count int `json:"count"`
	}
	_ = json.NewDecoder(res4.Body).Decode(&list1)
	if list1.Count != 1 {
		t.Fatalf("count want 1, got %d", list1.Count)
	}
	gr, err := http.Get(ts.URL + "/api/v1/capability-services/" + regOut.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Body.Close()
	if gr.StatusCode != http.StatusOK {
		t.Fatalf("get by id: %d", gr.StatusCode)
	}
	flt, err := http.Get(ts.URL + "/api/v1/capability-services?q=test")
	if err != nil {
		t.Fatal(err)
	}
	defer flt.Body.Close()
	if flt.StatusCode != http.StatusOK {
		t.Fatalf("list with q: %d", flt.StatusCode)
	}
}

func TestCapabilityServiceRegisterIDLookupAndListFilter(t *testing.T) {
	s := testServer(t)
	s.Cfg.ServiceRegistryToken = "regtest"
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	base := ts.URL + "/api/v1/capability-services"
	body := `{"name":"f","version":"0.0.1","base_url":"http://127.0.0.1:7","category":"news","status":"degraded"}`
	req, _ := http.NewRequest(http.MethodPost, base+"/register", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer regtest")
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("register: %d", res.StatusCode)
	}
	res1, _ := http.Get(base + "?category=news&status=degraded")
	if res1.StatusCode != http.StatusOK {
		t.Fatalf("list filter: %d", res1.StatusCode)
	}
	defer res1.Body.Close()
	var v struct {
		Count int `json:"count"`
	}
	_ = json.NewDecoder(res1.Body).Decode(&v)
	if v.Count != 1 {
		t.Fatalf("filter count want 1 got %d", v.Count)
	}
}

func TestCapabilityServiceRegisterBadBaseURL(t *testing.T) {
	s := testServer(t)
	s.Cfg.ServiceRegistryToken = "regtest"
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	body := `{"name":"x","version":"1","base_url":"ftp://a"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/capability-services/register", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer regtest")
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	if res == nil {
		t.Fatal("nil res")
	}
	res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", res.StatusCode)
	}
}

func TestCapabilityServicesRegisterWithoutToken(t *testing.T) {
	s := testServer(t)
	// Cfg has no ServiceRegistryToken
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	body := `{"name":"x","version":"1","base_url":"http://a"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/capability-services/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer anything")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotImplemented {
		t.Fatalf("want 501, got %d", res.StatusCode)
	}
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

func TestPatchAgentsMe(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Post(ts.URL+"/api/v1/agents", "application/json", bytes.NewReader([]byte(`{"name":"PatchAgent"}`)))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		t.Fatalf("register %d: %s", res.StatusCode, string(b))
	}
	var reg map[string]any
	if err := json.NewDecoder(res.Body).Decode(&reg); err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	key, _ := reg["api_key"].(string)
	if key == "" {
		t.Fatal("missing api_key")
	}
	patchBody := `{"display_name":"Shown","bio":"hello","metadata":{"geo_cluster":"EU","capabilities":["macro"]}}`
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/v1/agents/me", strings.NewReader(patchBody))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res2.Body)
		t.Fatalf("patch %d: %s", res2.StatusCode, string(b))
	}
	var me map[string]any
	if err := json.NewDecoder(res2.Body).Decode(&me); err != nil {
		t.Fatal(err)
	}
	if me["display_name"] != "Shown" {
		t.Fatalf("display_name: %v", me["display_name"])
	}
	if me["bio"] != "hello" {
		t.Fatalf("bio: %v", me["bio"])
	}
	md, ok := me["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata type %T", me["metadata"])
	}
	if md["geo_cluster"] != "EU" {
		t.Fatalf("metadata.geo_cluster: %v", md["geo_cluster"])
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
