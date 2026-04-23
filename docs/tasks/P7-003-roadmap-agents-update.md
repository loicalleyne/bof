# P7-003 — roadmap-agents-update: Document bof-mcp in AGENTS.md and ROADMAP.md

**Phase:** P7 — Crush Compatibility  
**Status:** Done  
**Created:** 2026-04-19  
**Spec:** `docs/specs/2026-04-19-bof-crush-compatibility-design.md`  
**Depends on:** P7-001 and P7-002 complete (so ROADMAP task statuses are accurate)

---

## Goal

Update `AGENTS.md` to document the bof-mcp exception to the "pure markdown" rule, and update `ROADMAP.md` to add the P7 phase and task table.

---

## In Scope

- `AGENTS.md`:
  - Update the Project Overview paragraph that states bof is "pure markdown — no build system, no binary, no runtime dependencies" to acknowledge `bof-mcp/` as the sole exception
  - Update the Repository Layout tree to include `bof-mcp/` with a short comment
  - Add `bof-mcp` to the Build Commands section with `cd bof-mcp && go build -o bof-mcp .`
  - Add bof-mcp to Available Tools & Services table
- `docs/planning/ROADMAP.md`:
  - Add P7 phase heading, task table (P7-001, P7-002, P7-003), and gate checklist

---

## Out of Scope

- Changing any skill file
- Changing any agent file
- Changing GLOSSARY.md or ONBOARDING.md
- Adding `bof-mcp` to the Key Dependencies table (no key dependencies beyond the MCP SDK — keep it simple)

---

## Files

| Path | Action | What |
|---|---|---|
| `AGENTS.md` | Modify | Pure-markdown exception note; repo layout entry; build command; tools table entry |
| `docs/planning/ROADMAP.md` | Modify | Add P7 phase section with task table and gate checklist |

---

## AGENTS.md Changes (exact edits)

### 1. Project Overview — update the "pure markdown" sentence

**Find:**
```
bof is **pure markdown** — no build system, no binary, no runtime dependencies.
```

**Replace with:**
```
bof is **pure markdown** — no build system, no binary, no runtime dependencies —
with one exception: `bof-mcp/` is a Go MCP server that provides Crush-compatible
agent dispatch. It has its own `go.mod` and must be built
separately (`cd bof-mcp && go build -o bof-mcp .`).
```

### 2. Repository Layout tree — add bof-mcp entry

Add after the `scripts/` entry:

```
├── bof-mcp/                               ← Go MCP server: Crush agent dispatch
│   ├── main.go
│   ├── tools.go
│   ├── runner.go
│   ├── adversarial.go
│   ├── dispatch.go
│   ├── models.go
│   ├── state.go
│   ├── go.mod
│   └── README.md
```

### 3. Build Commands section — add new section

Add a new `## Build Commands` section (bof has no existing one) after the Repository Layout section:

```markdown
## Build Commands

bof skills and agents are pure markdown — no compilation needed. The only exception:

### bof-mcp (Go MCP server)

```sh
cd bof-mcp
go build -o bof-mcp .
```

See `bof-mcp/README.md` for configuration instructions.
```

### 4. Available Tools & Services — add bof-mcp row

In the VS Code Copilot Chat Primitives table or as a new subsection, add:

```markdown
### bof-mcp (optional, for Crush and VS Code model routing)

| Tool | Purpose | Config |
|---|---|---|
| `adversarial_review` | Runs adversarial review via Crush; disabled with `--no-adversarial` | bof-mcp |
| `gate_review` | Checks `.adversarial/` verdicts | bof-mcp |
| `implementer_agent` | Dispatches ImplementerAgent role via Crush | bof-mcp |
| `spec_review` | Dispatches SpecReviewerAgent role via Crush | bof-mcp |
| `quality_review` | Dispatches CodeQualityReviewerAgent role via Crush | bof-mcp |

See `bof-mcp/README.md` for `crush.json` and `.vscode/mcp.json` configuration snippets.
```

---

## ROADMAP.md Changes

Append the following after the P6 gate checklist:

```markdown
---

### P7 Tasks (Crush Compatibility)

| Task | Status | Summary |
|---|---|---|
| P7-001 bof-mcp-server | ⬜ Ready | Go MCP server: 6 tools, embed, `--no-adversarial`, model cache, README |
| P7-002 crush-mode-skill-sections | ⬜ Ready | Add `## Crush Mode (bof-mcp)` to 3 skills |
| P7-003 roadmap-agents-update | ⬜ Ready | Document bof-mcp exception in AGENTS.md; add P7 to ROADMAP |

### P7 Gate Checklist
- [ ] `cd bof-mcp && go build -o bof-mcp .` exits 0
- [ ] `mcp tools ./bof-mcp` lists all 6 tools; `--no-adversarial` removes 2
- [ ] All 3 skill files contain `## Crush Mode (bof-mcp)` section
- [ ] `grep -rn "Bash(\|Task(\|TodoWrite" skills/` returns 0
- [ ] AGENTS.md "pure markdown" paragraph documents bof-mcp exception
- [ ] `bof-mcp/README.md` contains verbatim `crush.json` and `.vscode/mcp.json` snippets
```

---

## Acceptance Criteria

1. `AGENTS.md` Project Overview no longer claims bof is unconditionally "pure markdown"; it documents bof-mcp as the sole exception with the build command.
2. `AGENTS.md` Repository Layout tree includes a `bof-mcp/` entry with its files listed.
3. `AGENTS.md` contains a Build Commands section with `cd bof-mcp && go build -o bof-mcp .`.
4. `AGENTS.md` Available Tools & Services section includes the 6 bof-mcp tools.
5. `docs/planning/ROADMAP.md` contains a P7 section with the 3-task table and gate checklist.
6. No existing P6 or earlier content in ROADMAP.md is altered.
7. No existing content in AGENTS.md outside the four targeted areas is altered.

---

## Session Notes

- Apply AGENTS.md edits as surgical replacements to the exact strings identified above — do not rewrite whole sections.
- The ROADMAP.md P7 section is an append — add it after the P6 gate checklist, before any end-of-file content.
- P7-001 and P7-002 must be complete before marking this task Done, so that the task table statuses are accurate at the time of writing.
