# Trigger Test: writing-plans

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:writing-plans` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Write a plan to implement [feature].

Alternative triggers:
- "Create a task document for this feature"
- "Plan out the implementation of X"
- "Write a detailed implementation plan"

## Expected Behavior

1. Copilot activates the `writing-plans` skill.
2. A structured task document (P{n}-{nnn}-{slug}.md) is drafted.
3. Adversarial review is scheduled before handing off to implementation.
4. `bof:adversarial-review` is invoked before any `bof:subagent-driven-development`.

## Pass Criteria

- Skill activates and begins plan drafting.
- Plan output follows Esquisse P{n}-{nnn}-{slug}.md format.
- Adversarial review is NOT skipped.
- No plan is handed directly to `ImplementerAgent` without adversarial gate.
