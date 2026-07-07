# Overview

Self-update mechanism that lets the Acolyte binary detect newer GitHub releases, verify their integrity, and replace itself in place from the CLI. Ships `acolyte update` (replace, optionally restart) and `acolyte version`, plus a non-blocking startup check on `acolyte start` that surfaces an "outdated" warning through the existing notification pipeline.

# Details

- Embeds a build-time `var Version = "dev"` in `main.go`; the release workflow sets it via `-ldflags="-X main.Version=<tag>"`
- `acolyte version` prints the compiled `main.Version` (or `dev` for local builds)
- New `internal/updater` package (no third-party deps):
  - Talks to `GET https://api.github.com/repos/martins6/acolyte/releases/latest`
  - Maps `runtime.GOOS/GOARCH` to release assets named `acolyte_<os>_<arch>.tar.gz`
  - Computes `IsOutdated` by comparing the latest `tag_name` against `Version`, treating `dev` as "skip" so local builds are never flagged
  - `Apply` downloads the tarball and the published `checksums.txt` into `os.TempDir()`, verifies SHA256, then `os.Rename`s the new binary over the running path
  - `--restart` mode re-execs the new process with `exec.Command(exe, args...).Start()` followed by `os.Exit(0)` after the rename
  - Re-applies `codesign --force --sign - --deep` ad-hoc signing on macOS when `codesign` is on PATH (mirrors `local_install.sh`)
- New `acolyte update` cobra command (`cmd/update.go`) with flags:
  - `--check` only reports current vs. latest and exits
  - `--version vX.Y.Z` pins a specific release instead of `latest`
  - `--restart` (default `true`) re-execs after replace; `--restart=false` leaves the manual restart to the user
  - `--yes` skips the interactive "do you want to update?" prompt for non-interactive use
  - Clear permission-failure hint: reinstall via `install.sh` to `~/.local/bin` or run with `sudo`
- `acolyte start` spawns a goroutine after `bot.Initialize` succeeds that calls `updater.IsOutdated` with a 5-second `context.WithTimeout`; on a newer release it inserts a row via `database.InsertNotification` and logs to stdout. Errors are swallowed and never block startup.
- `.github/workflows/release.yml` generates a `dist/checksums.txt` (`sha256sum`-compatible) alongside the existing `dist/*.tar.gz` glob and uploads it via `softprops/action-gh-release@v2`
- README CLI table gains `update` and `version`; a short "Updating" subsection documents the flow and the checksum verification

# File Paths

- main.go
- cmd/root.go
- cmd/update.go
- cmd/version.go
- cmd/start.go
- internal/updater/updater.go
- internal/updater/updater_test.go
- .github/workflows/release.yml
- README.md
