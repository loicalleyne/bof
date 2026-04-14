# Trigger Test: subagent-driven-development

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:subagent-driven-development` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Use subagent-driven development to implement this task.

Alternative triggers:
- "Dispatch ImplementerAgent for this task"
- "Use the subagent workflow to build X"
- "Implement this using subagent-driven development"

## Expected Behavior

1. Copilot activates the `subagent-driven-development` skill.
2. Adversarial review verdict is checked before dispatching ImplementerAgent.
3. `runSubagent("ImplementerAgent", prompt)` is used to dispatch the implementer.
4. Spec and code quality review agents are dispatched after implementation.

## Pass Criteria

- Skill activates.
- Adversarial verdict is checked (FAILED blocks dispatch).
- `runSubagent` is used, not `Task(...)`.
- SpecReviewerAgent and CodeQualityReviewerAgent are invoked.
