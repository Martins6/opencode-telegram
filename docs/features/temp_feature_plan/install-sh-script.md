# Install Script Implementation Plan

## Problem

Create an `install.sh` script that allows users to install the `opencode-telegram` binary via a one-liner curl command from GitHub releases.

## Solution Overview

Create a robust shell script that:
1. Detects the user's OS and architecture
2. Fetches the latest release from GitHub API
3. Downloads the appropriate pre-built binary
4. Installs it to a suitable location (user's PATH)
5. Provides helpful messages for configuration

## Steps

### Step 1: Create the install.sh Script
Create a new `install.sh` file at the repository root with:
- OS detection (Linux, macOS)
- Architecture detection (amd64, arm64)
- GitHub API integration to fetch latest release
- Binary download and extraction logic
- Installation to `$HOME/.local/bin` or `/usr/local/bin`
- Version checking and upgrade capability
- Proper error handling

### Step 2: Define Release Asset Naming Convention
Establish consistent naming for release assets:
- Format: `opencode-telegram_{os}_{arch}.tar.gz`

### Step 3: Add GitHub Actions Workflow for Building Releases
Create `.github/workflows/release.yml` that:
- Builds binaries for all 4 platforms on tag push
- Creates GitHub releases with proper assets

### Step 4: Update README with Installation Instructions
Add installation section to README showing the curl command

## Verification

1. Test on Linux (amd64): `curl -sSL https://raw.githubusercontent.com/opencode-telegram/agent/main/install.sh | bash`
2. Test on macOS (arm64): `curl -sSL https://raw.githubusercontent.com/opencode-telegram/agent/main/install.sh | bash`
3. Verify binary works: `opencode-telegram --help`
4. Test version flag: `curl -sSL https://raw.githubusercontent.com/opencode-telegram/agent/main/install.sh | bash -s -- --version v1.0.0`
5. Verify installation path: `which opencode-telegram`
