# Overview

Shell script that enables one-line installation of the opencode-telegram binary via curl from GitHub releases.

# Details

- Detects OS (Linux, macOS) and architecture (amd64, arm64)
- Fetches latest release from GitHub API
- Downloads and extracts appropriate binary
- Installs to user PATH ($HOME/.local/bin or /usr/local/bin)
- Supports version specification for downgrades
- Provides helpful post-install messages

# File Paths

- install.sh - Main installation script
- local_install.sh - Local installation variant
- .github/workflows/release.yml - GitHub Actions workflow for building releases
