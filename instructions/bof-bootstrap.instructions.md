---
description: Always-on bof workflow bootstrap. Read project AGENTS.md first.
applyTo: "**"
---

# bof Workflow Bootstrap

## Priority Order

1. **Project `AGENTS.md` / `ONBOARDING.md`** (highest priority — project law)
2. **bof skills** (override default agent behavior)
3. **Default Copilot behavior** (lowest)

## Project Context (read first)

If `AGENTS.md` exists in the workspace root, read it before any response.
If `ONBOARDING.md` exists, read it. These are your project constitution.
If `GLOSSARY.md` exists, use its vocabulary — never invent alternative names for
domain concepts that already have names in the code.

## AST-Aided Navigation (prefer over file reads)

If `code_ast.duckdb` exists at the project root, use the `duckdb-code` skill
for structural questions (where is X defined, what calls X, what implements
interface Y) BEFORE reading source files. AST queries preserve context budget
for implementation. Fall back to `grep_search` / `read_file` only when no
cache exists or no cache hit.

Rule: **Use AST queries to find the right file, then read only that file.**

## Available bof Skills

Skills auto-trigger from their description. You do not need to name them. Core skills:

- **brainstorming** — MUST use before any feature work. HARD GATE: no code until design approved.
- **writing-plans** — use after the design spec is approved. Runs adversarial review before execution handoff.
- **subagent-driven-development** — use to execute plans with Implementer → SpecReviewer → QualityReviewer per task.
- **executing-plans** — use in single-session execution mode (no subagents).
- **test-driven-development** — IRON LAW: write the RED test first, always. Use during ALL implementation.
- **systematic-debugging** — use before ANY bug fix attempt. Phase 1 Investigation before any code change.
- **verification-before-completion** — use before ANY completion claim. Must run tests.
- **adversarial-review** — use to hostile-review a plan before implementation. Rotates reviewers.
- **using-git-worktrees** — use before starting feature implementation.
- **finishing-a-development-branch** — use when implementation is complete.
- **requesting-code-review** — use after completing major features.
- **receiving-code-review** — use when RECEIVING review feedback from a reviewer.
- **dispatching-parallel-agents** — use for 2+ independent failures or tasks.
- **writing-skills** — use to create or update bof skills.
- **mcp-cli** — use for on-demand MCP server queries.
- **using-tmux-for-interactive-commands** — use for interactive CLI tools (vim, git rebase -i, etc.).
- **duckdb-code** — use for AST-based structural codebase queries (if `code_ast.duckdb` present).

## SUBAGENT-STOP

If you are a subagent dispatched via `runSubagent()` for a specific task, DO NOT
activate full skill workflows unless explicitly instructed. Execute your specific
task. Do not re-read the bootstrap except for the tool name mapping section.

## Tool Name Reference

If you need to run a command or perform an action:

| What to do | Tool to use |
|---|---|
| Run a shell command | `run_in_terminal` |
| Dispatch a subagent | `runSubagent("AgentName", prompt)` |
| Track tasks | `manage_todo_list` |
| Ask the user a question | `vscode_askQuestions` |
| Read a file | `read_file` |
| Edit an existing file | `replace_string_in_file` |
| Create a new file | `create_file` |
| Search files by name | `file_search` |
| Search file contents | `grep_search` |
| List directory | `list_dir` |
| Fetch a web page | `fetch_webpage` |

**Banned tool names** (never use in skill steps or agent prompts):
`Bash`, `Task`, `TodoWrite`, `AskUserQuestion`, `Read`, `Write`, `Edit`

## Esquisse Integration

If the project uses Esquisse (has `AGENTS.md`, `GLOSSARY.md`, `docs/tasks/`):

- Spec files save to: `docs/specs/YYYY-MM-DD-{topic}-design.md`
- Plan tasks save to: `docs/tasks/P{n}-{nnn}-{slug}.md` (Esquisse task doc format)
- After each implementation task, follow the Esquisse completion protocol:
  - Update task doc `Status: Done` with session notes
  - Update `AGENTS.md` Common Mistakes if a new gotcha was found
  - Update `GLOSSARY.md` if new domain terms were introduced
  - Update `ONBOARDING.md` Key Files Map if new key files were created
  - Update `ROADMAP.md` task status
  - Append to `docs/planning/NEXT_STEPS.md` session log

If Esquisse documents are absent, save specs and plans to `docs/` with reasonable naming.
The workflow still applies; only the exact file paths differ.

## Adversarial Review Integration

- `writing-plans` outputs a plan → MUST run `bof:adversarial-review` before any execution handoff.
- `subagent-driven-development` checks `.adversarial/state.json` for a recent review verdict before dispatching the first ImplementerAgent.
- A FAILED verdict means revising the plan, not skipping review.
- For hard Stop-hook enforcement (blocking gate), run `bash scripts/init.sh` in your project to install Esquisse's workspace-level `.github/agents/` with `gate-review.sh`. bof's user-level agents provide advisory-only adversarial review.
