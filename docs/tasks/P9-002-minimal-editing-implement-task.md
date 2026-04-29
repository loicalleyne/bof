# P9-002: Minimal-Editing Reference Doc + implement-task Skill Update

## Status
<!-- One of: Draft | Ready | In Progress | In Review | Done | Blocked -->
Status: Done
Depends on: none
Blocks: none

## Summary
Creates `skills/implement-task/references/minimal-editing.md` as a standalone reference document encoding the minimal-editing principle and its enforcement procedure. Updates `skills/implement-task/SKILL.md` to add a diff-check step between Step 7 and Step 8. Updates `skills/implement-task/references/completion-protocol.md` to add the same pre-commit check.

## Problem
`skills/implement-task/SKILL.md` has no diff-review step. The "Follow Edit Order" and "Run Tests After Each Logical Unit" steps guide implementation, but there is no step that forces the agent to audit its own diff before committing. Over-editing goes undetected and unreverted. There is also no canonical reference document encoding the minimal-editing principle for the bof project.

## Solution
1. Create `skills/implement-task/references/minimal-editing.md` — a concise reference encoding: the core principle, the enforcement procedure (git commands), anti-patterns, and when to revert.
2. Add a new **Step 7b** to `skills/implement-task/SKILL.md` immediately after Step 7 (Run Tests After Each Logical Unit): "Diff-check before commit."
3. Add the same diff-check step to `skills/implement-task/references/completion-protocol.md` in the implementation loop section (before the commit implied by "Task Document" step 1).

## Scope
### In Scope
- Create `skills/implement-task/references/minimal-editing.md`
- Add Step 7b (diff-check) to `skills/implement-task/SKILL.md` between existing Step 7 and Step 8
- Add diff-check note to `skills/implement-task/references/completion-protocol.md`

### Out of Scope
- Changes to `agents/ImplementerAgent.agent.md` (P9-001)
- Changes to `bof-mcp/embedded/agents/ImplementerAgent.agent.md` (P9-001)
- Changes to reviewer agents (SpecReviewerAgent, CodeQualityReviewerAgent, Adversarial-r*)
- Updating `bof-mcp/embedded/` with any reference doc copy — the reference doc is for VS Code skill use only

## Prerequisites
- [ ] No blocking tasks

## Specification

### New file: `skills/implement-task/references/minimal-editing.md`

Full content:

```markdown
# Minimal Editing

Source: https://nrehiew.github.io/blog/minimal_editing/

## Core Principle

Every line you change must be either:
- Directly required by a specific item in the task's In Scope list, **or**
- A necessary structural consequence of a required change.

If neither applies, revert the line.

**Necessary structural consequences** (always allowed):
- Import statements added or removed because the changed code needs them
- `gofmt`/`goimports`-required formatting (blank lines between declarations, etc.)
- Interface method changes cascading from a modified type
- Test helper signature updates required by a changed function signature
- YAML/Markdown table alignment changes caused by adding a required row

## Enforcement Procedure

Before every commit, run:

\`\`\`sh
git diff           # see all unstaged changes
git diff --cached  # see staged changes
\`\`\`

For every changed hunk, ask: "which In Scope item requires this, or is this a necessary structural consequence?"

- Yes to either: keep the hunk.
- No to both: revert it.

\`\`\`sh
git restore --staged {file}    # unstage a file (working tree preserved)
git checkout -- {file}         # revert ALL unstaged changes in a file (only use when the
                               # file contains NOTHING required in unstaged changes)
\`\`\`

**Mixed-unstaged files:** If a file has both required and unnecessary changes with nothing staged yet, do NOT use `git checkout -- {file}`. Manually edit the file to remove only the unnecessary hunks, then stage the required changes with `git add -p {file}`.

## Anti-Patterns (revert these unless they are necessary structural consequences)

| Pattern | Example | Why it's a bug |
|---------|---------|----------------|
| Style normalization | Fixing indentation in a function you didn't touch | Inflates diff, hides real changes |
| Variable rename | Renaming `err` to `taskErr` "for clarity" | Not required; adds noise to review |
| Dead code removal | Deleting a commented-out block | Out of scope unless the task says so |
| Incidental refactor | Extracting a helper "since I was here anyway" | Introduces risk; belongs in a dedicated task |
| Import reordering | Resorting imports in a file you modified | Not required; `goimports` handles this at CI |
| Blank line normalization | Adding/removing blank lines "for readability" | Out of scope unless gofmt requires it |
| Docstring rewrite | Expanding a comment you didn't need to touch | Out of scope |
| YAML/Markdown whitespace | Normalising table column widths in unrelated sections | Out of scope |
| Trailing comma / semicolon | Adding trailing commas "for consistency" | Out of scope |
| Line ending change | CRLF → LF in a file you modified | Out of scope (fix in a dedicated chore commit) |

## What Counts as In Scope

Only changes explicitly listed in the task's `## In Scope` bullets. If the
In Scope list says "implement `Foo()`", only lines inside or directly called
by `Foo()` are in scope. Adjacent functions, test helpers not named in
Acceptance Criteria, and surrounding comments are out of scope unless
explicitly listed.

## The Exception for Blocking Pre-Existing Bugs

If a pre-existing bug is directly blocking the acceptance criterion (compilation
error, panicking test), fix it minimally and note it as:
```
// ASSUMPTION: fixed pre-existing bug in {function} — required for {criterion}
```
This is a structural consequence exception, not a license to refactor.
```

### Change to `skills/implement-task/SKILL.md`

After the current **Step 7 — Run Tests After Each Logical Unit** section and before **Step 8 — Bail-Out Trigger**, insert a new step:

```markdown
#### Step 7b — Diff-Check Before Commit

Before committing any logical unit, review the diff:

```sh
git diff --cached
```

For every changed hunk ask: "which item in the In Scope list requires this change?"

- If you can name it: keep it.
- If you cannot name it: revert it (`git restore --staged {file}` or `git checkout -- {file}`).

See `skills/implement-task/references/minimal-editing.md` for the full anti-pattern list.

A clean commit contains only lines directly traceable to an In Scope item.
```

### Change to `skills/implement-task/references/completion-protocol.md`

In **Section 1 (Task Document)**, before the "Set `Status: Done`" bullet, insert:

```markdown
- Before updating Status: run `git diff HEAD` on the task's files and verify no
  over-editing is present. Revert any hunk not traceable to an In Scope item.
  See `references/minimal-editing.md`.
```

## Files

| File | Action | Description |
|------|--------|-------------|
| `skills/implement-task/references/minimal-editing.md` | Create | Canonical minimal-editing reference: principle, enforcement procedure, anti-patterns |
| `skills/implement-task/SKILL.md` | Modify | Add Step 7b (diff-check) between Step 7 and Step 8 |
| `skills/implement-task/references/completion-protocol.md` | Modify | Add diff-check note before Status update in Section 1 |

## Design Principles
- **Reference doc is self-contained.** An agent reading only `minimal-editing.md` must understand what to do without any other context.
- **Anti-pattern table is the most important section.** Concrete examples are more effective than abstract rules for preventing LLM over-editing.
- **Step numbering continuity.** Insert Step 7b (not renumber Step 8 to 9) to avoid breaking cross-references in other documents that cite step numbers.
- **Exact git commands.** No vague "check your changes." The exact commands `git diff --cached`, `git restore --staged`, `git checkout --` must appear.

## Testing Strategy
- Verify by grep: `skills/implement-task/references/minimal-editing.md` exists and contains "Anti-Patterns".
- Verify by grep: `skills/implement-task/SKILL.md` contains "Step 7b" and "git diff --cached".
- Verify by grep: `skills/implement-task/references/completion-protocol.md` contains "over-editing" or "minimal-editing.md".
- Run bof invariant check: `grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion" skills/implement-task/` — 0 results.

## Acceptance Criteria
- [ ] `test -f skills/implement-task/references/minimal-editing.md` exits 0
- [ ] `grep -c "Anti-Pattern" skills/implement-task/references/minimal-editing.md` returns `>= 1`
- [ ] `grep -c "git diff --cached" skills/implement-task/references/minimal-editing.md` returns `>= 1`
- [ ] `grep -c "Step 7b" skills/implement-task/SKILL.md` returns `>= 1`
- [ ] `grep -c "git diff --cached" skills/implement-task/SKILL.md` returns `>= 1`
- [ ] `grep -c "minimal-editing\|over-editing" skills/implement-task/references/completion-protocol.md` returns `>= 1`
- [ ] `grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion" skills/implement-task/` — 0 results

## Session Notes
<!-- Append-only. Never overwrite. -->
<!-- 2026-04-28 — EsquissePlan — Planned. Three-file change: new reference doc + two skill modifications. Step 7b insertion chosen over Step 8 renumber to preserve cross-references. -->
<!-- 2026-04-28 — Completed. Created minimal-editing.md with principle, enforcement procedure, 10-row anti-pattern table, and structural-consequence exceptions; added Step 7b to SKILL.md; added diff-check note to completion-protocol.md. All acceptance criteria passed. -->
