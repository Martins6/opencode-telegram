---
description: Run the PR reviewer on the current branch's diff
---
Run a focused review of the current branch against its base.
1. Run `git diff $(git merge-base HEAD origin/main)...HEAD --stat` to see what changed. If the diff is large (>500 lines), focus on the files that changed most.
2. For each changed file, read the relevant sections and apply the same checks as the `pr-reviewer` agent: Go module layout, Conventional Commits, `gofmt` clean, `go vet` clean, cross-platform support, no hardcoded secrets, no comments without explicit ask, no `os.Exit` inside `RunE`, context-propagated exec calls, atomic file writes for runtime state.
3. Output a verdict: `LGTM`, `NEEDS CHANGES`, or `DISCUSS`, with file:line findings and concrete fixes.
Do not commit anything. Do not push. Just report.