# Trigger Test: adversarial-review

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:adversarial-review` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Run adversarial review on this plan.

Alternative triggers:
- "Send this plan through adversarial review"
- "Critically review this task document"
- "Apply adversarial review before implementation"

## Expected Behavior

1. Copilot activates the `adversarial-review` skill.
2. Rotation state is read from `.adversarial/state.json`.
3. The appropriate Adversarial-r* agent is dispatched: `runSubagent("Adversarial-r0", ...)`.
4. Verdict (PASSED / CONDITIONAL / FAILED) is written to `.adversarial/state.json`.
5. Verdict gates progression to `bof:subagent-driven-development`.

## Pass Criteria

- Skill activates.
- Correct rotation slot is selected (0 → 1 → 2 → 0...).
- `runSubagent("Adversarial-r*", prompt)` is used.
- Verdict is written to `.adversarial/state.json`.
- FAILED verdict blocks implementation dispatch.
