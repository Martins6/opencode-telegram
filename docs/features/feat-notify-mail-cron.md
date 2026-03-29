# Overview
SQLite-based notification and mail system that allows the OpenCode agent to send notifications and emails to the Telegram user. Includes internal Go-based scheduler for automated tasks and mail delivery triggers agent responses.

# Details
- Implements SQLite database at ~/.opencode-telegram/data.db
- Provides notify command: opencode-telegram notify "message" - delivers as "Notification (Agent Unaware)"
- Provides mail command: opencode-telegram mail send --sender X --subject Y --content Z
- All mails delivered immediately (no urgency levels)
- Mail delivery triggers OpenCode agent to respond with full mail details
- Internal Go-based scheduler executes shell commands at specified times
- Scheduler stores tasks in SQLite with cron/at expression support
- Background polling service checks every 10 seconds
- Scheduling skill guides agent on setting up scheduled tasks

# File Paths
- cmd/notify.go
- cmd/mail.go
- cmd/schedule.go
- internal/database/db.go
- internal/bot/notifier.go
- internal/scheduler/service.go
- internal/config/config.go
- skills/scheduling-tasks.md

# Configuration

To receive notifications, add your Telegram chat ID to the config file at `~/.opencode-telegram/config.toml`:

```toml
[notifications]
user_ids = [8347582793]  # Replace with your Telegram chat ID
```

To find your Telegram chat ID:
1. Message the bot
2. Check the bot logs - incoming message chat IDs are logged

