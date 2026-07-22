package opencode

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestErrOpenCodeMissingSentinel(t *testing.T) {
	if ErrOpenCodeMissing == nil {
		t.Fatal("ErrOpenCodeMissing must not be nil")
	}
	if !strings.Contains(ErrOpenCodeMissing.Error(), "opencode") {
		t.Errorf("sentinel message = %q, want it to mention opencode", ErrOpenCodeMissing.Error())
	}

	var err error = ErrOpenCodeMissing
	if !errors.Is(err, ErrOpenCodeMissing) {
		t.Errorf("errors.Is should detect ErrOpenCodeMissing sentinel")
	}
}

func TestRunOpenCodeMissingOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, _, err := RunOpenCode(context.Background(), t.TempDir(), []string{"session", "list"})
	if !errors.Is(err, ErrOpenCodeMissing) {
		t.Fatalf("expected ErrOpenCodeMissing, got %v", err)
	}
}

func TestSessionListMissingOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := SessionList(context.Background(), t.TempDir())
	if !errors.Is(err, ErrOpenCodeMissing) {
		t.Fatalf("expected ErrOpenCodeMissing, got %v", err)
	}
}

func TestSessionExportRejectsEmptyID(t *testing.T) {
	_, err := SessionExport(context.Background(), t.TempDir(), "")
	if !errors.Is(err, ErrInvalidSessionID) {
		t.Fatalf("expected ErrInvalidSessionID, got %v", err)
	}
}
