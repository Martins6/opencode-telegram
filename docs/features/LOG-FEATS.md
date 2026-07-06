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
