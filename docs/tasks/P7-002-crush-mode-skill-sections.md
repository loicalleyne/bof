# P7-002 — crush-mode-skill-sections: Add Crush Mode Sections to 3 Skills

**Phase:** P7 — Crush Compatibility  
**Status:** Done  
**Created:** 2026-04-19  
**Spec:** `docs/specs/2026-04-19-bof-crush-compatibility-design.md`  
**Depends on:** P7-001 must be complete so tool names in the sections match the registered tools

---

## Goal

Append `## Crush Mode (bof-mcp)` sections to the three bof skills that structurally depend on VS Code's `runSubagent` / `@agent` dispatch, making those skills fully functional on Crush and usable as a VS Code→Crush model-routing bridge.

---

## In Scope

- `skills/adversarial-review/SKILL.md` — append `## Crush Mode (bof-mcp)` section
- `skills/subagent-driven-development/SKILL.md` — append `## Crush Mode (bof-mcp)` section
- `skills/dispatching-parallel-agents/SKILL.md` — append `## Crush Mode (bof-mcp)` section

Exact section content for each skill is defined in spec Section 3.

---

## Out of Scope

- Changes to any other skill file
- Changes to `.agent.md` files
- Changes to `install_crush.sh` or the translation table — the new sections use Crush-native tool names directly (no translation needed for the Crush Mode sections since they reference MCP tools, not VS Code primitives)
- Adding Crush Mode sections to skills that already work on Crush without changes

---

## Files

| Path | Action | What |
|---|---|---|
| `skills/adversarial-review/SKILL.md` | Modify | Append `## Crush Mode (bof-mcp)` section per spec §3 |
| `skills/subagent-driven-development/SKILL.md` | Modify | Append `## Crush Mode (bof-mcp)` section per spec §3 |
| `skills/dispatching-parallel-agents/SKILL.md` | Modify | Append `## Crush Mode (bof-mcp)` section per spec §3 |

---

## Crush Mode Section Content

### `adversarial-review` — append verbatim:

```markdown
## Crush Mode (bof-mcp)

> **VS Code users:** Use the native `@Adversarial-r{slot}` dispatch path above.
> This section is only needed when running Crush, or when delegating to Crush from VS Code.

If bof-mcp is configured:

1. Call `discover_models` first to confirm available models.
2. Skip Steps 2–5 of this skill.
3. Call the `adversarial_review` MCP tool directly:
   - `plan_slug`: basename of the plan file (without `.md`)
   - `plan_content`: full text of the plan
   - `exclude_model`: your current implementing model ID (from `crush_info` tool; substring match is supported so you don't need the exact full string)
4. Read the verdict from the tool response. The tool writes `.adversarial/{plan_slug}.json`.
5. Apply the same PASSED / CONDITIONAL / FAILED response rules from Step 5.

**Coexistence with esquisse-mcp:** If esquisse-mcp is also configured, add
`--no-adversarial` to your bof-mcp server args to disable bof-mcp's
`adversarial_review` and `gate_review` tools and avoid name collisions.
```

### `subagent-driven-development` — append verbatim:

```markdown
## Crush Mode (bof-mcp)

> **VS Code users:** Use the native `runSubagent(...)` dispatch path above.
> This section is for Crush callers, or VS Code callers delegating to Crush for model access.

Replace each `runSubagent(...)` call with the corresponding bof-mcp tool:

| VS Code | bof-mcp tool | Notes |
|---|---|---|
| `runSubagent("ImplementerAgent", prompt)` | `implementer_agent` | Pass `model` param to select Crush model |
| `runSubagent("SpecReviewerAgent", prompt)` | `spec_review` | Same model or a smaller/faster one |
| `runSubagent("CodeQualityReviewerAgent", prompt)` | `quality_review` | Same model or a smaller/faster one |

**Before first task:** Call `discover_models` to confirm available models.
Use `adversarial_review` for the adversarial guard (or `gate_review` if the review already ran).
All other steps in this skill apply unchanged.
```

### `dispatching-parallel-agents` — append verbatim:

```markdown
## Crush Mode (bof-mcp)

> **VS Code users:** Use the native `runSubagent(...)` dispatch path above.

Crush does not support parallel agent dispatch. Perform each investigation
inline in the current session, one after another. No bof-mcp tool is required
for investigation-only work. If investigation tasks need to be run on a specific
model, use `implementer_agent` with `model` set appropriately.
```

---

## Acceptance Criteria

1. `skills/adversarial-review/SKILL.md` contains a `## Crush Mode (bof-mcp)` section with the `discover_models` first-call instruction and the `adversarial_review` tool call steps.
2. `skills/subagent-driven-development/SKILL.md` contains a `## Crush Mode (bof-mcp)` section with the tool substitution table mapping `runSubagent("ImplementerAgent",...)` → `implementer_agent` etc.
3. `skills/dispatching-parallel-agents/SKILL.md` contains a `## Crush Mode (bof-mcp)` section with the inline sequential investigation note.
4. All three sections include the `> **VS Code users:**` callout directing VS Code users to the native path.
5. `grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion\|CLAUDE\.md\|superpowers:" skills/` returns 0 results — no new tool-name violations introduced.
6. The existing body of each skill file is unchanged (only appending, no edits to existing content).

---

## Session Notes

- Append sections at the **end** of each SKILL.md, after all existing content, preceded by a blank line.
- Do not translate tool names in the Crush Mode sections — they already use MCP tool names (`adversarial_review`, `implementer_agent`, etc.) which are not VS Code primitives and do not appear in `install_crush.sh`'s translation table. This is correct and intentional.
- The `runSubagent(...)` references in the Crush Mode section header of `subagent-driven-development` appear in a VS Code context label ("VS Code users: use native `runSubagent(...)` path above") — this is not a tool call, it is documentation of the VS Code path. It does not violate the banned-tool-names invariant.
