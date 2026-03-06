# Overview

Core Telegram bot that enables users to interact with OpenCode AI agent from Telegram. Handles text messages, media uploads, slash commands, and enforces user access control.

# Details

- Implements Telegram Bot API using go-telegram/bot library
- Handles text messages and forwards to OpenCode server
- Processes media uploads: photos, audio, voice, documents, videos
- Implements slash commands: /set-agent, /set-model, /set-provider, /workspace, /help, /reset
- Restricts access to users defined in allowed_users config
- Supports both numeric user IDs and usernames in allowed_users (e.g., ["123456789", "username123"])
- Manages per-user conversation sessions and history
- Integrates with embedded OpenCode server
- Provides configuration management via config.toml

# File Paths

- cmd/root.go - CLI root command
- cmd/start.go - Start command to launch bot and server
- cmd/config.go - Config management commands
- cmd/new.go - Workspace initialization
- cmd/logs.go - Log viewing command
- internal/bot/client.go - Bot initialization
- internal/bot/handlers.go - Message handlers
- internal/bot/middleware.go - User authentication middleware
- internal/bot/commands.go - Slash command implementations
- internal/config/config.go - Configuration loading
- internal/opencode/client.go - OpenCode server client
- internal/opencode/server.go - OpenCode server management
- internal/media/downloader.go - Media file downloader
- internal/media/storage.go - Media file storage
- internal/session/manager.go - Session management
- internal/logger/logger.go - Logging system
- internal/workspace/files.go - Workspace file handling
