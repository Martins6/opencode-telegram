2026-03-09-12-01 | Added notify, mail, and cron features with SQLite-based notification system
2026-03-05-21-40 | Added Telegram bot core feature with message handling, media processing, and user whitelist
2026-03-05-21-40 | Added install script feature for easy binary deployment via curl
2026-03-05-22-00 | Added username support to allowed_users configuration
2026-03-06-14-00 | Added opencode run integration feature switching from HTTP server to direct command execution
2026-03-06-14-00 | Added progress updates feature for long-running operations
2026-03-07-00-00 | Added telegram-agent feature with custom opencode.json configuration
2026-03-07-00-00 | Added improved default prompts feature for workspace template files
2026-03-07-14-00 | Added AGENTS.md to workspace template with personality file references
2026-03-24-00-00 | Added internal Go-based scheduler replacing external cron/at commands
2026-03-24-00-00 | Added mail agent trigger feature - mail delivery now triggers OpenCode agent response
2026-03-24-00-00 | Removed mail urgency delivery mechanism - all mails delivered immediately
2026-03-24-00-00 | Removed urgency references from workspace templates and scheduler code
2026-03-07-14-30 | Added MAIN-PROMPTS folder reorganization for better workspace template structure
2026-06-03-14-55 | Added scheduler timezone gating with `schedule set` setup and bot.timezone in config.toml
2026-06-03-14-55 | Added runtime config hot-reload for defaults.agent/model/provider and bot.timezone in bot handlers and notifier
2026-06-03-14-55 | Documented the `schedule set` prerequisite in TOOLS.md template, scheduling skill, and feature doc
2026-07-05-21-42 | Added photo support to the Telegram bot - downloads photos, stores them under `<workspace>/downloads/images/`, and forwards a `File located at: <path>\n\nUser message: <caption>` prompt to OpenCode (audio/voice/document/video/sticker still ignored)
2026-07-05-22-00 | Upgraded default model from MiniMax-M2.7 (and M2.5 in docs) to MiniMax-M3 across `internal/config/config.go`, `internal/bot/handlers.go`, `internal/bot/notifier.go`, README example, and `feat-opencode-run-integration.md`
2026-07-07-10-26 | Renamed project from `opencode-telegram` to `acolyte` (GitHub repo, local folder, Go module path, binary name, workspace dir, OpenCode agent name, install scripts, CI, embedded workspace templates, docs). Clean break, no data migration.
2026-07-07-10-26 | Added timestamped filenames for downloaded Telegram media - new `GenerateTimestampedFilename(dir, ext)` helper in `internal/media/downloader.go` produces UTC-stamped names (`20060102_150405.jpg`) with `_1`/`_2` collision suffixes; `GetFilePath` now takes an extension instead of a filename; removed obsolete `FilenameFromTelegramPath` helper and its test.
2026-07-07-12-11 | Added self-update feature - `acolyte update`/`version` subcommands with SHA256-verified GitHub Releases replacement of the running binary, optional in-process restart, ad-hoc codesign on macOS, and a non-blocking 5s startup outdated-check in `acolyte start` that drops a row into the notifications table. Also extends `.github/workflows/release.yml` to publish a `checksums.txt`.
2026-07-21-11-18 | Added Hermes-style singleton user service with `acolyte start/stop/restart/status` controlling a per-user systemd (Linux) or launchd (macOS) service that invokes `acolyte __daemon`. Persisted active workspace path via `~/.acolyte/config.toml`, strict workspace validation that lists every missing template path with an `acolyte new <path>` hint, atomic runtime state in `~/.acolyte/.service/state.json` with `kill(pid,0)` liveness checking, and a service-aware `acolyte update` that restarts under the supervisor instead of re-execing in-process.
2026-07-21-11-18 | Added `acolyte session list` and `acolyte session <sessionID>` wrappers that shell out to `opencode session list` / `opencode export <id>` inside the active workspace via `internal/opencode/session.go` (with a `LookPath` guard and a 30s timeout).
2026-07-21-11-18 | Rewrote `acolyte logs` to `acolyte logs [N] [--date today|YYYY-MM-DD]` over `internal/logger/tail.go`; default N=10, multi-day newest-first scan until N entries, multi-line entries collapsed, non-daily files in `.logs/` skipped.
2026-07-22-09-06 | Added OpenCode-powered PR reviewer workflow (`.github/workflows/opencode.yml`, `.opencode/agent/pr-reviewer.md`, `.opencode/command/review.md`, `.opencode/opencode.json`) — auto-reviews PRs and responds to `/oc` or `/opencode` mentions on issues and review comments, using an Acolyte-aware review checklist (Go module layout, conventional commits, cross-platform service support, no `os.Exit` in `RunE`, context-propagated exec calls, atomic runtime-state writes).
