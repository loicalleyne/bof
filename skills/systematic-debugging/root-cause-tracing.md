# Root Cause Tracing

## Overview

Bugs often manifest deep in the call stack (wrong directory, wrong value, wrong
state). Your instinct is to fix where the error appears, but that's treating a
symptom.

**Core principle:** Trace backward through the call chain until you find the
original trigger, then fix at the source.

---

## When to Use

**Use when:**
- Error happens deep in execution (not at entry point)
- Stack trace shows long call chain
- Unclear where invalid data originated
- Need to find which test/code triggers the problem

---

## The Tracing Process

### 1. Observe the Symptom
```
Error: git init failed in /path/to/project/packages/core
```

### 2. Find Immediate Cause

What code directly causes this?
```go
// In Go:
cmd := exec.Command("git", "init")
cmd.Dir = projectDir  // ← projectDir is empty
```

### 3. Ask: What Called This?

```
createWorktree(projectDir, sessionID)
  → called by Session.InitializeWorkspace()
  → called by Session.Create()
  → called by test at Project.Create()
```

### 4. Keep Tracing Up

What value was passed?
- `projectDir = ""` (empty string!)
- Empty string as `Dir` resolves to `os.Getwd()`
- That's the source code directory!

### 5. Find Original Trigger

Where did the empty string come from?
```go
ctx := setupCoreTest() // Returns ctx with TempDir: ""
Project.Create("name", ctx.TempDir) // Accessed before setup!
```

**Root cause:** Variable initialization accessing empty value before setup ran.

---

## Adding Instrumentation

When you can't trace manually, add temporary diagnostic output:

```go
// Before the problematic operation:
func gitInit(directory string) error {
    _, file, line, _ := runtime.Caller(1)
    log.Printf("DEBUG git init: dir=%q cwd=%q caller=%s:%d",
        directory, mustGetwd(), file, line)
    // ... proceed
}
```

**In tests:** Use `fmt.Fprintf(os.Stderr, ...)` — test loggers may not show at failure.

**Run and capture:**
```sh
go test ./... 2>&1 | grep 'DEBUG git init'
```

**Analyze output:**
- Look for test file names in stack
- Find line number triggering the call
- Identify pattern (same test? same parameter?)

---

## Key Principle

**NEVER fix just where the error appears.** Trace back to find the original trigger.

Tracing chain:
1. Find immediate cause
2. Can trace one level up? → yes: keep tracing
3. Is this the source? → no: keep tracing
4. Found source → fix here
5. Also add defense-in-depth validation at each layer

---

## Stack Trace Tips

- **In tests:** Use `fmt.Fprintf(os.Stderr, ...)` — logger may be suppressed
- **Before operation:** Log before the dangerous operation, not after it fails
- **Include context:** Directory, cwd, environment variables, timestamps
- **In Go:** Use `debug.Stack()` or `runtime.Callers()` for programmatic stack capture
