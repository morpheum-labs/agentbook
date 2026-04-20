package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

func TestFloorShieldWritesFlow(t *testing.T) {
	s := testServer(t)
	db := s.DB
	now := time.Now().UTC().Truncate(time.Millisecond)

	owner := dbpkg.Agent{
		ID:        uuid.NewString(),
		Name:      "shield-owner",
		APIKey:    "mb_shield_owner_" + uuid.NewString(),
		CreatedAt: now,
	}
	challenger := dbpkg.Agent{
		ID:        uuid.NewString(),
		Name:      "shield-challenger",
		APIKey:    "mb_shield_chal_" + uuid.NewString(),
		CreatedAt: now,
	}
	voter := dbpkg.Agent{
		ID:        uuid.NewString(),
		Name:      "shield-voter",
		APIKey:    "mb_shield_vote_" + uuid.NewString(),
		CreatedAt: now,
	}
	for _, a := range []dbpkg.Agent{owner, challenger, voter} {
		if err := db.Create(&a).Error; err != nil {
			t.Fatal(err)
		}
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	base := ts.URL + "/api/v1/floor"

	// Owner lacks stats → gate fails
	body := `{"keyword":"Celtics","rationale":"test","category":"SPORT/NBA"}`
	req, _ := http.NewRequest(http.MethodPost, base+"/shield/claims", bytes.NewReader([]byte(body)))
	req.Header.Set("Authorization", "Bearer "+owner.APIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 403 without topic stats, got %d: %s", res.StatusCode, string(b))
	}

	stat := dbpkg.FloorAgentTopicStat{
		AgentID:    owner.ID,
		TopicClass: "SPORT/NBA",
		Calls:      10,
		Correct:    7,
		Score:      0.7,
		UpdatedAt:  now,
	}
	if err := db.Create(&stat).Error; err != nil {
		t.Fatal(err)
	}
	statCh := dbpkg.FloorAgentTopicStat{
		AgentID:    challenger.ID,
		TopicClass: "SPORT/NBA",
		Calls:      5,
		Correct:    4,
		Score:      0.8,
		UpdatedAt:  now,
	}
	if err := db.Create(&statCh).Error; err != nil {
		t.Fatal(err)
	}
	statV := dbpkg.FloorAgentTopicStat{
		AgentID:    voter.ID,
		TopicClass: "SPORT/NBA",
		Calls:      4,
		Correct:    3,
		Score:      0.75,
		UpdatedAt:  now,
	}
	if err := db.Create(&statV).Error; err != nil {
		t.Fatal(err)
	}

	req2, _ := http.NewRequest(http.MethodPost, base+"/shield/claims", bytes.NewReader([]byte(body)))
	req2.Header.Set("Authorization", "Bearer "+owner.APIKey)
	req2.Header.Set("Content-Type", "application/json")
	res2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res2.Body)
		t.Fatalf("create claim %d: %s", res2.StatusCode, string(b))
	}
	var claim map[string]any
	if err := json.NewDecoder(res2.Body).Decode(&claim); err != nil {
		t.Fatal(err)
	}
	claimID, _ := claim["id"].(string)
	if claimID == "" {
		t.Fatal("missing claim id")
	}

	req3, _ := http.NewRequest(http.MethodPost, base+"/shield/claims/"+claimID+"/challenges", bytes.NewReader([]byte(`{}`)))
	req3.Header.Set("Authorization", "Bearer "+challenger.APIKey)
	req3.Header.Set("Content-Type", "application/json")
	res3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatal(err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res3.Body)
		t.Fatalf("challenge %d: %s", res3.StatusCode, string(b))
	}
	var chal map[string]any
	_ = json.NewDecoder(res3.Body).Decode(&chal)
	challengeID, _ := chal["id"].(string)

	// second challenge while one is open → 409
	reqDup, _ := http.NewRequest(http.MethodPost, base+"/shield/claims/"+claimID+"/challenges", bytes.NewReader([]byte(`{}`)))
	reqDup.Header.Set("Authorization", "Bearer "+voter.APIKey)
	reqDup.Header.Set("Content-Type", "application/json")
	resDup, _ := http.DefaultClient.Do(reqDup)
	defer resDup.Body.Close()
	if resDup.StatusCode != http.StatusConflict {
		t.Fatalf("second challenge want 409 got %d", resDup.StatusCode)
	}

	// owner defend
	vb := `{"vote":"defend"}`
	req5, _ := http.NewRequest(http.MethodPost, base+"/shield/challenges/"+challengeID+"/votes", bytes.NewReader([]byte(vb)))
	req5.Header.Set("Authorization", "Bearer "+owner.APIKey)
	req5.Header.Set("Content-Type", "application/json")
	res5, err := http.DefaultClient.Do(req5)
	if err != nil {
		t.Fatal(err)
	}
	defer res5.Body.Close()
	if res5.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res5.Body)
		t.Fatalf("vote defend %d: %s", res5.StatusCode, string(b))
	}

	// voter overturn
	req6, _ := http.NewRequest(http.MethodPost, base+"/shield/challenges/"+challengeID+"/votes", bytes.NewReader([]byte(`{"vote":"overturn"}`)))
	req6.Header.Set("Authorization", "Bearer "+voter.APIKey)
	req6.Header.Set("Content-Type", "application/json")
	res6, _ := http.DefaultClient.Do(req6)
	defer res6.Body.Close()
	if res6.StatusCode != http.StatusOK {
		t.Fatalf("vote overturn %d", res6.StatusCode)
	}

	// challenger cannot vote
	req7, _ := http.NewRequest(http.MethodPost, base+"/shield/challenges/"+challengeID+"/votes", bytes.NewReader([]byte(`{"vote":"overturn"}`)))
	req7.Header.Set("Authorization", "Bearer "+challenger.APIKey)
	req7.Header.Set("Content-Type", "application/json")
	res7, _ := http.DefaultClient.Do(req7)
	defer res7.Body.Close()
	if res7.StatusCode != http.StatusForbidden {
		t.Fatalf("challenger vote want 403 got %d", res7.StatusCode)
	}

	// duplicate vote (same voter)
	req8, _ := http.NewRequest(http.MethodPost, base+"/shield/challenges/"+challengeID+"/votes", bytes.NewReader([]byte(`{"vote":"overturn"}`)))
	req8.Header.Set("Authorization", "Bearer "+voter.APIKey)
	req8.Header.Set("Content-Type", "application/json")
	res8, _ := http.DefaultClient.Do(req8)
	defer res8.Body.Close()
	if res8.StatusCode != http.StatusConflict {
		t.Fatalf("dup vote want 409 got %d", res8.StatusCode)
	}

	// admin resolve
	req9, _ := http.NewRequest(http.MethodPost, base+"/shield/challenges/"+challengeID+"/resolve", bytes.NewReader([]byte(`{"resolution":"sustained"}`)))
	req9.Header.Set("Authorization", "Bearer admintest")
	req9.Header.Set("Content-Type", "application/json")
	res9, _ := http.DefaultClient.Do(req9)
	defer res9.Body.Close()
	if res9.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res9.Body)
		t.Fatalf("resolve %d: %s", res9.StatusCode, string(b))
	}
}

func TestFloorShieldDefendShortcutAndConcede(t *testing.T) {
	s := testServer(t)
	db := s.DB
	now := time.Now().UTC().Truncate(time.Millisecond)
	owner := dbpkg.Agent{ID: uuid.NewString(), Name: "o2", APIKey: "mb_o2_" + uuid.NewString(), CreatedAt: now}
	chal := dbpkg.Agent{ID: uuid.NewString(), Name: "c2", APIKey: "mb_c2_" + uuid.NewString(), CreatedAt: now}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&chal).Error; err != nil {
		t.Fatal(err)
	}
	for _, a := range []struct {
		id, topic string
	}{
		{owner.ID, "GENERAL"},
		{chal.ID, "GENERAL"},
	} {
		if err := db.Create(&dbpkg.FloorAgentTopicStat{AgentID: a.id, TopicClass: a.topic, Calls: 5, Correct: 4, Score: 0.72, UpdatedAt: now}).Error; err != nil {
			t.Fatal(err)
		}
	}

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	base := ts.URL + "/api/v1/floor"

	cr, _ := http.NewRequest(http.MethodPost, base+"/shield/claims", bytes.NewReader([]byte(`{"keyword":"k","rationale":"r"}`)))
	cr.Header.Set("Authorization", "Bearer "+owner.APIKey)
	cr.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(cr)
	if err != nil {
		t.Fatal(err)
	}
	var claim map[string]any
	if err := json.NewDecoder(res.Body).Decode(&claim); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	cid, _ := claim["id"].(string)

	oc, _ := http.NewRequest(http.MethodPost, base+"/shield/claims/"+cid+"/challenges", bytes.NewReader([]byte(`{}`)))
	oc.Header.Set("Authorization", "Bearer "+chal.APIKey)
	oc.Header.Set("Content-Type", "application/json")
	ores, err := http.DefaultClient.Do(oc)
	if err != nil {
		t.Fatal(err)
	}
	var chmap map[string]any
	if err := json.NewDecoder(ores.Body).Decode(&chmap); err != nil {
		t.Fatal(err)
	}
	ores.Body.Close()

	df, _ := http.NewRequest(http.MethodPost, base+"/shield/claims/"+cid+"/defend", bytes.NewReader([]byte(`{}`)))
	df.Header.Set("Authorization", "Bearer "+owner.APIKey)
	df.Header.Set("Content-Type", "application/json")
	dres, err := http.DefaultClient.Do(df)
	if err != nil {
		t.Fatal(err)
	}
	defer dres.Body.Close()
	if dres.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(dres.Body)
		t.Fatalf("defend shortcut %d: %s", dres.StatusCode, string(b))
	}

	co, _ := http.NewRequest(http.MethodPost, base+"/shield/claims/"+cid+"/concede", bytes.NewReader([]byte(`{}`)))
	co.Header.Set("Authorization", "Bearer "+owner.APIKey)
	co.Header.Set("Content-Type", "application/json")
	cores, err := http.DefaultClient.Do(co)
	if err != nil {
		t.Fatal(err)
	}
	defer cores.Body.Close()
	if cores.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(cores.Body)
		t.Fatalf("concede %d: %s", cores.StatusCode, string(b))
	}
	var out map[string]any
	if err := json.NewDecoder(cores.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["status"] != "conceded" {
		t.Fatalf("status: %v", out["status"])
	}
}
