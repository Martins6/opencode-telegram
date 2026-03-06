package opencode

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/martins6/opencode-telegram/internal/logger"
)

type Runner struct {
	workspace string
	agent     string
	model     string
	provider  string
}

type RunResult struct {
	SessionID    string
	ResponseText string
}

func NewRunner(workspace, agent, model, provider string) *Runner {
	return &Runner{
		workspace: workspace,
		agent:     agent,
		model:     model,
		provider:  provider,
	}
}

func (r *Runner) Execute(sessionID, message string) (*RunResult, error) {
	args := []string{
		"run",
		"--model",
		fmt.Sprintf("%s/%s", r.provider, r.model),
		"--agent",
		r.agent,
		"--format",
		"json",
	}

	if sessionID != "" {
		args = append(args, "--continue")
		args = append(args, sessionID)
	}

	args = append(args, message)

	logger.LogDebug("Executing opencode run: opencode %s", strings.Join(args, " "))

	cmd := exec.Command("opencode", args...)
	cmd.Dir = r.workspace

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return nil, fmt.Errorf("opencode run failed: %s", stderr.String())
			}
			return nil, fmt.Errorf("opencode run failed: %w", err)
		}
	case <-time.After(300 * time.Second):
		cmd.Process.Kill()
		return nil, fmt.Errorf("opencode run timed out after 5 minutes")
	}

	output := stdout.String()
	logger.LogDebug("Opencode run output: %s", truncate(output, 500))

	return r.parseOutput(output)
}

func (r *Runner) parseOutput(output string) (*RunResult, error) {
	result := &RunResult{}
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			logger.LogDebug("Failed to parse JSON line: %s", line)
			continue
		}

		if _, ok := event["sessionID"]; !ok {
			continue
		}

		var sessionID string
		if err := json.Unmarshal(event["sessionID"], &sessionID); err == nil && result.SessionID == "" {
			result.SessionID = sessionID
		}

		partData, ok := event["part"]
		if !ok {
			continue
		}

		var part map[string]json.RawMessage
		if err := json.Unmarshal(partData, &part); err != nil {
			continue
		}

		partType, ok := part["type"]
		if !ok {
			continue
		}

		var typeStr string
		if err := json.Unmarshal(partType, &typeStr); err != nil {
			continue
		}

		if typeStr == "text" {
			textData, ok := part["text"]
			if !ok {
				continue
			}
			var text string
			if err := json.Unmarshal(textData, &text); err == nil {
				result.ResponseText += text
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.LogDebug("Scanner error: %v", err)
	}

	if result.SessionID == "" && result.ResponseText == "" {
		return nil, fmt.Errorf("no valid output from opencode run")
	}

	return result, nil
}
