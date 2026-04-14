# Trigger Test: test-driven-development

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:test-driven-development` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Use test-driven development to implement this feature.

Alternative triggers:
- "Write tests first, then implement X"
- "Follow TDD to build this"
- "Red-green-refactor workflow for X"

## Expected Behavior

1. Copilot activates the `test-driven-development` skill.
2. Failing test is written first (red).
3. Minimal implementation to pass (green).
4. Refactor while keeping tests green.
5. `testing-anti-patterns.md` is consulted.

## Pass Criteria

- Skill activates.
- Test is written before implementation code.
- `run_in_terminal` is used to run tests (not `Bash(...)`).
- Anti-patterns checklist is referenced.
