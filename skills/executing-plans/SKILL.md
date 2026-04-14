---
name: executing-plans
description: >
  Use when you have a written plan and want to execute it task by task without
  dispatching separate subagents. Triggers on: "execute the plan", "work through
  these tasks", "execute this plan yourself", "implement tasks one by one".
---

# Executing Plans

Execute a written plan by working through tasks in sequence within the current
session. Use this when tasks are tightly coupled, need shared context, or when
direct execution is preferred over subagent dispatch.

**Announce at start:** "I'm using `bof:executing-plans` to work through this plan."

---

## When to Use vs. Alternatives

**Use this skill when:**
- Tasks are tightly coupled (task 2 needs output from task 1)
- You need to maintain context across the whole implementation
- The plan has fewer than ~5 tasks

**Use `bof:subagent-driven-development` instead when:**
- Tasks are mostly independent
- You want review checkpoints per task (recommended for larger plans)
- You want to leverage specific models per task type

**Prerequisites:** Run `bof:using-git-worktrees` first to create an isolated feature branch.

---

## Setup

1. **Load the plan.** Read all task documents from `docs/tasks/` (or the plan file if non-Esquisse).

2. **Review the plan** before starting. Verify:
   - Tasks are in a sensible order
   - Dependencies are clear
   - Acceptance criteria are measurable

3. **Create todo list** with `manage_todo_list`. One item per task.

4. **Verify worktree** — run `git branch` and confirm you are on the feature branch.

---

## Execution Loop

For each task, in order:

1. **Mark in progress:**
   ```
   manage_todo_list({taskId: "...", status: "in-progress"})
   ```

2. **Follow the task steps.** Work through the task's implementation steps exactly.
   - Verify your understanding of "In Scope" and "Out of Scope" before coding
   - Run the project test command at intervals (from AGENTS.md)
   - Use `bof:test-driven-development` if tests are required

3. **Verify completion.** For each Acceptance Criterion:
   - Confirm it is satisfied
   - If test-based: run the specific test and confirm green

4. **Mark complete:**
   ```
   manage_todo_list({taskId: "...", status: "completed"})
   ```

---

## Stop Conditions

Stop execution and surface to the developer when:

- **Blocker hit:** Something in the environment or codebase prevents progress
- **Plan gap:** The plan says to do X but X is underdefined or contradictory
- **Unclear instruction:** After two interpretations, you are still unsure what was meant
- **Repeated verification failure:** A test is red and three attempts have not fixed it — do not thrash

When stopping: explain what was completed, what failed, and what the developer needs to decide.

---

## Completion

After all tasks are complete:

1. **Run the full test suite** for the project (from AGENTS.md).

2. **Esquisse completion protocol** (if `docs/tasks/` format is present):
   - Verify all task docs show `Status: Done`
   - Update `docs/planning/NEXT_STEPS.md` with session log
   - Update `docs/planning/ROADMAP.md` task statuses

3. **Invoke `bof:finishing-a-development-branch`** to merge, PR, or clean up.
