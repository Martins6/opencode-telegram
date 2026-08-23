# Overview

Adds two GitHub Actions workflows that wire the OpenCode `build` agent into the repository using the official `anomalyco/opencode/github` action: a slash-command trigger on `/opencode` or `/oc` in issue/PR comments, and an automatic code review that fires on PR open and every new commit.

# Details

- Adds `.github/workflows/opencode.yml` triggered by `issue_comment` (type `created`) and `pull_request_review_comment` (type `created`); only runs when the comment body contains `/oc` or `/opencode`
- Job uses `ubuntu-latest`, read-only permissions (`contents: read`, `pull-requests: read`, `issues: read`) plus `id-token: write` for OpenCode GitHub App OIDC, and explicitly pins `agent: build` with model `minimax-coding-plan/MiniMax-M3`
- Adds `.github/workflows/opencode-review.yml` triggered by `pull_request` types `[opened, synchronize, reopened, ready_for_review]`; skips drafts via `if: github.event.pull_request.draft == false` so `synchronize` re-runs the same review on every new commit
- Review job uses the same read-only permission set, the same model/agent, and a code-review-focused `prompt` covering code quality, Go idiom violations, bugs/race conditions/error-handling gaps, and security concerns (secret leakage, unsafe parsing)
- Auth uses the OpenCode GitHub App via OIDC; the API key is supplied via `${{ secrets.MINIMAX_API_KEY }}` — the secret name is written for `minimax-coding-plan` and must be verified against `opencode auth login` / models.dev
- Project-level `opencode.json` at the repo root pins the `minimax-coding-plan` provider and `MiniMax-M3` model so the opencode server started by the action resolves the model reliably in the runner (which has no local auth/config to fall back on); this is the same pattern that resolved [opencode#7958](https://github.com/anomalyco/opencode/issues/7958)
- README.md gains a "GitHub Workflow" subsection under Development covering app install, secret naming, slash-command usage, and auto-review behavior
- Workflow files mirror the official opencode docs pattern (`actions/checkout@v4` with `fetch-depth: 1` and `persist-credentials: false`, then `anomalyco/opencode/github@latest`)

# File Paths

- .github/workflows/opencode.yml
- .github/workflows/opencode-review.yml
- opencode.json
- README.md
