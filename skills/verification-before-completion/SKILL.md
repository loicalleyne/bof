---
name: verification-before-completion
description: >
  Use when about to claim work is complete, fixed, or passing. Required
  before committing, creating PRs, or marking any task done. Triggers on:
  "done", "complete", "fixed", "tests pass", "ready to commit", "all good",
  any success or completion claim.
---

# Verification Before Completion

## Overview

Claiming work is complete without verification is dishonesty, not efficiency.

**Core principle:** Evidence before claims, always.

**Violating the letter of this rule is violating the spirit of this rule.**

---

## The Iron Law

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE
```

If you haven't run the verification command in this message, you cannot claim it passes.

---

## The Gate Function

```
BEFORE claiming any status or expressing satisfaction:

1. IDENTIFY: What command proves this claim?
2. RUN: Execute the FULL command (fresh, complete) via `run_in_terminal` (VS Code) / `bash` (Crush)
3. READ: Full output, check exit code, count failures
4. VERIFY: Does output confirm the claim?
   - If NO: State actual status with evidence
   - If YES: State claim WITH evidence
5. ONLY THEN: Make the claim

Skip any step = lying, not verifying
```

---

## Common Failures

| Claim | Requires | Not Sufficient |
|-------|----------|----------------|
| Tests pass | Test command output: 0 failures | Previous run, "should pass" |
| Linter clean | Linter output: 0 errors | Partial check, extrapolation |
| Build succeeds | Build command: exit 0 | Linter passing, logs look good |
| Bug fixed | Test original symptom: passes | Code changed, assumed fixed |
| Regression test works | Red-green cycle verified | Test passes once |
| Agent completed | VCS diff shows changes | Agent reports "success" |
| Requirements met | Line-by-line checklist | Tests passing |

---

## Key Patterns

**Tests:**
```sh
# Run test command from AGENTS.md — use the project's actual command
go test ./...
# See: ok [package] — all packages. Then claim "all tests pass."
```

**Regression tests (TDD Red-Green):**
```
✅ Write test → Run (PASS) → Revert fix → Run (MUST FAIL) → Restore fix → Run (PASS)
❌ "I've written a regression test" (without running the red-green cycle)
```

**Build:**
```sh
go build ./...
# See: exit 0. Then claim "build succeeds."
# NOT: "linter passed" (linter doesn't check compilation)
```

**Requirements:**
```
✅ Re-read plan/task doc → Create checklist → Verify each item → Report gaps or completion
❌ "Tests pass, task complete" (without checking every acceptance criterion)
```

**Agent delegation:**
```
✅ Agent reports success → Check git diff → Verify changes exist → Report actual state
❌ Trust agent report without independent verification
```

---

## Red Flags — STOP

- Using "should", "probably", "seems to"
- Expressing satisfaction before verification ("Great!", "Perfect!", "Done!", etc.)
- About to commit/push/PR without verification
- Trusting agent success reports without checking
- Relying on partial verification
- Thinking "just this once"
- **ANY wording implying success without having run verification**

---

## Rationalization Prevention

| Excuse | Reality |
|--------|---------|
| "Should work now" | RUN the verification |
| "I'm confident" | Confidence ≠ evidence |
| "Just this once" | No exceptions |
| "Linter passed" | Linter ≠ compiler |
| "Agent said success" | Verify independently |
| "Partial check is enough" | Partial proves nothing |
| "Different words so rule doesn't apply" | Spirit over letter |

---

## When To Apply

**ALWAYS before:**
- Any variation of success/completion claims
- Any expression of satisfaction
- Committing, PR creation, task completion
- Moving to next task
- Delegating to agents

**Rule applies to:**
- Exact phrases
- Paraphrases and synonyms
- Implications of success
- ANY communication suggesting completion/correctness

---

## The Bottom Line

**No shortcuts for verification.**

Run the command. Read the output. THEN make the claim.

This is non-negotiable.
