package service

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

type fakeRunner struct {
	mu     sync.Mutex
	calls  []fakeCall
	lookOK map[string]bool
}

type fakeCall struct {
	Name string
	Args []string
	Out  []byte
	Err  error
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.calls {
		if c.Name != name {
			continue
		}
		if len(c.Args) != len(args) {
			continue
		}
		match := true
		for i, a := range args {
			if c.Args[i] != a {
				match = false
				break
			}
		}
		if match {
			return c.Out, nil, c.Err
		}
	}
	return nil, nil, errors.New("fakeRunner: no matching call registered")
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	if f.lookOK == nil {
		return "/usr/bin/" + file, nil
	}
	if ok, exists := f.lookOK[file]; exists && ok {
		return "/usr/bin/" + file, nil
	}
	return "", errors.New("fakeRunner: LookPath failed")
}

func TestNewPicksLinux(t *testing.T) {
	t.Setenv("ACOLYTE_TEST_GOOS", "linux")
	r := &fakeRunner{lookOK: map[string]bool{"systemctl": true}}
	m, err := New(r)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := m.(*systemdManager); !ok {
		t.Fatalf("expected systemdManager, got %T", m)
	}
}

func TestNewPicksDarwin(t *testing.T) {
	t.Setenv("ACOLYTE_TEST_GOOS", "darwin")
	r := &fakeRunner{lookOK: map[string]bool{"launchctl": true}}
	m, err := New(r)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := m.(*launchdManager); !ok {
		t.Fatalf("expected launchdManager, got %T", m)
	}
}

func TestNewErrorsOnOther(t *testing.T) {
	t.Setenv("ACOLYTE_TEST_GOOS", "windows")
	_, err := New(nil)
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

func TestSystemdRenderUnitHandlesSpacesInWorkspace(t *testing.T) {
	t.Setenv("ACOLYTE_TEST_GOOS", "linux")
	tmp := t.TempDir()
	ws := filepath.Join(tmp, "my work")
	if err := mkdir(ws); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(tmp, "acolyte")
	if err := writeFile(bin, "#!/bin/sh\n"); err != nil {
		t.Fatal(err)
	}
	unit, err := renderSystemdUnit(ServiceConfig{Workspace: ws, Binary: bin})
	if err != nil {
		t.Fatalf("renderSystemdUnit: %v", err)
	}
	if !contains(unit, `ExecStart=`) {
		t.Errorf("ExecStart missing")
	}
	if !contains(unit, `WorkingDirectory=`) {
		t.Errorf("WorkingDirectory missing")
	}
}

func TestSystemdTemplateQuoting(t *testing.T) {
	tpl, err := renderTemplate(systemdUnitTemplate, map[string]string{
		"Binary":    "/tmp/acolyte",
		"Workspace": "/tmp/has space",
		"Timeout":   "30s",
		"PathEnv":   "/usr/bin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(tpl, "/tmp/acolyte") {
		t.Errorf("missing binary")
	}
	if !contains(tpl, "/tmp/has space") {
		t.Errorf("missing workspace with space")
	}
}

func TestPlistRender(t *testing.T) {
	tmp := t.TempDir()
	ws := filepath.Join(tmp, "ws")
	if err := mkdir(ws); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(tmp, "acolyte")
	if err := writeFile(bin, "x"); err != nil {
		t.Fatal(err)
	}
	text, err := renderPlist(ServiceConfig{Workspace: ws, Binary: bin})
	if err != nil {
		t.Fatalf("renderPlist: %v", err)
	}
	if !contains(text, "com.martins6.acolyte") {
		t.Errorf("missing label")
	}
	if !contains(text, "__daemon") {
		t.Errorf("missing __daemon arg")
	}
	if !contains(text, "KeepAlive") {
		t.Errorf("missing KeepAlive block")
	}
}
