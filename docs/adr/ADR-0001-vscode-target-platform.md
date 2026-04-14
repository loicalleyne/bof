# ADR-0001: VS Code Copilot Chat as Target Platform

**Date:** 2026-04-13
**Status:** Accepted

---

## Context

The upstream project, `superpowers` (v5.0.7), targets Claude Code and Cursor. bof
is a port of superpowers. The question is: which AI coding assistant platform
should be the primary target for bof?

The user's primary development environment is VS Code with the GitHub Copilot Chat
extension. The user has evaluated the VS Code Copilot Chat skill system and confirmed
it provides full primitive support:

- `SKILL.md` with `description:` frontmatter for auto-triggered skills
- `.agent.md` files with `model:`, `tools:`, and `agents:` frontmatter
- `.instructions.md` files for always-on context injection
- `runSubagent("AgentName", prompt)` for named subagent dispatch
- `manage_todo_list` for per-session task tracking
- `vscode_askQuestions` for structured user interaction
- Lifecycle hooks (Preview): `SessionStart`, `SubagentStart`, `Stop`, etc.

---

## Decision

Target **VS Code Copilot Chat** as the sole supported platform for bof.

---

## Alternatives Considered

| Alternative | Reason Rejected |
|---|---|
| **Claude Code** (upstream) | Not the user's primary workflow; would be a mirror of superpowers, not an independent port |
| **Cursor** | Not in use; Cursor has different primitives than VS Code |
| **Gemini CLI** | No skill/agent system; would require a fundamentally different architecture |
| **Multi-platform (VS Code + Claude Code)** | Doubles the maintenance burden; skill content would have so many conditionals it would be unmaintainable |

---

## Consequences

**Positive:**
- Skills can assume VS Code Copilot Chat primitives: no conditional tool selection per platform.
- Agents can use `.agent.md` format which is VS Code-specific but well-supported.
- Bootstrap instructions file provides always-on context injection unavailable in Claude Code.
- Lifecycle hooks (SessionStart etc.) enable dynamic context that static instructions cannot provide.

**Negative:**
- bof is not usable with Claude Code or Cursor without manual adaptation.
- No plugin marketplace entry (VSIX packaging); installation is manual via `scripts/install.sh`.
- Hooks require `"chat.useCustomAgentHooks": true` (Preview feature) — may change with VS Code updates.

**Follow-up:**
- VSIX packaging deferred to post-P6. See Out of Scope in AGENTS.md.

---

## References

- superpowers upstream: https://github.com/obra/superpowers
- VS Code Copilot Chat skill docs: https://code.visualstudio.com/docs/copilot/copilot-customization
