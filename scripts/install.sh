#!/usr/bin/env bash

set -euo pipefail

REPO="squadbase/squadbase"
APP="squad"

: "${REPO_URL:=https://github.com/${REPO}/releases/download}"
: "${VERSION:=}"
: "${BINDIR:=/usr/local/bin}"

detect_os()   { uname -s | tr '[:upper:]' '[:lower:]'; }
detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)   echo "amd64"  ;;
    arm64|aarch64)  echo "arm64"  ;;
    armv7*)         echo "armv7"  ;;
    *) echo "unsupported arch: $(uname -m)" ; exit 1 ;;
  esac
}
OS="$(detect_os)"; ARCH="$(detect_arch)"

if [[ -z "$VERSION" || "$VERSION" == "latest" ]]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
              | grep -m1 '"tag_name":' | cut -d '"' -f4)"
fi

FILE="${APP}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="${REPO_URL}/${VERSION}/${FILE}"
SUM_URL="${REPO_URL}/${VERSION}/checksums.txt"

TMPDIR="$(mktemp -d)"
cleanup() { rm -rf "$TMPDIR"; }
trap cleanup EXIT

echo "➜ Downloading $FILE …"
curl -fsSL "$URL" -o "$TMPDIR/$FILE"

echo "➜ Verifying checksum …"
curl -fsSL "$SUM_URL" -o "$TMPDIR/checksums.txt"

checksum_cmd() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1"
  else
    shasum -a 256 "$1"
  fi
}

EXPECTED="$(grep " $FILE" "$TMPDIR/checksums.txt" | awk '{print $1}')"
ACTUAL="$(checksum_cmd "$TMPDIR/$FILE" | awk '{print $1}')"
[[ "$EXPECTED" == "$ACTUAL" ]] || { echo "❌ Checksum mismatch!" ; exit 1; }

echo "➜ Extracting …"
tar -C "$TMPDIR" -xzf "$TMPDIR/$FILE"

echo "➜ Installing to $BINDIR …"
install -m 755 "$TMPDIR/$APP" "$BINDIR/$APP"

echo "✅ Installed $APP $VERSION to $BINDIR"
