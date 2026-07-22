# Overview

Adds an OpenCode-powered PR reviewer workflow that comments on PRs and responds to `/oc` / `/opencode` mentions on issues and PR review comments. The reviewer runs the `pr-reviewer` agent adapted for the Acolyte CLI (Go, Cobra, Telegram bot gateway to OpenCode).

# Details

- GitHub Actions workflow at `.github/workflows/opencode.yml` triggers on `pull_request`, `issue_comment`, and `pull_request_review_comment`.
- Only runs on PR events when the PR is not a draft.
- On comment events the body must contain `/oc` or `/opencode`.
- Uses `anomalyco/opencode/github@latest` with model `minimax-coding-plan/MiniMax-M3` and agent `pr-reviewer`.
- Requires the `MINIMAX_API_KEY` repository secret.
- Custom `.opencode/agent/pr-reviewer.md` adapts the review checklist to Acolyte: Go module layout, Conventional Commits, `gofmt` / `go vet` / `go test`, cross-platform service support (Linux systemd / macOS launchd), no Windows-specific code, no `os.Exit` inside `RunE`, context propagation on `exec.Command`, atomic file writes, race-free package globals.
- `.opencode/command/review.md` is an interactive `/review` slash command for local reviews.
- `.opencode/opencode.json` sets the default model and the default agent (`build`).

# File Paths

- `.github/workflows/opencode.yml`
- `.opencode/opencode.json`
- `.opencode/agent/pr-reviewer.md`
- `.opencode/command/review.md`