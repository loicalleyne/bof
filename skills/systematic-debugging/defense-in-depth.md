# Defense-in-Depth Validation

## Overview

When you fix a bug caused by invalid data, adding validation at one place feels
sufficient. But that single check can be bypassed by different code paths,
refactoring, or mocks.

**Core principle:** Validate at EVERY layer data passes through. Make the bug
structurally impossible.

---

## Why Multiple Layers

Single validation: "We fixed the bug"
Multiple layers: "We made the bug impossible"

Different layers catch different cases:
- Entry validation catches most bugs
- Business logic catches edge cases
- Environment guards prevent context-specific dangers
- Debug logging helps when other layers fail

---

## The Four Layers

### Layer 1: Entry Point Validation

**Purpose:** Reject obviously invalid input at API boundary

```go
func CreateProject(name, workingDirectory string) (*Project, error) {
    if workingDirectory == "" {
        return nil, fmt.Errorf("createProject: workingDirectory cannot be empty")
    }
    info, err := os.Stat(workingDirectory)
    if err != nil {
        return nil, fmt.Errorf("createProject: workingDirectory does not exist: %s", workingDirectory)
    }
    if !info.IsDir() {
        return nil, fmt.Errorf("createProject: workingDirectory is not a directory: %s", workingDirectory)
    }
    // ... proceed
}
```

### Layer 2: Business Logic Validation

**Purpose:** Ensure data makes sense for this operation

```go
func InitializeWorkspace(projectDir, sessionID string) error {
    if projectDir == "" {
        return fmt.Errorf("initializeWorkspace: projectDir required")
    }
    // ... proceed
}
```

### Layer 3: Environment Guards

**Purpose:** Prevent dangerous operations in specific contexts

```go
func gitInit(directory string) error {
    // In tests, refuse git init outside temp directories
    if os.Getenv("GO_TEST_ENV") == "1" {
        tmpDir := os.TempDir()
        if !strings.HasPrefix(filepath.Clean(directory), filepath.Clean(tmpDir)) {
            return fmt.Errorf("refusing git init outside temp dir during tests: %s", directory)
        }
    }
    // ... proceed
}
```

### Layer 4: Debug Instrumentation

**Purpose:** Capture context for forensics

```go
func gitInit(directory string) error {
    _, file, line, _ := runtime.Caller(1)
    log.Printf("About to git init: dir=%q cwd=%q caller=%s:%d",
        directory, mustGetwd(), file, line)
    // ... proceed
}
```

---

## Applying the Pattern

When you find a bug:

1. **Trace the data flow** — Where does the bad value originate? Where is it used?
2. **Map all checkpoints** — List every point data passes through
3. **Add validation at each layer** — Entry, business, environment, debug
4. **Test each layer** — Try to bypass layer 1, verify layer 2 catches it

---

## Key Insight

All four layers are necessary. During testing, each layer catches bugs the
others miss:
- Different code paths bypass entry validation
- Mocks bypass business logic checks
- Edge cases on different platforms need environment guards
- Debug logging identifies structural misuse

**Don't stop at one validation point.** Add checks at every layer.
