# Condition-Based Waiting

## Overview

Flaky tests often guess at timing with arbitrary delays. This creates race
conditions where tests pass on fast machines but fail under load or in CI.

**Core principle:** Wait for the actual condition you care about, not a guess
about how long it takes.

---

## When to Use

**Use when:**
- Tests have arbitrary delays (`time.Sleep`, `setTimeout`, `sleep`)
- Tests are flaky (pass sometimes, fail under load)
- Tests timeout when run in parallel
- Waiting for async operations to complete

**Don't use when:**
- Testing actual timing behavior (debounce intervals, ticker periods)
- If you must use an arbitrary timeout, always document WHY

---

## Core Pattern

```go
// ❌ BEFORE: Guessing at timing
time.Sleep(50 * time.Millisecond)
result := getResult()
require.NotNil(t, result)

// ✅ AFTER: Waiting for condition
err := waitFor(func() bool { return getResult() != nil }, "result to appear", 5*time.Second)
require.NoError(t, err)
result := getResult()
require.NotNil(t, result)
```

---

## Quick Patterns

| Scenario | Pattern |
|----------|---------|
| Wait for event | `waitFor(func() bool { return findEvent(events, "DONE") != nil }, ...)` |
| Wait for state | `waitFor(func() bool { return machine.State() == "ready" }, ...)` |
| Wait for count | `waitFor(func() bool { return len(items) >= 5 }, ...)` |
| Wait for file | `waitFor(func() bool { _, err := os.Stat(path); return err == nil }, ...)` |
| Complex condition | `waitFor(func() bool { return obj.Ready() && obj.Value() > 10 }, ...)` |

---

## Implementation

Generic polling helper:

```go
// waitFor polls condition until it returns true or timeout is reached.
func waitFor(condition func() bool, description string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for {
        if condition() {
            return nil
        }
        if time.Now().After(deadline) {
            return fmt.Errorf("timeout waiting for %s after %v", description, timeout)
        }
        time.Sleep(10 * time.Millisecond) // Poll every 10ms
    }
}
```

---

## Common Mistakes

**Polling too fast:** `time.Sleep(time.Millisecond)` — wastes CPU
**Fix:** Poll every 10ms

**No timeout:** Loop forever if condition never met
**Fix:** Always include timeout with a clear error message

**Stale data:** Caching state before loop, checking cached value inside
**Fix:** Call getter inside the loop for fresh data

---

## When Arbitrary Timeout IS Correct

```go
// Tool ticks every 100ms — need 2 ticks to verify partial output
err := waitForState(manager, "TOOL_STARTED", 5*time.Second) // First: wait for condition
time.Sleep(200 * time.Millisecond) // Then: wait for timed behavior
// 200ms = 2 ticks at 100ms intervals — documented and justified
```

**Requirements for using arbitrary timeout:**
1. First wait for the triggering condition
2. Based on known timing interval (not guessing)
3. Comment explaining WHY the specific duration

---

## Real-World Impact

Condition-based waiting typically:
- Fixes flaky tests: pass rate 60% → 100%
- Makes test suite 30-40% faster (no unnecessary waiting)
- Eliminates race conditions on slow machines / CI
