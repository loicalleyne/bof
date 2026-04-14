---
name: test-driven-development
description: >
  Use when implementing any feature or bugfix, before writing implementation
  code. Triggers on: "implement this function", "write this feature", "fix this
  bug", "add this method", any request to produce production code.
---

# Test-Driven Development (TDD)

## Overview

Write the test first. Watch it fail. Write minimal code to pass.

**Core principle:** If you didn't watch the test fail, you don't know if it tests the right thing.

**Violating the letter of the rules is violating the spirit of the rules.**

---

## When to Use

**Always:**
- New features
- Bug fixes
- Refactoring
- Behavior changes

**Exceptions (ask the developer):**
- Throwaway prototypes
- Generated code
- Configuration files

Thinking "skip TDD just this once"? Stop. That's rationalization.

---

## The Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

Write code before the test? Delete it. Start over.

**No exceptions:**
- Don't keep it as "reference"
- Don't "adapt" it while writing tests
- Don't look at it
- Delete means delete

Implement fresh from tests. Period.

---

## Red-Green-Refactor

### RED — Write Failing Test

Write one minimal test showing what should happen.

**Good test (Go):**
```go
func TestRetryOperation_SucceedsAfterThreeAttempts(t *testing.T) {
    attempts := 0
    op := func() error {
        attempts++
        if attempts < 3 {
            return errors.New("fail")
        }
        return nil
    }
    err := RetryOperation(op, 3)
    require.NoError(t, err)
    require.Equal(t, 3, attempts)
}
```

Requirements:
- One behavior per test
- Clear, descriptive name
- Tests real code behavior (no mocks unless unavoidable)

### Verify RED — Watch It Fail

**MANDATORY. Never skip.**

```sh
# From AGENTS.md test command — use the project's actual test command:
go test -v -run TestRetryOperation_SucceedsAfterThreeAttempts ./...
```

Confirm:
- Test fails (not compile errors)
- Failure message is expected
- Fails because feature is missing (not a typo)

**Test passes?** You're testing existing behavior. Fix the test.
**Compile errors?** Fix them, re-run until the test runs and fails correctly.

### GREEN — Minimal Code

Write the simplest code that passes the test. Nothing more.

**Good:**
```go
func RetryOperation(fn func() error, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
    }
    return err
}
```

Don't add features, refactor other code, or "improve" beyond what the test requires.

### Verify GREEN — Watch It Pass

**MANDATORY.**

```sh
go test -v -run TestRetryOperation_SucceedsAfterThreeAttempts ./...
# All tests still passing:
go test ./...
```

Confirm:
- Test passes
- Other tests still pass
- Output is clean (no errors, no warnings)

**Test fails?** Fix code, not test.
**Other tests fail?** Fix them now.

### REFACTOR — Clean Up

After green only:
- Remove duplication
- Improve names
- Extract helpers

Keep tests green. Don't add behavior.

### Repeat

Next failing test for next feature.

---

## Flow

```
Write one failing test
        ↓
Run test → Does it fail correctly? → NO → Fix the test → re-run
                                  → YES
        ↓
Write minimal implementation
        ↓
Run test → Does it pass? → NO → Fix implementation → re-run
                         → YES
        ↓
Run all tests → Still passing? → NO → Fix it now
                               → YES
        ↓
Refactor (optional), keeping tests green
        ↓
[Next test]
```

---

## Example: Bug Fix

**Bug:** Empty email accepted as valid.

**RED**
```go
func TestValidateEmail_RejectsEmpty(t *testing.T) {
    err := ValidateEmail("")
    require.EqualError(t, err, "email required")
}
```

**Verify RED**
```sh
$ go test -v -run TestValidateEmail_RejectsEmpty ./...
--- FAIL: TestValidateEmail_RejectsEmpty
    expected error "email required", got nil
```

**GREEN**
```go
func ValidateEmail(email string) error {
    if strings.TrimSpace(email) == "" {
        return errors.New("email required")
    }
    return nil
}
```

**Verify GREEN**
```sh
$ go test -v -run TestValidateEmail_RejectsEmpty ./...
--- PASS: TestValidateEmail_RejectsEmpty
```

**REFACTOR** — extract if other fields need similar validation.

---

## Testing Anti-Patterns

See `testing-anti-patterns.md` in this skill directory for common pitfalls:
- Testing mock behavior instead of real behavior
- Adding test-only methods to production classes
- Mocking without understanding dependencies

---

## Common Rationalizations

| Excuse | Reality |
|--------|---------|
| "Too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after" | Tests passing immediately prove nothing. |
| "Tests after achieve same goals" | Tests-after = "what does this do?" Tests-first = "what should this do?" |
| "Already manually tested" | Ad-hoc ≠ systematic. No record, can't re-run. |
| "Deleting X hours is wasteful" | Sunk cost fallacy. Keeping unverified code is technical debt. |
| "Keep as reference, write tests first" | You'll adapt it. That's testing after. Delete means delete. |
| "TDD will slow me down" | TDD is faster than debugging. Pragmatic = test-first. |

---

## Red Flags — STOP and Start Over

- Code before test
- Test added after implementation
- Test passes immediately without implementation
- Can't explain why the test failed
- "I already manually tested it"
- "Tests after achieve the same purpose"
- "Keep as reference" or "adapt existing code"
- "This is different because..."

**All of these mean: Delete code. Start over with TDD.**

---

## Verification Checklist

Before marking work complete:

- [ ] Every new function/method has a test
- [ ] Watched each test fail before implementing
- [ ] Each test failed for expected reason (feature missing, not typo)
- [ ] Wrote minimal code to pass each test
- [ ] All tests pass
- [ ] Output clean (no errors, no warnings)
- [ ] Tests use real code (mocks only if unavoidable)
- [ ] Edge cases and errors covered

Can't check all boxes? You skipped TDD. Start over.

---

## When Stuck

| Problem | Solution |
|---------|----------|
| Don't know how to test | Write wished-for API. Write assertion first. Ask the developer. |
| Test too complicated | Design too complicated. Simplify interface. |
| Must mock everything | Code too coupled. Use dependency injection. |
| Test setup is huge | Extract helpers. Still complex? Simplify design. |

---

## Final Rule

```
Production code → test exists and failed first
Otherwise → not TDD
```

No exceptions without the developer's explicit permission.
