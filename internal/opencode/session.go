package opencode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

var (
	ErrOpenCodeMissing  = errors.New("opencode CLI not found in PATH")
	ErrInvalidSessionID = errors.New("invalid session ID")
)

const sessionCommandTimeout = 30 * time.Second

func RunOpenCode(ctx context.Context, workspace string, args []string) ([]byte, []byte, error) {
	if _, err := exec.LookPath("opencode"); err != nil {
		return nil, nil, ErrOpenCodeMissing
	}

	cmd := exec.CommandContext(ctx, "opencode", args...)
	cmd.Dir = workspace

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("opencode %s failed: %s", args, stderr.String())
		}
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("opencode %s failed: %w", args, err)
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

func SessionList(ctx context.Context, workspace string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, sessionCommandTimeout)
	defer cancel()

	stdout, _, err := RunOpenCode(ctx, workspace, []string{"session", "list"})
	if err != nil {
		return nil, fmt.Errorf("session list: %w", err)
	}
	return stdout, nil
}

func SessionExport(ctx context.Context, workspace, sessionID string) ([]byte, error) {
	if sessionID == "" {
		return nil, ErrInvalidSessionID
	}

	ctx, cancel := context.WithTimeout(ctx, sessionCommandTimeout)
	defer cancel()

	stdout, _, err := RunOpenCode(ctx, workspace, []string{"export", sessionID})
	if err != nil {
		return nil, fmt.Errorf("session export: %w", err)
	}
	return stdout, nil
}
