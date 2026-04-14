# Trigger Test: dispatching-parallel-agents

## Purpose

This file documents the trigger phrase and expected behavior for the
`bof:dispatching-parallel-agents` skill.

## Trigger Phrase

Paste this into VS Code Copilot Chat to activate the skill:

> Dispatch agents in parallel to work on these independent tasks.

Alternative triggers:
- "Run multiple agents at the same time on X, Y, Z"
- "Use parallel agents to handle these independent subtasks"
- "Split this work across parallel agents"

## Expected Behavior

1. Copilot activates the `dispatching-parallel-agents` skill.
2. Tasks are assessed for independence before parallelization.
3. `runSubagent` calls are made for each independent task.
4. Note: VS Code dispatches them sequentially (not truly concurrent).

## Pass Criteria

- Skill activates.
- Independence check is performed.
- `runSubagent` is used for each agent.
- Agent acknowledges sequential execution (VS Code limitation).
