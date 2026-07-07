# Overview
SQLite-based notification and mail system that allows the OpenCode agent to send notifications and emails to the Telegram user. Includes internal Go-based scheduler for automated tasks, mail delivery triggers agent responses, and a runtime config hot-reload layer that keeps the running daemon in sync with `config.toml` mutations.

# Details
- Implements SQLite database at ~/.acolyte/data.db
- Provides notify command: acolyte notify "message" - delivers as "Notification (Agent Unaware)"
- Provides mail command: acolyte mail send --sender X --subject Y --content Z
- All mails delivered immediately (no urgency levels)
- Mail delivery triggers OpenCode agent to respond with full mail details
- Internal Go-based scheduler executes shell commands at specified times
- Scheduler stores tasks in SQLite with cron/at expression support
- Background polling service checks every 5 seconds
- Scheduling skill guides agent on setting up scheduled tasks
- Timezone is mandatory for the scheduler: `bot.timezone` in `config.toml` is the only source of truth
- New `acolyte schedule set --timezone <IANA>` subcommand persists the timezone (one-time setup, idempotent)
- All other `schedule` subcommands (`add`, `list`, `delete`, `run`) are gated on `bot.timezone` being set; the gate failure is the setup prompt
- `notify` and `mail` are NOT gated - they have no time component
- `parseSchedule` accepts a `*time.Location`; `at HH:MM` / `once HH:MM` / cron expressions are resolved in the configured zone
- `executeTask` and `GetDueScheduledTasks(userID, cutoff)` thread the configured `time.Location` through so the SQLite cutoff and cron `Next()` use the user's zone (not the host's UTC)
- Runtime config hot-reload: `internal/bot/handlers.go` and `internal/bot/notifier.go` call `config.Load("")` on every message / mail tick so changes to `defaults.agent`, `defaults.model`, `defaults.provider`, and `bot.timezone` are picked up within microseconds, with no daemon restart
- Restart-only fields: `bot.token`, `bot.allowed_user_id`, `workspace.path` (token would require tearing down the Telegram client connection; intentionally not hot-reloaded)
- `config.GetLocation()` returns `(*time.Location, error)` and exposes `ErrTimezoneNotConfigured` for callers that want to gate on the zone

# File Paths
- cmd/notify.go
- cmd/mail.go
- cmd/schedule.go
- cmd/schedule_test.go
- internal/bot/client.go
- internal/bot/handlers.go
- internal/bot/notifier.go
- internal/bot/notifier_test.go
- internal/config/config.go
- internal/config/config_test.go
- internal/database/db.go
- internal/database/db_test.go
- internal/scheduler/service.go
- internal/scheduler/service_test.go
- internal/workspace/files.go
- skills/scheduling-tasks.md
- docs/features/feat-notify-mail-cron.md

# Configuration

To receive notifications, add your Telegram chat ID to the config file at `~/.acolyte/config.toml`:

```toml
[notifications]
user_ids = [8347582793]  # Replace with your Telegram chat ID
```

To find your Telegram chat ID:
1. Message the bot
2. Check the bot logs - incoming message chat IDs are logged

To enable scheduling, set the timezone once:

```bash
acolyte schedule set --timezone America/Sao_Paulo
```

This writes `[bot] timezone = "America/Sao_Paulo"` to `config.toml`. The running daemon picks it up on the next message / notifier / scheduler tick - no restart required. Re-run with a different IANA zone to change it; existing tasks are re-interpreted under the new zone on next fire.
