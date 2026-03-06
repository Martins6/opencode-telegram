# Overview

Switched from using `opencode serve` (HTTP server on port 4096) to using `opencode run` command-line execution. This resolves minimax-coding-plan provider configuration issues and provides a simpler integration.

# Details

- Executes `opencode run` commands directly using exec.Command
- Parses NDJSON output to extract session IDs and text responses
- Captures session ID from first response event
- Uses --continue flag for multi-turn conversations
- Adds /new-session command to start fresh conversations
- Changes default agent from coder to build
- Uses default model MiniMax-M2.5 with minimax-coding-plan provider
- Removes server startup from bot initialization
- Session state tracked via IsNewSession flag

# File Paths

- internal/opencode/runner.go
- internal/session/types.go
- internal/session/manager.go
- internal/bot/commands.go
- internal/bot/handlers.go
- cmd/start.go
