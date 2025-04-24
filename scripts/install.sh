#!/usr/bin/env bash

set -euo pipefail

REPO="squadbase/squadbase"
APP="squad"

: "${REPO_URL:=https://github.com/${REPO}/releases/download}"
: "${VERSION:=}"
: "${BINDIR:=/usr/local/bin}"

# os / arch
detect_os()   { uname -s | tr '[:upper:]' '[:lower:]'; }
detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)   echo "amd64"  ;;
    arm64|aarch64)  echo "arm64"  ;;
    armv7*)         echo "armv7"  ;;
    *) echo "unsupported arch: $(uname -m)"; exit 1 ;;
  esac
}
OS="$(detect_os)"; ARCH="$(detect_arch)"

# version
if [[ -z "$VERSION" || "$VERSION" == "latest" ]]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
              | grep -m1 '"tag_name":' | cut -d '"' -f4)"
  [[ -z "$VERSION" ]] && { echo "No release found"; exit 1; }
fi

# url
FILE="${APP}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="${REPO_URL}/${VERSION}/${FILE}"
SUM_URL="${REPO_URL}/${VERSION}/checksums.txt"

# download & checksum
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "➜ Downloading $FILE …"
curl -fsSL "$URL" -o "$TMPDIR/$FILE"

echo "➜ Verifying checksum …"
curl -fsSL "$SUM_URL" -o "$TMPDIR/checksums.txt"

checksum() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

EXPECTED="$(grep " $FILE" "$TMPDIR/checksums.txt" | awk '{print $1}')"
ACTUAL="$(checksum "$TMPDIR/$FILE")"
[[ "$EXPECTED" == "$ACTUAL" ]] || { echo "❌  checksum mismatch"; exit 1; }

# install
echo "➜ Extracting"
tar -C "$TMPDIR" -xzf "$TMPDIR/$FILE"

echo "➜ Installing to $BINDIR"
install -m 755 "$TMPDIR/$APP" "$BINDIR/$APP"

echo "✅  Installed $APP $VERSION → $BINDIR/$APP"
echo "ℹ︎  Run  '$APP --help'  to get started."
