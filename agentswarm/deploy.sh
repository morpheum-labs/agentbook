#!/bin/bash

# Build the agentswarm SPA and optionally copy to a static deploy directory.
# Usage: bash deploy.sh [destination_directory]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$SCRIPT_DIR"

if [ ! -f "$APP_DIR/package.json" ]; then
  echo "Error: package.json not found in agentswarm."
  exit 1
fi

cd "$APP_DIR"

# Baked in at build time (Vite). If VITE_API_URL is already set in the environment, that wins.
# Example: VITE_API_URL=https://api.example.com bash deploy.sh
export VITE_API_URL="${VITE_API_URL:-http://127.0.0.1:3456}"
echo "Building with VITE_API_URL=$VITE_API_URL"

bun run build

if [ ! -d "dist" ]; then
  echo "Error: dist was not created after build"
  exit 1
fi

cp -r vercel.json dist/vercel.json

DEST_DIR="${1:-/Users/hesdx/Documents/dump/cl-yinclawlaunry-i83-demo}"
mkdir -p "$DEST_DIR"
find "$DEST_DIR" -maxdepth 1 -not -name '.git' -not -path "$DEST_DIR" -exec rm -rf {} + 2>/dev/null || true
cp -r dist/* "$DEST_DIR/" 2>/dev/null || true

echo "Build completed. Artifacts: $DEST_DIR"
echo "Preview locally: cd \"$APP_DIR\" && bun run preview (after VITE_API_URL is set in .env or env)"


if git -C "$DEFAULT_DEPLOY" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Staging, committing, and pushing deploy repo at $DEFAULT_DEPLOY ..."
  git -C "$DEFAULT_DEPLOY" add -A
  if git -C "$DEFAULT_DEPLOY" diff --cached --quiet; then
    echo "No changes to commit in deploy repo."
  else
    MSG="${GIT_COMMIT_MSG:-Deploy $(date -u +%Y-%m-%dT%H:%M:%SZ)}"
    git -C "$DEFAULT_DEPLOY" commit -m "$MSG"
    git -C "$DEFAULT_DEPLOY" push
    echo "Pushed deploy repo."
  fi
else
  echo "Note: $DEFAULT_DEPLOY is not a git repository; skipping git add/commit/push."
fi

