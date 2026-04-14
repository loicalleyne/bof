#!/usr/bin/env bash
# uninstall.sh — Remove bof symlinks from ~/.copilot/
# Only removes symlinks that point back to this repo. Never removes manual files.
#
# Usage:
#   bash scripts/uninstall.sh
#   bash scripts/uninstall.sh --dry-run

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
  esac
done

# ─── Resolve .copilot target directory (same logic as install.sh) ────────────

resolve_copilot_dir() {
  if grep -qi microsoft /proc/version 2>/dev/null; then
    WIN_HOME="$(cmd.exe /c 'echo %USERPROFILE%' 2>/dev/null | tr -d '\r')"
    if [ -n "$WIN_HOME" ]; then
      COPILOT_DIR="$(wslpath "$WIN_HOME")/.copilot"
      return
    fi
  fi
  COPILOT_DIR="$HOME/.copilot"
}

resolve_copilot_dir

SKILLS_DIR="$COPILOT_DIR/skills"
AGENTS_DIR="$COPILOT_DIR/agents"
INSTRUCTIONS_DIR="$COPILOT_DIR/instructions"

# ─── Helpers ─────────────────────────────────────────────────────────────────

remove_if_bof_link() {
  local link="$1"
  if [ ! -L "$link" ]; then
    return  # Not a symlink — skip
  fi
  local target
  target="$(readlink "$link")"
  # Only remove if the symlink target is inside the bof repo
  if [[ "$target" == "$REPO_DIR"* ]]; then
    if [ "$DRY_RUN" -eq 1 ]; then
      echo "DRY-RUN: rm $link"
    else
      rm "$link"
      echo "REMOVED: $link"
    fi
  else
    echo "SKIP:    $link (points to $target — not a bof link)"
  fi
}

echo "Uninstalling bof from: $COPILOT_DIR"
[ "$DRY_RUN" -eq 1 ] && echo "(dry-run mode — no changes will be made)"
echo ""

# Skills
for skill_dir in "$REPO_DIR/skills"/*/; do
  skill_name="$(basename "$skill_dir")"
  remove_if_bof_link "$SKILLS_DIR/$skill_name"
done

# Agents
for agent_file in "$REPO_DIR/agents"/*.agent.md; do
  agent_name="$(basename "$agent_file")"
  remove_if_bof_link "$AGENTS_DIR/$agent_name"
done

# Instructions
for instr_file in "$REPO_DIR/instructions"/*.instructions.md; do
  instr_name="$(basename "$instr_file")"
  remove_if_bof_link "$INSTRUCTIONS_DIR/$instr_name"
done

echo ""
echo "Done."
