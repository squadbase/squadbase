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
    x86_64|amd64)  echo "x86_64" ;;   # GoReleaser と同じ表記
    arm64|aarch64) echo "arm64"  ;;
    armv7*)        echo "armv7"  ;;
    *) echo "unsupported arch: $(uname -m)"; exit 1 ;;
  esac
}
OS=$(detect_os); ARCH=$(detect_arch)

# -------- version --------
if [[ -z "$VERSION" || "$VERSION" == "latest" ]]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
            | grep -m1 '"tag_name":' | cut -d '"' -f4)
  [[ -z "$VERSION" ]] && { echo "No release found"; exit 1; }
fi

VERSION_NO_V="${VERSION#v}"
EXT=$([[ "$OS" == "windows" ]] && echo "zip" || echo "tar.gz")
FILE="${APP}_${VERSION_NO_V}_${OS}_${ARCH}.${EXT}"
URL="${REPO_URL}/${VERSION}/${FILE}"
SUM_URL="${REPO_URL}/${VERSION}/checksums.txt"

echo "▶︎ URL: $URL"

TMP=$(mktemp -d); trap 'rm -rf "$TMP"' EXIT
curl -fsSL "$URL" -o "$TMP/$FILE"           || { echo "❌ 404 ($FILE)"; exit 1; }
curl -fsSL "$SUM_URL" -o "$TMP/SUMS"        || { echo "❌ checksums.txt 取得失敗"; exit 1; }

# -------- checksum --------
hash_cmd() { command -v sha256sum >/dev/null && sha256sum "$1" || shasum -a 256 "$1"; }
EXPECTED=$(grep "$FILE$" "$TMP/SUMS" | awk '{print $1}')
ACTUAL=$(hash_cmd "$TMP/$FILE" | awk '{print $1}')
[[ "$EXPECTED" == "$ACTUAL" ]] || { echo "❌ checksum mismatch"; exit 1; }

# -------- extract & install --------
if [[ "$OS" == "windows" ]]; then
  unzip -q "$TMP/$FILE" -d "$TMP"
else
  tar -C "$TMP" -xzf "$TMP/$FILE"
fi
install -m 755 "$TMP/$APP" "$BINDIR/$APP"
echo "✅ Installed $APP $VERSION → $BINDIR/$APP"

