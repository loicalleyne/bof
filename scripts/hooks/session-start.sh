#!/usr/bin/env bash
# scripts/hooks/session-start.sh
# Runs at the start of every VS Code Copilot Chat session.
# Outputs a JSON object for VS Code to inject into the agent context.
#
# Security: This script reads only AGENTS.md and minimal git metadata.
# It never reads files containing credentials or forwards env vars with secrets.
# It does not execute user-provided code.

set -euo pipefail

# Resolve project root (where the chat session opened)
PROJECT_ROOT="${PWD:-$(pwd)}"

# Gather minimal context
PROJECT_NAME="$(basename "$PROJECT_ROOT")"
GIT_BRANCH="$(git -C "$PROJECT_ROOT" branch --show-current 2>/dev/null || echo 'N/A')"
HAS_AGENTS_MD=0
[ -f "$PROJECT_ROOT/AGENTS.md" ] && HAS_AGENTS_MD=1

# Adversarial review state (safe to read — only contains metadata, no secrets)
ADVERSARIAL_VERDICT=""
if [ -f "$PROJECT_ROOT/.adversarial/state.json" ]; then
  ADVERSARIAL_VERDICT="$(grep -o '"last_verdict"[[:space:]]*:[[:space:]]*"[^"]*"' \
    "$PROJECT_ROOT/.adversarial/state.json" 2>/dev/null \
    | grep -o '"[^"]*"$' | tr -d '"' || echo '')"
fi

# Emit context JSON for VS Code to inject
cat <<EOF
{
  "project": "${PROJECT_NAME}",
  "branch": "${GIT_BRANCH}",
  "hasAgentsMd": ${HAS_AGENTS_MD},
  "adversarialVerdict": "${ADVERSARIAL_VERDICT}"
}
EOF
