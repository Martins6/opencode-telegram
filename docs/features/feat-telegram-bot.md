# Overview

Core Telegram bot that enables users to interact with OpenCode AI agent from Telegram. Handles text messages, photo uploads, slash commands, and enforces user access control.

# Details

- Implements Telegram Bot API using go-telegram/bot library
- Handles text messages and forwards to OpenCode
- Accepts photo uploads: downloads them, stores them under `<workspace>/downloads/images/` with UTC-timestamped filenames (e.g. `20240706_143052.jpg`, with `_1`/`_2` collision suffixes for same-second duplicates), and forwards a `File located at: <path>\n\nUser message: <caption>` prompt to OpenCode
- Documents, audio, voice, video, stickers, contacts, and locations are ignored (only photos are wired in)
- Sender-side media (the bot sending photos/documents back to Telegram) is NOT implemented; all replies are plain text
- Implements slash commands: /set-agent, /set-model, /set-provider, /workspace, /help, /new-session
- Restricts access to the configured allowed user
- Manages per-user conversation sessions (in-memory)
- Integrates with the local `opencode` CLI via `opencode run` (NDJSON)

# File Paths

- cmd/root.go - CLI root command
- cmd/start.go - Start command to launch bot
- cmd/config.go - Config management commands
- cmd/new.go - Workspace initialization
- cmd/logs.go - Log viewing command
- internal/bot/client.go - Bot initialization
- internal/bot/handlers.go - Message handlers (text + photo)
- internal/bot/commands.go - Slash command implementations
- internal/config/config.go - Configuration loading
- internal/opencode/runner.go - OpenCode CLI runner
- internal/media/downloader.go - Telegram file downloader + prompt builder
- internal/session/manager.go - Session management
- internal/logger/logger.go - Logging system
- internal/workspace/template.go - Workspace file/directory templates