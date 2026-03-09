# Overview
SQLite-based notification and mail system that allows the OpenCode agent to send notifications and emails to the Telegram user. Includes cron skill for setting up automated tasks with proper notification handling.

# Details
- Implements SQLite database at ~/.opencode-telegram/data.db
- Provides notify command: opencode-telegram notify "message" - delivers as "Notification (Agent Unaware)"
- Provides mail command with subcommands: send, open, list
- Supports urgency levels: high (immediate), medium (after 1 hour), low (manual retrieval)
- Background polling service checks every 10 seconds
- Cron skill guides agent on setting up cron jobs with proper notification options

# File Paths
- cmd/notify.go
- cmd/mail.go
- internal/database/db.go
- internal/bot/notifier.go
- internal/config/config.go
- skills/cron.md
