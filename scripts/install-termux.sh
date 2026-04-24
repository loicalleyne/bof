#!/usr/bin/env bash
# install-termux.sh — Install bof skills into Crush on Termux
# Idempotent: safe to run multiple times. Skips existing links.
#
# Assumes bof-mcp is installed separately (e.g. via `go install`).
#
# Usage:
#   bash scripts/install-termux.sh             # Install skills into Crush config
#   bash scripts/install-termux.sh --dry-run   # Show what would be created
#
# Crush config directory: $HOME/.config/crush
# Skills target:          $HOME/.config/crush/skills/
# Agents target:          $HOME/.config/crush/agents/
#
# Termux note: $HOME resolves to /data/data/com.termux/files/home correctly.
# Standard Linux symlinks (ln -sf) work normally in Termux.

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
  esac
done

# ─── Verify Termux environment ────────────────────────────────────────────────

if [ ! -d "/data/data/com.termux" ]; then
  echo "ERROR: This script is intended for Termux only." >&2
  echo "       For WSL/Linux use: bash scripts/install.sh" >&2
  exit 1
fi

# ─── Resolve Crush config directory ──────────────────────────────────────────

CRUSH_CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/crush"
SKILLS_DIR="$CRUSH_CONFIG_DIR/skills"

# ─── Helpers ─────────────────────────────────────────────────────────────────

link_dir_or_dry() {
  local src="${1%/}" dst="${2%/}"
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: link dir  $src → $dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  if [ -e "$dst" ] || [ -L "$dst" ]; then
    echo "EXISTS:  $dst"
    return
  fi
  ln -sf "$src" "$dst"
  echo "LINKED:  $dst → $src"
}

mkdir_or_dry() {
  local dir="$1"
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: mkdir -p $dir"
    return
  fi
  mkdir -p "$dir"
}

# ─── Pre-flight ───────────────────────────────────────────────────────────────

echo "Installing bof into Crush on Termux"
echo "Source repo:   $REPO_DIR"
echo "Crush config:  $CRUSH_CONFIG_DIR"
echo "Skills target: $SKILLS_DIR"
[ "$DRY_RUN" -eq 1 ] && echo "(dry-run mode — no changes will be made)"
echo ""

mkdir_or_dry "$SKILLS_DIR"

# ─── Skills ───────────────────────────────────────────────────────────────────

for skill_dir in "$REPO_DIR/skills"/*/; do
  skill_name="$(basename "$skill_dir")"
  link_dir_or_dry "$skill_dir" "$SKILLS_DIR/$skill_name"
done

# ─── Done ─────────────────────────────────────────────────────────────────────

echo ""
echo "Done."
echo ""
echo "Verify with:"
echo "  ls $SKILLS_DIR"
echo ""
echo "Note: bof-mcp must be installed separately."
echo "      Add it to your crush.json under the 'mcp' section:"
echo ""
echo '  "bof": {'
echo '    "type": "stdio",'
echo '    "command": "/data/data/com.termux/files/home/go/bin/bof-mcp",'
echo '    "args": ["--project-root", "."]'
echo '  }'
