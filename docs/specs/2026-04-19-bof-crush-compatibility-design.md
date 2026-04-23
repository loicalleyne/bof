# bof Crush Compatibility Design

**Date:** 2026-04-19  
**Status:** Approved  
**Phase:** P7

---

## Problem Statement

bof skills are authored in VS Code Copilot Chat terms. Three VS Code-only primitives block full Crush compatibility:

1. `runSubagent("AgentName", prompt)` — dispatches named `.agent.md` agents; no Crush equivalent
2. `@Adversarial-r{slot}` agent dispatch — same mechanism
3. `.agent.md` definitions — VS Code Copilot Chat-specific format; Crush does not load them

Additionally, Crush can access models not available to VS Code Copilot (local Ollama, OpenRouter, provider-specific). A VS Code agent could delegate subagent work to Crush to use those models — but there is currently no bridge.

`scripts/install_crush.sh` (P6-007) already handles tool-name translation (`run_in_terminal` → `bash`, `manage_todo_list` → `todos`, etc.) for all other skills. The structural agent-dispatch gap is separate and requires an MCP server.

---

## Goals

1. All 22 bof skills work correctly on Crush.
2. VS Code Copilot Chat agents can delegate subagent tasks to Crush via MCP, accessing Crush-exclusive models.
3. bof-mcp is self-contained for adversarial review; does not depend on esquisse-mcp being present.
4. bof-mcp and esquisse-mcp can coexist without tool name conflicts when both are configured.
5. Self-configuration documentation is structured so Crush's `crush_config` skill can configure bof-mcp from the README alone.

---

## Non-Goals

- Windows-native build support (WSL only, matching esquisse-mcp pattern)
- Crush `.agent.md` equivalent definitions
- Auto-detecting esquisse-mcp at runtime (explicit opt-out flag, not a probe)
- A GUI, web UI, or persistent daemon

---

## Architecture

### Component 1: `bof/bof-mcp/` — Go MCP server

A new subdirectory of the bof repo. No separate repository. This is the sole exception to bof's "pure markdown" rule; AGENTS.md and ROADMAP.md will document it explicitly.

**Module path:** `github.com/loicalleyne/bof/bof-mcp`

**Binary:** `bof-mcp`

**Transport:** stdio (standard MCP pattern, same as esquisse-mcp)

**CLI:**
```
bof-mcp [--project-root <path>] [--no-adversarial] [--default-model <model-id>]
```

| Flag | Default | Purpose |
|---|---|---|
| `--project-root` | `$PWD` | Root for `.adversarial/` state files and `skills/adversarial-review/references/` |
| `--no-adversarial` | off | Skip registering `adversarial_review` and `gate_review` tools; use when esquisse-mcp is also configured to avoid tool name collision |
| `--default-model` | `""` (required in tool call if not set) | Default Crush model ID for dispatch tools when `model` param is omitted |

**Environment variable equivalents:**

| Env var | Flag equivalent |
|---|---|
| `BOF_PROJECT_ROOT` | `--project-root` |
| `BOF_NO_ADVERSARIAL` | `--no-adversarial` (set to `1` to enable) |
| `BOF_DEFAULT_MODEL` | `--default-model` |

Flags take precedence over env vars.

---

### Component 2: Tool Registry

#### 2a. Adversarial review tools (disabled by `--no-adversarial`)

**`adversarial_review`**

Runs bof's adversarial review protocol against a plan. Embeds bof's own `skills/adversarial-review/references/` protocol and report template via `//go:embed`. Uses a 5-slot default model pool, configurable with the `BOF_MODELS` environment variable. Runs `crush run --model {model} --quiet` with the full review prompt piped via stdin.

Input schema:
```json
{
  "plan_slug":     "string — used as .adversarial/{slug}.json state file name",
  "plan_content":  "string — full text of the plan to review",
  "rounds":        "int, optional — number of review rounds (default 3, max 10)",
  "exclude_model": "string, optional — model ID to exclude from pool (e.g. your implementing agent's model)"
}
```

Output: worst-case verdict written to `.adversarial/{plan_slug}.json`, tool returns verdict summary.

**`gate_review`**

Checks whether all `.adversarial/*.json` state files have a `PASSED` or `CONDITIONAL` last_verdict. Returns a summary; non-zero exit plan count blocks handoff.

Input schema: none (uses `--project-root`).

---

#### 2b. Agent dispatch tools (always registered)

These tools replace `runSubagent("AgentName", prompt)` for Crush callers and provide a model-routing bridge for VS Code callers.

All three tools share the same pattern:
- Accept a `prompt` string and an optional `model` string
- If `model` is omitted and `--default-model` is configured, use the default
- If neither is set, return an error: "model is required"
- Before invoking, strip YAML frontmatter from the embedded `.agent.md` preamble (everything from `---` through the closing `---` inclusive) so only the markdown body is sent to the model
- Invoke `crush run --model {model} --quiet` with the stripped role preamble prepended to the prompt, piped via stdin
- Return the full output

**`implementer_agent`**

Dispatches the ImplementerAgent role. The preamble is the full content of `agents/ImplementerAgent.agent.md` (embedded via `//go:embed`).

Input schema:
```json
{
  "prompt": "string — full task document text plus any cross-task context",
  "model":  "string, optional — Crush model ID (e.g. 'gemini-2.5-pro'); uses default if omitted"
}
```

**`spec_review`**

Dispatches the SpecReviewerAgent role. Preamble embedded from `agents/SpecReviewerAgent.agent.md`.

Input schema: same as `implementer_agent`.

**`quality_review`**

Dispatches the CodeQualityReviewerAgent role. Preamble embedded from `agents/CodeQualityReviewerAgent.agent.md`.

Input schema: same as `implementer_agent`.

---


---

### Component 3: Skill `## Crush Mode (bof-mcp)` sections

Three skills receive a new `## Crush Mode (bof-mcp)` section appended to their body. The source-of-truth files in `bof/skills/` are updated; `install_crush.sh` copies and translates them as normal.

**`adversarial-review`**

```markdown
## Crush Mode (bof-mcp)

> **VS Code users:** Use the native `@Adversarial-r{slot}` dispatch path above.
> This section is only needed when running Crush, or when delegating to Crush from VS Code.

If bof-mcp is configured:

1. Skip Steps 2–5 of this skill.
2. Call the `adversarial_review` MCP tool directly:
   - `plan_slug`: basename of the plan file (without `.md`)
   - `plan_content`: full text of the plan
   - `exclude_model`: your current implementing model ID (from `crush_info` tool)
3. Read the verdict from the tool response. The tool writes `.adversarial/{plan_slug}.json`.
4. Apply the same PASSED / CONDITIONAL / FAILED response rules from Step 5.

**Coexistence with esquisse-mcp:** If esquisse-mcp is also configured, add
`--no-adversarial` to your bof-mcp server args to disable bof-mcp's
`adversarial_review` and `gate_review` tools and avoid name collisions.
```

**`subagent-driven-development`**

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

Use `adversarial_review` for the adversarial guard (or `gate_review` if review already ran).
All other steps in this skill apply unchanged.
```

**`dispatching-parallel-agents`**

```markdown
## Crush Mode (bof-mcp)

> **VS Code users:** Use the native `runSubagent(...)` dispatch path above.

Crush does not support parallel agent dispatch. Perform each investigation
inline in the current session, one after another. No bof-mcp tool is required
for investigation-only work. If investigation tasks need to be run on a specific
model, use `implementer_agent` with `model` set appropriately.
```

---

### Component 4: `bof-mcp/README.md`

Structured for `crush_config` skill self-configuration. Sections in order:

1. **What bof-mcp does** — one paragraph
2. **Prerequisites** — `crush` binary in PATH; Go 1.22+ to build
3. **Build**
   ```sh
   cd bof-mcp && go build -o bof-mcp .
   # or from repo root:
   go build -o bof-mcp ./bof-mcp/
   ```
4. **Configure in Crush** — verbatim `crush.json` snippet:
   ```json
   "mcpServers": {
     "bof": {
       "command": "/absolute/path/to/bof-mcp",
       "args": ["--project-root", "/absolute/path/to/project"],
       "env": {}
     }
   }
   ```
5. **If you also use esquisse-mcp** — add `"--no-adversarial"` to `args`:
   ```json
   "args": ["--project-root", "/absolute/path/to/project", "--no-adversarial"]
   ```
6. **Configure in VS Code** — verbatim `.vscode/mcp.json` snippet (same structure, different key)
7. **Using a default model** — add `"--default-model", "gemini-2.5-pro"` to `args`
8. **Tool reference table** — name, description, required params, optional params

---

## Implementation Notes

- Pattern: identical to `esquisse-mcp`. Read `esquisse/esquisse-mcp/runner.go` and `adversarial.go` as primary reference.
- Embed agent markdown files: `//go:embed ../agents/ImplementerAgent.agent.md` etc. (paths relative to `bof-mcp/` package)
- Embed adversarial references: `//go:embed ../skills/adversarial-review/references/*.md`
- Frontmatter stripping: strip from first `---` through the next `---\n` (inclusive) before using embedded `.agent.md` content as a preamble; if no closing `---` is found, use the full content as-is
- Module: `bof-mcp/go.mod` is a standalone module (`github.com/loicalleyne/bof/bof-mcp`). No `go.work` file needed — bof-mcp is the only Go code in the repo. Build from `bof-mcp/` directory or via `go build ./bof-mcp/` from repo root (the latter requires `-C bof-mcp` or `cd` first since there is no workspace)

---

## File Map

| Path | Action | What |
|---|---|---|
| `bof-mcp/main.go` | Create | CLI entry point, flag parsing, server startup |
| `bof-mcp/tools.go` | Create | `registerTools`, tool handler wiring |
| `bof-mcp/runner.go` | Create | `RunCrush` — `crush run` subprocess invocation (same as esquisse-mcp) |
| `bof-mcp/adversarial.go` | Create | `adversarial_review` and `gate_review` tool handlers |
| `bof-mcp/dispatch.go` | Create | `implementer_agent`, `spec_review`, `quality_review` handlers |
| `bof-mcp/models.go` | Create | Adversarial model pool management |
| `bof-mcp/state.go` | Create | `.adversarial/*.json` read/write (same as esquisse-mcp) |
| `bof-mcp/go.mod` | Create | Standalone module `github.com/loicalleyne/bof/bof-mcp`; no `go.work` needed |
| `bof-mcp/README.md` | Create | Self-configuration docs (Section 4 above) |
| `skills/adversarial-review/SKILL.md` | Modify | Append `## Crush Mode (bof-mcp)` section |
| `skills/subagent-driven-development/SKILL.md` | Modify | Append `## Crush Mode (bof-mcp)` section |
| `skills/dispatching-parallel-agents/SKILL.md` | Modify | Append `## Crush Mode (bof-mcp)` section |
| `AGENTS.md` | Modify | Document bof-mcp as the sole exception to "pure markdown" rule |
| `docs/planning/ROADMAP.md` | Modify | Add P7 phase and task table |

---

## Phase P7 Task Breakdown

| Task | Slug | Summary |
|---|---|---|
| P7-001 | `bof-mcp-server` | Go MCP server: 5 tools, embed, flags, README |
| P7-002 | `crush-mode-skill-sections` | Add `## Crush Mode (bof-mcp)` to 3 skills |
| P7-003 | `roadmap-agents-update` | Update AGENTS.md pure-markdown exception + ROADMAP.md P7 table |

---

## Spec Self-Review

- No placeholders left blank
- No contradictions between tool descriptions and implementation notes
- Scope is bounded: 6 tools, 3 skill edits, 1 README
- Out-of-scope items are explicit
- crush_config skill can configure bof-mcp from README Section 4 alone (verbatim snippets, no inference required)
- `--no-adversarial` conflict resolution is documented in both the README and the `adversarial-review` Crush Mode section
