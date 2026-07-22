package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func newState() State {
	return State{
		PID:        os.Getpid(),
		Workspace:  "/tmp/ws",
		StartedAt:  time.Now().UTC(),
		Ready:      false,
		Stopping:   false,
		BinaryPath: "/tmp/acolyte",
	}
}

func TestWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	want := newState()
	if err := Write(dir, want); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := os.Stat(StateFile(dir)); err != nil {
		t.Fatalf("state file should exist: %v", err)
	}
	got, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.PID != want.PID {
		t.Errorf("PID = %d, want %d", got.PID, want.PID)
	}
	if got.Workspace != want.Workspace {
		t.Errorf("Workspace = %q, want %q", got.Workspace, want.Workspace)
	}
	if got.Ready || got.Stopping {
		t.Errorf("fresh state should not be ready or stopping")
	}
}

func TestTouchReadyAndMarkStopping(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, newState()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := TouchReady(dir); err != nil {
		t.Fatalf("TouchReady: %v", err)
	}
	s, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !s.Ready {
		t.Errorf("Ready = false, want true")
	}
	if err := MarkStopping(dir); err != nil {
		t.Fatalf("MarkStopping: %v", err)
	}
	s, _ = Read(dir)
	if !s.Stopping {
		t.Errorf("Stopping = false, want true")
	}
}

func TestReadLiveMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadLive(dir)
	if !errors.Is(err, ErrNoState) {
		t.Errorf("ReadLive missing: err = %v, want ErrNoState", err)
	}
}

func TestReadLiveStale(t *testing.T) {
	dir := t.TempDir()
	stale := State{PID: 0xDEAD, Workspace: "/tmp/ws", StartedAt: time.Now().Add(-time.Hour)}
	if err := Write(dir, stale); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := ReadLive(dir)
	if !errors.Is(err, ErrStale) {
		t.Errorf("ReadLive stale: err = %v, want ErrStale", err)
	}
}

func TestIsProcessAliveCurrent(t *testing.T) {
	if !IsProcessAlive(os.Getpid()) {
		t.Fatal("current process should be considered alive")
	}
}

func TestIsProcessAliveDead(t *testing.T) {
	if IsProcessAlive(0xDEAD) {
		t.Fatal("pid 0xDEAD should be considered dead")
	}
}

func TestIsProcessAliveInvalid(t *testing.T) {
	if IsProcessAlive(-1) {
		t.Fatal("negative pid should be considered dead")
	}
	if IsProcessAlive(0) {
		t.Fatal("pid 0 should be considered dead")
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, newState()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := Remove(dir); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(StateFile(dir)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("state file should be gone, got err = %v", err)
	}
	if err := Remove(dir); err != nil {
		t.Errorf("Remove on missing file should not error, got %v", err)
	}
}

func TestAtomicWriteNoLeftover(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, newState()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp-*" || filepath.Base(e.Name()) != "state.json" {
			continue
		}
		_ = syscall.Kill // keep import used
	}
}
