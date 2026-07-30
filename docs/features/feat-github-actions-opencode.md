# Overview

GitHub Actions workflow that lets maintainers (and bots) invoke OpenCode directly from issue and pull request review comments. A `/oc` or `/opencode` mention in a comment triggers a one-shot OpenCode run with the comment body as the prompt and the comment author's token, all scoped to the current repository.

# Details

- Triggered by `issue_comment` and `pull_request_review_comment` events with `types: [created]`
- Comment body must contain ` /oc`, start with `/oc`, contain ` /opencode`, or start with `/opencode` (the leading space disambiguates `/oc` from paths or words like `/octocat`)
- Runs on `ubuntu-latest` with `id-token: write`, `contents: read`, `pull-requests: read`, and `issues: read` permissions
- Checks out the repo with `actions/checkout@v6` using `persist-credentials: false` so the runner token is not left in `.git/config`
- Calls `anomalyco/opencode/github@latest` (the official OpenCode GitHub Action) with `model: minimax-coding-plan/MiniMax-M3`
- Reads the `MINIMAX_API_KEY` from repository secrets and forwards it to the action
- No file paths inside the Go codebase — pure CI configuration under `.github/`

# File Paths

- .github/workflows/opencode.yml