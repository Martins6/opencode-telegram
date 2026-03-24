2026-03-07-14-30 | Fixed JSONL parsing issue where embedded newlines in JSON string values caused "No response from OpenCode" errors
2026-03-07-14-45 | Removed /reset command - use /new-session to reset conversation while keeping user settings
2026-03-07-14-45 | Removed log truncation to capture full OpenCode output without character limits
2026-03-05-21-40 | Fixed config.toml not being created when running `opencode-telegram new` command
2026-03-05-22-00 | Fixed config set command overwriting entire config file instead of merging
2026-03-05-22-00 | Fixed allowed_users config array parsing issue with JSON array strings
2026-03-06-14-00 | Fixed empty response handling in OpenCode client to return meaningful error
2026-03-06-14-00 | Fixed model object format to use providerID and modelID fields
2026-03-06-14-00 | Fixed message response polling to retrieve responses correctly
2026-03-06-14-00 | Fixed excessive polling by implementing smarter polling intervals
2026-03-24-00-00 | Fixed mail content parsing - strings.Fields now properly handles quoted arguments with spaces
2026-03-24-00-00 | Fixed scheduled notify command - extended matching to handle bare "notify" commands
2026-03-24-00-00 | Fixed schedule add and notify subprocess - added config.Load() in PersistentPreRun/PreRun
2026-03-24-00-00 | Fixed notify and mail CLI commands - proper chat ID resolution and error handling
2026-03-24-00-00 | Fixed scheduled notifications not appearing - wrong user ID in scheduler/notifier
2026-03-24-00-00 | Fixed notify mail scheduler - command detection now handles -m flag format
2026-03-24-00-00 | Fixed cron skill and tool visibility - updated TOOL.md and AGENTS.md workspace files
2026-03-24-00-00 | Fixed database initialization - workspace path now passed to database.Init()
2026-03-24-00-00 | Fixed database error - created .opencode-telegram directory before opening database
2026-03-06-14-00 | Fixed OpenCode server communication with proper session management
