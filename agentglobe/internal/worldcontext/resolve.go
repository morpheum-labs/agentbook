// Package worldcontext resolves a local worldmon service base and applies RSS library query defaults
// for agentglobe's public world-context proxy and related callers.
package worldcontext

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

// ResolveServiceBase returns the worldmon HTTP base URL: WORLDMON_BASE_URL, or a capability_services
// world_monitor base_url, following the same rules as the previous MCP get_world_context resolver.
func ResolveServiceBase(gdb *gorm.DB) (string, error) {
	if v := strings.TrimSpace(os.Getenv("WORLDMON_BASE_URL")); v != "" {
		return strings.TrimRight(v, "/"), nil
	}
	if gdb == nil {
		return "", fmt.Errorf("database not available: set WORLDMON_BASE_URL (or use a process with a DB to resolve world_monitor from the capability registry)")
	}
	var rows []db.CapabilityService
	_ = gdb.Where("LOWER(category) = LOWER(?)", "world_monitor").Order("name ASC, base_url ASC").Find(&rows).Error
	for i := range rows {
		if rows[i].IsHealthy(2 * time.Minute) {
			return strings.TrimRight(strings.TrimSpace(rows[i].BaseURL), "/"), nil
		}
	}
	for i := range rows {
		if bu := strings.TrimSpace(rows[i].BaseURL); bu != "" {
			return strings.TrimRight(bu, "/"), nil
		}
	}
	return "", fmt.Errorf("set WORLDMON_BASE_URL (or register a world_monitor service in the capability registry)")
}

// ApplyRssLibQuery sets list-feed-digest style defaults from config rss_lib: rss_library (file path) or
// rss_library_url (https). Later query merges should be applied by the caller to allow overrides.
func ApplyRssLibQuery(q url.Values, resolvedRss string) {
	rss := strings.TrimSpace(resolvedRss)
	if rss == "" {
		return
	}
	low := strings.ToLower(rss)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") {
		q.Set("rss_library_url", rss)
		return
	}
	q.Set("rss_library", rss)
}
