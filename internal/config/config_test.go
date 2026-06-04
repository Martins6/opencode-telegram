package config

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestGetLocationNotConfigured(t *testing.T) {
	orig := globalConfig
	defer func() { globalConfig = orig }()

	globalConfig = nil
	_, err := GetLocation()
	if !errors.Is(err, ErrTimezoneNotConfigured) {
		t.Errorf("got %v, want ErrTimezoneNotConfigured", err)
	}

	globalConfig = &Config{Bot: BotConfig{Timezone: ""}}
	_, err = GetLocation()
	if !errors.Is(err, ErrTimezoneNotConfigured) {
		t.Errorf("got %v, want ErrTimezoneNotConfigured", err)
	}
}

func TestGetLocationInvalid(t *testing.T) {
	orig := globalConfig
	defer func() { globalConfig = orig }()

	globalConfig = &Config{Bot: BotConfig{Timezone: "Not/A/Real/Zone"}}
	_, err := GetLocation()
	if err == nil {
		t.Fatal("expected error for invalid timezone, got nil")
	}
	if errors.Is(err, ErrTimezoneNotConfigured) {
		t.Errorf("got %v, want a LoadLocation error (not ErrTimezoneNotConfigured)", err)
	}
}

func TestGetLocationValid(t *testing.T) {
	orig := globalConfig
	defer func() { globalConfig = orig }()

	globalConfig = &Config{Bot: BotConfig{Timezone: "America/Sao_Paulo"}}
	loc, err := GetLocation()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.String() != "America/Sao_Paulo" {
		t.Errorf("got %q, want America/Sao_Paulo", loc.String())
	}
	now := time.Now().In(loc)
	if now.Location().String() != "America/Sao_Paulo" {
		t.Errorf("time.Now().In(loc).Location() = %q, want America/Sao_Paulo", now.Location().String())
	}
}

func TestLoadConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(cfgPath, []byte(`
[bot]
token = "t"
allowed_user_id = "u"
timezone = "America/Sao_Paulo"

[defaults]
agent = "a"
model = "m"
provider = "p"

[workspace]
path = "/tmp/test"
`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if _, err := Load(cfgPath); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("Load failed under concurrency: %v", err)
	}

	cfg := Get()
	if cfg == nil {
		t.Fatal("Get() returned nil after concurrent Loads")
	}
	if cfg.Bot.Timezone != "America/Sao_Paulo" {
		t.Errorf("after concurrent Loads: timezone = %q, want America/Sao_Paulo", cfg.Bot.Timezone)
	}
}
