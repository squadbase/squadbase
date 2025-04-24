#!/bin/sh
set -e

GITHUB_REPO="squadbase/squadbase"
INSTALL_DIR="${HOME}/bin"
BINARY_NAME="squad"

# Detect system information
detect_os() {
  case "$(uname -s)" in
    Darwin)
      echo "darwin"
      ;;
    Linux)
      echo "linux"
      ;;
    CYGWIN*|MINGW*|MSYS*)
      echo "windows"
      ;;
    *)
      echo "Unsupported system: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)
      echo "x86_64"
      ;;
    arm64|aarch64)
      echo "arm64"
      ;;
    *)
      echo "Unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

# Get latest release
get_latest_release() {
  curl --silent "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | 
  grep '"tag_name":' | 
  sed -E 's/.*"([^"]+)".*/\1/'
}

# Main process
main() {
  OS=$(detect_os)
  ARCH=$(detect_arch)
  VERSION=$(get_latest_release)

  # Display version
  echo "Installing version: ${VERSION}"

  # Create installation directory
  mkdir -p "${INSTALL_DIR}"

  # Set extension and archive type for Windows
  if [ "$OS" = "windows" ]; then
    ARCHIVE_EXT="zip"
    BINARY_EXT=".exe"
  else
    ARCHIVE_EXT="tar.gz"
    BINARY_EXT=""
  fi

  # Build archive name (following goreleaser naming convention)
  ARCHIVE_NAME="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}${BINARY_EXT}"
  DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}.${ARCHIVE_EXT}"

  echo "Downloading: ${DOWNLOAD_URL}"
  
  # Create temporary directory
  TMP_DIR=$(mktemp -d)
  TMP_FILE="${TMP_DIR}/${ARCHIVE_NAME}.${ARCHIVE_EXT}"
  
  # Download
  curl -L -o "${TMP_FILE}" "${DOWNLOAD_URL}"
  
  # Extract
  if [ "$OS" = "windows" ]; then
    unzip -o "${TMP_FILE}" -d "${TMP_DIR}"
  else
    tar -xzf "${TMP_FILE}" -C "${TMP_DIR}"
  fi
  
  # Install binary
  BINARY_PATH="${TMP_DIR}/${BINARY_NAME}${BINARY_EXT}"
  if [ -f "${BINARY_PATH}" ]; then
    install -m 755 "${BINARY_PATH}" "${INSTALL_DIR}/${BINARY_NAME}${BINARY_EXT}"
    echo "Installed ${BINARY_NAME} to ${INSTALL_DIR}"
  else
    echo "Binary not found" >&2
    exit 1
  fi
  
  # Clean up temporary files
  rm -rf "${TMP_DIR}"
  
  # Check if installation directory is in PATH
  if ! echo "$PATH" | grep -q "${INSTALL_DIR}"; then
    echo ""
    echo "NOTE: ${INSTALL_DIR} is not in your PATH"
    echo "Add the following line to your .profile, .bash_profile, or .zshrc:"
    echo "export PATH=\"\$HOME/bin:\$PATH\""
    echo "Then restart your terminal or run: source ~/.$(basename "$SHELL")rc"
  fi
  
  echo ""
  echo "Installation complete! Verify with '${BINARY_NAME} --version'"
}

main