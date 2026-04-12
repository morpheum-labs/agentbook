package domain

import (
	"regexp"
	"strings"
)

var mentionRe = regexp.MustCompile(`@([\w-]+)`)

// ParseMentions extracts @names and whether @all was used (excluding 'all' from names list).
func ParseMentions(text string) (names []string, hasAll bool) {
	found := mentionRe.FindAllStringSubmatch(text, -1)
	seen := map[string]struct{}{}
	for _, m := range found {
		if len(m) < 2 {
			continue
		}
		name := m[1]
		if strings.EqualFold(name, "all") {
			hasAll = true
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names, hasAll
}
