package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	acolyteruntime "github.com/martins6/acolyte/internal/runtime"
)

var ErrLaunchdOperationFailed = errors.New("launchctl operation failed")

type launchdManager struct {
	runner Runner
}

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", LaunchdLabel+".plist")
}

func (l *launchdManager) UnitPath() string { return "" }

func (l *launchdManager) PlistPath() string { return plistPath() }

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{.Label}}</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.Binary}}</string>
    <string>__daemon</string>
  </array>
  <key>WorkingDirectory</key>
  <string>{{.Workspace}}</string>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <dict>
    <key>Crashed</key>
    <true/>
    <key>SuccessfulExit</key>
    <false/>
  </dict>
  <key>EnvironmentVariables</key>
  <dict>
    <key>PATH</key>
    <string>{{.PathEnv}}</string>
  </dict>
  <key>StandardOutPath</key>
  <string>{{.StdoutLog}}</string>
  <key>StandardErrorPath</key>
  <string>{{.StderrLog}}</string>
  <key>ProcessType</key>
  <string>Background</string>
</dict>
</plist>`

func renderPlist(cfg ServiceConfig) (string, error) {
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
	stdoutLog := filepath.Join(ws, ".logs", "launchd.out.log")
	stderrLog := filepath.Join(ws, ".logs", "launchd.err.log")
	return renderTemplate(plistTemplate, map[string]string{
		"Label":     LaunchdLabel,
		"Binary":    bin,
		"Workspace": ws,
		"PathEnv":   pathEnv,
		"StdoutLog": stdoutLog,
		"StderrLog": stderrLog,
	})
}

func (l *launchdManager) writePlist(cfg ServiceConfig) error {
	text, err := renderPlist(cfg)
	if err != nil {
		return err
	}
	path := l.PlistPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cfg.Workspace, ".logs"), 0o755); err != nil {
		return err
	}
	return atomicWrite(path, []byte(text))
}

func (l *launchdManager) Install(ctx context.Context, cfg ServiceConfig) error {
	if err := l.writePlist(cfg); err != nil {
		return err
	}
	return nil
}

func (l *launchdManager) Uninstall(ctx context.Context) error {
	l.runLaunchctl(ctx, "bootout", guiSpec(), plistPath())
	if err := os.Remove(plistPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l *launchdManager) Enable(ctx context.Context) error {
	_, stderr, err := l.runner.Run(ctx, "launchctl", "enable", guiSpec())
	if err != nil && !strings.Contains(string(stderr), "already") {
		return fmt.Errorf("%w: launchctl enable: %s", ErrLaunchdOperationFailed, trim(string(stderr)))
	}
	return nil
}

func (l *launchdManager) Disable(ctx context.Context) error {
	_, stderr, err := l.runner.Run(ctx, "launchctl", "disable", guiSpec())
	if err != nil {
		return fmt.Errorf("%w: launchctl disable: %s", ErrLaunchdOperationFailed, trim(string(stderr)))
	}
	return nil
}

func (l *launchdManager) Start(ctx context.Context) error {
	if _, _, err := l.runner.Run(ctx, "launchctl", "print", guiSpec()); err == nil {
		return nil
	}
	if err := l.Enable(ctx); err != nil {
		return err
	}
	return l.runLaunchctl(ctx, "bootstrap", guiSpec(), plistPath())
}

func (l *launchdManager) Stop(ctx context.Context) error {
	if _, _, err := l.runner.Run(ctx, "launchctl", "print", guiSpec()); err != nil {
		return nil
	}
	return l.runLaunchctl(ctx, "bootout", guiSpec(), plistPath())
}

func (l *launchdManager) Restart(ctx context.Context) error {
	if _, _, err := l.runner.Run(ctx, "launchctl", "print", guiSpec()); err == nil {
		return l.runLaunchctl(ctx, "kickstart", "-k", guiSpec())
	}
	return l.Start(ctx)
}

func (l *launchdManager) Status(ctx context.Context) (ServiceStatus, error) {
	st := ServiceStatus{}
	_, _, printErr := l.runner.Run(ctx, "launchctl", "print", guiSpec())
	st.Loaded = printErr == nil
	_, _, enableErr := l.runner.Run(ctx, "launchctl", "print-disabled", guiSpec())
	st.Enabled = enableErr == nil
	st.Autostart = st.Enabled
	if state, err := acolyteruntime.ReadLive(runtimeDir()); err == nil {
		st.PID = state.PID
		st.Workspace = state.Workspace
	}
	if !st.Loaded {
		st.Reason = "service is not loaded"
	}
	return st, nil
}

func (l *launchdManager) Ready(ctx context.Context) (ServiceStatus, error) {
	st, err := l.Status(ctx)
	if err != nil {
		return st, err
	}
	if !st.Loaded {
		return st, fmt.Errorf("service is not loaded")
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

func (l *launchdManager) WaitReady(ctx context.Context, timeout time.Duration) (ServiceStatus, error) {
	deadline := time.Now().Add(timeout)
	var last ServiceStatus
	var lastErr error
	for {
		st, err := l.Ready(ctx)
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

func (l *launchdManager) runLaunchctl(ctx context.Context, args ...string) error {
	_, stderr, err := l.runner.Run(ctx, "launchctl", args...)
	if err != nil {
		if isBenignLaunchdErr(string(stderr)) {
			return nil
		}
		return fmt.Errorf("%w: launchctl %s: %s", ErrLaunchdOperationFailed, strings.Join(args, " "), trim(string(stderr)))
	}
	return nil
}

func guiSpec() string {
	if v := os.Getenv("ACOLYTE_TEST_UID"); v != "" {
		return "gui/" + v
	}
	cmd := exec.Command("id", "-u")
	out, err := cmd.Output()
	if err != nil {
		return "gui/" + os.Getenv("USER")
	}
	return "gui/" + string(bytes.TrimSpace(out))
}

func isBenignLaunchdErr(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "already loaded") ||
		strings.Contains(s, "service is disabled") ||
		strings.Contains(s, "not found")
}
