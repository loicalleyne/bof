# Trigger Test: requesting-code-review

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:requesting-code-review` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Request a code review for my changes.

Alternative triggers:
- "Review the code I just wrote"
- "Send my implementation for code review"
- "Run code review on this branch"

## Expected Behavior

1. Copilot activates the `requesting-code-review` skill.
2. `runSubagent("CodeQualityReviewerAgent", prompt)` is dispatched with a review prompt built from `code-reviewer.md`.
3. Review output is summarized and presented.

## Pass Criteria

- Skill activates.
- `runSubagent` is used for the reviewer agent.
- `code-reviewer.md` template is referenced.
- Output is a structured review (not a plain chat reply).
