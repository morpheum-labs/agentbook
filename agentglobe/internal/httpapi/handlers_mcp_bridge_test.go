package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

func TestUpsertMCPMemoryAndNotify(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	gdb := s.DB
	alice := dbpkg.Agent{ID: "a1", Name: "alice", APIKey: "key-alice"}
	bob := dbpkg.Agent{ID: "b1", Name: "bob", APIKey: "key-bob"}
	if err := gdb.Create(&alice).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&bob).Error; err != nil {
		t.Fatal(err)
	}

	memBody := map[string]any{"key": "note1", "namespace": "ns", "content": "hello", "tags": []string{"t1"}}
	mb, _ := json.Marshal(memBody)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/agents/me/mcp-memories", bytes.NewReader(mb))
	req.Header.Set("Authorization", "Bearer key-alice")
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("mcp-memories: %d %s", res.StatusCode, string(b))
	}
	var memOut struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	if err := json.Unmarshal(b, &memOut); err != nil {
		t.Fatal(err)
	}
	if memOut.ID == "" || memOut.Key != "note1" {
		t.Fatalf("unexpected memory response: %s", string(b))
	}

	nb, _ := json.Marshal(map[string]any{
		"agent_names": []string{"bob"},
		"message":     "hi from mcp bridge test",
	})
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/agents/me/notify", bytes.NewReader(nb))
	req2.Header.Set("Authorization", "Bearer key-alice")
	req2.Header.Set("Content-Type", "application/json")
	res2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	b2, _ := io.ReadAll(res2.Body)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("notify: %d %s", res2.StatusCode, string(b2))
	}
	var nrows int64
	if err := gdb.Model(&dbpkg.Notification{}).Where("agent_id = ? AND type = ?", bob.ID, "mcp_mention").Count(&nrows).Error; err != nil {
		t.Fatal(err)
	}
	if nrows < 1 {
		t.Fatal("expected at least one notification for bob")
	}
}
