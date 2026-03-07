# OpenCode Telegram Agent

A Telegram bot that acts as a gateway to an OpenCode server, allowing users to interact with the OpenCode AI agent directly from Telegram. It functions a bit like OpenClaw.

## Installation

### Prerequisites

**Both OpenCode CLI AND this binary are required to run the bot.**

- [OpenCode CLI](https://github.com/anomalyco/opencode) installed (`opencode` command available)

### Quick Install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/martins6/opencode-telegram/main/install.sh | bash
```

This will install the binary to `~/.local/bin`. Make sure to add it to your PATH:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Or install system-wide with sudo:

```bash
curl -sSL https://raw.githubusercontent.com/martins6/opencode-telegram/main/install.sh | bash -s -- -p /usr/local/bin
```

Install a specific version:

```bash
curl -sSL https://raw.githubusercontent.com/martins6/opencode-telegram/main/install.sh | bash -s -- -v v0.1.0
```

### Development Installation (local_install.sh)

For developers who want to build from source, use `local_install.sh`:

```bash
./local_install.sh
```

**Requirements:**

- Go 1.21+ installed
- OpenCode CLI installed (`opencode` command available)

This script builds the binary from local source code instead of downloading from GitHub releases.

## Quick Start

### 1. Create a Telegram Bot

1. Open Telegram and search for @BotFather
2. Send `/newbot` to create a new bot
3. Follow the instructions and get your bot token

### 2. Initialize Workspace

```bash
opencode-telegram new
```

This creates a workspace at `~/.opencode-telegram/` with template files.

### 3. Configure the Bot

Use the `opencode auth login`, choose your provider. Then also set it for `opencode-telegram`. Please, check [models.dev](https://models.dev/) for the standard reference of names for providers and models used in OpenCode. You can also inspect the logs for any mismatch of the naming.

To set the specific settings for your bot, you can use the following commands:

```bash
# Set your bot token
opencode-telegram config set bot.token "YOUR_BOT_TOKEN"

# Add your Telegram user ID to allowed users
opencode-telegram config set bot.allowed_users [123456789, "YourUserName"]

# Optional: Set default agent
opencode-telegram config set defaults.agent "telegram-agent"

# Optional: Set default model
opencode-telegram config set defaults.model "MiniMax-M2.5"

# Optional: Set default provider
opencode-telegram config set defaults.provider "minimax-coding-plan"

# View current configuration
opencode-telegram config list
```

### 4. Start the Bot

```bash
opencode-telegram start
```

This will:

1. Initialize the Telegram bot
2. Begin handling messages

Press `Ctrl+C` to stop gracefully.

## Configuration

The configuration file is stored at `~/.opencode-telegram/config.toml`:

```toml
[bot]
token = "YOUR_BOT_TOKEN"
allowed_users = [123456789]

[workspace]
path = "~/.opencode-telegram/"

[defaults]
agent = "telegram-agent"
model = "MiniMax-M2.5"
provider = "minimax-coding-plan"
```

### Configuration Commands

| Command                                      | Description            |
| -------------------------------------------- | ---------------------- |
| `opencode-telegram config set <key> <value>` | Set a config value     |
| `opencode-telegram config get <key>`         | Get a config value     |
| `opencode-telegram config list`              | List all config values |

### Config Keys

| Key                 | Description                       |
| ------------------- | --------------------------------- |
| `bot.token`         | Telegram bot token                |
| `bot.allowed_users` | List of allowed Telegram user IDs |
| `workspace.path`    | Workspace directory path          |
| `defaults.agent`    | Default agent name                |
| `defaults.model`    | Default LLM model                 |
| `defaults.provider` | Default LLM provider              |

## Slash Commands

| Command                | Description        | Example                        |
| ---------------------- | ------------------ | ------------------------------ |
| `/start`               | Start the bot      | `/start`                       |
| `/help`                | Show help message  | `/help`                        |
| `/set-agent <name>`    | Set active agent   | `/set-agent coder`             |
| `/set-model <name>`    | Set LLM model      | `/set-model claude-sonnet-4-5` |
| `/set-provider <name>` | Set LLM provider   | `/set-provider anthropic`      |
| `/workspace <path>`    | Set workspace path | `/workspace /path/to/folder`   |

## Workspace Structure

```
~/.opencode-telegram/
├── AGENTS.md           # Agent definitions
├── SOUL.md            # System operator behavior
├── USER.md            # User information
├── IDENTITY.md        # Model identity
├── BOOTSTRAP.md       # First-time setup
├── TOOLS.md           # Tool definitions
├── downloads/
│   ├── images/        # Downloaded images
│   ├── audio/        # Downloaded audio
│   ├── documents/    # Downloaded documents
│   └── videos/       # Downloaded videos
├── conversations/    # Per-user conversation history
└── .logs/
    └── YYYY-MM-DD.log # Daily log files
```

## Logging

Logs are stored in `~/.opencode-telegram/.logs/` with format `YYYY-MM-DD.log`.

### View Logs

```bash
# View today's logs
opencode-telegram logs today

# View specific date
opencode-telegram logs 2026-03-05
```

Log format:

```
[INPUT]  [2026-03-05 10:30:45] User 123456789: "Hello"
[OUTPUT] [2026-03-05 10:30:46] Response text...
```

Log retention is 30 days (auto-cleanup).

## Media Handling

When you send media to the bot:

1. The bot downloads the file from Telegram
2. Saves it to `workspace/downloads/{type}/`
3. Constructs a prompt with the file path
4. Sends to OpenCode server with context

### Supported Media Types

| Telegram Type | Workspace Folder       |
| ------------- | ---------------------- |
| Photo         | `downloads/images/`    |
| Audio         | `downloads/audio/`     |
| Voice         | `downloads/audio/`     |
| Document      | `downloads/documents/` |
| Video         | `downloads/videos/`    |

## CLI Commands

| Command                         | Description                 |
| ------------------------------- | --------------------------- |
| `opencode-telegram start`       | Start bot + OpenCode server |
| `opencode-telegram config`      | Configuration management    |
| `opencode-telegram new [path]`  | Initialize new workspace    |
| `opencode-telegram logs [date]` | View logs                   |

## Development

### Project Structure

```
.
├── cmd/                    # CLI commands
│   ├── root.go
│   ├── start.go
│   ├── config.go
│   ├── new.go
│   └── logs.go
├── internal/
│   ├── bot/               # Telegram bot
│   ├── config/            # Configuration
│   ├── logger/            # Logging
│   ├── media/             # Media handling
│   ├── opencode/          # OpenCode integration
│   ├── session/           # Session management
│   └── workspace/         # Workspace templates
├── main.go
└── go.mod
```

### Dependencies

- `github.com/go-telegram/bot` - Telegram Bot API
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration
- `github.com/google/uuid` - UUID generation

## License

MIT
