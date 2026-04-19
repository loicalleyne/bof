# GLOSSARY.md — bof

All terms are drawn from the actual names used in bof's files and referenced by
bof skills and agents. If a file uses a different name from the definition here,
the file is wrong.

| Term | Definition |
|------|-----------|
| **skill** | A `SKILL.md` file installed in `~/.copilot/skills/{name}/`. VS Code Copilot Chat auto-triggers the skill when the user's message matches the `description:` frontmatter field. |
| **skill trigger** | The mechanism by which VS Code Copilot Chat activates a skill: the user message is matched against the `description:` field in `SKILL.md` frontmatter. No explicit invocation is needed. |
| **SKILL.md** | The file that defines a skill. Must have YAML frontmatter with at minimum `name:` and `description:`. Lives at `~/.copilot/skills/{skill-name}/SKILL.md`. |
| **agent file** | A `.agent.md` file installed in `~/.copilot/agents/`. Defines a named subagent dispatched via `runSubagent()`. Must have frontmatter: `name:`, `model:`, `tools:`. |
| **bootstrap injection** | The mechanism by which `bof-bootstrap.instructions.md` (an `.instructions.md` file with `applyTo: "**"`) injects always-on context into every VS Code Copilot Chat session without an explicit user action. |
| **subagent dispatch** | Calling `runSubagent("AgentName", prompt)` to hand off a bounded task to a named agent. The dispatching agent waits for the result before continuing. |
| **bof namespace** | The `bof:` prefix used in cross-skill references (e.g., `bof:brainstorming`, `bof:writing-plans`). Distinguishes bof skills from built-in or third-party skills. |
| **tool mapping** | The correspondence between Claude Code tool names (upstream) and VS Code Copilot Chat tool names (bof). Documented in `docs/specs/tool-mapping.md`. Any upstream tool name appearing in a bof skill is a bug. |
| **spec compliance review** | The review performed by `SpecReviewerAgent`: verifies that the implementation matches the task specification exactly — no missing requirements, no extra features. Returns ✅ compliant or ❌ with specific issues. |
| **code quality review** | The review performed by `CodeQualityReviewerAgent`: assesses clean code, test coverage, and maintainability AFTER spec compliance passes. Severity: Critical / Important / Minor. |
| **worktree isolation** | The practice of creating a git worktree in `.worktrees/` for each feature branch so that main checkout and feature work do not share the same directory. Managed by `bof:using-git-worktrees`. |
| **Esquisse integration** | bof's optional compatibility with the Esquisse project framework. When Esquisse documents (`AGENTS.md`, `GLOSSARY.md`, `docs/tasks/`) are present, bof skills use Esquisse task doc format and follow the Esquisse completion protocol. Skills degrade gracefully when Esquisse is absent. |
| **bof-mcp** | The Go MCP server in `bof/bof-mcp/`. Exposes 6 MCP tools that give Crush (and VS Code via MCP) the equivalent of VS Code's `runSubagent` dispatch plus `discover_models` for Crush model discovery. The only non-markdown component of bof. Built with `cd bof-mcp && go build -o bof-mcp .`. |
| **Crush Mode** | A `## Crush Mode (bof-mcp)` section appended to bof skills that structurally depend on `runSubagent`. Describes how to replace VS Code agent dispatch with the equivalent bof-mcp MCP tool (`implementer_agent`, `spec_review`, `quality_review`, `adversarial_review`). |
| **instructions file** | An `.instructions.md` file in `~/.copilot/instructions/`. Injected into every Copilot Chat session matching its `applyTo:` glob. bof uses one: `bof-bootstrap.instructions.md`. |
| **adversarial review** | A hostile, cross-model critique of a plan using the 7-attack protocol. Performed by `Adversarial-r*` agents. Produces a verdict: PASSED, CONDITIONAL, or FAILED. |
| **rotation slot** | The adversarial reviewer model selected by `iteration % 3` from `.adversarial/state.json`. Slot 0 → GPT-4.1; slot 1 → Claude Opus 4.6; slot 2 → GPT-4o. |
| **bof:adversarial-review** | The `skills/adversarial-review/SKILL.md` orchestrator skill. Reads rotation state, dispatches the appropriate `Adversarial-r*`, writes the verdict report. |
| **advisory-only gate** | bof's adversarial review is advisory: it produces a verdict file and the writing-plans/SDD skills enforce it via prompt instructions, but there is no OS-level blocking. Hard enforcement requires Esquisse's workspace-level `gate-review.sh` Stop hook. |
| **AST cache** | The `code_ast.duckdb` file at a project root, built by the `sitting_duck` DuckDB extension. Used by the `duckdb-code` skill for structural queries. Gitignored; rebuilt with `bash scripts/rebuild-ast.sh`. |
| **trigger test** | A `.md` file in `tests/triggers/` that documents the exact prompt to paste into Copilot Chat, the expected observable behavior, and the things that must NOT happen, for a specific skill. |
| **ImplementerAgent** | The bof subagent that implements a single task document following TDD. Reports `DONE`, `DONE_WITH_CONCERNS`, `BLOCKED`, or `NEEDS_CONTEXT`. Has full read/write/execute tools. |
| **SpecReviewerAgent** | The bof subagent that verifies implementation against the task spec. Read-only. Returns ✅ compliant or ❌ with specific issue list. |
| **CodeQualityReviewerAgent** | The bof subagent that reviews implementation for code quality after spec compliance passes. Read-only. Returns severity-labelled verdict. |
| **install.sh** | `scripts/install.sh` — symlinks bof skills, agents, instructions, and hooks into `~/.copilot/`. WSL-aware: resolves the correct Windows `.copilot` path. |
| **junction** | A Windows NTFS directory reparse point. Used as a fallback by `install.sh` when `ln -s` is unavailable (requires no Windows Developer Mode). Created via `cmd.exe /C mklink /J`. |
| **hooks** | VS Code Copilot Chat Preview lifecycle events: `SessionStart`, `SubagentStart`, `PostToolUse`, `Stop`, etc. Configured in `~/.copilot/hooks/` via a JSON file. Require `chat.useCustomAgentHooks: true` in VS Code settings. |

