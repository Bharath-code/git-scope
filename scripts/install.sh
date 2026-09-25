#!/bin/sh
# Git-Scope Installer Script
# Usage: curl -sSL https://raw.githubusercontent.com/Bharath-code/git-scope/main/scripts/install.sh | sh
#
# This script detects your OS and architecture, downloads the appropriate
# git-scope binary, and installs it to ~/.local/bin

set -e

REPO="Bharath-code/git-scope"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="git-scope"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    printf "${GREEN}[INFO]${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)   echo "linux" ;;
        Darwin*)  echo "darwin" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *)        error "Unsupported OS: $(uname -s)" ;;
    esac
}

# Detect Architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)  echo "arm64" ;;
        armv7l)         echo "arm" ;;
        i386|i686)      echo "386" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get the latest release version from GitHub API
get_latest_version() {
    curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

main() {
    info "Git-Scope Installer"
    info "==================="

    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)

    if [ -z "$VERSION" ]; then
        error "Failed to get latest version. Check your internet connection."
    fi

    info "Detected: ${OS}/${ARCH}"
    info "Latest version: ${VERSION}"

    # Construct download URL
    # Assumes release assets are named like: git-scope_1.2.0_linux_amd64.tar.gz
    ARCHIVE_NAME="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

    info "Downloading from: ${DOWNLOAD_URL}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf ${TMP_DIR}" EXIT

    # Download and extract
    if ! curl -fsSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/archive.tar.gz"; then
        error "Failed to download. Check if the release exists for your platform."
    fi

    CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
    if ! curl -fsSL "${CHECKSUMS_URL}" -o "${TMP_DIR}/checksums.txt"; then
        error "Failed to download checksums.txt; refusing to install an unverified binary."
    fi
    EXPECTED=$(grep " ${ARCHIVE_NAME}\$" "${TMP_DIR}/checksums.txt" | awk '{print $1}')
    if [ -z "$EXPECTED" ]; then
        error "No checksum listed for ${ARCHIVE_NAME}."
    fi
    if command -v sha256sum >/dev/null 2>&1; then
        ACTUAL=$(sha256sum "${TMP_DIR}/archive.tar.gz" | awk '{print $1}')
    else
        ACTUAL=$(shasum -a 256 "${TMP_DIR}/archive.tar.gz" | awk '{print $1}')
    fi
    if [ "$EXPECTED" != "$ACTUAL" ]; then
        error "Checksum mismatch for ${ARCHIVE_NAME} (expected ${EXPECTED}, got ${ACTUAL})."
    fi
    info "Checksum verified."

    tar xzf "${TMP_DIR}/archive.tar.gz" -C "${TMP_DIR}"

    # Create install directory if it doesn't exist
    mkdir -p "${INSTALL_DIR}"

    # Move binary
    if [ -f "${TMP_DIR}/${BINARY_NAME}" ]; then
        mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        # Sometimes the binary is in a subdirectory
        FOUND_BINARY=$(find "${TMP_DIR}" -name "${BINARY_NAME}" -type f | head -n 1)
        if [ -n "$FOUND_BINARY" ]; then
            mv "$FOUND_BINARY" "${INSTALL_DIR}/${BINARY_NAME}"
        else
            error "Binary not found in archive. Please report this issue."
        fi
    fi

    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    info "Installed to: ${INSTALL_DIR}/${BINARY_NAME}"

    # Check if install dir is in PATH
    if ! echo "$PATH" | grep -q "${INSTALL_DIR}"; then
        warn "~/.local/bin is not in your PATH."
        warn "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        printf "\n  export PATH=\"\$HOME/.local/bin:\$PATH\"\n\n"
    fi

    info "Installation complete! Run 'git-scope' to get started."
    printf "\n"
    ${INSTALL_DIR}/${BINARY_NAME} --version 2>/dev/null || true
}

main
