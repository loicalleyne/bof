#!/usr/bin/env bash
# uninstall.sh — Remove bof links from ~/.copilot/
# Handles both Linux symlinks and Windows junctions/hard links.
# Only removes entries installed by bof. Never removes manual files.
#
# Usage:
#   bash scripts/uninstall.sh
#   bash scripts/uninstall.sh --dry-run

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DRY_RUN=0
USE_WINDOWS_LINKS=0

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
      USE_WINDOWS_LINKS=1
      return
    fi
  fi
  COPILOT_DIR="$HOME/.copilot"
}

resolve_copilot_dir

SKILLS_DIR="$COPILOT_DIR/skills"
AGENTS_DIR="$COPILOT_DIR/agents"
INSTRUCTIONS_DIR="$COPILOT_DIR/instructions"

to_win_path() { wslpath -w "${1%/}"; }

# ─── Helpers ─────────────────────────────────────────────────────────────────

# Remove a directory junction (Windows) or symlink (Linux).
remove_dir_link() {
  local path="${1%/}"
  [ -e "$path" ] || return 0
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: remove dir link $path"
    return
  fi
  if [ "$USE_WINDOWS_LINKS" -eq 1 ]; then
    powershell.exe -NoProfile -NonInteractive -Command \
      "Remove-Item -Path '$(to_win_path "$path")' -Force" > /dev/null 2>&1 \
      && echo "REMOVED: $path" || echo "SKIP:    $path (not a junction or already gone)"
  else
    if [ -L "$path" ] && [[ "$(readlink "$path")" == "$REPO_DIR"* ]]; then
      rm "$path" && echo "REMOVED: $path"
    else
      echo "SKIP:    $path"
    fi
  fi
}

# Remove a hard link / file installed by bof.
remove_file_link() {
  local path="${1%/}"
  [ -e "$path" ] || return 0
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: remove file link $path"
    return
  fi
  if [ "$USE_WINDOWS_LINKS" -eq 1 ]; then
    powershell.exe -NoProfile -NonInteractive -Command \
      "Remove-Item -Path '$(to_win_path "$path")' -Force" > /dev/null 2>&1 \
      && echo "REMOVED: $path" || echo "SKIP:    $path (could not remove)"
  else
    if [ -L "$path" ] && [[ "$(readlink "$path")" == "$REPO_DIR"* ]]; then
      rm "$path" && echo "REMOVED: $path"
    else
      echo "SKIP:    $path"
    fi
  fi
}

echo "Uninstalling bof from: $COPILOT_DIR"
[ "$DRY_RUN" -eq 1 ] && echo "(dry-run mode — no changes will be made)"
echo ""

# Skills (directories → junctions on Windows)
for skill_dir in "$REPO_DIR/skills"/*/; do
  skill_name="$(basename "$skill_dir")"
  remove_dir_link "$SKILLS_DIR/$skill_name"
done

# Agents (files → hard links on Windows)
for agent_file in "$REPO_DIR/agents"/*.agent.md; do
  agent_name="$(basename "$agent_file")"
  remove_file_link "$AGENTS_DIR/$agent_name"
done

# Instructions (files → hard links on Windows)
for instr_file in "$REPO_DIR/instructions"/*.instructions.md; do
  instr_name="$(basename "$instr_file")"
  remove_file_link "$INSTRUCTIONS_DIR/$instr_name"
done

echo ""
echo "Done."
