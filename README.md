# bof

A port of [superpowers](https://github.com/obra/superpowers) for **VS Code Copilot Chat**.

bof ships as a repository of `SKILL.md` files and `.agent.md` files that install into `~/.copilot/skills/` and `~/.copilot/agents/` via a symlink-based `scripts/install.sh`. When used alongside the [Esquisse](https://github.com/loicalleyne/esquisse) project framework, bof provides a complete end-to-end developer workflow: brainstorm → spec → plan → adversarial review → subagent implementation → code review → merge.

---

## Requirements

- **VS Code Copilot Chat v1.115+** — required for `SKILL.md` auto-trigger, `.agent.md` support, and `runSubagent`
- **WSL (Ubuntu)** — recommended for Windows users; install.sh auto-detects WSL and links to the correct Windows `.copilot` directory
- For lifecycle hooks: `"chat.useCustomAgentHooks": true` in VS Code settings

---

## Installation

```sh
# From WSL (recommended on Windows):
git clone https://github.com/loicalleyne/bof
cd bof
bash scripts/install.sh

# Verify:
ls ~/.copilot/skills/ | grep brainstorming
```

To uninstall:
```sh
bash scripts/uninstall.sh
```

Both scripts support `--dry-run` to preview changes without making them.

---

## Skills

| Skill | Trigger |
|-------|---------|
| `brainstorming` | "brainstorm", "explore the idea of" |
| `writing-plans` | "write a plan", "create task documents" |
| `executing-plans` | "execute the plan", "work through these tasks" |
| `subagent-driven-development` | "execute the plan with subagents", "implement these tasks" |
| `dispatching-parallel-agents` | "investigate these failures in parallel" |
| `adversarial-review` | "review this plan", "adversarial review" |
| `requesting-code-review` | "request code review", "review my changes" |
| `receiving-code-review` | "I received code review feedback" |
| `test-driven-development` | "write tests first", "TDD" |
| `verification-before-completion` | "verify before closing", "final check" |
| `systematic-debugging` | "test is failing", "debug this bug" |
| `using-git-worktrees` | "create a worktree", "start feature work" |
| `finishing-a-development-branch` | "implementation is complete", "merge this branch" |
| `writing-skills` | "write a skill", "create a SKILL.md" |
| `mcp-cli` | "use MCP server on demand", "discover MCP tools" |
| `using-tmux-for-interactive-commands` | "run interactive CLI", "vim programmatically" |

---

## Agents

| Agent | Role |
|-------|------|
| `ImplementerAgent` | Implements a single task document. Dispatched by SDD. |
| `SpecReviewerAgent` | Read-only spec compliance reviewer. |
| `CodeQualityReviewerAgent` | Read-only code quality reviewer. |
| `Adversarial-r0` | GPT-4.1 hostile plan reviewer (slot 0) |
| `Adversarial-r1` | Claude Opus 4.6 hostile plan reviewer (slot 1) |
| `Adversarial-r2` | GPT-4o hostile plan reviewer (slot 2) |

---

## Workflow

```
brainstorming
  → writing-plans + adversarial-review (required gate)
    → using-git-worktrees
      → subagent-driven-development
          loop: ImplementerAgent → SpecReviewerAgent → CodeQualityReviewerAgent
      → finishing-a-development-branch
```

---

## Esquisse Integration

When used in an [Esquisse](https://github.com/loicalleyne/esquisse)-initialized project:

- Plans go in `docs/tasks/P{n}-{nnn}-{slug}.md` (Esquisse task format)
- Adversarial review writes to `.adversarial/state.json` and `reports/`
- ImplementerAgent follows the full Esquisse completion protocol
- AST cache is copied at worktree creation and rebuilt incrementally

Without Esquisse, all skills degrade gracefully to single-file plans.

---

## Key Rules

1. **Tool names are VS Code only.** `run_in_terminal`, `runSubagent`, `manage_todo_list`, `vscode_askQuestions` — never Claude Code equivalents.
2. **Namespace is `bof:`** — not `superpowers:`.
3. **Read AGENTS.md** in the project root before starting any task.

---

## Credits

Ported from [superpowers v5.0.7](https://github.com/obra/superpowers) by Jesse Vincent, with VS Code adaptations and Esquisse integration.
