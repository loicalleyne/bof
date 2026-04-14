# Trigger Test: using-git-worktrees

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:using-git-worktrees` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Create a git worktree for this task.

Alternative triggers:
- "Set up a worktree for working on this branch"
- "Use git worktrees to isolate this feature"
- "Create an isolated workspace using git worktrees"

## Expected Behavior

1. Copilot activates the `using-git-worktrees` skill.
2. A new branch is created: `git worktree add ../project-task-slug task/slug`.
3. AST cache is copied from the main worktree to the new one.
4. Work proceeds in the new worktree.
5. On completion, worktree is removed cleanly.

## Pass Criteria

- Skill activates.
- `run_in_terminal` used (not `Bash(...)`).
- AST copy step is present (Step 3 in skill).
- Worktree created outside the main repo directory.
- Cleanup step is included.
