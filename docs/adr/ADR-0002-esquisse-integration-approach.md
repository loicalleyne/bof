# ADR-0002: Esquisse Integration Approach

**Date:** 2026-04-13
**Status:** Accepted

---

## Context

bof ships as a standalone skill library. However, the user also uses the
**Esquisse** project framework, which defines:

- A directory structure (`docs/tasks/`, `docs/specs/`, `docs/adr/`, etc.)
- A task document schema (Status, Goal, In Scope, Out of Scope, Files, Acceptance Criteria)
- A completion protocol (update AGENTS.md, GLOSSARY.md, ROADMAP.md, NEXT_STEPS.md after each task)
- Phase gates with `scripts/gate-check.sh`
- AST caching with `code_ast.duckdb` and `scripts/rebuild-ast.sh`
- Adversarial review infrastructure with `.github/agents/PlanD.agent.md`, `gate-review.sh`

The question is: how tightly should bof couple to Esquisse?

---

## Decision

**Reference Esquisse conventions in bof skills, but degrade gracefully when Esquisse is absent.**

Specifically:

- Spec save path: `docs/specs/YYYY-MM-DD-{topic}-design.md` — only used if the dir exists; skill creates it if not.
- Plan task path: `docs/tasks/P{n}-{nnn}-{slug}.md` — Esquisse format recommended; skill falls back to a flat plan doc if no `docs/tasks/` directory exists.
- AGENTS.md reading: all skills instruct "read AGENTS.md first"; they do not require it to exist.
- Completion protocol: ImplementerAgent runs Esquisse completion steps only if Esquisse documents are present.
- AST cache: skills use `code_ast.duckdb` if present; skip silently if absent.

bof does NOT:
- Require Esquisse to be installed for bof skills to function
- Require Esquisse documents to be present for skill triggers to work
- Modify Esquisse's FRAMEWORK.md or scripts

---

## Alternatives Considered

| Alternative | Reason Rejected |
|---|---|
| **No Esquisse coupling** (pure standalone) | Loses valuable task doc format, completion protocol, and phase gate integration; user uses Esquisse and wants the workflow to be consistent |
| **Hard Esquisse requirement** | Forces all bof users to adopt Esquisse; appropriate for user but wrong for a general-purpose port |
| **Merge bof into Esquisse** | Blurs the boundary between project framework (Esquisse) and workflow skills library (bof); different audiences and different update cadences |

---

## Consequences

**Positive:**
- bof works out of the box for new users without Esquisse.
- When used with Esquisse, bof skills produce properly structured task docs and follow the full completion protocol.
- Skills like `writing-plans` benefit from Esquisse task doc format when available, without requiring it.

**Negative:**
- Skills must detect Esquisse presence at runtime (check if `docs/tasks/` exists) rather than assuming.
- Skills have two code paths: Esquisse-aware and fallback. Documentation must describe both.

**Enforcement:**
- The `bof-bootstrap.instructions.md` teaches the agent about Esquisse integration.
- Individual skills include conditional guidance: "If `docs/tasks/` exists, use Esquisse task doc format; otherwise create a flat plan file."

---

## References

- Esquisse FRAMEWORK.md: `../esquisse/FRAMEWORK.md`
- bof integration notes in bootstrap: `instructions/bof-bootstrap.instructions.md`
