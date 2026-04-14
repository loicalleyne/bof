#!/usr/bin/env bash
# scripts/hooks/subagent-start.sh
# Runs at the start of each subagent invocation.
# Outputs a JSON object with context to inject into the subagent.
#
# Security: Reads only AGENTS.md and .adversarial/state.json (metadata only).
# Never reads credentials files or forwards environment variables with secrets.

set -euo pipefail

PROJECT_ROOT="${PWD:-$(pwd)}"

# Subagent name (provided by VS Code as COPILOT_AGENT_NAME env var when hooks run)
AGENT_NAME="${COPILOT_AGENT_NAME:-unknown}"

# Check if project has Esquisse task structure
HAS_TASKS_DIR=0
[ -d "$PROJECT_ROOT/docs/tasks" ] && HAS_TASKS_DIR=1

# Check for AST cache
HAS_AST_CACHE=0
[ -f "$PROJECT_ROOT/code_ast.duckdb" ] && HAS_AST_CACHE=1

cat <<EOF
{
  "agentName": "${AGENT_NAME}",
  "hasTasksDir": ${HAS_TASKS_DIR},
  "hasAstCache": ${HAS_AST_CACHE},
  "reminder": "Read AGENTS.md before starting. Use edit not write. Use run_in_terminal not Bash."
}
EOF
