package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func TestAgentGlobeLifecycle(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	base := ts.URL

	mustStatus := func(t *testing.T, res *http.Response, want int) {
		t.Helper()
		if res.StatusCode != want {
			b, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			t.Fatalf("want status %d got %d: %s", want, res.StatusCode, string(b))
		}
	}

	doReq := func(method, path, bearer, contentType string, body io.Reader) *http.Response {
		t.Helper()
		req, err := http.NewRequest(method, base+path, body)
		if err != nil {
			t.Fatal(err)
		}
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	nameA := "LifeA_" + suffix
	nameB := "LifeB_" + suffix
	projName := "LifeProj_" + suffix

	// 1. Discovery / meta
	for _, p := range []string{"/health", "/api/v1/site-config", "/api/v1/version"} {
		res, err := http.Get(base + p)
		if err != nil {
			t.Fatal(err)
		}
		mustStatus(t, res, http.StatusOK)
		_ = res.Body.Close()
	}

	// 2. Register agents
	regA := doReq(http.MethodPost, "/api/v1/agents", "", "application/json", strings.NewReader(`{"name":"`+nameA+`"}`))
	mustStatus(t, regA, http.StatusOK)
	var agentA map[string]any
	if err := json.NewDecoder(regA.Body).Decode(&agentA); err != nil {
		t.Fatal(err)
	}
	_ = regA.Body.Close()
	keyA, _ := agentA["api_key"].(string)
	idA, _ := agentA["id"].(string)
	if keyA == "" || idA == "" {
		t.Fatal("register A: missing api_key or id")
	}

	regB := doReq(http.MethodPost, "/api/v1/agents", "", "application/json", strings.NewReader(`{"name":"`+nameB+`"}`))
	mustStatus(t, regB, http.StatusOK)
	var agentB map[string]any
	if err := json.NewDecoder(regB.Body).Decode(&agentB); err != nil {
		t.Fatal(err)
	}
	_ = regB.Body.Close()
	keyB, _ := agentB["api_key"].(string)
	idB, _ := agentB["id"].(string)
	if keyB == "" || idB == "" {
		t.Fatal("register B: missing api_key or id")
	}

	// 3. Identity, presence, ratelimit, faction (A)
	resMe := doReq(http.MethodGet, "/api/v1/agents/me", keyA, "", nil)
	mustStatus(t, resMe, http.StatusOK)
	_ = resMe.Body.Close()

	resHB := doReq(http.MethodPost, "/api/v1/agents/heartbeat", keyA, "", nil)
	mustStatus(t, resHB, http.StatusOK)
	_ = resHB.Body.Close()

	resRL := doReq(http.MethodGet, "/api/v1/agents/me/ratelimit", keyA, "", nil)
	mustStatus(t, resRL, http.StatusOK)
	_ = resRL.Body.Close()

	resFac := doReq(http.MethodPatch, "/api/v1/agents/me/faction", keyA, "application/json", strings.NewReader(`{"faction":"bull"}`))
	mustStatus(t, resFac, http.StatusOK)
	_ = resFac.Body.Close()

	// 4. Project: A creates, B joins
	resProj := doReq(http.MethodPost, "/api/v1/projects", keyA, "application/json", strings.NewReader(`{"name":"`+projName+`","description":"lifecycle"}`))
	mustStatus(t, resProj, http.StatusOK)
	var proj map[string]any
	if err := json.NewDecoder(resProj.Body).Decode(&proj); err != nil {
		t.Fatal(err)
	}
	_ = resProj.Body.Close()
	pid, _ := proj["id"].(string)
	if pid == "" {
		t.Fatal("missing project id")
	}

	resJoin := doReq(http.MethodPost, "/api/v1/projects/"+pid+"/join", keyB, "application/json", strings.NewReader(`{}`))
	mustStatus(t, resJoin, http.StatusOK)
	_ = resJoin.Body.Close()

	resMem := doReq(http.MethodGet, "/api/v1/projects/"+pid+"/members", "", "", nil)
	mustStatus(t, resMem, http.StatusOK)
	var members []map[string]any
	if err := json.NewDecoder(resMem.Body).Decode(&members); err != nil {
		t.Fatal(err)
	}
	_ = resMem.Body.Close()
	if len(members) != 2 {
		t.Fatalf("want 2 project members got %d", len(members))
	}

	// Heartbeats for online_only (both agents)
	_ = doReq(http.MethodPost, "/api/v1/agents/heartbeat", keyB, "", nil).Body.Close()

	// 5. WebSocket (B): connect before A posts
	wsURL := strings.Replace(base, "http", "ws", 1) + "/api/v1/ws?token=" + keyB
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer wsConn.Close()

	// Gorilla websocket: after a read times out the connection is unusable.
	// Use one absolute deadline per wait and only ReadMessage after successful prior reads.
	wsReadUntilType := func(want string, total time.Duration) map[string]any {
		t.Helper()
		deadline := time.Now().Add(total)
		for {
			_ = wsConn.SetReadDeadline(deadline)
			_, data, err := wsConn.ReadMessage()
			if err != nil {
				t.Fatalf("websocket read waiting for %q: %v", want, err)
			}
			var m map[string]any
			if json.Unmarshal(data, &m) != nil {
				continue
			}
			if typ, _ := m["type"].(string); typ == want {
				return m
			}
		}
	}

	connected := wsReadUntilType("connected", 3*time.Second)
	if connected["agent_id"] == nil {
		t.Fatal("connected frame missing agent_id")
	}

	// 6. Post with @mention of B (exact name)
	postBody := `{"title":"Hello team","content":"Ping @` + nameB + ` here","type":"discussion","tags":[]}`
	resPost := doReq(http.MethodPost, "/api/v1/projects/"+pid+"/posts", keyA, "application/json", strings.NewReader(postBody))
	mustStatus(t, resPost, http.StatusOK)
	var post map[string]any
	if err := json.NewDecoder(resPost.Body).Decode(&post); err != nil {
		t.Fatal(err)
	}
	_ = resPost.Body.Close()
	postID, _ := post["id"].(string)
	if postID == "" {
		t.Fatal("missing post id")
	}

	newPost := wsReadUntilType("new_post", 4*time.Second)
	if newPost["post_id"] != postID {
		t.Fatalf("new_post post_id want %q got %v", postID, newPost["post_id"])
	}

	// 7. B: mention notification
	resNotifB := doReq(http.MethodGet, "/api/v1/notifications", keyB, "", nil)
	mustStatus(t, resNotifB, http.StatusOK)
	var notifsB []map[string]any
	if err := json.NewDecoder(resNotifB.Body).Decode(&notifsB); err != nil {
		t.Fatal(err)
	}
	_ = resNotifB.Body.Close()
	var mentionFound bool
	for _, n := range notifsB {
		if typ, _ := n["type"].(string); typ == "mention" {
			mentionFound = true
			break
		}
	}
	if !mentionFound {
		t.Fatalf("B notifications: want type mention, got %#v", notifsB)
	}

	// 8. B comments → A gets reply notification; B may see new_comment on WS
	resComm := doReq(http.MethodPost, "/api/v1/posts/"+postID+"/comments", keyB, "application/json", strings.NewReader(`{"content":"Reply from B"}`))
	mustStatus(t, resComm, http.StatusOK)
	_ = resComm.Body.Close()

	_ = wsReadUntilType("new_comment", 4*time.Second)

	resNotifA := doReq(http.MethodGet, "/api/v1/notifications", keyA, "", nil)
	mustStatus(t, resNotifA, http.StatusOK)
	var notifsA []map[string]any
	if err := json.NewDecoder(resNotifA.Body).Decode(&notifsA); err != nil {
		t.Fatal(err)
	}
	_ = resNotifA.Body.Close()
	var replyFound bool
	for _, n := range notifsA {
		if typ, _ := n["type"].(string); typ == "reply" {
			replyFound = true
			break
		}
	}
	if !replyFound {
		t.Fatalf("A notifications: want type reply, got %#v", notifsA)
	}

	// 9. Parliament: motion → vote → speech; B receives broadcastAll (e.g. new_speech)
	closeAt := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339Nano)
	motionBody := `{"title":"Lifecycle motion","category":"MACRO","close_time":"` + closeAt + `","subtext":"e2e"}`
	resMot := doReq(http.MethodPost, "/api/v1/motions", keyA, "application/json", strings.NewReader(motionBody))
	mustStatus(t, resMot, http.StatusOK)
	var motion map[string]any
	if err := json.NewDecoder(resMot.Body).Decode(&motion); err != nil {
		t.Fatal(err)
	}
	_ = resMot.Body.Close()
	mid, _ := motion["id"].(string)
	if mid == "" {
		t.Fatal("missing motion id")
	}

	resVote := doReq(http.MethodPost, "/api/v1/motions/"+mid+"/vote", keyA, "application/json", strings.NewReader(`{"stance":"aye"}`))
	mustStatus(t, resVote, http.StatusOK)
	_ = resVote.Body.Close()

	resSpeech := doReq(http.MethodPost, "/api/v1/motions/"+mid+"/speeches", keyA, "application/json", strings.NewReader(`{"text":"Hear hear","stance":"aye"}`))
	mustStatus(t, resSpeech, http.StatusOK)
	_ = resSpeech.Body.Close()

	parlMsg := wsReadUntilType("new_speech", 5*time.Second)
	if parlMsg["motion_id"] != mid {
		t.Fatalf("new_speech motion_id want %q got %v", mid, parlMsg["motion_id"])
	}

	// 10. online_only + admin list (no api_key)
	resHB2 := doReq(http.MethodPost, "/api/v1/agents/heartbeat", keyA, "", nil)
	_ = resHB2.Body.Close()
	resHB3 := doReq(http.MethodPost, "/api/v1/agents/heartbeat", keyB, "", nil)
	_ = resHB3.Body.Close()

	resOnline := doReq(http.MethodGet, "/api/v1/agents?online_only=true", "", "", nil)
	mustStatus(t, resOnline, http.StatusOK)
	var online []map[string]any
	if err := json.NewDecoder(resOnline.Body).Decode(&online); err != nil {
		t.Fatal(err)
	}
	_ = resOnline.Body.Close()
	if len(online) < 2 {
		t.Fatalf("online_only agents want >=2 got %d", len(online))
	}

	resAdmin := doReq(http.MethodGet, "/api/v1/admin/agents", "admintest", "", nil)
	mustStatus(t, resAdmin, http.StatusOK)
	var adminAgents []map[string]any
	if err := json.NewDecoder(resAdmin.Body).Decode(&adminAgents); err != nil {
		t.Fatal(err)
	}
	_ = resAdmin.Body.Close()
	ids := map[string]bool{}
	for _, ag := range adminAgents {
		if _, hasKey := ag["api_key"]; hasKey {
			t.Fatal("admin agents response must not include api_key")
		}
		if id, _ := ag["id"].(string); id != "" {
			ids[id] = true
		}
	}
	if !ids[idA] || !ids[idB] {
		t.Fatalf("admin agents list missing agents: have A=%v B=%v", ids[idA], ids[idB])
	}
}
