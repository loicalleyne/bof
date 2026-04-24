---
name: writing-skills
description: Use when creating new bof skills, editing existing skills, or verifying skills work before deployment. Triggers on "create a skill", "write a SKILL.md", "test this skill", "improve this skill".
---

# Writing Skills

## Overview

**Writing skills IS Test-Driven Development applied to process documentation.**

**bof skills live in `~/.copilot/skills/{name}/SKILL.md`** — installed via `bof/scripts/install.sh`.

You write test cases (pressure scenarios with subagents), watch them fail (baseline behavior without the skill), write the skill (documentation), watch tests pass (agents comply), and refactor (close loopholes).

**Core principle:** If you didn't watch an agent fail without the skill, you don't know if the skill teaches the right thing.

**REQUIRED BACKGROUND:** You MUST understand [`bof:test-driven-development`](../test-driven-development/SKILL.md) before using this skill. That skill defines the fundamental RED-GREEN-REFACTOR cycle. This skill adapts TDD to documentation.

---

## What is a Skill?

A **skill** is a reference guide for proven techniques, patterns, or tools. Skills help future agent instances find and apply effective approaches.

**Skills are:** Reusable techniques, patterns, tools, reference guides.

**Skills are NOT:** Narratives about how you solved a problem once, code templates, project-specific conventions (put those in `AGENTS.md`).

---

## TDD Mapping for Skills

| TDD Concept | Skill Creation |
|-------------|----------------|
| **Test case** | Pressure scenario with subagent |
| **Production code** | Skill document (SKILL.md) |
| **Test fails (RED)** | Agent violates rule without skill (baseline) |
| **Test passes (GREEN)** | Agent complies with skill present |
| **Refactor** | Close loopholes while maintaining compliance |
| **Write test first** | Run baseline scenario BEFORE writing skill |
| **Watch it fail** | Document exact rationalizations agent uses |
| **Minimal code** | Write skill addressing those specific violations |
| **Watch it pass** | Verify agent now complies |
| **Refactor cycle** | Find new rationalizations → plug → re-verify |

---

## SKILL.md Structure

**Frontmatter (YAML) — VS Code Copilot Chat:**

VS Code parses **only `name` and `description`**. All other frontmatter fields are silently ignored.

- `name`: letters, numbers, and hyphens only; ≤64 chars
- `description`: ≤1024 chars — **the only auto-trigger mechanism**. Third-person; describes ONLY **when to use** (NOT what it does)
  - Start with "Use when…" to focus on triggering conditions
  - Include specific symptoms, situations, and contexts
  - Include "DO NOT USE when…" guards — there is no separate `not:` field; embed them here
  - **NEVER summarize the skill's process or workflow** — description summarizing workflow creates a shortcut the agent will take instead of reading the full skill body
  - Keep under 500 characters if possible

```yaml
---
name: skill-name-with-hyphens
description: >
  Use when [specific triggering conditions and symptoms]. Include
  the exact phrases a user would type to invoke this workflow.
  DO NOT USE when [overlapping domain — name the correct skill].
---
```

**Charmbracelet Crush** additionally parses three optional fields:

```yaml
---
name: skill-name-with-hyphens
description: >
  Use when …
license: MIT                                    # optional: SPDX identifier
compatibility: "VS Code Copilot Chat 1.115+ / Crush"    # optional: ≥00 chars
metadata:                                       # optional: map[string]string
  category: workflow
---
```

**Fields that are NOT supported in either runtime** (silently dropped — do not add):
`triggers:`, `not:`, `tools_required:`, `updated:`

Document required tools in `## Prerequisites` body using dual-platform notation:
`**Tools (VS Code / Crush):** \`read_file\`/\`view\`, \`create_file\`/\`write\``
Document last-updated date as a footer: `*Last updated: YYYY-MM-DD*`

**Body structure:**
```markdown
# Skill Name

## Overview
Core principle in 1-3 sentences.

## When to Use
- Symptom-based triggers
- When NOT to use

## Core Pattern / Process
Steps, flowchart only if decision is non-obvious.

## Quick Reference
Table or bullets for scanning.

## Common Mistakes
What goes wrong + fixes.
```

---

## Tool Name Rule (bof invariant)

Before marking any skill Done, scan for banned Claude Code / Cursor tool names:

```
BANNED (grep for these in any skill you write):
  Bash(
  Task(
  TodoWrite
  AskUserQuestion
  CLAUDE.md
  Claude Code
  superpowers:
  docs/superpowers/plans/
  @import (do not use @import directives)
  "your human partner"
```

Replace with VS Code Copilot Chat equivalents from `docs/specs/tool-mapping.md`.

```sh
# Verify before marking Done:
grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion\|CLAUDE\.md\|superpowers:" skills/
```

---

## Installation (bof skills)

bof skills are installed via symlinks from `/bof/skills/{name}/` → `~/.copilot/skills/{name}/`.

```sh
# After creating or updating a skill:
bash bof/scripts/install.sh        # creates/updates symlinks (idempotent)
# Open a NEW VS Code Copilot Chat session — existing sessions do NOT reload
```

The `description:` field is matched against user messages at the start of each
response. A new VS Code session is required after any change to `description:`.

---

## The Iron Law (Same as TDD)

```
NO SKILL WITHOUT A FAILING TEST FIRST
```

This applies to NEW skills AND EDITS to existing skills.

Write skill before testing? Delete it. Start over.

**No exceptions:**
- Not for "simple additions"
- Not for "just adding a section"
- Not for bof tool-name updates (still test that the skill triggers and the agent follows the VS Code process)

---

## RED-GREEN-REFACTOR for Skills

### RED: Baseline (no skill present)

Run a pressure scenario with a subagent WITHOUT the skill loaded. Document:
- What choices did the agent make?
- What rationalizations did it use (verbatim)?
- Which pressures triggered violations?

### GREEN: Write Minimal Skill

Write the skill addressing those specific rationalizations. Don't add content for hypothetical cases.

Run same scenarios WITH skill. Agent should now comply.

### REFACTOR: Close Loopholes

Agent found new rationalization? Add explicit counter. Re-test until bulletproof.

---

## Skill Creation Checklist

Use `manage_todo_list` (VS Code) / `todos` (Crush) to track each item.

**RED Phase — Write Failing Test:**
- [ ] Create 3+ pressure scenarios (for discipline-enforcing skills)
- [ ] Run scenarios WITHOUT skill — document baseline behavior verbatim
- [ ] Identify patterns in rationalizations/failures

**GREEN Phase — Write Minimal Skill:**
- [ ] `name:` uses only letters, numbers, hyphens
- [ ] YAML frontmatter has `name:` and `description:`
- [ ] `description:` starts with "Use when..."
- [ ] `description:` written in third person
- [ ] No banned tool names anywhere in the skill body
- [ ] `bof:` namespace for all cross-skill references (not `superpowers:`)
- [ ] All cross-references use skill name only: "invoke `bof:test-driven-development`"
- [ ] Run scenarios WITH skill — verify agents comply

**REFACTOR Phase — Close Loopholes:**
- [ ] Build rationalization table from all test iterations
- [ ] Add explicit counters for each rationalization
- [ ] Create Red Flags list if discipline-enforcing
- [ ] Re-test until bulletproof

**Quality Checks:**
- [ ] `grep -rn "Bash(\|Task(\|CLAUDE\.md\|superpowers:" skills/{name}/` returns 0
- [ ] No `@import` directives
- [ ] Verify trigger: paste `description:` text into Copilot Chat → skill activates
- [ ] Install: `bash scripts/install.sh` → verify link in `~/.copilot/skills/`

---

## Cross-Referencing Other Skills

Use the `bof:` namespace:

```markdown
✅ Good: "**REQUIRED BACKGROUND:** You MUST understand [`bof:test-driven-development`](../test-driven-development/SKILL.md)"
✅ Good: "Invoke [`bof:systematic-debugging`](../systematic-debugging/SKILL.md)"
❌ Bad: "See skills/systematic-debugging/SKILL.md"
❌ Bad: "superpowers:test-driven-development"
❌ Bad: "@skills/test-driven-development/SKILL.md"  (@ syntax force-loads and burns context)
```

---

## When to Create a Skill

**Create when:**
- Technique wasn't intuitively obvious
- You'd reference this again across projects
- Pattern applies broadly (not project-specific)

**Don't create for:**
- One-off solutions
- Standard practices well-documented elsewhere
- Project-specific conventions (put in `AGENTS.md`)

---

## The Bottom Line

Creating bof skills IS TDD for process documentation.

Same Iron Law: No skill without failing test first.
Same cycle: RED (baseline) → GREEN (write skill) → REFACTOR (close loopholes).

If you follow TDD for code, follow it for skills. It's the same discipline applied to documentation.
