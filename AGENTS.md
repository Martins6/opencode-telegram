# Project: Telegram OpenCode Agent

A Telegram bot that acts as a gateway to an OpenCode server, enabling users to interact with the OpenCode AI agent directly from Telegram.

## Core Tech

- **Go 1.21+** - Backend language
- **go-telegram/bot** - Telegram Bot API
- **spf13/cobra** - CLI framework
- **OpenCode CLI** - AI agent server (runs locally on port 4096)

## Features

- Text messaging with OpenCode
- Media support (images, audio, documents, video)
- Slash commands for agent/model/provider configuration
- Per-user session management
- Structured logging with 30-day retention

## Documentation

See `docs/features/` for deeper documentation.
