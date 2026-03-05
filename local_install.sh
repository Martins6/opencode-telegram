#!/bin/bash

# OpenCode Telegram Agent Local Installation Script
# Builds from local source code instead of cloning from remote

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_info() { echo -e "${YELLOW}ℹ${NC} $1"; }
print_cmd() { echo -e "${BLUE}$1${NC}"; }

get_install_dir() {
    LOCAL_DIR="$HOME/.local/bin"
    GO_BIN="$(go env GOPATH 2>/dev/null)/bin" || "$HOME/go/bin"
    
    if [ -d "$LOCAL_DIR" ] || mkdir -p "$LOCAL_DIR" 2>/dev/null; then
        echo "$LOCAL_DIR"
        return 0
    fi
    
    if [ -d "$GO_BIN" ]; then
        echo "$GO_BIN"
        return 0
    fi
    
    if mkdir -p "$GO_BIN" 2>/dev/null; then
        echo "$GO_BIN"
        return 0
    fi
    
    if [ -w /usr/local/bin ]; then
        echo "/usr/local/bin"
        return 0
    fi
    
    return 1
}

main() {
    print_info "Installing OpenCode Telegram Agent from local source"
    print_info "====================================================="
    echo ""

    if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
        print_error "Not in a valid OpenCode Telegram Agent source directory"
        print_info "Run this script from the opencode-telegram source root"
        exit 1
    fi

    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        print_info "Install Go from: https://golang.org/doc/install"
        exit 1
    fi

    TARGET_DIR=$(get_install_dir)
    if [ $? -ne 0 ]; then
        print_error "Could not find a suitable installation directory."
        exit 1
    fi

    print_info "Building from local source..."
    go build -ldflags "-s -w" -o "opencode-telegram" .
    print_success "Built opencode-telegram"

    if command -v codesign &> /dev/null; then
        codesign --force --sign - --deep "opencode-telegram" 2>/dev/null || true
        if codesign -v "opencode-telegram" 2>/dev/null; then
            print_success "Code signing verified"
        else
            print_error "Code signing verification failed (binary may still work)"
        fi
    fi

    print_info "Installing to $TARGET_DIR..."
    rm -f "$TARGET_DIR/opencode-telegram"
    cp -f "opencode-telegram" "$TARGET_DIR/"
    chmod +x "$TARGET_DIR/opencode-telegram"
    
    if command -v codesign &> /dev/null; then
        codesign --force --sign - --deep "$TARGET_DIR/opencode-telegram" 2>/dev/null || true
    fi
    
    print_info "Verifying binary execution..."
    if "$TARGET_DIR/opencode-telegram" --help >/dev/null 2>&1; then
        print_success "Binary executes correctly"
    else
        print_error "Binary verification failed - it may have been killed by macOS"
        print_info "Try running: codesign --force --sign - --deep $TARGET_DIR/opencode-telegram"
    fi
    
    print_success "Installed to $TARGET_DIR/opencode-telegram"

    rm -f "opencode-telegram"

    if [ -x "$TARGET_DIR/opencode-telegram" ]; then
        print_success "Installation complete!"
        echo ""
        echo "Quick start:"
        print_cmd "  opencode-telegram new         # Initialize workspace"
        print_cmd "  opencode-telegram config set bot.token \"YOUR_TOKEN\"  # Set bot token"
        print_cmd "  opencode-telegram start       # Start the bot"
        echo ""
        
        if [[ ":$PATH:" != *":$TARGET_DIR:"* ]]; then
            print_info "Add $TARGET_DIR to your PATH:"
            print_cmd "  export PATH=\"\$PATH:$TARGET_DIR\""
            echo ""
            print_info "Add this to your ~/.bashrc or ~/.zshrc to make it permanent."
        fi
    else
        print_error "Installation verification failed"
        exit 1
    fi
}

main "$@"
