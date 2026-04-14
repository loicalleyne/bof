#!/usr/bin/env bash
# install.sh — Install bof skills and agents into ~/.copilot/
# Idempotent: safe to run multiple times. Skips existing symlinks.
#
# Usage:
#   bash scripts/install.sh             # Uses default ~/.copilot location
#   bash scripts/install.sh --dry-run   # Show what would be created
#
# WSL note: This script detects whether ~/.copilot should point to the
# Windows user directory (when running in WSL) or the Linux home directory.

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DRY_RUN=0

# Parse flags
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
  esac
done

# ─── Resolve .copilot target directory ───────────────────────────────────────

resolve_copilot_dir() {
  # In WSL, ~/.copilot must point to the Windows user directory because
  # VS Code reads from the Windows filesystem, not the WSL Linux home.
  if grep -qi microsoft /proc/version 2>/dev/null; then
    # Running under WSL — resolve Windows USERPROFILE
    WIN_HOME="$(cmd.exe /c 'echo %USERPROFILE%' 2>/dev/null | tr -d '\r')"
    if [ -n "$WIN_HOME" ]; then
      COPILOT_DIR="$(wslpath "$WIN_HOME")/.copilot"
      return
    fi
    echo "WARN: Could not resolve Windows USERPROFILE. Falling back to \$HOME/.copilot" >&2
  fi
  COPILOT_DIR="$HOME/.copilot"
}

resolve_copilot_dir

SKILLS_DIR="$COPILOT_DIR/skills"
AGENTS_DIR="$COPILOT_DIR/agents"

# ─── Helpers ─────────────────────────────────────────────────────────────────

link_or_dry() {
  local src="$1" dst="$2"
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: ln -sf $src $dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  if [ -L "$dst" ]; then
    echo "EXISTS:  $dst"
  elif [ -e "$dst" ]; then
    echo "SKIP:    $dst (not a symlink — manual file present, not overwriting)"
  else
    ln -sf "$src" "$dst"
    echo "LINKED:  $dst → $src"
  fi
}

mkdir_or_dry() {
  local dir="$1"
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: mkdir -p $dir"
    return
  fi
  mkdir -p "$dir"
}

# ─── Pre-flight ──────────────────────────────────────────────────────────────

echo "Installing bof into: $COPILOT_DIR"
echo "Source repo:         $REPO_DIR"
[ "$DRY_RUN" -eq 1 ] && echo "(dry-run mode — no changes will be made)"
echo ""

mkdir_or_dry "$SKILLS_DIR"
mkdir_or_dry "$AGENTS_DIR"

# ─── Skills ──────────────────────────────────────────────────────────────────

for skill_dir in "$REPO_DIR/skills"/*/; do
  skill_name="$(basename "$skill_dir")"
  dst_dir="$SKILLS_DIR/$skill_name"
  # Link the whole skill directory (so all files in it are available)
  link_or_dry "$skill_dir" "$dst_dir"
done

# ─── Agents ──────────────────────────────────────────────────────────────────

for agent_file in "$REPO_DIR/agents"/*.agent.md; do
  agent_name="$(basename "$agent_file")"
  link_or_dry "$agent_file" "$AGENTS_DIR/$agent_name"
done

# ─── Instructions ─────────────────────────────────────────────────────────────

INSTRUCTIONS_DIR="$COPILOT_DIR/instructions"
mkdir_or_dry "$INSTRUCTIONS_DIR"

for instr_file in "$REPO_DIR/instructions"/*.instructions.md; do
  instr_name="$(basename "$instr_file")"
  link_or_dry "$instr_file" "$INSTRUCTIONS_DIR/$instr_name"
done

# ─── Done ────────────────────────────────────────────────────────────────────

echo ""
echo "Done."
echo ""
echo "Verify with:"
echo "  ls $SKILLS_DIR"
echo "  ls $AGENTS_DIR"
echo "  ls $INSTRUCTIONS_DIR"
