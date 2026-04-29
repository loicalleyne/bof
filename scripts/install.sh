#!/usr/bin/env bash
# install.sh — Install bof skills and agents into ~/.copilot/
#              Optionally also sync bof skills into ~/.config/crush/skills/ (--crush)
# Idempotent: safe to run multiple times. Skips existing links.
#
# Usage:
#   bash scripts/install.sh             # VS Code Copilot Chat only
#   bash scripts/install.sh --crush     # also sync bof skills into crush
#   bash scripts/install.sh --dry-run   # show what would be created (both targets)
#
# WSL note: When targeting the Windows filesystem, Linux symlinks (ln -sf) are
# NOT followed by Windows/VS Code — they appear as 0 KB opaque files.
# This script detects WSL + Windows target and uses Windows-native links:
#   - Directories → NTFS junction (mklink /J, no admin required)
#   - Files       → NTFS hard link (mklink /H, same drive, no admin required)
# The crush target (~/.config/crush/skills/) is on the Linux filesystem and
# always uses regular ln -sf symlinks regardless of platform.

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DRY_RUN=0
INSTALL_CRUSH=0
USE_WINDOWS_LINKS=0  # set to 1 when targeting Windows NTFS from WSL

# Parse flags
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    --crush)   INSTALL_CRUSH=1 ;;
  esac
done

# ─── Resolve .copilot target directory ───────────────────────────────────────

resolve_copilot_dir() {
  if grep -qi microsoft /proc/version 2>/dev/null; then
    # Running under WSL — resolve Windows USERPROFILE
    WIN_HOME="$(cmd.exe /c 'echo %USERPROFILE%' 2>/dev/null | tr -d '\r')"
    if [ -n "$WIN_HOME" ]; then
      COPILOT_DIR="$(wslpath "$WIN_HOME")/.copilot"
      USE_WINDOWS_LINKS=1  # target is Windows NTFS — must use mklink
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

# Convert a WSL path (/mnt/c/...) to a Windows path (C:\...) for PowerShell.
# Strips trailing slash first — a trailing backslash in Windows paths breaks quoting.
to_win_path() {
  wslpath -w "${1%/}"
}

# Link a DIRECTORY. Uses NTFS junction on Windows, ln -sf on Linux.
link_dir_or_dry() {
  local src="${1%/}" dst="${2%/}"
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: link dir $src → $dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  if [ -e "$dst" ] || [ -L "$dst" ]; then
    echo "EXISTS:  $dst"
    return
  fi
  if [ "$USE_WINDOWS_LINKS" -eq 1 ]; then
    local win_src win_dst
    win_src="$(to_win_path "$src")"
    win_dst="$(to_win_path "$dst")"
    powershell.exe -NoProfile -NonInteractive -Command \
      "New-Item -ItemType Junction -Path '${win_dst}' -Target '${win_src}' | Out-Null" \
      && echo "JUNCTION: $dst → $src" \
      || echo "ERROR:    failed to create junction for $dst"
  else
    ln -sf "$src" "$dst"
    echo "LINKED:  $dst → $src"
  fi
}

# Link a FILE. Uses NTFS hard link on Windows, ln -sf on Linux.
link_file_or_dry() {
  local src="${1%/}" dst="${2%/}"
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: link file $src → $dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  if [ -e "$dst" ] || [ -L "$dst" ]; then
    echo "EXISTS:  $dst"
    return
  fi
  if [ "$USE_WINDOWS_LINKS" -eq 1 ]; then
    local win_src win_dst
    win_src="$(to_win_path "$src")"
    win_dst="$(to_win_path "$dst")"
    powershell.exe -NoProfile -NonInteractive -Command \
      "New-Item -ItemType HardLink -Path '${win_dst}' -Target '${win_src}' | Out-Null" \
      && echo "HARDLINK: $dst → $src" \
      || echo "ERROR:    failed to create hard link for $dst"
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
  # Directories: junction on Windows, symlink on Linux
  link_dir_or_dry "$skill_dir" "$dst_dir"
done

# ─── Agents ──────────────────────────────────────────────────────────────────

for agent_file in "$REPO_DIR/agents"/*.agent.md; do
  agent_name="$(basename "$agent_file")"
  link_file_or_dry "$agent_file" "$AGENTS_DIR/$agent_name"
done

# ─── Instructions ─────────────────────────────────────────────────────────────

INSTRUCTIONS_DIR="$COPILOT_DIR/instructions"
mkdir_or_dry "$INSTRUCTIONS_DIR"

for instr_file in "$REPO_DIR/instructions"/*.instructions.md; do
  instr_name="$(basename "$instr_file")"
  link_file_or_dry "$instr_file" "$INSTRUCTIONS_DIR/$instr_name"
done

# ─── Crush skills (optional, --crush flag) ────────────────────────────────────
# Links each bof skill into ~/.config/crush/skills/.
# Non-bof skills already in that directory are left untouched.
# Stale bof dirs (real directories, not symlinks) are replaced with symlinks.

if [ "$INSTALL_CRUSH" -eq 1 ]; then
  CRUSH_SKILLS_DIR="$HOME/.config/crush/skills"
  echo ""
  echo "Syncing bof skills into: $CRUSH_SKILLS_DIR"
  mkdir -p "$CRUSH_SKILLS_DIR"

  for skill_dir in "$REPO_DIR/skills"/*/; do
    skill_name="$(basename "$skill_dir")"
    dst="$CRUSH_SKILLS_DIR/$skill_name"

    if [ "$DRY_RUN" -eq 1 ]; then
      if [ -L "$dst" ]; then
        echo "DRY-RUN: already linked $dst"
      elif [ -d "$dst" ]; then
        echo "DRY-RUN: replace stale dir $dst → $skill_dir"
      else
        echo "DRY-RUN: link $dst → $skill_dir"
      fi
      continue
    fi

    if [ -L "$dst" ]; then
      # Already a symlink — leave it (idempotent)
      echo "EXISTS:  $dst"
    elif [ -d "$dst" ]; then
      # Stale real directory from old install — replace with symlink
      rm -rf "$dst"
      ln -sf "$skill_dir" "$dst"
      echo "UPDATED: $dst → $skill_dir"
    else
      ln -sf "$skill_dir" "$dst"
      echo "LINKED:  $dst → $skill_dir"
    fi
  done
fi

# ─── Done ────────────────────────────────────────────────────────────────────

echo ""
echo "Done."
echo ""
echo "Verify with:"
echo "  ls $SKILLS_DIR"
if [ "$INSTALL_CRUSH" -eq 1 ]; then
  echo "  ls $HOME/.config/crush/skills"
fi
echo "  ls $AGENTS_DIR"
echo "  ls $INSTRUCTIONS_DIR"
