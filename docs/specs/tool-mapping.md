# Tool Mapping: Claude Code → VS Code Copilot Chat

This document is the canonical reference for translating upstream `superpowers`
skill content (authored for Claude Code / Cursor) to VS Code Copilot Chat
terminology. Any skill containing a banned term from the left column is a bug.

---

## Tool Name Mapping

| Claude Code Tool | VS Code Copilot Chat Tool | Notes |
|---|---|---|
| `Bash(command)` | `run_in_terminal` | Standard terminal execution |
| `Task("AgentName", prompt)` | `runSubagent("AgentName", prompt)` | Dispatches a named `.agent.md` agent; sequential |
| `TodoWrite` | `manage_todo_list` | Task tracking; persists within session |
| `AskUserQuestion(...)` | `vscode_askQuestions` | Structured Q&A with types options, multi-select |
| `Read(file)` | `read_file` | File read with line range support |
| `Edit(file, ...)` | `replace_string_in_file` | Single targeted replacement |
| `MultiEdit(file, ...)` | `multi_replace_string_in_file` | Multiple replacements per file |
| `Write(file, content)` | `create_file` | Create new file |
| `WebFetch(url)` | `fetch_webpage` | Web content retrieval |
| `glob(pattern)` | `file_search` | File find by glob pattern |
| `grep(pattern)` / `Grep(...)` | `grep_search` | Text search with regex support |
| `ls(dir)` | `list_dir` | Directory listing |
| (implicit) | `semantic_search` | Natural language code search |
| `get_terminal_output(id)` | `get_terminal_output` | Same name — direct port |
| (none) | `vscode_listCodeUsages` | Find all usages of a symbol |
| (none) | `get_errors` | Get compile/lint errors |

---

## Skill Invocation Mapping

| Claude Code / Superpowers | VS Code Copilot Chat / bof |
|---|---|
| `Skill("skill-name")` (explicit) | Description-based auto-trigger (implicit) |
| `"invoke superpowers:brainstorming"` | `"invoke bof:brainstorming"` (cross-reference) |
| `"superpowers:"` prefix | `"bof:"` prefix |

---

## Path / File Convention Mapping

| Superpowers Path | bof + Esquisse Path |
|---|---|
| `docs/superpowers/specs/YYYY-MM-DD-{topic}-design.md` | `docs/specs/YYYY-MM-DD-{topic}-design.md` |
| `docs/superpowers/plans/YYYY-MM-DD-{feature}.md` | `docs/tasks/P{n}-{nnn}-{slug}.md` (Esquisse task doc format) |
| `CLAUDE.md` references in skills | `AGENTS.md` (Esquisse convention) |
| `~/.config/superpowers/worktrees` | `~/.config/bof/worktrees` |

---

## Document Convention Mapping

| Superpowers | bof |
|---|---|
| `CLAUDE.md` | `AGENTS.md` |
| `"superpowers:code-reviewer"` agent | `SpecReviewerAgent` then `CodeQualityReviewerAgent` |
| Single monolithic plan file | Multiple Esquisse task docs per feature |
| `@import` directives for companion files | "See `{companion-file}.md` in this skill directory" |
| Visual companion (built-in server) | `napkin` skill (external) |

---

## Agent Frontmatter: Key Differences

| Concern | Claude Code | VS Code Copilot Chat |
|---|---|---|
| Model selection | `model:` string | `model:` array (priority list with `(copilot)` suffix) |
| Tool list | Different names | `read`, `search`, `write`, `execute/runInTerminal`, `execute/getTerminalOutput`, `vscode/memory` |
| Target platform | (implicit) | `target: vscode` |
| User-invocable | (implicit) | `user-invocable: true/false` |

### VS Code Model Name Reference (as of 2026)

| Model | `model:` array value |
|---|---|
| Claude Sonnet 4.6 | `'Claude Sonnet 4.6 (copilot)'` |
| Claude Opus 4.6 | `'claude-opus-4-6 (copilot)'` |
| Claude Haiku 4.5 | `'Claude Haiku 4.5 (copilot)'` |
| GPT-4.1 | `'gpt-4.1 (copilot)'` |
| GPT-4o | `'gpt-4o (copilot)'` |
| Gemini 3 Flash | `'Gemini 3 Flash (Preview) (copilot)'` |
| Auto (VS Code picks) | `'Auto (copilot)'` |

---

## Banned Terms Checklist

Before marking any skill task Done, scan for and eliminate these:

```
BANNED TERMS (grep for these):
  Bash(
  Task(
  TodoWrite
  AskUserQuestion
  CLAUDE.md
  Claude Code
  superpowers:
  docs/superpowers/plans/
  @import
  agentskills.io     (unless linking to reference — keep if harmless)
  "your human partner"    (replace with "the developer" or "your partner")
  visual-companion.md     (replace with napkin skill reference)
```

Run before each skill is marked complete:
```sh
grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion\|CLAUDE\.md\|superpowers:" skills/
```
