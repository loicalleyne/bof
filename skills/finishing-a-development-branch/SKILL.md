---
name: finishing-a-development-branch
description: >
  Use when implementation is complete, all tests pass, and you need to decide
  how to integrate the work. Triggers on: "implementation is complete", "merge
  this branch", "create a PR", "done with this feature", "clean up worktree".
---

# Finishing a Development Branch

Guide completion of development work by presenting clear options and handling
the chosen workflow.

**Core principle:** Verify tests → Present options → Execute choice → Clean up.

**Announce at start:** "I'm using `bof:finishing-a-development-branch` to complete this work."

---

## The Process

### Step 1: Verify Tests

**Before presenting options, verify tests pass:**

```sh
# Use the test command from AGENTS.md (examples):
go test ./...
pytest
npm test
cargo test
```

**If tests fail:**
```
Tests failing (<N> failures). Must fix before completing:

[Show failures]

Cannot proceed with merge/PR until tests pass.
```

Stop. Don't proceed to Step 2.

**If tests pass:** Continue to Step 2.

---

### Step 2: Determine Base Branch

```sh
git merge-base HEAD main 2>/dev/null || git merge-base HEAD master 2>/dev/null
```

Or ask: "This branch split from main — is that correct?"

---

### Step 3: Present Options

Present exactly these 4 options using `vscode_askQuestions`:

```
Implementation complete. What would you like to do?

1. Merge back to <base-branch> locally
2. Push and create a Pull Request
3. Keep the branch as-is (I'll handle it later)
4. Discard this work

Which option?
```

**Don't add explanation** — keep options concise.

---

### Step 4: Execute Choice

#### Option 1: Merge Locally

```sh
git checkout <base-branch>
git pull
git merge <feature-branch>

# Verify tests on merged result
<test command from AGENTS.md>

# If tests pass
git branch -d <feature-branch>
```

Then: Cleanup worktree (Step 5)

#### Option 2: Push and Create PR

```sh
git push -u origin <feature-branch>

gh pr create \
  --title "<title>" \
  --body "## Summary
- <bullet 1>
- <bullet 2>

## Test Plan
- [ ] <verification step>"
```

Then: Cleanup worktree (Step 5)

#### Option 3: Keep As-Is

Report: "Keeping branch `<name>`. Worktree preserved at `<path>`."

**Do NOT cleanup worktree.**

#### Option 4: Discard

**Confirm first via `vscode_askQuestions`:**
```
This will permanently delete:
- Branch <name>
- All commits: <commit-list>
- Worktree at <path>

Type 'discard' to confirm.
```

Wait for exact confirmation before proceeding.

If confirmed:
```sh
git checkout <base-branch>
git branch -D <feature-branch>
```

Then: Cleanup worktree (Step 5)

---

### Step 5: Cleanup Worktree

**For Options 1, 2, 4:**

Check if in a worktree:
```sh
git worktree list | grep "$(git branch --show-current)"
```

If yes:
```sh
git worktree remove <worktree-path>
```

**For Option 3:** Keep worktree.

---

## Quick Reference

| Option | Merge | Push | Keep Worktree | Remove Branch |
|--------|-------|------|---------------|---------------|
| 1. Merge locally | ✓ | — | — | ✓ |
| 2. Create PR | — | ✓ | ✓ | — |
| 3. Keep as-is | — | — | ✓ | — |
| 4. Discard | — | — | — | ✓ (force) |

---

## Common Mistakes

**Skipping test verification:**
- Problem: Merge broken code, create failing PR
- Fix: Always verify tests before offering options

**Open-ended questions:**
- Problem: "What should I do next?" → ambiguous
- Fix: Present exactly 4 structured options

**Automatic worktree cleanup:**
- Problem: Remove worktree when it might still be needed (Options 2, 3)
- Fix: Only cleanup for Options 1 and 4

**No confirmation for discard:**
- Problem: Accidentally delete work
- Fix: Require typed "discard" confirmation

---

## Red Flags

**Never:**
- Proceed with failing tests
- Merge without verifying tests on merged result
- Delete work without typed confirmation
- Force-push without explicit developer request

**Always:**
- Verify tests before offering options
- Present exactly 4 options
- Get typed confirmation for Option 4
- Clean up worktree for Options 1 & 4 only

---

## Integration

**Called by:**
- `bof:subagent-driven-development` — After all tasks complete
- `bof:executing-plans` — After all batches complete

**Pairs with:**
- `bof:using-git-worktrees` — Cleans up the worktree created by that skill
