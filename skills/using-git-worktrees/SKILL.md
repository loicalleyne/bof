---
name: using-git-worktrees
description: >
  Use when starting feature work that needs isolation from the current workspace,
  or before executing implementation plans. Triggers on: "create a worktree",
  "set up isolated workspace", "start work on new feature branch", "before
  implementing the plan".
---

# Using Git Worktrees

Create isolated workspaces sharing the same repository, allowing work on
multiple branches simultaneously without switching.

**Core principle:** Systematic directory selection + safety verification = reliable isolation.

**Announce at start:** "I'm using `bof:using-git-worktrees` to set up an isolated workspace."

---

## Directory Selection Process

Follow this priority order:

### 1. Check Existing Directories

```sh
ls -d .worktrees 2>/dev/null     # Preferred (hidden)
ls -d worktrees 2>/dev/null      # Alternative
```

**If found:** Use that directory. If both exist, `.worktrees` wins.

### 2. Check AGENTS.md

```sh
grep -i "worktree.*director" AGENTS.md 2>/dev/null
```

**If preference specified:** Use it without asking.

### 3. Ask Developer

If no directory exists and no AGENTS.md preference:

```
No worktree directory found. Where should I create worktrees?

1. .worktrees/ (project-local, hidden)
2. ~/.config/bof/worktrees/<project-name>/ (global location)

Which would you prefer?
```

---

## Safety Verification

### For Project-Local Directories (.worktrees or worktrees)

**MUST verify directory is ignored before creating worktree:**

```sh
git check-ignore -q .worktrees 2>/dev/null || git check-ignore -q worktrees 2>/dev/null
```

**If NOT ignored:**

Fix immediately:
1. Add appropriate line to `.gitignore`
2. Commit the change
3. Proceed with worktree creation

**Why critical:** Prevents accidentally committing worktree contents to the repository.

### For Global Directory (~/.config/bof/worktrees)

No `.gitignore` verification needed — it's outside the project entirely.

---

## Creation Steps

### 1. Detect Project Name

```sh
project=$(basename "$(git rev-parse --show-toplevel)")
```

### 2. Create Worktree

```sh
# Determine full path based on location choice
case $LOCATION in
  .worktrees|worktrees)
    path="$LOCATION/$BRANCH_NAME"
    ;;
  ~/.config/bof/worktrees/*)
    path="$HOME/.config/bof/worktrees/$project/$BRANCH_NAME"
    ;;
esac

# Create worktree with new branch
git worktree add "$path" -b "$BRANCH_NAME"
cd "$path"
```

### 3. Copy AST Cache (if present)

```sh
# If code_ast.duckdb exists at project root, copy it to the worktree.
# This gives the implementer agent a stale-but-usable AST cache.
# The ImplementerAgent will rebuild it incrementally on task completion.
if [ -f "$(git rev-parse --show-toplevel)/code_ast.duckdb" ]; then
  cp "$(git rev-parse --show-toplevel)/code_ast.duckdb" "$path/code_ast.duckdb"
fi
```

### 4. Run Project Setup

Auto-detect and run appropriate setup:

```sh
# Go
if [ -f go.mod ]; then go mod download; fi

# Node.js
if [ -f package.json ]; then npm install; fi

# Python (uv)
if [ -f pyproject.toml ]; then uv sync; fi

# Rust
if [ -f Cargo.toml ]; then cargo build; fi
```

### 5. Verify Clean Baseline

Run the project test command (from AGENTS.md) to confirm the worktree starts clean:

```sh
# Use the test command from AGENTS.md — examples:
go test ./...
pytest
npm test
```

**If tests fail:** Report failures, ask whether to proceed or investigate first.

**If tests pass:** Report ready.

### 6. Report Location

```
Worktree ready at <full-path>
Tests passing (<N> tests, 0 failures)
AST cache: copied from main (stale — will be rebuilt incrementally)
Ready to implement <feature-name>
```

---

## Quick Reference

| Situation | Action |
|-----------|--------|
| `.worktrees/` exists | Use it (verify ignored) |
| `worktrees/` exists | Use it (verify ignored) |
| Both exist | Use `.worktrees/` |
| Neither exists | Check AGENTS.md → Ask developer |
| Directory not ignored | Add to `.gitignore` + commit |
| Tests fail during baseline | Report failures + ask |
| `code_ast.duckdb` exists | Copy to worktree |

---

## Common Mistakes

**Skipping ignore verification:**
- Problem: Worktree contents get tracked, pollute git status
- Fix: Always use `git check-ignore` before creating project-local worktree

**Assuming directory location:**
- Problem: Creates inconsistency, violates project conventions
- Fix: Follow priority: existing > AGENTS.md > ask

**Proceeding with failing tests:**
- Problem: Can't distinguish new bugs from pre-existing issues
- Fix: Report failures, get explicit permission to proceed

**Not copying AST cache:**
- Problem: ImplementerAgent has no AST assistance until a full rebuild
- Fix: Copy `code_ast.duckdb` at creation time (Step 3 above)

---

## Integration

**Called by:**
- [`bof:brainstorming`](../brainstorming/SKILL.md) — REQUIRED when design is approved and implementation follows
- [`bof:subagent-driven-development`](../subagent-driven-development/SKILL.md) — REQUIRED before executing any tasks
- [`bof:executing-plans`](../executing-plans/SKILL.md) — REQUIRED before executing any tasks

**Pairs with:**
- [`bof:finishing-a-development-branch`](../finishing-a-development-branch/SKILL.md) — REQUIRED for cleanup after work is complete
