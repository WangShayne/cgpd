#!/bin/sh
set -e

REPO="WangShayne/cgpd"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="cgpd"

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) echo "unknown" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) echo "unknown" ;;
    esac
}

main() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        echo "Error: Unsupported OS ($OS) or architecture ($ARCH)"
        exit 1
    fi

    if [ "$OS" = "windows" ]; then
        FILENAME="${BINARY_NAME}-${OS}-${ARCH}.exe"
    else
        FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${FILENAME}"

    echo "Detected: ${OS}/${ARCH}"
    echo "Downloading ${FILENAME}..."

    TEMP_FILE=$(mktemp)
    trap 'rm -f "$TEMP_FILE"' EXIT

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$TEMP_FILE" "$DOWNLOAD_URL"
    else
        echo "Error: curl or wget required"
        exit 1
    fi

    chmod +x "$TEMP_FILE"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TEMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        trap - EXIT
    else
        echo "Installing to ${INSTALL_DIR} (requires sudo)..."
        sudo mv "$TEMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        trap - EXIT
    fi

    echo "Installed: $(command -v $BINARY_NAME)"
    echo "Version: $($BINARY_NAME --version 2>/dev/null || echo 'installed')"
}

main
