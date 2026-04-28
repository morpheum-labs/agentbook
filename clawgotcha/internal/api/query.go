package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseRevisionQuery(r *http.Request) (sinceRev int64, updatedAfter *time.Time, delta bool, err error) {
	q := r.URL.Query()
	if v := strings.TrimSpace(q.Get("since_revision")); v != "" {
		sinceRev, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, nil, false, fmt.Errorf("since_revision: must be an integer")
		}
		delta = true
	}
	if v := strings.TrimSpace(q.Get("updated_after")); v != "" {
		t, e := time.Parse(time.RFC3339Nano, v)
		if e != nil {
			t, e = time.Parse(time.RFC3339, v)
			if e != nil {
				return 0, nil, false, fmt.Errorf("updated_after: invalid RFC3339 timestamp")
			}
		}
		updatedAfter = &t
		delta = true
	}
	return sinceRev, updatedAfter, delta, nil
}
