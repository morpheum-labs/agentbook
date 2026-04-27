// Package prompt composes and decomposes MiroClaw modular prompt files
// (IDENTITY.md, SOUL.md, USER.md) to and from a single system_prompt with fixed markers.
package prompt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// MarkerIdentity begins the identity block in a combined system prompt.
	MarkerIdentity = "=== IDENTITY ==="
	// MarkerSoul begins the soul block in a combined system prompt.
	MarkerSoul = "=== SOUL ==="
	// MarkerUser begins the user block in a combined system prompt.
	MarkerUser = "=== USER ==="
)

// Pack is the three workspace parts (trimmed from disk when composing).
type Pack struct {
	Identity string
	Soul     string
	User     string
}

// File names in the MiroClaw workspace (relative to workspace root).
const (
	FilenameIdentity = "IDENTITY.md"
	FilenameSoul     = "SOUL.md"
	FilenameUser     = "USER.md"
)

// ErrNotModular is returned when a prompt does not contain the three MiroClaw
// section markers in the required order.
var ErrNotModular = errors.New("system prompt has no MiroClaw section markers (expected " + MarkerIdentity + ", " + MarkerSoul + ", " + MarkerUser + " in that order)")

// Compose reads the three files from a MiroClaw workspace and returns a single
// system_prompt string with [MarkerIdentity] / [MarkerSoul] / [MarkerUser] boundaries.
// Missing file paths return a wrapped error. Empty file contents are allowed.
func Compose(workspaceDir string) (string, error) {
	p, err := ReadPackStrict(workspaceDir)
	if err != nil {
		return "", err
	}
	return p.CombinedString(), nil
}

// ReadPackStrict reads all three files; if any is missing, it returns the first read error.
func ReadPackStrict(workspaceDir string) (Pack, error) {
	var p Pack
	for _, pair := range []struct {
		filename string
		out      *string
	}{
		{FilenameIdentity, &p.Identity},
		{FilenameSoul, &p.Soul},
		{FilenameUser, &p.User},
	} {
		path := filepath.Join(workspaceDir, pair.filename)
		data, err := os.ReadFile(path)
		if err != nil {
			return Pack{}, fmt.Errorf("read %s: %w", pair.filename, err)
		}
		*pair.out = strings.TrimSpace(string(data))
	}
	return p, nil
}

// CombinedString returns the single DB value for this pack (markers + ordering fixed).
func (p *Pack) CombinedString() string {
	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n\n%s\n%s\n",
		MarkerIdentity, p.Identity,
		MarkerSoul, p.Soul,
		MarkerUser, p.User,
	)
}

// ParseSections returns the three sections for a DB system_prompt, or
// an error if required markers or section order is missing.
func ParseSections(fullPrompt string) (Pack, error) {
	fullPrompt = strings.ReplaceAll(fullPrompt, "\r\n", "\n")
	lines := strings.Split(fullPrompt, "\n")

	const (
		phaseNone = iota
		identity
		soul
		user
	)
	phase := phaseNone
	var idLines, soulLines, userLines []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == MarkerIdentity {
			if err := requirePhase(phase, phaseNone, "IDENTITY"); err != nil {
				return Pack{}, err
			}
			phase = identity
			continue
		}
		if t == MarkerSoul {
			if err := requirePhase(phase, identity, "SOUL"); err != nil {
				return Pack{}, err
			}
			phase = soul
			continue
		}
		if t == MarkerUser {
			if err := requirePhase(phase, soul, "USER"); err != nil {
				return Pack{}, err
			}
			phase = user
			continue
		}
		switch phase {
		case identity:
			idLines = append(idLines, line)
		case soul:
			soulLines = append(soulLines, line)
		case user:
			userLines = append(userLines, line)
		}
	}
	if phase != user {
		return Pack{}, ErrNotModular
	}
	return Pack{
		Identity: strings.TrimSpace(strings.Join(idLines, "\n")),
		Soul:     strings.TrimSpace(strings.Join(soulLines, "\n")),
		User:     strings.TrimSpace(strings.Join(userLines, "\n")),
	}, nil
}

func requirePhase(have, want int, name string) error {
	if have != want {
		return fmt.Errorf("unexpected %s marker: section order must be %s, %s, %s",
			name, MarkerIdentity, MarkerSoul, MarkerUser)
	}
	return nil
}

// Decompose writes IDENTITY.md, SOUL.md, and USER.md under workspaceDir
// from a single DB system_prompt. The prompt must contain the three section markers
// in order; otherwise it returns [ErrNotModular] or a parse error.
func Decompose(fullPrompt, workspaceDir string) error {
	pack, err := ParseSections(fullPrompt)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return fmt.Errorf("mkdir workspace: %w", err)
	}
	for _, pair := range []struct {
		filename string
		content  string
	}{
		{FilenameIdentity, pack.Identity},
		{FilenameSoul, pack.Soul},
		{FilenameUser, pack.User},
	} {
		path := filepath.Join(workspaceDir, pair.filename)
		if err := os.WriteFile(path, []byte(pair.content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", pair.filename, err)
		}
	}
	return nil
}
