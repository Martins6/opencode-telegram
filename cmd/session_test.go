package cmd

import (
	"strings"
	"testing"

	"github.com/martins6/acolyte/internal/config"
	"github.com/spf13/cobra"
)

func TestSessionCommandRegistered(t *testing.T) {
	var session *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "session" {
			session = c
			break
		}
	}
	if session == nil {
		t.Fatal("expected rootCmd to register a 'session' subcommand")
	}
	if session.Short != "Inspect OpenCode sessions" {
		t.Errorf("session.Short = %q, want %q", session.Short, "Inspect OpenCode sessions")
	}
}

func TestSessionSubcommands(t *testing.T) {
	var session *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "session" {
			session = c
			break
		}
	}
	if session == nil {
		t.Fatal("session subcommand not registered")
	}

	var hasList, hasExport bool
	for _, sub := range session.Commands() {
		switch sub.Name() {
		case "list":
			hasList = true
			if !strings.Contains(sub.Short, "List") {
				t.Errorf("session list Short = %q, want it to mention List", sub.Short)
			}
		case "export":
			hasExport = true
			if !strings.Contains(sub.Use, "<sessionID>") {
				t.Errorf("session export Use = %q, want it to mention <sessionID>", sub.Use)
			}
		}
	}

	if !hasList {
		t.Error("session subcommand is missing 'list' child")
	}
	if !hasExport {
		t.Error("session subcommand is missing 'export' child")
	}
}

func TestGlobalAccessorReturnsConfigGet(t *testing.T) {
	if Global() != config.Get() {
		t.Errorf("Global() should return the same pointer as config.Get()")
	}
}
