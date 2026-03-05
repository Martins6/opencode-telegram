# OpenCode Telegram Agent - Implementation Plan

## Summary
Build a Telegram bot gateway that allows users to interact with an OpenCode AI agent directly from Telegram. The bot must handle text messages, media uploads (images, audio, documents, videos), and provide configuration via slash commands while restricting access to only allowed users.

## Solution Overview
Create a Go CLI application using Cobra for command structure and Viper for configuration management. The bot will use the `go-telegram/bot` library for Telegram Bot API integration, start an embedded OpenCode server locally when launched, handle user messages and media, forward them to the OpenCode server, manage per-user sessions and conversation history, and restrict access to only users defined in the allowed list.

## Steps to Solve

### Step 1: Project Initialization & Dependencies
- Set up the Go project structure with all required dependencies
- Files: go.mod, main.go, cmd/root.go, cmd/start.go, cmd/config.go, cmd/new.go, cmd/logs.go
- Dependencies: github.com/go-telegram/bot, github.com/spf13/cobra, github.com/spf13/viper, github.com/google/uuid

### Step 2: Configuration Management
- Implement configuration struct and file handling with TOML support
- Files: internal/config/config.go, internal/config/types.go
- Config Structure: bot.token, bot.allowed_users, workspace.path, opencode.port, opencode.password, defaults.agent, defaults.model, defaults.provider

### Step 3: Telegram Bot Core
- Set up the bot initialization, update handlers, and user filtering middleware
- Files: internal/bot/client.go, internal/bot/handlers.go, internal/bot/middleware.go
- Key Components: Initialize bot, user whitelist middleware, message handlers for all media types, slash commands

### Step 4: Slash Commands Implementation
- Implement all Telegram slash commands: /set-agent, /set-model, /set-provider, /workspace, /help, /reset
- Files: internal/bot/commands.go

### Step 5: OpenCode Server Integration
- Implement HTTP client to communicate with OpenCode server
- Files: internal/opencode/client.go, internal/opencode/types.go, internal/opencode/server.go
- Endpoints: POST /session/:id/message, GET /global/health, GET /session, POST /session

### Step 6: Media Handling
- Implement media download, storage, and prompt construction
- Files: internal/media/downloader.go, internal/media/storage.go, internal/media/types.go
- Media Types: Photo, Audio, Voice, Document, Video

### Step 7: User Session Management
- Manage per-user conversation state and settings
- Files: internal/session/manager.go, internal/session/store.go, internal/session/types.go

### Step 8: Logging System
- Implement structured logging with file rotation
- Files: internal/logger/logger.go, internal/logger/formatter.go

### Step 9: Workspace Template
- Create workspace initialization with template files
- Files: internal/workspace/template.go, internal/workspace/files.go
- Template Files: AGENTS.md, SOUL.md, USER.md, IDENTITY.md, BOOTSTRAP.md, TOOLS.md

### Step 10: CLI Commands
- Implement remaining CLI commands (config, new, logs)
- Files: cmd/config.go, cmd/new.go, cmd/logs.go

### Step 11: Graceful Shutdown
- Handle process termination cleanly
- Files: cmd/start.go - Add signal handling

## Challenges
1. OpenCode Server Lifecycle - Use Go's exec package with proper process management
2. Rate Limiting - Implement message queuing and respect retry_after headers
3. User Session Persistence - Store session data in JSON files
4. Media File Size Limits - Validate file sizes before download
5. Concurrent Message Handling - Use Go channels and mutexes

## Risk Assessment
- Level: Medium
- Key Risks: Server startup failure, rate limiting, permission issues, network timeouts, memory leaks

## Fallback Plan
- Remote Server Option: Add config option to connect to existing OpenCode server
- Simplified Messaging: If OpenCode API unavailable, respond with error
- Queue-Based Processing: Queue messages and retry on failure
