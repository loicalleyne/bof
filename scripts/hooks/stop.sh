#!/usr/bin/env bash
# scripts/hooks/stop.sh
# Runs at session Stop.
# Checks for a pending FAILED adversarial verdict and prints an advisory warning.
#
# Always exits 0 (advisory only). Hard gates are enforced by the skills themselves.
# Never writes files or modifies state.

set -uo pipefail

PROJECT_ROOT="${PWD:-$(pwd)}"
STATE_FILE="$PROJECT_ROOT/.adversarial/state.json"

# Fast-pass if no adversarial state
[ -f "$STATE_FILE" ] || exit 0

# Read the verdict using basic shell (no jq dependency)
VERDICT=$(grep -o '"last_verdict"[[:space:]]*:[[:space:]]*"[^"]*"' "$STATE_FILE" 2>/dev/null | head -1 | sed 's/.*"last_verdict"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' || echo "")

case "$VERDICT" in
  "FAILED")
    echo "[bof] WARNING: The last adversarial review verdict was FAILED." >&2
    echo "[bof] The plan should NOT proceed to implementation until the verdict is PASSED or CONDITIONAL." >&2
    echo "[bof] See .adversarial/state.json for details." >&2
    ;;
  "CONDITIONAL")
    echo "[bof] REMINDER: The last adversarial review verdict was CONDITIONAL." >&2
    echo "[bof] Address all conditions before dispatching ImplementerAgent." >&2
    ;;
  *)
    # PASSED or unknown — no action needed
    ;;
esac

exit 0
