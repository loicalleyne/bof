# Trigger Test: finishing-a-development-branch

## Purpose

This file documents the trigger phrase and expected behavior for
the `bof:finishing-a-development-branch` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Finish this development branch and prepare it for merge.

Alternative triggers:
- "This branch is done — prepare it for merge"
- "Clean up and finish the feature branch"
- "I'm ready to merge — help me finish the branch"

## Expected Behavior

1. Copilot activates the `finishing-a-development-branch` skill.
2. `bof:verification-before-completion` is run first.
3. `vscode_askQuestions` confirms merge target and strategy.
4. Branch is tidied (commit squash optional, PR description drafted).
5. PR is opened or merge is executed.

## Pass Criteria

- Skill activates.
- Verification is run before any merge action.
- `vscode_askQuestions` used for merge confirmation.
- `run_in_terminal` used for git commands.
- No `Bash(...)` tool names in output.
