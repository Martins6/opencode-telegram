#!/bin/bash
set -euo pipefail

REPO_OWNER="martins6"
REPO_NAME="opencode-telegram"
BINARY_NAME="opencode-telegram"
INSTALL_PREFIX="${INSTALL_PREFIX:-$HOME/.local/bin}"

VERSION=""
FORCE_VERSION=""
PREFIX_SET_BY_USER=""

usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Installs the opencode-telegram binary from GitHub releases.

OPTIONS:
    -v, --version VERSION    Install specific version
    -p, --prefix PREFIX      Installation directory (default: ~/.local/bin)
    -h, --help               Show this help message

EXAMPLES:
    $0                                    # Install latest version
    curl -sSL https://raw.githubusercontent.com/martins6/opencode-telegram/main/install.sh | bash
    $0 -v v1.0.0                          # Install specific version
    $0 -p /usr/local/bin                  # Install to /usr/local/bin

EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                FORCE_VERSION="$2"
                shift 2
                ;;
            -p|--prefix)
                PREFIX_SET_BY_USER="1"
                INSTALL_PREFIX="$2"
                shift 2
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

detect_os() {
    local os
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux*)
            echo "linux"
            ;;
        darwin*)
            echo "darwin"
            ;;
        *)
            echo "Unsupported OS: $os" >&2
            exit 1
            ;;
    esac
}

detect_arch() {
    local arch
    arch="$(uname -m | tr '[:upper:]' '[:lower:]')"
    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            echo "Unsupported architecture: $arch" >&2
            exit 1
            ;;
    esac
}

get_latest_version() {
    local version
    version=$(curl -sSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4)
    if [[ -z "$version" ]]; then
        echo "Error: Could not fetch latest release version" >&2
        exit 1
    fi
    echo "$version"
}

get_asset_url() {
    local os="$1"
    local arch="$2"
    local version="$3"
    local asset_name="${BINARY_NAME}_${os}_${arch}.tar.gz"
    
    local url
    url=$(curl -sSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/tags/${version}" | \
        grep -o "\"browser_download_url\": *\"[^\"]*${asset_name}[^\"]*\"" | \
        head -n1 | \
        cut -d'"' -f4)
    
    if [[ -z "$url" ]]; then
        echo "Error: Could not find release asset for ${os}/${arch} version ${version}" >&2
        echo "Expected asset: ${asset_name}" >&2
        exit 1
    fi
    
    echo "$url"
}

install_binary() {
    local url="$1"
    local install_dir="$2"
    local tmp_dir
    tmp_dir=$(mktemp -d)
    
    trap "rm -rf $tmp_dir" EXIT
    
    echo "Downloading $url..."
    curl -sSL "$url" -o "${tmp_dir}/binary.tar.gz"
    
    echo "Extracting..."
    tar -xzf "${tmp_dir}/binary.tar.gz" -C "$tmp_dir"
    
    mkdir -p "$install_dir"
    
    local binary_path="${tmp_dir}/${BINARY_NAME}"
    if [[ ! -f "$binary_path" ]]; then
        binary_path=$(find "$tmp_dir" -name "$BINARY_NAME" -type f | head -n1)
    fi
    
    if [[ -z "$binary_path" || ! -f "$binary_path" ]]; then
        echo "Error: Binary not found in archive" >&2
        exit 1
    fi
    
    cp "$binary_path" "${install_dir}/${BINARY_NAME}"
    chmod +x "${install_dir}/${BINARY_NAME}"
    
    echo "Installed to ${install_dir}/${BINARY_NAME}"
}

check_in_path() {
    local binary_path="$1"
    local binary_name="$2"
    
    if command -v "$binary_name" &>/dev/null; then
        local current_path
        current_path=$(command -v "$binary_name")
        if [[ "$current_path" != "$binary_path" ]]; then
            echo ""
            echo "Warning: A different version of $binary_name is already installed at: $current_path"
            echo "The new version is installed at: $binary_path"
            echo "You may want to remove the old version or update your PATH."
        fi
    fi
}

main() {
    parse_args "$@"
    
    local os
    local arch
    os=$(detect_os)
    arch=$(detect_arch)
    
    if [[ -z "$VERSION" ]]; then
        echo "Fetching latest version..."
        VERSION=$(get_latest_version)
    fi
    
    echo "Installing ${BINARY_NAME} ${VERSION} for ${os}/${arch}..."
    
    local asset_url
    asset_url=$(get_asset_url "$os" "$arch" "$VERSION")
    
    install_binary "$asset_url" "$INSTALL_PREFIX"
    
    check_in_path "${INSTALL_PREFIX}/${BINARY_NAME}" "$BINARY_NAME"
    
    echo ""
    echo "Installation complete!"
    echo ""
    
    if [[ -z "$PREFIX_SET_BY_USER" && "$INSTALL_PREFIX" == "$HOME/.local/bin" ]]; then
        if [[ ! ":$PATH:" == *":$HOME/.local/bin:"* ]]; then
            echo "IMPORTANT: Add ~/.local/bin to your PATH if not already present:"
            echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
            echo ""
        fi
    fi
    
    echo "To get started, configure your bot token:"
    echo "  ${BINARY_NAME} config set bot.token \"YOUR_BOT_TOKEN\""
    echo ""
    echo "Then start the bot:"
    echo "  ${BINARY_NAME} start"
}

main "$@"
