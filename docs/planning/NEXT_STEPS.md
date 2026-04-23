# NEXT_STEPS.md — bof planning

## Session Log

### 2026-04-19 — P7-001 bof-mcp-server Complete

**What was done:** Created `bof/bof-mcp/` — a standalone Go MCP stdio server exposing 5 tools: `adversarial_review`, `gate_review`, `implementer_agent`, `spec_review`, `quality_review`.

**Key decision:** Go's `//go:embed` prohibits `..` in patterns. Worked around by creating `bof-mcp/embedded/` with copies of the agent `.agent.md` files and adversarial review references. Binary is self-contained. Updated `AGENTS.md` Common Mistakes with this gotcha.

**Files created:**
- `bof-mcp/go.mod`
- `bof-mcp/main.go`
- `bof-mcp/tools.go`
- `bof-mcp/runner.go`
- `bof-mcp/adversarial.go`
- `bof-mcp/dispatch.go`
- `bof-mcp/models.go`
- `bof-mcp/state.go`
- `bof-mcp/README.md`
- `bof-mcp/embedded/agents/ImplementerAgent.agent.md` (copy)
- `bof-mcp/embedded/agents/SpecReviewerAgent.agent.md` (copy)
- `bof-mcp/embedded/agents/CodeQualityReviewerAgent.agent.md` (copy)
- `bof-mcp/embedded/references/task-review-protocol.md` (copy)
- `bof-mcp/embedded/references/report-template.md` (copy)

**Next up:** P7-002 (crush-mode skill sections), P7-003 (ROADMAP + AGENTS.md P7 update).

---

## Blocked / Parked

None.
