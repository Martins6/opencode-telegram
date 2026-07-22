package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupValidWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dirs := []string{
		"downloads/images",
		"downloads/audio",
		"downloads/documents",
		"downloads/videos",
		"conversations",
		".logs",
		".opencode/agents",
		".opencode/skills",
		"MAIN-PROMPTS",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	files := []string{
		"opencode.json",
		"AGENTS.md",
		"MAIN-PROMPTS/SOUL.md",
		"MAIN-PROMPTS/USER.md",
		"MAIN-PROMPTS/IDENTITY.md",
		"MAIN-PROMPTS/BOOTSTRAP.md",
		"MAIN-PROMPTS/TOOLS.md",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(root, f), []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}
	return root
}

func TestStrictValidateOk(t *testing.T) {
	root := setupValidWorkspace(t)
	if err := StrictValidate(root); err != nil {
		t.Fatalf("StrictValidate returned error for valid workspace: %v", err)
	}
}

func TestStrictValidateMissingDirectoryAndFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "downloads/images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	err := StrictValidate(root)
	if err == nil {
		t.Fatal("expected error for incomplete workspace")
	}
	msg := err.Error()
	for _, want := range []string{
		"downloads/audio",
		"downloads/documents",
		"downloads/videos",
		"conversations",
		".logs",
		".opencode/agents",
		".opencode/skills",
		"MAIN-PROMPTS",
		"opencode.json",
		"MAIN-PROMPTS/SOUL.md",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message missing %q\n---\n%s\n---", want, msg)
		}
	}
	if !strings.Contains(msg, "acolyte new") {
		t.Errorf("error message missing suggestion 'acolyte new'\n---\n%s\n---", msg)
	}
}

func TestStrictValidateNotDirectory(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := StrictValidate(file); err == nil {
		t.Fatal("expected error for non-directory workspace")
	}
}

func TestStrictValidateNonexistent(t *testing.T) {
	if err := StrictValidate("/nonexistent/path/xyz/abc"); err == nil {
		t.Fatal("expected error for nonexistent workspace")
	}
}
