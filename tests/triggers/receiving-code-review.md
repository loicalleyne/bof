# Trigger Test: receiving-code-review

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:receiving-code-review` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Help me respond to this code review feedback.

Alternative triggers:
- "I received a code review. Help me address the comments."
- "Process these code review comments"
- "Address the feedback from the code review"

## Expected Behavior

1. Copilot activates the `receiving-code-review` skill.
2. Review comments are parsed and categorized (must fix / should fix / optional).
3. A structured response plan is produced.
4. Changes are applied with `replace_string_in_file`.
5. PR replies drafted using `gh pr review --comment` syntax.

## Pass Criteria

- Skill activates.
- Comments are categorized (not all treated equally).
- `replace_string_in_file` is used (not direct file write commands).
- Uses "the developer" (not "your human partner").
