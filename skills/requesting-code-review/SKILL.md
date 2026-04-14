---
name: requesting-code-review
description: >
  Use when completing tasks, implementing major features, or before merging to
  verify work meets requirements. Triggers on: "request code review", "review
  my changes", "review this implementation", "code review before merge".
---

# Requesting Code Review

Dispatch a code review subagent to catch issues before they cascade. The
reviewer gets precisely crafted context for evaluation — never your session
history. This keeps the reviewer focused on the work product, not your thought
process, and preserves your own context for continued work.

**Core principle:** Review early, review often.

---

## When to Request Review

**Mandatory:**
- After each task in `bof:subagent-driven-development`
- After completing a major feature
- Before merge to main

**Optional but valuable:**
- When stuck (fresh perspective)
- Before refactoring (baseline check)
- After fixing a complex bug

---

## How to Request

**1. Get git SHAs:**
```sh
BASE_SHA=$(git rev-parse origin/main)   # or the merge base
HEAD_SHA=$(git rev-parse HEAD)
git log --oneline "$BASE_SHA".."$HEAD_SHA"
```

**2. Dispatch code reviewer:**

```
runSubagent("CodeQualityReviewerAgent", reviewPrompt)
```

Fill in the template at [skills/requesting-code-review/code-reviewer.md](../requesting-code-review/code-reviewer.md).

**Placeholders:**
- `{WHAT_WAS_IMPLEMENTED}` — what you just built
- `{PLAN_OR_REQUIREMENTS}` — what it should do (task doc path or inline spec)
- `{BASE_SHA}` — starting commit
- `{HEAD_SHA}` — ending commit

**3. Act on feedback:**
- Fix Critical issues immediately
- Fix Important issues before proceeding
- Acknowledge Minor issues (defer or fix)
- Push back if reviewer is wrong (with technical reasoning)

---

## Integration with Workflows

**Subagent-Driven Development:**
- Review after each task via the SDD loop (already built in)
- Direct use of this skill: for ad-hoc mid-task reviews

**Executing Plans:**
- Review after each batch (3-4 tasks)
- Apply fixes, then continue

**Ad-Hoc Development:**
- Review before merge
- Review when stuck on a bug

---

## Red Flags

**Never:**
- Skip review because "it's simple"
- Ignore Critical issues
- Proceed with unfixed Important issues
- Dismiss pushback without technical reasoning

**If reviewer is wrong:**
- Push back with technical reasoning
- Show tests or code that prove the current approach works
- Request clarification on the reviewer's reasoning

---

See review prompt template at: [skills/requesting-code-review/code-reviewer.md](../requesting-code-review/code-reviewer.md)
