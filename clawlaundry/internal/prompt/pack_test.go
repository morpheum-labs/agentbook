package prompt

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestComposeDecomposeRoundTrip(t *testing.T) {
	d := t.TempDir()
	identity := "I am a helpful bot.\n\n- List item"
	soul := "Be concise."
	user := "User region."
	for _, c := range []struct {
		filename, content string
	}{
		{FilenameIdentity, identity},
		{FilenameSoul, soul},
		{FilenameUser, user},
	} {
		err := os.WriteFile(filepath.Join(d, c.filename), []byte(c.content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}
	combined, err := Compose(d)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	out := filepath.Join(t.TempDir(), "w")
	if err := Decompose(combined, out); err != nil {
		t.Fatalf("decompose: %v", err)
	}
	bI, _ := os.ReadFile(filepath.Join(out, FilenameIdentity))
	bS, _ := os.ReadFile(filepath.Join(out, FilenameSoul))
	bU, _ := os.ReadFile(filepath.Join(out, FilenameUser))
	// Composed from TrimSpace of files; decomposed sections are trim(join(lines)).
	// Re-read round-trip: identity/soul/user match trimmed originals.
	if string(bI) != identity {
		t.Errorf("identity got %q want %q", bI, identity)
	}
	if string(bS) != soul {
		t.Errorf("soul got %q want %q", bS, soul)
	}
	if string(bU) != user {
		t.Errorf("user got %q want %q", bU, user)
	}
}

func TestParseSectionErrors(t *testing.T) {
	_, err := ParseSections("not modular")
	if !errors.Is(err, ErrNotModular) {
		t.Fatalf("expected ErrNotModular, got %v", err)
	}
	_, err = ParseSections("=== IDENTITY ===\nhi\n=== SOUL ===\n")
	if !errors.Is(err, ErrNotModular) {
		t.Fatalf("incomplete, expected ErrNotModular, got %v", err)
	}
}

func TestEmptySections(t *testing.T) {
	p := Pack{}
	combined := p.CombinedString()
	got, err := ParseSections(combined)
	if err != nil {
		t.Fatal(err)
	}
	if got.Identity != "" || got.Soul != "" || got.User != "" {
		t.Errorf("all empty, got %#v", got)
	}
}

func TestReadPackStrict_Missing(t *testing.T) {
	_, err := ReadPackStrict(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing files")
	}
}
