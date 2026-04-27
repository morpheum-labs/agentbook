//go:build live

package newsapi

import (
	"context"
	"net/url"
	"os"
	"testing"
)

// Live tests call https://newsapi.org and require API_KEY in the environment.
// Run: API_KEY=yourkey go test -tags=live -v

func TestLiveV2TopHeadlines(t *testing.T) {
	key := os.Getenv("API_KEY")
	if key == "" {
		t.Skip("set API_KEY to run live tests")
	}
	c, err := New(key)
	if err != nil {
		t.Fatal(err)
	}
	u := url.Values{"country": []string{"us"}, "pageSize": []string{"1"}}
	res, _, err := c.V2.TopHeadlines(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "ok" {
		t.Fatalf("status: %q", res.Status)
	}
}
