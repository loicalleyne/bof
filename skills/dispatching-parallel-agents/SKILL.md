---
name: dispatching-parallel-agents
description: >
  Use when multiple independent areas of the codebase need to be investigated or
  fixed simultaneously. Triggers on: "investigate these failures in parallel",
  "dispatch parallel agents", "investigate multiple independent issues at once",
  "run parallel subagents".
---

# Dispatching Parallel Agents

Investigate multiple independent problems simultaneously by dispatching focused
read-only subagents, then synthesizing their results.

**Important: `runSubagent` in VS Code Copilot Chat is sequential**, not truly
parallel. "Parallel" here means dispatching multiple agents in quick succession
without waiting for one to finish before starting the next — VS Code schedules them.
The practical benefit is reduced total wall-clock time on investigations.

---

## When to Use

**Use when:**
- 3+ test files failing with clearly different root causes
- Multiple independent subsystems broken (e.g. auth, database, and rendering — no shared code)
- Need to gather information from multiple isolated areas simultaneously
- Investigation workload would take too long sequentially

**Do NOT use when:**
- Failures share a common root cause (one agent seeing the real issue is faster)
- Agents need the full project context to make sense of their findings
- Work is exploratory — unknown territory needs breadth-first, not divide-and-conquer
- Any agent needs to read or write shared state (file locks, DB state)
- The work involves *implementing*, not investigating — use `bof:subagent-driven-development`

---

## Process

### 1. Identify Independent Domains

List the independent areas to investigate. Each domain must be:
- **Self-contained:** agent can gather full context without needing other agents' findings
- **Focused:** one area, one question, specific output requested

Example decomposition:
```
Failures to investigate:
- Auth middleware tests (auth/, middleware/)
- Database migration tests (internal/db/migrations/)
- Renderer snapshot tests (internal/ui/renderer/)
```

### 2. Create Focused Prompts

Each subagent prompt must be:
- **Self-contained:** include all context (project root, relevant files, what to look for)
- **Specific output format:** tell the agent exactly what to return (findings, suspected cause, files involved)
- **Read-only scope:** investigating agents never write files

Template:
```
You are investigating [SPECIFIC PROBLEM] in [PROJECT NAME].

Project root: [PATH]
Area to investigate: [files/packages/subsystem]

Your goal: [specific question to answer]

Please examine the failing tests, trace the cause, and report:
1. Root cause (1-2 sentences)
2. Files involved (list)
3. Suggested fix approach (1-2 sentences)

Do NOT make any changes. Investigation only.
```

### 3. Dispatch All Agents

Dispatch in rapid succession using `runSubagent`:

```
runSubagent("ExploreD", authInvestigationPrompt)
runSubagent("ExploreD", dbInvestigationPrompt)
runSubagent("ExploreD", rendererInvestigationPrompt)
```

Use `ExploreD` for read-only investigations. Use `ImplementerAgent` only if an agent
needs to make changes (but then sequence them — don't dispatch multiple implementers
in parallel, that creates conflicts).

### 4. Synthesize Results

After all agents return:

1. **Check for conflicts:** Do any findings contradict each other? Does one root
   cause explain multiple failures?
2. **Check for missed connections:** Did agents independently find the same root cause?
   (If yes, the problem was not as independent as assumed — revise your plan.)
3. **Merge into unified understanding:** Write a summary of all root causes and
   their relationships.

### 5. Execute Fixes

Fix issues sequentially, not in parallel:
- One `ImplementerAgent` per fix, in order of dependency
- After all fixes: run the full test suite
- Verify all originally-failing tests now pass

---

## Red Flags

- Dispatching `ImplementerAgent` to multiple areas in parallel — creates file conflicts
- Giving agents overlapping file scope — findings will conflict
- Skipping synthesis — acting on each agent's finding independently misses interactions
- Using parallel dispatch for exploratory work where scope is unknown
