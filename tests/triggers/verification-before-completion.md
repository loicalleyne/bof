# Trigger Test: verification-before-completion

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:verification-before-completion` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Verify the implementation is complete before marking this done.

Alternative triggers:
- "Run the verification checklist"
- "Verify before finishing this task"
- "Check that everything works before completing"

## Expected Behavior

1. Copilot activates the `verification-before-completion` skill.
2. Tests are run with `run_in_terminal`.
3. Build is verified clean.
4. Checklist confirms: tests pass, no TODO stubs, no broken imports.
5. Only after all checks pass does the task get marked complete.

## Pass Criteria

- Skill activates.
- `run_in_terminal` is used to run tests and build.
- Checklist is explicitly run through.
- No step is skipped.
- Task not marked complete until all checks pass.
