# Overview

Improves the default markdown templates created when running `opencode-telegram new`. The templates provide better prompts and instructions for the OpenCode agent.

# Details

- Updates SOUL.md: Core behavior - helpful, friendly, concise but not boring, warm/caring
- Updates IDENTITY.md: Personal assistant role, friendly/helpful, uses emoji but not excessively
- Updates USER.md: Empty template for user information storage
- Updates TOOLS.md: Empty placeholder for future tool instructions
- Updates BOOTSTRAP.md: Greets warmly with emojis, asks setup questions, deletes self after execution
- Removes detailed AGENTS.md (replaced by opencode.json)
- Simplifies tools documentation

# File Paths

- internal/workspace/files.go
