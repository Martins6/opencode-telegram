package daemon

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newValidWorkspace(t *testing.T) string {
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
	for _, f := range []string{"opencode.json", "AGENTS.md", "MAIN-PROMPTS/SOUL.md"} {
		if err := os.WriteFile(filepath.Join(root, f), nil, 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}
	return root
}

type fakeExec struct{ lookPathErr error }

func (f fakeExec) LookPath(file string) (string, error) {
	if f.lookPathErr != nil {
		return "", f.lookPathErr
	}
	return "/usr/bin/" + file, nil
}

func TestRuntimeRunRejectsInvalidWorkspace(t *testing.T) {
	_, err := New(Options{Workspace: "/definitely/does/not/exist/abc/xyz"})
	if err != nil {
		t.Fatalf("New should accept any path string; got %v", err)
	}

	rt, err := New(Options{Workspace: "/definitely/does/not/exist/abc/xyz"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rt.exec = fakeExec{}
	err = rt.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid workspace")
	}
	if !errors.Is(err, err) {
		return
	}
}

func TestRuntimeRunValidationFails(t *testing.T) {
	rt, err := New(Options{Workspace: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rt.exec = fakeExec{}
	if err := rt.Run(context.Background()); err == nil {
		t.Fatal("expected validation error for empty workspace")
	}
}

func TestRuntimeCancels(t *testing.T) {
	root := newValidWorkspace(t)
	t.Setenv("HOME", root)

	rt, err := New(Options{Workspace: root})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rt.exec = fakeExec{}
	rt.binaryGet = func() (string, error) { return "/test/acolyte", nil }

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("runtime did not return after cancel")
	}
}

func TestAbsoluteWorkspace(t *testing.T) {
	tmp := t.TempDir()
	rel, err := filepath.Rel(tmp, tmp)
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(rel)
	if err != nil {
		t.Fatal(err)
	}
	rt, err := New(Options{Workspace: rel})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if rt.opts.Workspace != abs {
		t.Errorf("opts.Workspace = %q, want %q", rt.opts.Workspace, abs)
	}
}
