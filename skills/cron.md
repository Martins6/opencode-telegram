# Cron Skill

## Overview

This skill guides the OpenCode agent on how to set up cron jobs for repetitive automated tasks and how to ensure the results reach the Telegram user properly.

## When to Use Cron

Use cron jobs for:
- Daily reports (weather, news digests, system health)
- Periodic backups or cleanup tasks
- Scheduled notifications or reminders
- Any repetitive task that runs automatically in the background

## Cron Basics

### Crontab Format

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday = 0)
│ │ │ │ │
* * * * * command
```

### Common Examples

- `0 9 * * *` - Daily at 9 AM
- `0 * * * *` - Every hour
- `0 0 * * *` - Midnight every day
- `*/15 * * * *` - Every 15 minutes

## Notification Options

When setting up cron jobs, you MUST ensure the output reaches the user. Choose the appropriate method:

### 1. Use `opencode-telegram notify` for:
- Silent background tasks
- Status updates the user doesn't need to act on
- Error alerts or system monitoring
- Any output the user just needs to be aware of

Example:
```bash
0 9 * * * /path/to/backup.sh > /tmp/backup.log 2>&1 && opencode-telegram notify "Daily backup completed successfully"
```

### 2. Use `opencode-telegram mail` for:
- Tasks where the user should review the output
- Information the user needs to act on
- Reports that should appear in the working context

Example:
```bash
0 8 * * * /path/to/news_digest.sh | opencode-telegram mail send --sender "cron@system" --subject "Daily News Digest" --content "$(cat)" --urgency medium
```

## User Preference Questions

When setting up cron jobs, ALWAYS ask the user:

1. **Notification Preference**:
   - "Do you want me to notify you when this completes (notify), or do you want to see the full output in your working context (mail)?"

2. **Urgency Level** (if using mail):
   - "Should this be high urgency (appears immediately), medium (appears after 1 hour), or low (you'll retrieve it manually)?"

3. **Timing**:
   - "What time/frequency works best for you?"

4. **Command Verification**:
   - Show the cron entry before adding it
   - Ask if the user wants to test it once manually first

## Examples

### Example 1: Daily Weather Report
User: "Every morning at 7 AM, I want a weather summary"

1. Ask: "Do you want to see the weather in your working context (mail) or just get a notification when it's ready (notify)?"
2. If mail: "What urgency level - high (immediate), medium (1 hour), or low (manual retrieval)?"
3. Set up cron:
```bash
0 7 * * * /path/to/weather.sh | opencode-telegram mail send --sender "weather@cron" --subject "Morning Weather" --content "$(cat)" --urgency high
```

### Example 2: Disk Space Monitoring
User: "Alert me if disk space is low"

1. This should use notify since it's an alert
2. Set up cron:
```bash
0 */4 * * * df -h | awk '$5 > 90 {print "Disk usage at " $5}' | xargs -I {} opencode-telegram notify "{}"
```

### Example 3: GitHub Notifications
User: "Summarize my GitHub notifications daily"

1. Ask: "Do you want to see the full summary in your working context or just a notification?"
2. Set up accordingly with mail or notify

## Important Notes

- Always redirect stderr to stdout (`2>&1`) so errors are captured
- Use absolute paths in cron jobs
- Cron environment is minimal - test commands manually first
- For long-running tasks, consider using `nohup` or `screen`
- The Telegram bot must be running for notifications to be delivered

## Skill Usage

When a user mentions:
- "Schedule something"
- "Set up a cron job"
- "Run this periodically"
- "Remind me every..."
- Any repetitive task request

Apply this skill to:
1. Understand the task and frequency
2. Ask user preferences for notification
3. Create the cron entry
4. Test if possible
5. Confirm with user
