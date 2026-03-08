# Overview

Improves the default markdown templates created when running `opencode-telegram new`. The templates provide better prompts and instructions for the OpenCode agent.

# Details

- Updates SOUL.md: Core behavior - helpful, friendly, concise but not boring, warm/caring
- Updates IDENTITY.md: Personal assistant role, friendly/helpful, uses emoji but not excessively
- Updates USER.md: Empty template for user information storage
- Updates TOOLS.md: Empty placeholder for future tool instructions
- Updates BOOTSTRAP.md: Greets warmly with emojis, asks setup questions, deletes self after execution
- Adds AGENTS.md: References personality files using @{FILE} syntax, documents workspace structure
- Removes detailed AGENTS.md (replaced by opencode.json) - later added back as lightweight reference
- Simplifies tools documentation
- Reorganizes prompt files into MAIN-PROMPTS/ subdirectory for better structure
- Updates opencode.json to use native `instructions` field with glob pattern `MAIN-PROMPTS/*.md`
- Updates AGENTS.md to reference files in MAIN-PROMPTS/ directory

# File Paths

- internal/workspace/files.go
- internal/workspace/template.go
