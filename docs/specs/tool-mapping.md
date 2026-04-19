# Tool Mapping: Claude Code → VS Code Copilot Chat → Crush

This document is the canonical reference for translating upstream `superpowers`
skill content (authored for Claude Code / Cursor) to VS Code Copilot Chat
terminology. Any skill containing a banned term from the left column is a bug.

The Crush column documents equivalents for
[charm.land/crush](https://github.com/charmbracelet/crush) — skills symlinked
into `~/.config/crush/skills/` are loaded by Crush and the LLM should use the
Crush tool name when running inside Crush.

---

## Tool Name Mapping

| Claude Code Tool | VS Code Copilot Chat Tool | Crush Tool | Notes |
|---|---|---|---|
| `Bash(command)` | `run_in_terminal` | `bash` | Standard terminal execution |
| `Task("AgentName", prompt)` | `runSubagent("AgentName", prompt)` | *(inline — no subagent dispatch)* | VS Code only; Crush implements inline |
| `TodoWrite` | `manage_todo_list` | `todos` | Task tracking; persists within session |
| `AskUserQuestion(...)` | `vscode_askQuestions` | *(ask inline in response)* | VS Code only; Crush asks as plain text |
| `Read(file)` | `read_file` | `view` | File read with line range support |
| `Edit(file, ...)` | `replace_string_in_file` | `edit` | Single targeted replacement |
| `MultiEdit(file, ...)` | `multi_replace_string_in_file` | `multiedit` | Multiple replacements per file |
| `Write(file, content)` | `create_file` | `write` | Create/overwrite file |
| `WebFetch(url)` | `fetch_webpage` | `web_fetch` | Web content retrieval |
| `glob(pattern)` | `file_search` | `glob` | File find by glob pattern |
| `grep(pattern)` / `Grep(...)` | `grep_search` | `grep` | Text search with regex support |
| `ls(dir)` | `list_dir` | `ls` | Directory listing |
| (implicit) | `semantic_search` | `grep` | Natural language → text search in Crush |
| `get_terminal_output(id)` | `get_terminal_output` | `job_output` | Background job output |
| (none) | `vscode_listCodeUsages` | `lsp_references` | Find all usages of a symbol |
| (none) | `get_errors` | `lsp_diagnostics` | Get compile/lint/LSP errors |

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

---

## Crush Tool Quick-Reference

When writing skills that should work in both VS Code and Crush, use
`tool-name` (VS Code) / `tool-name` (Crush) notation inline, or a
`**Tools (VS Code / Crush):**` header in the Prerequisites section.

| Scenario | VS Code | Crush |
|---|---|---|
| Run a shell command | `run_in_terminal` | `bash` |
| Read a file | `read_file` | `view` |
| Edit a file (targeted replace) | `replace_string_in_file` | `edit` |
| Multiple replacements | `multi_replace_string_in_file` | `multiedit` |
| Create / overwrite file | `create_file` | `write` |
| Search by filename | `file_search` | `glob` |
| Search file contents | `grep_search` | `grep` |
| List directory | `list_dir` | `ls` |
| Fetch a web page | `fetch_webpage` | `web_fetch` |
| Task tracking | `manage_todo_list` | `todos` |
| Compile/lint errors | `get_errors` | `lsp_diagnostics` |
| Find symbol usages | `vscode_listCodeUsages` | `lsp_references` |
| Structured Q&A with user | `vscode_askQuestions` | *(ask inline as plain text)* |
| Dispatch named subagent | `runSubagent("Name", prompt)` | *(implement inline)* |
