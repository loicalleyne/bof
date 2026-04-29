# P9-001: Minimal-Editing Diff Enforcement — ImplementerAgent

## Status
<!-- One of: Draft | Ready | In Progress | In Review | Done | Blocked -->
Status: Done
Depends on: none
Blocks: none

## Summary
Adds a mandatory diff-review step to ImplementerAgent's Implementation Rules and TDD Cycle commit step, enforcing that every changed line is traceable to a specific In Scope item before it is committed. Prevents "while I was there" over-editing.

## Problem
ImplementerAgent already says "preserve existing code" but gives no enforcement mechanism. Agents routinely change lines for style, rename variables, or clean up surrounding code when implementing a fix — none of which are required by the task. These changes inflate diffs, complicate reviews, and introduce unintended regressions. There is no step that forces the agent to check its own diff before committing.

## Solution
Add two targeted additions to `agents/ImplementerAgent.agent.md` and its embedded copy `bof-mcp/embedded/agents/ImplementerAgent.agent.md`:

1. A new bullet in `## Implementation Rules` that names the diff-review procedure with exact git commands.
2. Expand TDD Cycle step 6 from "Commit" to "Diff-check, then commit" with the explicit check and revert commands.

## Scope
### In Scope
- Add diff-review bullet to `## Implementation Rules` in `agents/ImplementerAgent.agent.md`
- Expand TDD Cycle step 6 in `agents/ImplementerAgent.agent.md`
- Apply identical changes to `bof-mcp/embedded/agents/ImplementerAgent.agent.md`

### Out of Scope
- Changes to `skills/implement-task/SKILL.md` or its references (P9-002)
- Updating `bof-mcp/embedded/agents/ImplementerAgent.agent.md` for other out-of-sync sections (Completion Protocol Tier 1/Tier 2 split) — separate task
- Adding a `minimal-editing.md` reference document (P9-002)
- Changes to any reviewer agent or other agent file

## Prerequisites
- [ ] No blocking tasks

## Specification

### Change 1: New bullet in `## Implementation Rules`

In both files, after the existing "Preserve existing code." paragraph, insert:

```markdown
- **Review your diff before every commit.** Run `git diff` (unstaged) and
  `git diff --cached` (staged) before committing. For every changed hunk ask:
  "is this change required by a specific item in the In Scope list OR is it
  a necessary structural consequence of a required change?" Necessary
  structural consequences are allowed: added/removed imports, gofmt-required
  blank lines, cascading type changes in interfaces, test helper signature
  updates. Everything else — style fixes, variable renames, dead-code removal,
  incidental refactors — must be reverted. To revert staged hunks:
  `git restore --staged {file}`. To revert unstaged hunks when the file
  contains ONLY unnecessary changes: `git checkout -- {file}`. If the file
  contains both required and unnecessary unstaged changes with nothing staged
  yet, do NOT use `git checkout -- {file}` — manually edit the file to remove
  only the unnecessary hunks, then stage the result with `git add -p {file}`.
```

### Change 2: Expand TDD Cycle step 6

Locate the `## TDD Cycle (per feature/function)` section. The current step 6 reads:

```
6. Commit
```

Replace with:

```
6. **Diff-check, then commit:** run `git diff --cached`, verify every staged
   hunk traces to a specific In Scope item, unstage anything that doesn't
   (`git restore --staged {file}`), then commit with a semantic message.
```

### Sync requirement

Both files must have byte-identical content for these two sections after the change. The implementer must apply the changes to `agents/ImplementerAgent.agent.md` first, then apply the identical edits to `bof-mcp/embedded/agents/ImplementerAgent.agent.md`. After both edits, verify the sections match:

```sh
diff \
  <(grep -A6 "Review your diff" agents/ImplementerAgent.agent.md) \
  <(grep -A6 "Review your diff" bof-mcp/embedded/agents/ImplementerAgent.agent.md)
# must produce no output
```

## Files

| File | Action | Description |
|------|--------|-------------|
| `agents/ImplementerAgent.agent.md` | Modify | Add diff-review bullet to Implementation Rules; expand TDD Cycle step 6 |
| `bof-mcp/embedded/agents/ImplementerAgent.agent.md` | Modify | Identical changes to keep embedded copy in sync |

## Design Principles
- **Minimal change to the agent instructions.** Add exactly two insertions; do not restructure surrounding text.
- **Exact commands, not descriptions.** The agent must see `git diff --cached`, `git restore --staged`, `git checkout --` — not vague "check your diff."
- **Sync is mandatory.** The embedded copy is what bof-mcp embeds into its binary. An out-of-sync copy is a bug; the sync diff check is part of acceptance criteria.

## Testing Strategy
- Verify by grep: both files contain "git diff --cached" in the Implementation Rules section.
- Verify by grep: TDD Cycle step 6 in both files contains "Diff-check".
- Run the sync diff command above — must produce no output.
- Run the bof invariant check: `grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion" agents/` — must return 0.

## Acceptance Criteria
- [ ] `grep -c "git diff --cached" agents/ImplementerAgent.agent.md` returns `>= 1`
- [ ] `grep -c "Diff-check" agents/ImplementerAgent.agent.md` returns `>= 1`
- [ ] `grep -c "git diff --cached" bof-mcp/embedded/agents/ImplementerAgent.agent.md` returns `>= 1`
- [ ] `grep -c "Diff-check" bof-mcp/embedded/agents/ImplementerAgent.agent.md` returns `>= 1`
- [ ] `diff <(grep -A6 "Review your diff" agents/ImplementerAgent.agent.md) <(grep -A6 "Review your diff" bof-mcp/embedded/agents/ImplementerAgent.agent.md)` — produces no output
- [ ] `grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion" agents/ImplementerAgent.agent.md` — 0 results

## Session Notes
<!-- Append-only. Never overwrite. -->
<!-- 2026-04-28 — EsquissePlan — Planned. Two-file change: main agent + embedded copy. Sync diff check is the key acceptance criterion. -->
<!-- 2026-04-28 — Completed. Added diff-review bullet after 'Preserve existing code.' paragraph and expanded TDD Cycle step 6 to 'Diff-check, then commit' in both ImplementerAgent files; sync diff confirmed no output. -->
