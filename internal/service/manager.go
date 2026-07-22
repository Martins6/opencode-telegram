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
	"text/template"
	"time"
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform: Acolyte requires Linux (systemd user) or macOS (launchd)")
	ErrManagerUnavailable  = errors.New("service manager is not available on this system")
	ErrUnitNotLoaded       = errors.New("service is not loaded")
)

const (
	UnitLabel      = "acolyte"
	LaunchdLabel   = "com.martins6.acolyte"
	systemdTimeout = "30s"
)

var GOOS = currentGOOS

func currentGOOS() string {
	if v := os.Getenv("ACOLYTE_TEST_GOOS"); v != "" {
		return v
	}
	return runtimeGOOS()
}

// runtimeGOOS is a package-level indirection so tests can override.
var runtimeGOOS = func() string { return detectGOOS() }

func detectGOOS() string { return "" } // overridden by platform_init.go on each OS

type ServiceConfig struct {
	Workspace string
	Binary    string
}

type ServiceStatus struct {
	Loaded    bool
	Enabled   bool
	Autostart bool
	PID       int
	Workspace string
	Reason    string
}

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (stdout []byte, stderr []byte, err error)
	LookPath(file string) (string, error)
}

type osRunner struct{}

func (osRunner) Run(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func (osRunner) LookPath(file string) (string, error) { return exec.LookPath(file) }

type Manager interface {
	Install(ctx context.Context, cfg ServiceConfig) error
	Uninstall(ctx context.Context) error
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Restart(ctx context.Context) error
	Status(ctx context.Context) (ServiceStatus, error)
	Ready(ctx context.Context) (ServiceStatus, error)
	WaitReady(ctx context.Context, timeout time.Duration) (ServiceStatus, error)
	UnitPath() string
	PlistPath() string
}

func New(r Runner) (Manager, error) {
	if r == nil {
		r = osRunner{}
	}
	switch GOOS() {
	case "linux":
		if _, err := r.LookPath("systemctl"); err != nil {
			return nil, fmt.Errorf("%w (systemctl not found in PATH)", ErrManagerUnavailable)
		}
		return &systemdManager{runner: r}, nil
	case "darwin":
		if _, err := r.LookPath("launchctl"); err != nil {
			return nil, fmt.Errorf("%w (launchctl not found in PATH)", ErrManagerUnavailable)
		}
		return &launchdManager{runner: r}, nil
	default:
		return nil, ErrUnsupportedPlatform
	}
}

func absPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("empty path")
	}
	return filepath.Abs(p)
}

func renderTemplate(tpl string, data any) (string, error) {
	var buf bytes.Buffer
	if err := template.Must(template.New("t").Parse(tpl)).Execute(&buf, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
