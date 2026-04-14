# Implementer Prompt Template

Use this as the structure for `runSubagent("ImplementerAgent", prompt)` calls.
Fill in the bracketed sections from your current context.

---

```
You are implementing [TASK TITLE] for [PROJECT NAME].

## Task Document

[PASTE FULL TASK DOCUMENT TEXT HERE — Status, Goal, Background, In Scope, Out of Scope, Files table, Acceptance Criteria, Notes]

## Project Context

Project root: [ABSOLUTE PATH TO PROJECT ROOT]
AGENTS.md path: [PATH]/AGENTS.md

Key project constraints from AGENTS.md:
[PASTE RELEVANT SECTIONS — code conventions, common mistakes to avoid, invariants that apply to this task]

## Cross-Task Context

[Describe any dependencies from prior tasks, e.g.:]
"Task 1 (P{n}-{nnn}) defined the following interfaces that this task implements: ..."
"The following types were added in Task 1: ..."

## AST Cache

[If project has code_ast.duckdb:]
AST cache is available at: [PATH]/code_ast.duckdb
Use the duckdb-code skill for structural queries before editing files.

## Status Report Format

End your session with one of:
- `STATUS: DONE` — all acceptance criteria met, tests pass
- `STATUS: DONE_WITH_CONCERNS — [describe concern]` — done but note an issue
- `STATUS: NEEDS_CONTEXT — [what is needed]` — cannot proceed without more info
- `STATUS: BLOCKED — [describe blocker]` — fundamental obstacle
```
