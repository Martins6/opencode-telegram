package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

var (
	ErrNoState = errors.New("runtime state not found")
	ErrStale   = errors.New("runtime state belongs to a process that is no longer running")
)

type State struct {
	PID        int       `json:"pid"`
	Workspace  string    `json:"workspace"`
	StartedAt  time.Time `json:"started_at"`
	Ready      bool      `json:"ready"`
	Stopping   bool      `json:"stopping"`
	BinaryPath string    `json:"binary_path,omitempty"`
}

func StateFile(dir string) string {
	return filepath.Join(dir, "state.json")
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acolyte", ".service"), nil
}

func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if err := syscall.Kill(pid, syscall.Signal(0)); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return false
		}
		if errno, ok := err.(syscall.Errno); ok && errno == syscall.EPERM {
			return true
		}
		return false
	}
	return true
}

func lockPath(target string) string {
	dir := filepath.Dir(target)
	base := filepath.Base(target)
	return filepath.Join(dir, "."+base+".lock")
}

func writeAtomic(target string, data []byte) error {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure state dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(target)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp state: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp state: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp state: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("rename temp state: %w", err)
	}
	cleanup = false
	return nil
}

func encode(s State) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func Write(dir string, s State) error {
	data, err := encode(s)
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	return writeAtomic(StateFile(dir), data)
}

func Read(dir string) (State, error) {
	var s State
	data, err := os.ReadFile(StateFile(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, ErrNoState
		}
		return s, fmt.Errorf("read state: %w", err)
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse state: %w", err)
	}
	return s, nil
}

func ReadLive(dir string) (State, error) {
	s, err := Read(dir)
	if err != nil {
		return s, err
	}
	if !IsProcessAlive(s.PID) {
		return s, ErrStale
	}
	return s, nil
}

func updateField(dir, field string, value any) error {
	s, err := Read(dir)
	if err != nil {
		return err
	}
	switch field {
	case "ready":
		s.Ready = value.(bool)
	case "stopping":
		s.Stopping = value.(bool)
	}
	return Write(dir, s)
}

func TouchReady(dir string) error {
	return updateField(dir, "ready", true)
}

func MarkStopping(dir string) error {
	return updateField(dir, "stopping", true)
}

func Remove(dir string) error {
	err := os.Remove(StateFile(dir))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func StringPID(pid int) string { return strconv.Itoa(pid) }
