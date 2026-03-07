# Overview

Creates a custom "telegram-agent" in opencode.json and updates the default agent configuration. This replaces the previous AGENTS.md-based approach with a more structured OpenCode configuration.

# Details

- Adds OpenCodeConfigContent constant for opencode.json with telegram-agent definition
- Updates template.go to create opencode.json instead of AGENTS.md
- Changes default agent from "coder" to "telegram-agent" in config
- Defines telegram-agent with prompt instructions to read SOUL.md, IDENTITY.md, TOOLS.md, USER.md, BOOTSTRAP.md
- Includes timestamp awareness in agent prompt
- Agent prompt instructs to execute BOOTSTRAP.md then delete it

# File Paths

- internal/workspace/files.go
- internal/workspace/template.go
- internal/config/config.go
