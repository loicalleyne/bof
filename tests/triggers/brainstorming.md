# Trigger Test: brainstorming

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:brainstorming` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> I need to brainstorm ideas for [topic]. Let's explore options.

Alternative triggers:
- "Help me brainstorm approaches to X"
- "Let's brainstorm possible solutions"
- "I want to explore ideas before writing a plan"

## Expected Behavior

1. Copilot activates the `brainstorming` skill.
2. A structured brainstorming session begins using `vscode_askQuestions`.
3. Ideas are organized into a brief list or mind-map-style output.
4. Session ends with a prompt to move to `bof:writing-plans`.

## Pass Criteria

- Skill activates (no default response pattern).
- `vscode_askQuestions` is used for structured prompting.
- No `Bash(...)` or `Task(...)` tool names appear in output.
- Output uses VS Code tool names only.
