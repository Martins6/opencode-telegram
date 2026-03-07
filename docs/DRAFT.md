# OpenCode Telegram Agent - Draft Document

## Overview

A Telegram bot that acts as a gateway to an OpenCode server, allowing users to interact with the OpenCode agent directly from Telegram. The bot will handle text, media (images, audio, files), and provide configuration via slash commands.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         TELEGRAM USER                               │
└─────────────────────────────────────────────────────────────────────┘
                                 │
                                 │ Telegram API (MTProto)
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    TELEGRAM BOT (Go + Cobra)                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │   Commands   │  │   Handlers   │  │  Media Store │              │
│  │  /set-agent  │  │  Text/Msg    │  │  Downloads   │              │
│  │  /set-model  │  │  Processing  │  │  to Workspace│              │
│  │  /set-provider│ │              │  │              │              │
│  │  /help       │  │              │  │              │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  │ HTTP/WebSocket
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    OPENCODE SERVER (Local)                         │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  │    Sessions      │  │     Agents       │  │      Tools       │  │     Skills       │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘  └──────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      WORKSPACE (Local FS)                          │
│  ├── AGENTS.md         (Agent definitions/project rules)           │
│  ├── SOUL.md           (System operator - LLM behavior)           │
│  ├── USER.md           (Critical user information)                │
│  ├── IDENTITY.md       (User's model identity)                    │
│  ├── BOOTSTRAP.md      (First-time setup walkthrough)             │
│  ├── TOOLS.md          (Tool definitions)                         │
│  ├── .opencode/        (OpenCode local config)                    │
│  │   ├── agents/       (Custom agent configs)                     │
│  │   └── skills/       (Custom skills)                             │
│  ├── .logs/            (Daily log files YYYY-MM-DD.log)          │
│  ├── downloads/         (Media files from Telegram)                │
│  │   ├── images/                                               │
│  │   ├── audio/                                                │
│  │   └── documents/                                            │
│  └── conversations/     (Chat history per user)                   │
└─────────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Telegram Bot (Go + Cobra CLI)

**CLI Structure (Cobra):**

```
opencode-telegram/
├── cmd/
│   ├── root.go           # Root command
│   ├── start.go          # Start bot + OpenCode server
│   ├── config.go         # Config management
│   │   └── set.go        # config set subcommand
│   ├── new.go            # Initialize new workspace from template
│   └── logs.go           # View logs command
├── internal/
│   ├── bot/
│   │   ├── handlers.go   # Message handlers
│   │   ├── commands.go   # Slash command handlers
│   │   └── middleware.go # Auth, rate limiting
│   ├── telegram/
│   │   ├── client.go     # Telegram Bot API client
│   │   └── types.go      # Type conversions
│   ├── opencode/
│   │   ├── client.go     # HTTP client to OpenCode server
│   │   └── types.go      # Request/response types
│   ├── media/
│   │   ├── downloader.go # Download images/audio/files
│   │   └── storage.go    # Workspace file management
│   ├── config/
│   │   └── config.go     # Configuration struct
│   └── logger/
│       └── logger.go     # Structured logging
└── main.go
```

**CLI Commands:**

| Command                                         | Description                                         |
| ----------------------------------------------- | --------------------------------------------------- |
| `opencode-telegram start`                       | Start bot + OpenCode server in workspace            |
| `opencode-telegram config set workspace <path>` | Set workspace path (default: ~/.opencode-telegram/) |
| `opencode-telegram new`                         | Initialize new workspace from GitHub template       |
| `opencode-telegram logs [today\|YYYY-MM-DD]`    | View logs from specific date                        |

**Start Command Output:**

```
[INPUT]  [2026-03-05 10:30:45] User 123456789: "Hello, help me with this code..."
[DEBUG]  [2026-03-05 10:30:45] POST /session/new - 200 OK
[OUTPUT] [2026-03-05 10:30:46] Here's your request.
```

**Log Storage:**

- Location: `<workspace>/.logs/`
- Format: `YYYY-MM-DD.log`
- Retention: 30 days (auto-cleanup)

**Start Command Behavior:**

1. Reads workspace path from config (default: `~/.opencode-telegram/`)
2. Changes working directory to workspace
3. Starts OpenCode server: `opencode serve --port 4096 --hostname 127.0.0.1`
4. Starts Telegram bot process
5. Outputs structured logs to console and file simultaneously

**Config File (~/.opencode-telegram/config.toml):**

```toml
[bot]
token = "TELEGRAM_BOT_TOKEN"
admin_ids = [123456789]  # Supports IDs and usernames: ["123456789", "username"]

[workspace]
path = "~/.opencode-telegram/"

[opencode]
port = "4096"  # Other OpenCode settings (API key, auth) must be configured via `opencode auth login`

[defaults]
agent = "coder"
model = "MiniMax2.5"
provider = "minimax"
```

**CLI Config Commands:**

| Command                                          | Description                                   |
| ------------------------------------------------ | --------------------------------------------- | ---------------------------- |
| `opencode-telegram start`                        | Start bot + OpenCode server in workspace      |
| `opencode-telegram config set bot.token`         | Set bot token                                 |
| `opencode-telegram config set bot.admin_ids`     | Set admin IDs (supports IDs and usernames)    |
| `opencode-telegram config set workspace.path`    | Set workspace path                            |
| `opencode-telegram config set opencode.port`     | Set OpenCode server port                      |
| `opencode-telegram config set defaults.agent`    | Set default agent                             |
| `opencode-telegram config set defaults.model`    | Set default model                             |
| `opencode-telegram config set defaults.provider` | Set default provider                          |
| `opencode-telegram config get <key>`             | Get config value                              |
| `opencode-telegram config list`                  | List all config values                        |
| `opencode-telegram new`                          | Initialize new workspace from GitHub template |
| `opencode-telegram logs [today                   | YYYY-MM-DD]`                                  | View logs from specific date |

**Telegram Bot Slash Commands:**

| Command         | Description        | Example                        |
| --------------- | ------------------ | ------------------------------ |
| `/set-agent`    | Set active agent   | `/set-agent planner`           |
| `/set-model`    | Set LLM model      | `/set-model claude-sonnet-4-5` |
| `/set-provider` | Set LLM provider   | `/set-provider anthropic`      |
| `/workspace`    | Set workspace path | `/workspace /path/to/folder`   |
| `/help`         | Show help          |                                |

### 2. Media Handling Flow

```
Telegram Message (with media)
         │
         ▼
┌─────────────────────┐
│  Detect Media Type  │
│  (Photo/Audio/Doc)  │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│  Download to temp   │
│  via Telegram API   │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│  Move to Workspace │
│  /downloads/{type}/ │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│  Construct prompt: │
│  "File X located    │
│  at workspace/      │
│  downloads/..."     │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│  Send to OpenCode   │
│  Server with context│
└─────────────────────┘
```

**Media Type Mapping:**
| Telegram Type | Workspace Folder | File Extension |
|---------------|------------------|----------------|
| Photo | `downloads/images/` | `.jpg`, `.png`, etc. |
| Audio | `downloads/audio/` | `.mp3`, `.ogg`, etc. |
| Voice | `downloads/audio/` | `.ogg` |
| Document | `downloads/documents/` | Original extension |
| Video | `downloads/videos/` | `.mp4`, etc. |

### 3. OpenCode Server Integration

**HTTP Endpoints (from OpenCode Server):**

```
POST /session/:id/message   # Send message, get response
GET  /global/health         # Health check
GET  /session               # List sessions
POST /session               # Create new session
GET  /agent                 # List available agents
GET  /config/providers      # List providers and models
```

**Request Format:**

```json
{
  "parts": [
    {
      "type": "text",
      "text": "Analyze the image in workspace/downloads/images/photo_123.jpg"
    }
  ],
  "agent": "coder",
  "model": "claude-3-5-sonnet",
  "provider": "anthropic"
}
```

**Response Format:**

```json
{
  "info": { "id": "msg_123", "role": "assistant", ... },
  "parts": [
    { "type": "text", "text": "Here's the analysis...", ... }
  ]
}
```

### 4. Workspace Structure (Template)

**AGENTS.md:**

```markdown
# Agents

## coder

You are an expert programmer...

## planner

You are a strategic planner...

## custom

Define your own agent here...
```

**SOUL.md:**

```markdown
# System Operator

You are a helpful AI assistant...

[Define the LLM's behavior, personality, response style, and core instructions here]
```

**USER.md:**

```markdown
# User Information

[Critical information about the user: name, preferences, timezone, important contexts]

## Preferences

- Language: en
- Response style: concise

## Context

- [Any important context about the user that the LLM should know]
```

**IDENTITY.md:**

```markdown
# Identity

[Define the model's identity/persona here]

## Persona

- Name: [Assistant Name]
- Role: [e.g., Senior Developer, Technical Writer]
- Background: [Personality traits, experience, expertise]

## Voice & Tone

- [How should the model communicate?]
```

**BOOTSTRAP.md:**

```markdown
# Bootstrap Setup

[First-time setup walkthrough - executed when user starts fresh]

## Welcome Questions

1. Who am I?
   - Ask for user's name
   - Ask for user's role/background

2. Who are you?
   - Introduce the AI assistant
   - Define the working relationship

3. What are we building?
   - Ask about the project context
   - Set up initial workspace

## Setup Steps

1. Greet user warmly
2. Ask identity questions (name, role, experience)
3. Configure model identity based on responses
4. Save to IDENTITY.md and USER.md
5. Confirm setup is complete
```

**TOOLS.md:**

```markdown
# Tools

## bash

Execute shell commands...

## read

Read files from workspace...

## write

Write files to workspace...
```

### 5. OpenCode Local Configuration

OpenCode supports local workspace-specific configuration via the `.opencode/` directory. This allows per-workspace customization of agents and skills.

**Directory Structure:**

```
.opencode/
├── agents/              # Custom agent definitions
│   └── <agent-name>/
│       ├── agent.json  # Agent configuration
│       └── instructions.md
└── skills/             # Custom skills
    └── <skill-name>/
        └── SKILL.md    # Skill definition with frontmatter
```

**Alternative Locations (also supported):**

- `.agents/skills/<name>/SKILL.md`
- `.claude/skills/<name>/SKILL.md`

**Global Fallback:**

- `~/.config/opencode/skills/` - Global skills
- `~/.agents/skills/` - Global agents/skills

**How Discovery Works:**

- OpenCode walks up from the current working directory to find `.opencode/`
- Loads matching `skills/*/SKILL.md` from `.opencode/`, `.agents/`, and `.claude/`
- Global definitions loaded from `~/.config/opencode/skills/`

**AGENTS.md (Simpler Alternative):**
For simpler project-specific rules, you can use `AGENTS.md` in the workspace root (similar to Cursor rules). This contains instructions included in the LLM's context.

## User Flow

```
┌──────────┐     ┌──────────────┐     ┌─────────────────┐
│  User    │────▶│  Telegram    │────▶│  Process Text   │
│  sends   │     │  Bot         │     │  or Media       │
│  message │     │              │     │                 │
└──────────┘     └──────────────┘     └────────────────┘
                                             │
                                             ▼
┌──────────┐     ┌──────────────┐     ┌─────────────────┐
│  User    │◀────│  Telegram    │◀────│  OpenCode       │
│  receives│     │  Bot         │     │  Server         │
│  response│     │              │     │                 │
└──────────┘     └──────────────┘     └─────────────────┘
```
