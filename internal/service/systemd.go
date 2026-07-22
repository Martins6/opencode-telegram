package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	acolyteruntime "github.com/martins6/acolyte/internal/runtime"
)

var ErrSystemdOperationFailed = errors.New("systemctl operation failed")

type systemdManager struct {
	runner Runner
}

func (s *systemdManager) UnitPath() string {
	return unitFilePath()
}

func (s *systemdManager) PlistPath() string { return "" }

func unitFilePath() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(v, "systemd", "user", UnitLabel+".service")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", UnitLabel+".service")
}

const systemdUnitTemplate = `[Unit]
Description=Acolyte Telegram bot gateway to OpenCode
After=network-online.target
Wants=network-online.target

[Service]
Type=exec
ExecStart={{.Binary}} __daemon
WorkingDirectory={{.Workspace}}
Restart=on-failure
RestartSec=10s
TimeoutStopSec={{.Timeout}}
Environment=PATH={{.PathEnv}}

[Install]
WantedBy=default.target
`

func renderSystemdUnit(cfg ServiceConfig) (string, error) {
	ws, err := absPath(cfg.Workspace)
	if err != nil {
		return "", err
	}
	bin, err := absPath(cfg.Binary)
	if err != nil {
		return "", err
	}
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		pathEnv = "/usr/local/bin:/usr/bin:/bin"
	}
	return renderTemplate(systemdUnitTemplate, map[string]string{
		"Binary":    bin,
		"Workspace": ws,
		"Timeout":   systemdTimeout,
		"PathEnv":   pathEnv,
	})
}

func (s *systemdManager) writeUnit(cfg ServiceConfig) error {
	unit, err := renderSystemdUnit(cfg)
	if err != nil {
		return err
	}
	path := s.UnitPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return atomicWrite(path, []byte(unit))
}

func (s *systemdManager) Install(ctx context.Context, cfg ServiceConfig) error {
	if err := s.writeUnit(cfg); err != nil {
		return err
	}
	return s.runSystemctl(ctx, "daemon-reload")
}

func (s *systemdManager) Uninstall(ctx context.Context) error {
	s.runSystemctl(ctx, "disable", "--now", UnitLabel)
	s.runSystemctl(ctx, "stop", UnitLabel)
	path := s.UnitPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return s.runSystemctl(ctx, "daemon-reload")
}

func (s *systemdManager) Enable(ctx context.Context) error {
	return s.runSystemctl(ctx, "enable", UnitLabel)
}

func (s *systemdManager) Disable(ctx context.Context) error {
	return s.runSystemctl(ctx, "disable", UnitLabel)
}

func (s *systemdManager) Start(ctx context.Context) error {
	return s.runSystemctl(ctx, "start", UnitLabel)
}

func (s *systemdManager) Stop(ctx context.Context) error {
	return s.runSystemctl(ctx, "stop", UnitLabel)
}

func (s *systemdManager) Restart(ctx context.Context) error {
	return s.runSystemctl(ctx, "restart", UnitLabel)
}

func (s *systemdManager) Status(ctx context.Context) (ServiceStatus, error) {
	st := ServiceStatus{}
	activeOut, _, _ := s.runner.Run(ctx, "systemctl", "--user", "is-active", UnitLabel)
	enabledOut, _, _ := s.runner.Run(ctx, "systemctl", "--user", "is-enabled", UnitLabel)
	st.Loaded = trim(string(activeOut)) == "active"
	st.Enabled = trim(string(enabledOut)) == "enabled"
	st.Autostart = st.Enabled
	st.Reason = trim(string(activeOut))

	if state, err := acolyteruntime.ReadLive(runtimeDir()); err == nil {
		st.PID = state.PID
		if state.Workspace != "" {
			st.Workspace = state.Workspace
		}
	}
	return st, nil
}

func (s *systemdManager) Ready(ctx context.Context) (ServiceStatus, error) {
	st, err := s.Status(ctx)
	if err != nil {
		return st, err
	}
	if !st.Loaded {
		return st, fmt.Errorf("service is not active")
	}
	state, err := acolyteruntime.Read(runtimeDir())
	if err != nil {
		return st, fmt.Errorf("read runtime state: %w", err)
	}
	if !state.Ready {
		return st, fmt.Errorf("worker not yet ready")
	}
	return st, nil
}

func (s *systemdManager) WaitReady(ctx context.Context, timeout time.Duration) (ServiceStatus, error) {
	deadline := time.Now().Add(timeout)
	var last ServiceStatus
	var lastErr error
	for {
		st, err := s.Ready(ctx)
		if err == nil {
			return st, nil
		}
		last = st
		lastErr = err
		if time.Now().After(deadline) {
			return last, lastErr
		}
		select {
		case <-ctx.Done():
			return last, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func (s *systemdManager) runSystemctl(ctx context.Context, args ...string) error {
	all := append([]string{"--user"}, args...)
	_, stderr, err := s.runner.Run(ctx, "systemctl", all...)
	if err != nil {
		return fmt.Errorf("%w: systemctl %s: %s", ErrSystemdOperationFailed, strings.Join(all, " "), trim(string(stderr)))
	}
	return nil
}

func runtimeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".acolyte", ".service")
}

func trim(s string) string {
	for _, c := range []string{"\n", "\r", " "} {
		s = strings.TrimPrefix(s, c)
		s = strings.TrimSuffix(s, c)
	}
	return strings.TrimSpace(s)
}
