package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSingletonConfigPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	got, err := SingletonConfigPath()
	if err != nil {
		t.Fatalf("SingletonConfigPath: %v", err)
	}
	want := filepath.Join(tmp, ".acolyte", "config.toml")
	if got != want {
		t.Errorf("SingletonConfigPath = %q, want %q", got, want)
	}
}

func TestLoadIfExistsMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	missing := filepath.Join(tmp, "nope.toml")

	_, err := LoadIfExists(missing)
	if !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("LoadIfExists missing: err = %v, want ErrConfigNotFound", err)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Errorf("LoadIfExists should not create the singleton file when missing; stat err = %v", err)
	}
}

func TestLoadIfExistsExisting(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	dir := filepath.Join(tmp, ".acolyte")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "config.toml")
	contents := `[bot]
token = "t"
allowed_user_id = "u"
timezone = "UTC"

[defaults]
agent = "a"
model = "m"
provider = "p"

[workspace]
path = "/tmp/test"
`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := LoadIfExists(path)
	if err != nil {
		t.Fatalf("LoadIfExists: %v", err)
	}
	if cfg.Bot.Token != "t" || cfg.Workspace.Path != "/tmp/test" {
		t.Errorf("unexpected cfg: %+v", cfg)
	}
}

func TestWriteWorkspacePathRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	desired := filepath.Join(tmp, "my-ws")
	if err := WriteWorkspacePath(desired); err != nil {
		t.Fatalf("WriteWorkspacePath: %v", err)
	}

	path, err := SingletonConfigPath()
	if err != nil {
		t.Fatalf("SingletonConfigPath: %v", err)
	}

	cfg, err := LoadIfExists(path)
	if err != nil {
		t.Fatalf("LoadIfExists: %v", err)
	}
	abs, _ := filepath.Abs(desired)
	if cfg.Workspace.Path != abs {
		t.Errorf("workspace.path = %q, want %q", cfg.Workspace.Path, abs)
	}
}

func TestWriteWorkspacePathPreservesExistingKeys(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path, err := SingletonConfigPath()
	if err != nil {
		t.Fatalf("SingletonConfigPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	contents := `[bot]
token = "secret"
allowed_user_id = "u"
timezone = "UTC"

[defaults]
agent = "myagent"
model = "m"
provider = "p"

[workspace]
path = "/tmp/orig"
`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := WriteWorkspacePath("/tmp/new"); err != nil {
		t.Fatalf("WriteWorkspacePath: %v", err)
	}

	cfg, err := LoadIfExists(path)
	if err != nil {
		t.Fatalf("LoadIfExists: %v", err)
	}
	if cfg.Bot.Token != "secret" {
		t.Errorf("token lost: %q", cfg.Bot.Token)
	}
	if cfg.Defaults.Agent != "myagent" {
		t.Errorf("agent lost: %q", cfg.Defaults.Agent)
	}
	abs, _ := filepath.Abs("/tmp/new")
	if cfg.Workspace.Path != abs {
		t.Errorf("workspace.path = %q, want %q", cfg.Workspace.Path, abs)
	}
}
