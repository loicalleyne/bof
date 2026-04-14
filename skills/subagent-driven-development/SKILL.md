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

---

## When to Use vs. Alternatives

**Use this skill when:**
- You have a written implementation plan (in `docs/tasks/` or a plan file)
- Tasks are mostly independent
- You want review checkpoints per task

**Use `bof:executing-plans` instead when:**
- Tasks are tightly coupled and need shared context
- You prefer a single-session batch execution

**Prerequisites:** Run `bof:using-git-worktrees` first to create an isolated feature branch.

---

## Adversarial Review Guard

**Before dispatching the first ImplementerAgent**, check for a recent adversarial review:

```sh
# Check if a review exists and is dated after the plan was last modified:
ls -la .adversarial/reports/ 2>/dev/null || echo "No reviews found"
cat .adversarial/state.json 2>/dev/null || echo "No state found"
```

If no review report exists (or the plan has been modified since the last review):
invoke `bof:adversarial-review` and wait for PASSED or CONDITIONAL verdict before
proceeding. A FAILED verdict requires revising the plan.

This check runs once — before the first task. Do not re-run between tasks.

---

## Setup

1. **Read all task documents** from `docs/tasks/` (or the plan file) upfront.
   Extract the full text of each task with its context.

2. **Create todo list** with `manage_todo_list`. One item per task.

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
manage_todo_list({taskId: "task-slug", status: "completed"})
```

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

1. **Esquisse completion protocol** (if Esquisse docs present):
   - Verify all task docs show `Status: Done`
   - `AGENTS.md` Common Mistakes updated (if any)
   - `GLOSSARY.md` updated (if new terms)
   - `ROADMAP.md` task statuses updated
   - `docs/planning/NEXT_STEPS.md` session log updated

2. **Full implementation review:**
   ```
   runSubagent("CodeQualityReviewerAgent", fullImplementationReviewPrompt)
   ```
   Review the entire implementation, not just the last task.

3. **Invoke `bof:finishing-a-development-branch`** to merge, PR, or clean up.

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
- `implementer-prompt.md` — base context for ImplementerAgent dispatch
- `spec-reviewer-prompt.md` — base context for SpecReviewerAgent dispatch
- `code-quality-reviewer-prompt.md` — base context for CodeQualityReviewerAgent dispatch
