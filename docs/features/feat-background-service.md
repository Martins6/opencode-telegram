# Overview

Adds a Hermes-style singleton per-user background service so a single Acolyte instance owns one workspace per OS user, with public lifecycle commands wrapping a hidden `acolyte __daemon` worker invoked by systemd (Linux) or launchd (macOS). Adds session inspection wrappers, tail-style log inspection, strict workspace validation, persisted workspace path under `~/.acolyte/config.toml`, atomic runtime state, and service-aware self-update.

# Details

- Linux uses systemd user service (`Type=exec`, `Restart=on-failure`, `WantedBy=default.target`); macOS uses a LaunchAgent plist (`RunAtLoad`, crash-only `KeepAlive`).
- `acolyte start [--workspace PATH]` validates and persists the workspace, installs/enables/starts the service, and waits up to 5s for runtime readiness.
- `acolyte stop [--forever]` stops the service; with `--forever` it also disables autostart.
- `acolyte restart` restarts the running service and waits for readiness.
- `acolyte status` prints running/stopped, workspace, autostart, PID; exits 3 when stopped.
- `acolyte session list` runs `opencode session list` inside the workspace; `acolyte session <id>` runs `opencode export <id>`.
- `acolyte logs [N] [--date today|YYYY-MM-DD]` returns the most recent N entries (default 10). Multi-line messages are counted as a single entry. Non-daily files in `.logs/` are ignored.
- Workspace validation is strict: missing template files/directories are listed in a single error message with a `acolyte new <path>` hint.
- The runtime state (`~/.acolyte/.service/state.json`) is written atomically; PID liveness is checked via `kill(pid, 0)`.
- Self-update now restarts via the service manager instead of re-execing inside the same process.
- Boot/install is fatal if `systemctl --user` / `launchctl` are unavailable; clear `ErrManagerUnavailable` and `ErrUnsupportedPlatform` errors are exposed.

# File Paths

- `cmd/start.go`
- `cmd/stop.go` / `cmd/service.go`
- `cmd/daemon.go`
- `cmd/session.go`
- `cmd/logs.go`
- `cmd/update.go`
- `internal/config/workspace.go`
- `internal/config/config.go`
- `internal/workspace/template.go`
- `internal/runtime/runtime.go`
- `internal/daemon/runtime.go`
- `internal/service/manager.go`
- `internal/service/systemd.go`
- `internal/service/launchd.go`
- `internal/service/platform_linux.go`
- `internal/service/platform_darwin.go`
- `internal/service/platform_other.go`
- `internal/opencode/session.go`
- `internal/logger/tail.go`
- `internal/scheduler/service.go`
- `internal/bot/notifier.go`
