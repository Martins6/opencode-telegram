package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/martins6/acolyte/internal/service"
	"github.com/spf13/cobra"
)

type stubManager struct {
	service.Manager
	installCalled   bool
	enableCalled    bool
	startCalled     bool
	stopCalled      bool
	restartCalled   bool
	statusReturn    service.ServiceStatus
	waitReadyReturn service.ServiceStatus
	waitReadyErr    error
}

func (s *stubManager) Install(ctx context.Context, cfg service.ServiceConfig) error {
	s.installCalled = true
	return nil
}
func (s *stubManager) Enable(ctx context.Context) error  { s.enableCalled = true; return nil }
func (s *stubManager) Disable(ctx context.Context) error { return nil }
func (s *stubManager) Start(ctx context.Context) error   { s.startCalled = true; return nil }
func (s *stubManager) Stop(ctx context.Context) error    { s.stopCalled = true; return nil }
func (s *stubManager) Restart(ctx context.Context) error {
	s.restartCalled = true
	return nil
}
func (s *stubManager) Status(ctx context.Context) (service.ServiceStatus, error) {
	return s.statusReturn, nil
}
func (s *stubManager) WaitReady(ctx context.Context, t time.Duration) (service.ServiceStatus, error) {
	return s.waitReadyReturn, s.waitReadyErr
}

func findCmd(name string) *cobra.Command {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestStatusCmdFormat(t *testing.T) {
	t.Setenv("ACOLYTE_TEST_GOOS", "linux")
	stub := &stubManager{
		statusReturn: service.ServiceStatus{
			Loaded:    true,
			Enabled:   true,
			Autostart: true,
			PID:       12345,
			Workspace: "/tmp/ws",
		},
	}
	original := serviceFactory
	serviceFactory = func() (service.Manager, error) { return stub, nil }
	defer func() { serviceFactory = original }()

	cmd := findCmd("status")
	if cmd == nil {
		t.Fatal("status command not registered")
	}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}
	output := buf.String()
	for _, want := range []string{"Status:    running", "Workspace: /tmp/ws", "Autostart: enabled", "PID:       12345"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\n---\n%s\n---", want, output)
		}
	}
}
