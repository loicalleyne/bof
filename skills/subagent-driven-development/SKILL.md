---
name: subagent-driven-development
description: >
  Use when executing implementation plans with independent tasks in the current
  session. Triggers on: "execute the plan", "start implementation with subagents",
  "implement these tasks", "run subagent-driven development".
---

# Subagent-Driven Development

Execute a plan by dispatching a fresh subagent per task, with two-stage review
after each: spec compliance first, then code quality.

**Core principle:** Fresh subagent per task + two-stage review (spec then quality) = high quality, fast iteration.

**Announce at start:** "I'm using `bof:subagent-driven-development` to execute this plan."

> **Crush:** `runSubagent` is not available in Crush. Use the **bof-mcp tools** instead — `implementer_agent`, `spec_review`, `quality_review`. See the **Crush Mode** section at the bottom of this skill. The full per-task review loop is supported; only the call syntax differs.

---

## When to Use vs. Alternatives

**Use this skill when:**
- You have a written implementation plan (in `docs/tasks/` or a plan file)
- Tasks are mostly independent
- You want review checkpoints per task

**Use [`bof:executing-plans`](../executing-plans/SKILL.md) instead when:**
- Tasks are tightly coupled and need shared context
- You prefer inline single-session execution with no review checkpoints
- bof-mcp is not available

**Prerequisites:** Run [`bof:using-git-worktrees`](../using-git-worktrees/SKILL.md) first to create an isolated feature branch.

---

## Adversarial Review Guard

**Before dispatching the first ImplementerAgent**, check for a recent adversarial review:

```sh
# Check if a review exists and is dated after the plan was last modified:
ls -la .adversarial/reports/ 2>/dev/null || echo "No reviews found"
cat .adversarial/state.json 2>/dev/null || echo "No state found"
```

If no review report exists (or the plan has been modified since the last review):
invoke [`bof:adversarial-review`](../adversarial-review/SKILL.md) and wait for PASSED or CONDITIONAL verdict before
proceeding. A FAILED verdict requires revising the plan.

This check runs once — before the first task. Do not re-run between tasks.

---

## Setup

1. **Read all task documents** from `docs/tasks/` (or the plan file) upfront.
   Extract the full text of each task with its context.

2. **Create todo list** with `manage_todo_list` (VS Code) / `todos` (Crush). One item per task.

3. **Verify worktree is active** (`git branch` should show feature branch, not main).

---

## Per-Task Loop

For each task (in order, one at a time — no parallel implementation):

### Step 1: Dispatch ImplementerAgent

```
runSubagent("ImplementerAgent", taskPrompt)
```

The `taskPrompt` must include:
- Full task document text (Status, Goal, In Scope, Out of Scope, Files, Acceptance Criteria)
- Project AGENTS.md content (or path to it)
- Current working directory / project root
- Any cross-task context needed (e.g. "Task 2 depends on the types defined in Task 1")

**Never** make the implementer read the plan file itself — provide the text directly.

If ImplementerAgent asks questions: answer completely, then re-dispatch.

### Step 2: Handle ImplementerAgent Status

| Status | Action |
|--------|--------|
| `DONE` | Proceed to SpecReviewerAgent |
| `DONE_WITH_CONCERNS` | Read concerns. If about correctness/scope: resolve before review. If observation only: proceed. |
| `NEEDS_CONTEXT` | Provide missing context, re-dispatch ImplementerAgent |
| `BLOCKED` | See Handling BLOCKED below |

### Step 3: Dispatch SpecReviewerAgent

```
runSubagent("SpecReviewerAgent", specReviewPrompt)
```

The `specReviewPrompt` must include:
- Full task document text
- List of files that were changed (from ImplementerAgent's report)
- Git SHAs for the review range: `git log --oneline -5`

If SpecReviewerAgent returns **❌ issues:**
- Re-dispatch ImplementerAgent with the specific issues to fix
- Then re-dispatch SpecReviewerAgent
- Repeat until ✅ COMPLIANT

### Step 4: Dispatch CodeQualityReviewerAgent

**Only after SpecReviewerAgent returns ✅ COMPLIANT.**

```
runSubagent("CodeQualityReviewerAgent", qualityReviewPrompt)
```

If CodeQualityReviewerAgent returns **Critical** or **Important** issues:
- Re-dispatch ImplementerAgent with those specific issues
- Then re-dispatch CodeQualityReviewerAgent
- Repeat until ✅ APPROVED (Minor issues may be deferred)

### Step 5: Mark Task Complete

```
   manage_todo_list / todos({taskId: "task-slug", status: "completed"})
```

Then **update the task document directly** (the orchestrating session does this, not the subagent):
- Set `Status: Done`
- Add a Session Notes entry: what was implemented, any deviations from the plan, gotchas encountered
- If a new gotcha was found: add it to `AGENTS.md` Common Mistakes now, before the next task

Proceed to next task.

---

## Handling BLOCKED

When ImplementerAgent reports BLOCKED:

1. If it's a **context problem**: provide more context, re-dispatch
2. If the task requires **more reasoning**: note it and proceed to next task; blocked task gets a follow-up
3. If the task is **too large**: break it down and create sub-tasks
4. If the **plan itself is wrong**: escalate to the developer

Never ignore BLOCKED or force the same approach to retry without making a structural change.

---

## After All Tasks Complete

**Do not proceed to finishing-a-development-branch until all of the following are done, in order.**

### Step A: Esquisse Completion Protocol (mandatory if Esquisse docs present)

The orchestrating session performs these directly — do not delegate to a subagent:

1. **Update every task doc to `Status: Done`** — use `replace_string_in_file` on each task file. Use the task document schema from `SCHEMAS.md §3`. Add a Session Notes entry: date, what was implemented, deviations from plan, gotchas. Do not leave any task at `Ready` or `In Progress`.
2. **Update `AGENTS.md` Common Mistakes** — if any new gotcha was found during the session that was not already added per-task, add it now with the `[Category]` tag format.
3. **Update `GLOSSARY.md`** — if any new domain terms were introduced, add them using the existing alphabetical section format.
4. **Update `ROADMAP.md`** — change each completed task row to `Done`.
5. **Append to `docs/planning/NEXT_STEPS.md`** — add a session log entry (see `SCHEMAS.md §6` for format): date, tasks completed, open decisions, blocked items. Use `replace_string_in_file` to append; do not overwrite the file.

### Step B: Full Implementation Review

**VS Code:**
```
runSubagent("CodeQualityReviewerAgent", fullImplementationReviewPrompt)
```

**Crush (bof-mcp):**
```
quality_review(code_content: <full diff or changed file contents>, model: <model>)
```

Review the entire implementation across all tasks, not just the last task. If Critical or Important issues are found, fix them before proceeding.

### Step C: Finish the branch

**Invoke [`bof:finishing-a-development-branch`](../finishing-a-development-branch/SKILL.md)** to merge, PR, or clean up.

---

## Red Flags

**Never:**
- Start without adversarial review check
- Dispatch implementation subagent on main/master branch without explicit user consent
- Skip spec compliance review
- Proceed to code quality review before spec compliance is ✅
- Dispatch multiple ImplementerAgents in parallel (creates conflicts)
- Skip review loops when issues are found
- Accept "close enough" on spec compliance
- Accept DONE_WITH_CONCERNS without reading the concerns

---

## Prompt Templates

See in this directory:
- [`implementer-prompt.md`](implementer-prompt.md) — base context for ImplementerAgent dispatch
- [`spec-reviewer-prompt.md`](spec-reviewer-prompt.md) — base context for SpecReviewerAgent dispatch
- [`code-quality-reviewer-prompt.md`](code-quality-reviewer-prompt.md) — base context for CodeQualityReviewerAgent dispatch

**Document schemas** referenced during this skill:
- Task document format → [`SCHEMAS.md §3`](../../SCHEMAS.md)
- Session log / NEXT_STEPS entry format → [`SCHEMAS.md §6`](../../SCHEMAS.md)

## Crush Mode (bof-mcp)

> **VS Code users:** Use the native `runSubagent(...)` dispatch path above.
> This section is for Crush, where `runSubagent` is unavailable. bof-mcp provides equivalent dispatch via MCP tools.

### Tool mapping

| VS Code | bof-mcp tool | Key params |
|---|---|---|
| `runSubagent("ImplementerAgent", prompt)` | `implementer_agent` | `task_content` (full task doc text), `model` |
| `runSubagent("SpecReviewerAgent", prompt)` | `spec_review` | `spec_content` (task doc + changed files list), `model` |
| `runSubagent("CodeQualityReviewerAgent", prompt)` | `quality_review` | `code_content` (diff or changed file contents), `model` |
| [`bof:adversarial-review`](../adversarial-review/SKILL.md) guard | `gate_review` / `adversarial_review` | `plan_slug` matches `.adversarial/{slug}.json` |

### Per-task loop in Crush

**Step 1 — Implement:**
```
implementer_agent(
  task_content: "<full task document text including Goal, Scope, Files, Acceptance Criteria>",
  model: "<provider/model>"
)
```
Wait for completion. If the tool returns a BLOCKED or NEEDS_CONTEXT status, resolve and re-call.

**Step 2 — Spec review:**
```
spec_review(
  spec_content: "<task document text>\n\nChanged files:\n<list from implementer output>",
  model: "<provider/model>"
)
```
If issues found, fix inline and re-call until clean.

**Step 3 — Quality review:**
```
quality_review(
  code_content: "<git diff or full content of changed files>",
  model: "<provider/model>"
)
```
If Critical or Important issues found, fix and re-call.

**Adversarial guard (before first task):**
```
gate_review()                          # if adversarial_review already ran for this plan
# or
adversarial_review(
  plan_slug: "<slug matching .adversarial/{slug}.json>",
  plan_content: "<full plan text>"
)
```

All non-dispatch steps (todo tracking, task doc updates, AGENTS.md updates, ROADMAP.md) apply unchanged.
