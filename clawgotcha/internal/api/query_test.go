package api

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseRevisionQuery(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v1/agents?since_revision=5", nil)
	since, after, delta, err := parseRevisionQuery(r)
	if err != nil {
		t.Fatal(err)
	}
	if since != 5 || after != nil || !delta {
		t.Fatalf("got since=%d after=%v delta=%v", since, after, delta)
	}

	r2 := httptest.NewRequest("GET", "/api/v1/agents?updated_after=2024-01-02T15:04:05Z", nil)
	since2, after2, delta2, err := parseRevisionQuery(r2)
	if err != nil {
		t.Fatal(err)
	}
	if since2 != 0 || after2 == nil || !delta2 {
		t.Fatal("expected updated_after delta")
	}
	if !after2.Equal(time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)) {
		t.Fatalf("time: %v", after2)
	}

	r3 := httptest.NewRequest("GET", "/api/v1/agents", nil)
	since3, after3, delta3, err := parseRevisionQuery(r3)
	if err != nil {
		t.Fatal(err)
	}
	if since3 != 0 || after3 != nil || delta3 {
		t.Fatal("expected no delta")
	}
}
