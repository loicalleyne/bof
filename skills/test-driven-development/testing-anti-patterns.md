# Testing Anti-Patterns

**Load this reference when:** writing or changing tests, adding mocks, or tempted to add test-only methods to production code.

## Overview

Tests must verify real behavior, not mock behavior. Mocks are a means to isolate, not the thing being tested.

**Core principle:** Test what the code does, not what the mocks do.

**Following strict TDD prevents these anti-patterns.**

## The Iron Laws

```
1. NEVER test mock behavior
2. NEVER add test-only methods to production classes
3. NEVER mock without understanding dependencies
```

---

## Anti-Pattern 1: Testing Mock Behavior

**The violation:**
```go
// ❌ BAD: Testing that the mock exists
func TestRendersUserWidget(t *testing.T) {
    // using a FakeUserService that records calls
    _, err := RenderPage(fakeUserService)
    require.NoError(t, err)
    // This asserts the mock was called, not real behavior:
    require.True(t, fakeUserService.WasCalled())
}
```

**Why this is wrong:**
- You're verifying the mock works, not that the component works
- Test passes when mock is present, fails when it's not
- Tells you nothing about real behavior

**Gate function:**
```
BEFORE asserting on any mock element:
  Ask: "Am I testing real component behavior or just mock existence?"
  IF testing mock existence:
    STOP - Delete the assertion or unmock the component
    Test real behavior instead
```

---

## Anti-Pattern 2: Test-Only Methods in Production

**The violation:**
```go
// ❌ BAD: Destroy() only used in tests
type Session struct { ... }

func (s *Session) Destroy() error {
    // Dangerous if called in production!
    return s.store.Delete(s.id)
}

// In tests:
defer session.Destroy()
```

**Why this is wrong:**
- Production type polluted with test-only code
- Confuses object lifecycle with entity lifecycle
- Violates YAGNI and separation of concerns

**The fix:** Put cleanup in test utilities, not in the production type:
```go
// ✅ GOOD: Test utility handles cleanup
func cleanupSession(t *testing.T, store Store, id string) {
    t.Helper()
    require.NoError(t, store.Delete(id))
}
// In tests:
defer cleanupSession(t, store, session.ID)
```

**Gate function:**
```
BEFORE adding any method to a production type:
  Ask: "Is this method only called from test files?"
  IF yes: STOP — put it in a test utility instead
```

---

## Anti-Pattern 3: Mocking Without Understanding

**The violation:**
```go
// ❌ BAD: Mock breaks test logic
func TestDetectsDuplicateKey(t *testing.T) {
    // This mock prevents the side effect the test depends on!
    store := &mockStore{writeErr: nil}
    // store.Write is never called because we mocked it away...
    // so the duplicate check can't detect what was written before.
    err := AddKey(store, "key1")
    require.NoError(t, err)
    err = AddKey(store, "key1") // Should error — but won't
    require.Error(t, err)       // FAILS for the wrong reason
}
```

**The fix:** Mock at the correct level. Understand what the test depends on before mocking.

```
Gate function:
BEFORE mocking any method:
  STOP - Don't mock yet
  1. Ask: "What side effects does the real method have?"
  2. Ask: "Does this test depend on any of those side effects?"
  3. Run test with real implementation FIRST, observe behavior
  THEN add minimal mocking at the right level

  Red flags:
    - "I'll mock this to be safe"
    - "This might be slow, better mock it"
    - Mocking without understanding the dependency chain
```

---

## Anti-Pattern 4: Incomplete Mocks

**The violation:**
```go
// ❌ BAD: Partial mock — only fields you think you need
type mockResponse struct {
    Status string
    UserID string
    // Missing: metadata that downstream code uses!
}
```

**Why this is wrong:** Tests pass but integration fails when code accesses missing fields.

**The fix:** Mirror the real structure completely.

```go
// ✅ GOOD: Match the real API structure
type mockResponse struct {
    Status   string
    UserID   string
    Metadata ResponseMetadata // Include ALL fields the system may consume
}
```

---

## Anti-Pattern 5: Tests as Afterthought

```
✅ Implementation complete
❌ No tests written
"Ready for testing" ← VIOLATION
```

Testing is part of implementation, not optional follow-up. TDD would have caught this. Can't claim complete without tests.

---

## Red Flags

- Assertion checks for `*-mock` or `*-fake` identifiers
- Methods only called in test files
- Mock setup is >50% of the test body
- Test fails when you remove a mock (you're testing the mock)
- Can't explain why a mock is needed
- Mocking "just to be safe"

---

## Quick Reference

| Anti-Pattern | Fix |
|---|---|
| Assert on mock elements | Test real component or unmock it |
| Test-only methods in production | Move to test utilities |
| Mock without understanding | Understand dependencies first, mock minimally |
| Incomplete mocks | Mirror real API completely |
| Tests as afterthought | TDD — tests first |
| Over-complex mocks | Consider integration tests instead |

---

## The Bottom Line

**Mocks are tools to isolate, not things to test.**

If TDD reveals you're testing mock behavior, you've gone wrong.

Fix: Test real behavior, or question why you're mocking at all.
