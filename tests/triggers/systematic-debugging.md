# Trigger Test: systematic-debugging

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:systematic-debugging` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Help me debug this issue systematically.

Alternative triggers:
- "Debug this error using systematic debugging"
- "Trace the root cause of this bug"
- "Something is broken — let's debug it step by step"

## Expected Behavior

1. Copilot activates the `systematic-debugging` skill.
2. `root-cause-tracing.md` protocol is applied.
3. Hypothesis → test → confirm loop is followed.
4. Defense-in-depth checks are applied after fix.
5. A regression test is written before the fix is applied.

## Pass Criteria

- Skill activates.
- Root cause is identified before fix is applied.
- Regression test written first.
- `run_in_terminal` used to test hypotheses.
- `bof:test-driven-development` is referenced for the regression test step.
