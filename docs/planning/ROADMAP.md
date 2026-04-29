# ROADMAP — bof

## Current Phase: P0 — Foundation

**Target:** Project constitution complete; all directory scaffolding in place; directory structure matches plan; 3 ADRs written.

---

### P0 Tasks

| Task | Status | Summary |
|------|--------|---------|
| P0-001 AGENTS.md | ✅ Done | Full project constitution: tool rules, invariants, scope, security |
| P0-002 GLOSSARY.md | ✅ Done | 20 canonical terms |
| P0-003 ONBOARDING.md | ✅ Done | Mental model, workflow diagram, key files map |
| P0-004 Directory scaffold + tool-mapping.md | ✅ Done | All dirs created; docs/specs/tool-mapping.md complete |
| P0-005 ADR-0001, ADR-0002, ADR-0003 | ✅ Done | Platform, Esquisse integration, AST worktree strategy |
| P0-006 ROADMAP.md | ✅ Done | This file |

### P0 Gate Checklist
- [x] All P0 tasks Done
- [x] `grep -l "description:" skills/*/SKILL.md` — N/A (no skills yet; gate runs at P2 end)
- [x] AGENTS.md reflects actual current structure
- [x] Directory tree matches plan in AGENTS.md

---

### P1 Tasks (Bootstrap + Meta)

| Task | Status | Summary |
|------|--------|---------|
| P1-001 bof-bootstrap.instructions.md | ⬜ Ready | Always-on bootstrap with skill list, AST hints, Esquisse integration |
| P1-002 writing-skills SKILL.md | ⬜ Ready | Port from superpowers; VS Code tool name adaptation |

### P1 Gate Checklist
- [ ] `instructions/bof-bootstrap.instructions.md` exists with correct `applyTo: "**"` frontmatter
- [ ] Opening a new VS Code session with a project AGENTS.md → agent reads it first
- [ ] All P1 task docs Status: Done

---

### P2 Tasks (Core Development Workflow Skills)

| Task | Status | Summary |
|------|--------|---------|
| P2-001 brainstorming | ⬜ Ready | Port; napkin skill substitution; vscode_askQuestions |
| P2-002 writing-plans | ⬜ Ready | Port; adversarial review gate; Esquisse task doc format |
| P2-003 test-driven-development | ⬜ Ready | Port; testing-anti-patterns.md |
| P2-004 verification-before-completion | ⬜ Ready | Port; minor adaptations |

### P2 Gate Checklist
- [ ] All 4 skill SKILL.md files present
- [ ] `grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion\|CLAUDE\.md\|superpowers:" skills/` returns 0 results
- [ ] brainstorming trigger prompt activates skill; HARD-GATE present
- [ ] writing-plans runs adversarial review before execution handoff

---

### P3 Tasks (Agent Definitions + Subagent Workflow)

| Task | Status | Summary |
|------|--------|---------|
| P3-001 ImplementerAgent.agent.md | ⬜ Ready | TDD implementer; AST update at completion |
| P3-002 SpecReviewerAgent.agent.md | ⬜ Ready | Read-only spec compliance |
| P3-003 CodeQualityReviewerAgent.agent.md | ⬜ Ready | Read-only code quality |
| P3-004 subagent-driven-development | ⬜ Ready | Controller skill; adversarial guard; 3-agent loop |
| P3-005 executing-plans | ⬜ Ready | Single-session execution mode |
| P3-006 dispatching-parallel-agents | ⬜ Ready | Parallel subagent dispatch |
| P3-007 adversarial-review SKILL.md + Adversarial agents | ⬜ Ready | Rotating cross-model plan reviewer |

### P3 Gate Checklist
- [ ] All 3 agent files present with correct frontmatter
- [ ] All 3 Adversarial-r* agent files present; read-only tools only
- [ ] adversarial-review references/ dir populated
- [ ] subagent-driven-development checks for adversarial verdict before first ImplementerAgent dispatch
- [ ] ImplementerAgent runs `rebuild-ast.sh --incremental` at task completion

---

### P4 Tasks (Code Review + Quality Skills)

| Task | Status | Summary |
|------|--------|---------|
| P4-001 requesting-code-review | ⬜ Ready | Dispatches CodeQualityReviewerAgent |
| P4-002 receiving-code-review | ⬜ Ready | Handles incoming review feedback |
| P4-003 systematic-debugging | ⬜ Ready | Phase 1 Investigation before any fix |

### P4 Gate Checklist
- [ ] requesting-code-review dispatches via runSubagent (no Bash/Task)
- [ ] systematic-debugging prevents fix before investigation phase

---

### P5 Tasks (Git Workflow Skills)

| Task | Status | Summary |
|------|--------|---------|
| P5-001 using-git-worktrees | ⬜ Ready | Copy-at-creation AST cache; deps; baseline tests |
| P5-002 finishing-a-development-branch | ⬜ Ready | 4-option menu; vscode_askQuestions confirm |

### P5 Gate Checklist
- [ ] using-git-worktrees copies `code_ast.duckdb` from main root if present
- [ ] finishing-a-development-branch presents 4 options via vscode_askQuestions

---

### P6 Tasks (Lab Skills + Install Infrastructure)

| Task | Status | Summary |
|------|--------|---------|
| P6-001 Verify mcp-cli | ⬜ Ready | Check vs superpowers-lab source |
| P6-002 Verify tmux skill | ⬜ Ready | Check vs superpowers-lab source |
| P6-003 install.sh + uninstall.sh | ⬜ Ready | WSL-aware symlinker + hooks |
| P6-004 README.md | ⬜ Ready | Install guide, skill list, Esquisse notes |
| P6-005 hooks/bof-hooks.json.tpl + scripts | ⬜ Ready | SessionStart, SubagentStart, Stop hooks |
| P6-006 tests/triggers/ | ⬜ Ready | 14 manual trigger test checklists |
| P6-007 install_crush.sh | ⬜ Ready | Idempotent skill installer for Crush; symlinks skills/ into ~/.config/crush/skills/ |

### P6 Gate Checklist
- [ ] `bash scripts/install.sh` runs idempotently from WSL
- [ ] New VS Code session shows bof skills after install
- [ ] All 14 tests/triggers/*.md files present
- [ ] hooks/bof-hooks.json.tpl present; stop.sh checks verification + adversarial review

---

## Completed Phases

None yet.

---

## P7 — Crush Compatibility

| Task | Status | What |
|---|---|---|
| P7-001 bof-mcp-server | ✅ Done | Go MCP server: 6 tools, embed, `--no-adversarial`, model cache, README |
| P7-002 crush-mode-skill-sections | ✅ Done | Skills: Crush-mode sections for each skill |
| P7-003 roadmap-agents-update | ✅ Done | ROADMAP + AGENTS.md P7 updates |

### P7 Gate Checklist
- [x] P7-001 `bof-mcp` binary builds clean (`go build -o bof-mcp .`)
- [x] P7-002 Crush-mode sections added to skills
- [x] P7-003 ROADMAP and AGENTS.md updated

---

## P8 — Termux / Android Support

| Task | Status | What |
|---|---|---|
| P8-001 install_termux.sh | ⬜ Ready | Termux/Android install: skill copy+translate, jq crush.json merge, bof-mcp build |

### P8 Gate Checklist
- [ ] `test -x scripts/install_termux.sh`
- [ ] `bash scripts/install_termux.sh --dry-run` exits 0 without creating files
- [ ] After real run: installed SKILL.md files contain `agent` not `runSubagent`
- [ ] After real run: `options.skills_paths` in crush.json contains skills dir (exactly once)

---

## P9 — Agent Behavior Hardening

**Target:** Enforce minimal-editing discipline in ImplementerAgent and implement-task skill to prevent agents from over-editing when making fixes.

| Task | Status | Summary |
|------|--------|---------|
| [P9-001-minimal-editing-implementer](../tasks/P9-001-minimal-editing-implementer.md) | ✅ Done | Diff-review bullet + expanded TDD Cycle step 6 in ImplementerAgent and embedded copy |
| [P9-002-minimal-editing-implement-task](../tasks/P9-002-minimal-editing-implement-task.md) | ✅ Done | `minimal-editing.md` reference doc + Step 7b in implement-task skill |

### P9 Gate Checklist
- [x] `grep -c "git diff --cached" agents/ImplementerAgent.agent.md` ≥ 1
- [x] `grep -c "git diff --cached" bof-mcp/embedded/agents/ImplementerAgent.agent.md` ≥ 1
- [x] `diff <(grep -A6 "Review your diff" agents/ImplementerAgent.agent.md) <(grep -A6 "Review your diff" bof-mcp/embedded/agents/ImplementerAgent.agent.md)` — no output
- [x] `test -f skills/implement-task/references/minimal-editing.md`
- [x] `grep -c "Step 7b" skills/implement-task/SKILL.md` ≥ 1
- [x] `grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion" agents/ skills/implement-task/` — 0 results
- [x] All P9 task docs Status: Done

---

## Phase Gate — Universal Criteria (all phases)

```
- [ ] All tasks in phase have Status: Done
- [ ] `grep -rn "Bash(\|Task(\|TodoWrite\|AskUserQuestion\|CLAUDE\.md\|superpowers:" skills/` returns 0
- [ ] All skill SKILL.md files have correct YAML frontmatter (name, description)
- [ ] All agent .agent.md files have name, model, tools, target: vscode
- [ ] AGENTS.md Common Mistakes updated with any new findings
- [ ] GLOSSARY.md updated with any new terms
- [ ] ONBOARDING.md Key Files Map current
- [ ] NEXT_STEPS.md session log updated
```

