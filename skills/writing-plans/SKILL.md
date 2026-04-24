---
name: writing-plans
description: >
  Use when you have a spec or requirements for a multi-step task, before
  touching any code. Triggers on: "write an implementation plan", "plan this
  feature", "create tasks for", "break this down into steps".
---

# Writing Plans

## Overview

Write comprehensive implementation plans. Document everything needed: which files to touch for each task, code, testing, docs, how to test it. Give the whole plan as bite-sized tasks. DRY. YAGNI. TDD. Frequent commits.

**Announce at start:** "I'm using the `bof:writing-plans` skill to create the implementation plan."

**Save plans to:** `docs/tasks/P{n}-{nnn}-{slug}.md` (Esquisse task doc format — one file per logical unit).
- If `docs/tasks/` does not exist: create it, or use `docs/<feature>-plan.md` as fallback.
- For Esquisse projects: each logical unit becomes a separate `docs/tasks/P{n}-{nnn}-{slug}.md` following the Esquisse task doc schema (Status, Goal, In Scope, Out of Scope, Files, Acceptance Criteria, Session Notes).

---

## Scope Check

If the spec covers multiple independent subsystems, suggest breaking into separate plans — one per subsystem. Each plan should produce working, testable software on its own.

---

## File Structure Mapping

Before defining tasks, map which files will be created or modified and what each is responsible for.

**If `code_ast.duckdb` exists (do this first):**

Use the `duckdb-code` skill to answer these questions before writing any task:
- \"What files define types related to `{feature area}`?\" \u2014 find the integration points
- \"What calls `{function to change}`?\" \u2014 find all call sites that will need updating
- \"What implements `{interface to extend}`?\" \u2014 find all types that must be updated
- \"What does `{package}` export?\" \u2014 verify the API surface before designing around it

Record results in the Files table. This preserves context budget for the
planning itself and prevents missed call sites in the Files table.

**If no AST cache:** use `grep_search`/`grep` + `file_search`/`glob` to locate integration points, then read only file headers and interface sections.

---

## Task Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" — step
- "Run it to make sure it fails" — step
- "Implement the minimal code to pass the test" — step
- "Run the tests and verify they pass" — step
- "Commit" — step

---

## Esquisse Task Doc Format

For Esquisse projects, each task doc has this structure:

```markdown
# P{n}-{nnn}: {slug}

**Status:** Ready
**Phase:** {n}
**Goal:** One sentence describing the observable outcome.

## In Scope
- Exact list of what changes (file names, function names)

## Out of Scope
- At least 2-3 explicit exclusions

## Files
| File | Action | What |
|------|--------|------|
| `path/to/file.go` | Create | New type X |
| `path/to/other.go` | Modify | Add method Y |

## Acceptance Criteria
- [ ] `TestFunctionBehavior` passes
- [ ] `TestEdgeCase` passes
- [ ] `go build ./...` succeeds

## Session Notes
<!-- Append dated entries here after each work session -->
```

---

## Non-Esquisse Plan Header

For non-Esquisse projects, every plan starts with:

```markdown
# [Feature Name] Implementation Plan

> **For agentic workers:** REQUIRED SKILL: Use [`bof:subagent-driven-development`](../subagent-driven-development/SKILL.md)
> (recommended) or [`bof:executing-plans`](../executing-plans/SKILL.md) to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

---
```

---

## No Placeholders

Every step must contain the actual content needed. These are **plan failures** — never write:
- "TBD", "TODO", "implement later", "fill in details"
- "Add appropriate error handling" / "add validation" / "handle edge cases"
- "Write tests for the above" (without actual test code)
- "Similar to Task N" (repeat the code — tasks may be read out of order)
- Steps that describe what to do without showing how
- References to types or functions not defined in any task

---

## Self-Review

After writing the complete plan, run this checklist (inline — no subagent dispatch):

1. **Spec coverage:** Skim each requirement in the spec. Can you point to a task that implements it? List gaps.
2. **Placeholder scan:** Search the plan for red flags from "No Placeholders" above. Fix them.
3. **Type consistency:** Do type names, method signatures, and property names match between tasks? Inconsistencies between Task 3 and Task 7 are bugs.

If issues found, fix inline. If a spec requirement has no task, add the task.

---

## Adversarial Review Gate (REQUIRED before execution)

After saving the plan and the self-review passes, run [`bof:adversarial-review`](../adversarial-review/SKILL.md):

> "Plan is complete and self-review passed. Running adversarial review before execution."

**Do NOT dispatch [`bof:subagent-driven-development`](../subagent-driven-development/SKILL.md) or [`bof:executing-plans`](../executing-plans/SKILL.md) until the adversarial review verdict is PASSED or CONDITIONAL.**

- FAILED verdict → revise the plan addressing the reviewer's objections → re-run [`bof:adversarial-review`](../adversarial-review/SKILL.md)
- CONDITIONAL verdict → address blocking conditions, confirm resolution, then proceed
- PASSED verdict → proceed to execution handoff

---

## Execution Handoff

After adversarial review passes, offer execution choice:

> "Plan complete. Adversarial review: [VERDICT]. Two execution options:
>
> **1. Subagent-Driven (recommended)** — Fresh subagent per task, spec + quality review between tasks, fast iteration.
>
> **2. Inline Execution** — Execute tasks in this session using `bof:executing-plans`, batch execution with checkpoints.
>
> Which approach?"

Use `vscode_askQuestions` (VS Code) / ask inline (Crush) with options `["Subagent-Driven (recommended)", "Inline Execution"]`.

**If Subagent-Driven chosen:** Invoke [`bof:subagent-driven-development`](../subagent-driven-development/SKILL.md).
**If Inline Execution chosen:** Invoke [`bof:executing-plans`](../executing-plans/SKILL.md).
