# Trigger Test: writing-skills

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:writing-skills` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Write a new skill for [workflow or capability].

Alternative triggers:
- "Create a SKILL.md for X"
- "Write a bof skill that teaches agents to do X"
- "Port this upstream skill to VS Code Copilot Chat"

## Expected Behavior

1. Copilot activates the `writing-skills` skill.
2. A SKILL.md is created with correct YAML frontmatter (`name:`, `description:`).
3. All tool names use VS Code conventions (no `Bash`, no `Task`, no `write`).
4. `description:` field is written to match real user trigger phrases.
5. Skill is placed in `skills/{name}/SKILL.md`.

## Pass Criteria

- Skill activates.
- Output is a valid SKILL.md with correct frontmatter.
- No banned tool names in output.
- `description:` field is trigger-phrase-optimized.
- File is placed at correct path.
