# bof — Document Schemas

Canonical schemas for all artifact types in the bof project.
When a file, agent, or skill references a schema, this document is the source of truth.

---

## 1. Adversarial Review State (`.adversarial/state.json`)

**Location:** `.adversarial/state.json` in the project root (gitignored).

**Written by:** Only `Adversarial-r*` agents. Never EsquissePlan, the
`adversarial-review` skill, a human, nor any hook script.

**Read by:** `adversarial-review` skill (rotation slot), `stop.sh` (verdict gate),
`session-start.sh` (context injection), `subagent-driven-development` skill
(implementation gate).

### Required Fields

```json
{
  "iteration":        <int>,        // Increments on every review. Default 0 if file absent.
  "last_model":       "<string>",   // Model name that produced the most recent verdict.
  "last_verdict":     "PASSED|CONDITIONAL|FAILED",
  "last_review_date": "YYYY-MM-DD"
}
```

These four fields are the **minimum required**. Any write that omits or renames
one of them will cause `stop.sh` to read `last_verdict` as empty and emit a
warning as if no review has been performed.

### Full Schema (with optional history)

```json
{
  "iteration":        3,
  "last_model":       "GPT-4.1 (copilot)",
  "last_verdict":     "PASSED",
  "last_review_date": "2026-04-13",
  "plan":             "docs/tasks/P1-001-example.md",
  "history": [
    {
      "iteration":             <int>,
      "reviewer":              "<string>",
      "verdict":               "PASSED|CONDITIONAL|FAILED",
      "summary":               "<one sentence>",
      "prior_issues_resolved": [],
      "new_issues":            []
    }
  ]
}
```

### Field Rules

| Field | Type | Allowed values | Written by |
|-------|------|---------------|-----------|
| `iteration` | integer ≥ 0 | Any non-negative integer | Adversarial-r* |
| `last_model` | string | Any model name string | Adversarial-r* |
| `last_verdict` | enum | `"PASSED"`, `"CONDITIONAL"`, `"FAILED"` | Adversarial-r* |
| `last_review_date` | string | ISO 8601 date `YYYY-MM-DD` | Adversarial-r* |
| `plan` | string | Relative path to the task/plan doc reviewed | Adversarial-r* |
| `history` | array | Array of `HistoryEntry` objects (optional) | Adversarial-r* |

### Rotation Mapping

`slot = iteration % 3`:

| Slot | Agent | Model |
|------|-------|-------|
| 0 | `Adversarial-r0` | GPT-4.1 (copilot) |
| 1 | `Adversarial-r1` | Claude Opus 4.6 (copilot) |
| 2 | `Adversarial-r2` | GPT-4o (copilot) |

### Recovery (manual write)

If the file is corrupt or a session gets stuck, restore it with exactly the
required four fields and set `"last_verdict": "FAILED"` to force a new review
before implementation proceeds.

---

## 2. Hook Output: Session Context (`scripts/hooks/session-start.sh`)

**Event:** `SessionStart`

**Consumed by:** VS Code Copilot Chat context injection.

```json
{
  "project":            "<string>",   // basename of the project root directory
  "branch":             "<string>",   // current git branch, or "N/A" if not a git repo
  "hasAgentsMd":        0 | 1,        // 1 if AGENTS.md exists at project root
  "adversarialVerdict": "<string>"    // value of last_verdict from state.json, or "" if absent
}
```

| Field | Type | Source |
|-------|------|--------|
| `project` | string | `basename "$PWD"` |
| `branch` | string | `git branch --show-current` |
| `hasAgentsMd` | 0 or 1 | file existence check |
| `adversarialVerdict` | string | `last_verdict` from `.adversarial/state.json`, empty string if absent |

---

## 3. Hook Output: Subagent Context (`scripts/hooks/subagent-start.sh`)

**Event:** `AgentStart`

**Consumed by:** VS Code Copilot Chat context injection for each subagent dispatch.

```json
{
  "agentName":   "<string>",   // value of $COPILOT_AGENT_NAME env var, or "unknown"
  "hasTasksDir": 0 | 1,        // 1 if docs/tasks/ directory exists
  "hasAstCache": 0 | 1,        // 1 if code_ast.duckdb exists at project root
  "reminder":    "<string>"    // Fixed advisory string (read AGENTS.md, use correct tool names)
}
```

| Field | Type | Source |
|-------|------|--------|
| `agentName` | string | `$COPILOT_AGENT_NAME` environment variable |
| `hasTasksDir` | 0 or 1 | directory existence check for `docs/tasks/` |
| `hasAstCache` | 0 or 1 | file existence check for `code_ast.duckdb` |
| `reminder` | string | static string (never user-sourced) |

---

## 4. Hook Configuration (`.github/hooks/hooks.json`)

**Canonical location:** `.github/hooks/hooks.json` in the project root.
This is the file VS Code reads. `hooks/bof-hooks.json.tpl` is a legacy template
and is superseded by this file.

```json
{
  "hooks": {
    "Stop": [
      {
        "type": "command",
        "command": "wsl bash ./scripts/gate-review.sh"
      }
    ],
    "SessionStart": [
      {
        "type": "command",
        "command": "wsl bash ./scripts/hooks/session-start.sh"
      }
    ],
    "AgentStart": [
      {
        "type": "command",
        "command": "wsl bash ./scripts/hooks/subagent-start.sh"
      }
    ]
  }
}
```

**Format rules:**
- `hooks` is an **object** keyed by event name — not an array with `event` property.
- Each event value is an array of hook objects: `{ "type": "command", "command": "..." }`.
- Commands use `wsl bash` prefix on Windows to ensure WSL bash is invoked.

**Supported event values** (VS Code Copilot Chat):

| Event | When it fires | Script |
|-------|--------------|--------|
| `Stop` | When the agent session ends | `scripts/gate-review.sh` (outputs `{"continue":true}` or exits 2) |
| `SessionStart` | Once at the beginning of a new chat session | `scripts/hooks/session-start.sh` (outputs context JSON) |
| `AgentStart` | Each time a subagent is dispatched via `runSubagent` | `scripts/hooks/subagent-start.sh` (outputs context JSON) |

**Prerequisite:** VS Code setting `"chat.useCustomAgentHooks": true` must be
enabled for the hooks file to be read.
