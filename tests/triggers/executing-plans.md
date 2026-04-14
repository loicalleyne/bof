# Trigger Test: executing-plans

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:executing-plans` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Execute this plan: [plan content or file path]

Alternative triggers:
- "Work through this task document"
- "Start executing P1-001"
- "Implement the steps in this plan"

## Expected Behavior

1. Copilot activates the `executing-plans` skill.
2. `manage_todo_list` is used to track steps.
3. Steps are executed in order with status updates.
4. `bof:verification-before-completion` is invoked before marking done.

## Pass Criteria

- Skill activates with a plan as input.
- `manage_todo_list` is used (not a text-only checklist).
- Steps are executed sequentially.
- Verification step is present before completion.
