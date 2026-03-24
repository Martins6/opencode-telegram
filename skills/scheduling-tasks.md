# Scheduling Tasks Skill

## Overview

This skill guides the OpenCode agent on how to set up scheduled tasks using the built-in scheduler. The scheduler executes shell commands and automatically handles output:
- **On success**: Output is sent to mail (visible to AI in context)
- **On failure**: Notification is sent to user

## When to Use This Skill

Use scheduled tasks for:
- Daily reports (weather, news digests, system health)
- Periodic backups or cleanup tasks
- Scheduled notifications or reminders
- One-time reminders ("remind me in 5 minutes")
- Any task that runs automatically in the background

## Built-in Scheduler Commands

Use `opencode-telegram schedule` to manage scheduled tasks.

### Any Bash Command

**Any bash command or script can be scheduled.** For example:
- `curl https://api.example.com` - HTTP requests
- `git pull` - Git operations
- `npm run build` - Build commands
- `echo "hello"` - Simple commands
- `python script.py` - Python scripts

### Add a Scheduled Task

```bash
# One-time schedules
opencode-telegram schedule add -s "in 30m" -c "echo hello"
opencode-telegram schedule add -s "at 09:00" -c "backup.sh"
opencode-telegram schedule add -s "once 14:30" -c "task.sh"

# Cron schedules (recurring)
opencode-telegram schedule add -s "0 9 * * *" -c "morning-task.sh"
opencode-telegram schedule add -s "*/15 * * * *" -c "check-status.sh"
```

### Schedule Expression Formats

#### One-time:
- `in 30m` - Run in 30 minutes from now
- `at HH:MM` - Run at specific time today (e.g., "at 09:00", "at 14:30")
- `once HH:MM` - Same as "at" (run once at specific time)
- `now + 1h` - Run after duration (e.g., "now + 1h30m")

#### Cron (recurring):
Standard cron format: `minute hour day-of-month month day-of-week`

- `0 9 * * *` - Daily at 9 AM
- `0 * * * *` - Every hour
- `0 0 * * *` - Midnight every day
- `*/15 * * * *` - Every 15 minutes
- `0 9 * * 1-5` - Weekdays at 9 AM

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-s, --schedule` | Schedule expression (cron or one-time) | (required) |
| `-c, --command` | Shell command to execute | (required) |
| `-d, --dir` | Working directory | workspace |
| `--on-success` | Action on success: `mail` or `notify` | `mail` |
| `--on-failure` | Action on failure: `notify` or `ignore` | `notify` |

### Manage Scheduled Tasks

```bash
# List all scheduled tasks
opencode-telegram schedule list

# Delete a scheduled task
opencode-telegram schedule delete <id>

# Run a task immediately (for testing)
opencode-telegram schedule run <id>
```

## Output Handling

The scheduler handles output automatically:

- **On Success**: Command output is sent to mail with subject "Scheduled Task: [command]"
   - When mail is delivered, the AI agent is triggered and responds to the user
   - The agent will proactively engage with the user about the mail content
   - Useful for reports, logs, data processing results

- **On Failure**: A notification is sent to the user
   - Includes the error message
   - User is immediately alerted to issues
   - Note: Notifications are silent (agent is NOT triggered)

You can customize this behavior:
- `--on-success notify` - Send output via notification instead of mail (silent)
- `--on-failure ignore` - Don't notify on failure

## User Preference Questions

When setting up scheduled tasks, ALWAYS ask the user:

1. **Timing**:
   - "When should this happen?"
   - For one-time: "In how many minutes/hours?" or "At what time?"
   - For recurring: "Daily? Weekly? At what time?"

2. **Output Preference**:
   - "Do you want to see the full output in your working context (mail), or just get a notification when it completes?"

3. **Working Directory**:
   - "Should this run in a specific directory, or the default workspace?"

4. **Command Verification**:
   - Show the command before adding it
   - Ask if the user wants to test it once manually first

## Examples

### Example 1: Daily Weather Report
User: "Every morning at 7 AM, I want a weather summary"

1. Ask: "Do you want to see the full weather output (mail) or just a notification when it's ready?"
2. Ask: "What directory should this run in?"
3. Set up:
```bash
opencode-telegram schedule add -s "0 7 * * *" -c "weather-script.sh" --dir /path/to/scripts
```

### Example 2: One-Time Reminder
User: "Remind me to call mom in 30 minutes"

1. This is a one-time task
2. Set up:
```bash
opencode-telegram schedule add -s "in 30m" -c "echo 'Time to call mom!'"
```

### Example 3: Disk Space Monitoring
User: "Alert me if disk space is low every 4 hours"

1. This should use notify on failure to alert the user
2. Set up:
```bash
opencode-telegram schedule add -s "0 */4 * * *" -c "df -h | awk '\$5 > 90 {print \"Disk at \" \$5}'"
```

### Example 4: GitHub Notifications
User: "Summarize my GitHub notifications daily at 8 AM"

1. Ask: "Do you want to see the full summary in your working context (mail) or just a notification?"
2. Set up:
```bash
opencode-telegram schedule add -s "0 8 * * *" -c "github-summary.sh"
```

### Example 5: Backup Script with Custom Output
User: "Run my backup script daily at midnight and notify me"

1. User wants notification on success (not mail)
2. Set up:
```bash
opencode-telegram schedule add -s "0 0 * * *" -c "backup.sh" --on-success notify
```

## Important Notes

- The Telegram bot must be running for scheduled tasks to execute
- Scheduled tasks are stored in SQLite and persist across bot restarts
- Working directory defaults to the workspace folder (~/.opencode-telegram)
- Use absolute paths in commands for reliability
- Test commands manually before scheduling

## Skill Usage

When a user mentions:
- "Schedule something"
- "Set up a cron job"
- "Run this periodically"
- "Remind me every..."
- "Remind me in X minutes/hours"
- "Run once at..."
- "Set up a task to run automatically"
- Any task scheduling request

Apply this skill to:
1. Understand the task and timing
2. Ask user preferences for output handling
3. Create the scheduled task using `opencode-telegram schedule add`
4. Test if possible
5. Confirm with user
