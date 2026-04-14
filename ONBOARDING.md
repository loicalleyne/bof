# ONBOARDING.md — bof

## Mental Model

bof is a skill library that gives VS Code Copilot Chat structured developer workflows. It is a port of [`superpowers`](https://github.com/obra/superpowers) using VS Code Copilot Chat primitives. Each skill is a `SKILL.md` file that auto-triggers when a user's message matches its `description:` field — no plugin system, no marketplace, just files in `~/.copilot/skills/`. Agents (`.agent.md` files) are named subagent definitions dispatched via `runSubagent()` for isolated tasks. The bootstrap instructions file (`bof-bootstrap.instructions.md`) is injected into every session to prime the agent with workflow context and project reading priorities. Together, skills + agents + instructions create a fully structured development workflow from feature ideation through code review and merge.

## Read Order

1. [`AGENTS.md`](AGENTS.md) — project constitution: tool name rules, invariants, security, scope
2. This file — mental model, data flow, key files map
3. [`GLOSSARY.md`](GLOSSARY.md) — canonical vocabulary (what each term means when used in skills)
4. [`docs/specs/tool-mapping.md`](docs/specs/tool-mapping.md) — the complete Claude Code → VS Code tool name mapping (read before porting any upstream content)
5. [`docs/planning/ROADMAP.md`](docs/planning/ROADMAP.md) — current phase and open tasks
6. Individual task doc from `docs/tasks/` — when implementing a specific phase task

## How VS Code Copilot Chat Loads bof

```
~/.copilot/
├── skills/
│   └── brainstorming/        ← symlink → bof/skills/brainstorming/
│       └── SKILL.md          ← description: field triggers auto-activation
├── agents/
│   └── ImplementerAgent.agent.md  ← symlink → bof/agents/ImplementerAgent.agent.md
├── instructions/
│   └── bof-bootstrap.instructions.md  ← injected into EVERY session (applyTo: "**")
└── hooks/
    └── bof-hooks.json        ← lifecycle event handlers (Preview feature)
```

When a user opens a new Copilot Chat session:
1. VS Code reads all `.instructions.md` files matching `applyTo:` — bootstrap is injected
2. User sends a message → VS Code matches against all `SKILL.md` `description:` fields
3. Matching skill content is provided to the model as context
4. Model responds following the skill's workflow
5. If skill dispatches subagents, `runSubagent("AgentName", prompt)` is used

## Workflow Data Flow

```
Feature Idea
    │
    ▼
bof:brainstorming ──── HARD GATE: no code until design approved ────
    │  Asks questions via vscode_askQuestions
    │  Writes: docs/specs/YYYY-MM-DD-{topic}-design.md
    ▼
bof:writing-plans
    │  Reads spec → maps structure → writes docs/tasks/P{n}-{nnn}-{slug}.md
    │  Runs bof:adversarial-review → blocks on FAILED verdict
    ▼
bof:using-git-worktrees
    │  Creates branch + worktree, installs deps, copies AST cache,
    │  runs baseline tests
    ▼
bof:subagent-driven-development (controller)
    │  For each task in docs/tasks/:
    │    runSubagent(ImplementerAgent)      ← TDD: RED→GREEN→REFACTOR, commit
    │    runSubagent(SpecReviewerAgent)     ← read-only spec check
    │    runSubagent(CodeQualityReviewerAgent) ← read-only quality check
    │  If issues → ImplementerAgent fixes → re-review
    ▼
bof:requesting-code-review
    │  runSubagent(CodeQualityReviewerAgent) for full implementation
    ▼
bof:finishing-a-development-branch
    │  Runs tests, presents merge/PR/keep/discard options
    ▼
Esquisse Completion Protocol (if Esquisse in use)
    Task doc: Status → Done
    AGENTS.md: update Common Mistakes
    GLOSSARY.md: new terms
    ROADMAP.md: task status
    NEXT_STEPS.md: session log
```

## Key Files Map

| File | Role |
|------|------|
| `AGENTS.md` | Project constitution: tool name rules, scope, invariants, security |
| `GLOSSARY.md` | Canonical vocabulary for all terms used in skills |
| `docs/specs/tool-mapping.md` | Complete Claude Code → VS Code Copilot Chat tool mapping |
| `docs/planning/ROADMAP.md` | Phase-by-phase implementation plan with task status |
| `docs/adr/ADR-0001-*.md` | Why VS Code, not Claude Code |
| `docs/adr/ADR-0002-*.md` | Esquisse integration approach (graceful degradation) |
| `docs/adr/ADR-0003-*.md` | AST cache copy-at-creation + `--incremental` strategy |
| `skills/brainstorming/SKILL.md` | Feature design workflow — HARD GATE before any code |
| `skills/writing-plans/SKILL.md` | Task breakdown with adversarial review gate |
| `skills/subagent-driven-development/SKILL.md` | Controller: dispatches Implementer → SpecReviewer → QualityReviewer per task |
| `skills/adversarial-review/SKILL.md` | Rotating cross-model hostile plan reviewer |
| `skills/test-driven-development/SKILL.md` | IRON LAW: RED test first, always |
| `skills/verification-before-completion/SKILL.md` | Must pass before any completion claim |
| `agents/ImplementerAgent.agent.md` | Full read/write/execute — TDD implementer |
| `agents/SpecReviewerAgent.agent.md` | Read-only — spec compliance checker |
| `agents/CodeQualityReviewerAgent.agent.md` | Read-only — code quality reviewer |
| `agents/Adversarial-r0.agent.md` | Read-only — GPT-4.1, hostile plan reviewer, slot 0 |
| `agents/Adversarial-r1.agent.md` | Read-only — Claude Opus 4.6, hostile plan reviewer, slot 1 |
| `agents/Adversarial-r2.agent.md` | Read-only — GPT-4o, hostile plan reviewer, slot 2 |
| `instructions/bof-bootstrap.instructions.md` | Always-on context injection; skill list; project reading priority |
| `hooks/bof-hooks.json.tpl` | Lifecycle hooks: SessionStart context, SubagentStart context, Stop verification |
| `scripts/install.sh` | WSL-aware symlinker: bof/ → ~/.copilot/ |
| `tests/triggers/` | Manual trigger test checklists (one per skill) |

## Invariants

1. Every SKILL.md triggers correctly in VS Code Copilot Chat (description field is accurate).
2. No banned upstream tool names (Bash, Task, TodoWrite, AskUserQuestion) appear in any skill.
3. Read-only agents (SpecReviewer, CodeQualityReviewer, Adversarial-r*) never write files.
4. `bof:adversarial-review` verdict gates execution handoff from `writing-plans` and from `subagent-driven-development`.
5. `install.sh` is idempotent.

