#!/usr/bin/env bash
# pgview demo — quick-start script
# Works on macOS, Linux, and WSL2 (Windows).
# Requirements: Docker Desktop, Colima, or Podman — nothing else.
set -euo pipefail

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="https://raw.githubusercontent.com/sibasismukherjee/pgview/main/demo"

# ── helpers ──────────────────────────────────────────────────────────────────

say()  { printf '\033[1;36m→\033[0m %s\n' "$*"; }
ok()   { printf '\033[1;32m✓\033[0m %s\n' "$*"; }
die()  { printf '\033[1;31m✗\033[0m %s\n' "$*" >&2; exit 1; }

# ── download support files if missing ────────────────────────────────────────

download() {
  local file="$1"
  if [ ! -f "$DEMO_DIR/$file" ]; then
    say "fetching $file..."
    if command -v curl &>/dev/null; then
      curl -fsSL "$BASE_URL/$file" -o "$DEMO_DIR/$file"
    elif command -v wget &>/dev/null; then
      wget -qO "$DEMO_DIR/$file" "$BASE_URL/$file"
    else
      die "Neither curl nor wget found. Install one and try again."
    fi
  fi
}

download docker-compose.yml
download seed.sql

# ── detect container runtime ─────────────────────────────────────────────────

COMPOSE=""
if command -v docker &>/dev/null; then
  COMPOSE="docker compose"
elif command -v podman &>/dev/null; then
  COMPOSE="podman compose"
else
  echo ""
  echo "  No container runtime found. Install one of:"
  echo "    macOS:   brew install --cask docker          (Docker Desktop)"
  echo "    macOS:   brew install colima docker          (Colima, no GUI)"
  echo "    Linux:   curl -fsSL https://get.docker.com | sh"
  echo "    WSL2:    enable Docker Desktop WSL2 integration"
  echo ""
  exit 1
fi

# ── on macOS, start Colima if it is installed but not running ─────────────────

if command -v colima &>/dev/null && ! colima status 2>/dev/null | grep -q "Running"; then
  say "starting Colima..."
  colima start
fi

# ── start the database ────────────────────────────────────────────────────────

say "starting demo database (postgres:16-alpine on :5433)..."
cd "$DEMO_DIR"
$COMPOSE up -d

say "waiting for PostgreSQL to be ready..."
until $COMPOSE exec db pg_isready -U postgres -d demodb -q 2>/dev/null; do
  printf '.'; sleep 1
done
printf '\n'
ok "demo database ready on localhost:5433"

# ── download pgview binary ────────────────────────────────────────────────────

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"

TAG=""
if command -v curl &>/dev/null; then
  TAG=$(curl -sfL "https://api.github.com/repos/sibasismukherjee/pgview/releases/latest" \
        | grep '"tag_name"' | cut -d'"' -f4)
elif command -v wget &>/dev/null; then
  TAG=$(wget -qO- "https://api.github.com/repos/sibasismukherjee/pgview/releases/latest" \
        | grep '"tag_name"' | cut -d'"' -f4)
fi

[ -z "$TAG" ] && die "Could not fetch latest release tag (network issue?)."

BINARY="pgview_${TAG}_${OS}_${ARCH}"
say "downloading pgview ${TAG} for ${OS}/${ARCH}..."

if command -v curl &>/dev/null; then
  curl -sfL "https://github.com/sibasismukherjee/pgview/releases/download/${TAG}/${BINARY}" \
    -o "$DEMO_DIR/pgview"
else
  wget -qO "$DEMO_DIR/pgview" \
    "https://github.com/sibasismukherjee/pgview/releases/download/${TAG}/${BINARY}"
fi
chmod +x "$DEMO_DIR/pgview"
ok "pgview ${TAG} ready"

# ── macOS Gatekeeper ──────────────────────────────────────────────────────────

if [ "$OS" = "darwin" ]; then
  xattr -d com.apple.quarantine "$DEMO_DIR/pgview" 2>/dev/null || true
fi

# ── launch ────────────────────────────────────────────────────────────────────

echo ""
echo "  Connect info (for reference):"
echo "    host=localhost:5433  user=postgres  pass=demo  db=demodb"
echo ""
say "launching pgview..."
echo ""

"$DEMO_DIR/pgview" \
  -url      localhost:5433 \
  -username postgres \
  -password demo \
  -dbname   demodb \
  -sslmode  disable
