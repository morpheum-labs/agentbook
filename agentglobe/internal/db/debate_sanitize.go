package db

import (
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

// Rune limits applied at persistence (GORM BeforeSave) so debate content is safe even if APIs forget to validate.
const (
	DebateMaxTitleRunes            = 500
	DebateMaxPostRunes            = 32000
	DebateMaxThreadBodyRunes      = 32000
	DebateMaxReportDetailRunes    = 8000
	DebateMaxModerationNotesRunes = 2000
	DebateMaxReasonPublicRunes    = 2000
	DebateMaxReasonCodeRunes      = 64
	DebateMaxImposedByRunes       = 128
)

var debateStrictHTML = bluemonday.StrictPolicy()

// SanitizeDebatePlain strips HTML/scripts, removes NUL and most C0 controls (keeps \n \r \t),
// then truncates by rune count. Use for any user-authored debate text stored in the DB.
func SanitizeDebatePlain(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = debateStrictHTML.Sanitize(s)
	var b strings.Builder
	n := 0
	for _, r := range s {
		if r == 0 {
			continue
		}
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			continue
		}
		if maxRunes > 0 && n >= maxRunes {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

// SanitizeDebateToken keeps lowercase [a-z0-9_-] for stable identifiers (reason_code, action, categories).
func SanitizeDebateToken(s string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = DebateMaxReasonCodeRunes
	}
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			if b.Len() >= maxRunes {
				break
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeDebateEnum(s string, allowed []string, fallback string) string {
	s = strings.ToLower(strings.TrimSpace(SanitizeDebatePlain(s, 64)))
	for _, a := range allowed {
		if s == a {
			return a
		}
	}
	return fallback
}
