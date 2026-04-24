# ONBOARDING.md — bof-mcp

## Read This First

You are working in `bof-mcp`, a small Go MCP stdio server that connects AI agents using the bof framework to LLMs via the `crush` CLI tool.

Before writing any code, read:
1. [`AGENTS.md`](AGENTS.md) — invariants, security rules, common mistakes
2. [`GLOSSARY.md`](GLOSSARY.md) — exact term definitions
3. [`README.md`](README.md) — build, configuration, tool reference
4. [`llms.txt`](llms.txt) — concise API index

---

## Mental Model

```
Crush Agent (or VS Code via MCP)
    │
    │  MCP stdio
    ▼
bof-mcp (this binary)
    │
    ├─► implementer_agent / spec_review / quality_review
    │       └─► frontmatter strip embedded .agent.md
    │       └─► crush run --model {id} --quiet (prompt via stdin)
    │
    ├─► adversarial_review
    │       └─► read .adversarial/{slug}.json (iteration state)
    │       └─► buildModelPool() → buildRotationOrder(pool, rounds)
    │       └─► runOneRound() per round → crush run --model {model} --quiet (prompt via stdin)
    │       └─► worstVerdict(verdicts) → write .adversarial/{slug}.json
    │
    ├─► gate_review
    │       └─► scan .adversarial/*.json → check last_verdict
    │
    
```

---

## Request Lifecycle

### `implementer_agent` / `spec_review` / `quality_review`

1. Handler built in `newImplementerHandler` / `newSpecReviewHandler` / `newQualityReviewHandler` (in `dispatch.go`) at server start.
2. Embedded agent `.md` loaded via `//go:embed embedded/agents/{Name}.agent.md`; frontmatter stripped at startup (not per-call).
3. Handler called with `{implementer,specReview,qualityReview}Input`.
4. `resolveModel(input.Model, defaultModel)` — uses request model or falls back to `--default-model`.
5. Prompt = stripped agent instructions + preamble (project root, role) + caller's task/spec/code content.
6. `RunCrush(ctx, model, prompt)` invoked — `strings.NewReader(prompt)` → `cmd.Stdin`.
7. Full combined output returned as MCP text response.

### `adversarial_review`

1. Handler built in `newAdversarialHandler` (in `adversarial.go`); `buildModelPool()` called once at handler construction.
2. `validateSlug(input.PlanSlug)` — rejects path traversal.
3. `ReadState(projectRoot, input.PlanSlug)` — reads `.adversarial/{slug}.json` or zero value.
4. `excludeModelFilter(pool, input.ExcludeModel)` — removes caller's model (exact case-insensitive match); fail-open if would empty pool.
5. `rounds = effectiveRounds(input.Rounds)` — clamps to [1, 50], default 5.
6. `rotOrder = buildRotationOrder(effectivePool, rounds)` — family-interleaved shuffle across N rounds.
7. For each round: embedded reference content (7-attack protocol) prepended to preamble; `runOneRound(ctx, pool, rotOrder[i], preamble, planContent)` called (stdin-based, fallback through pool on unavailable).
8. Verdict extracted from output or from written report file via regex.
9. After all rounds: `worstVerdict(verdicts)` determines final verdict; state updated with `Iteration += rounds`, written atomically via `WriteState`.
10. Multi-round summary returned.

### `gate_review`

1. `filepath.Glob(.adversarial/*.json)` in project root.
2. Files filtered to those directly in the directory (not in `reports/` subdir).
3. Each file: unmarshal `last_verdict` field.
4. Any FAILED or missing verdict → `blocked=true`.
5. `strict=true`: CONDITIONAL also blocks.



---

## Key Files

| File | What to read first |
|------|--------------------|
| `main.go` | Server startup order: flags → crush check → server → register tools → Run |
| `runner.go` | `RunCrush` — the single subprocess invocation function |
| `dispatch.go` | `stripFrontmatter`, `newImplementerHandler`, `resolveModel` |
| `adversarial.go` | `loadReferenceContent`, `newAdversarialHandler` |
| `models.go` | `buildModelPool`, `buildRotationOrder`, `runOneRound`, `worstVerdict` |
| `state.go` | `validateSlug`, `ReadState`, `WriteState` |
| `tools.go` | `registerTools` — where `--no-adversarial` gates tool registration |
| `embedded/` | Source files for `//go:embed` (must not use `..`) |

---

## Startup Order

```
1. Parse flags (--project-root, --no-adversarial, --default-model)
2. Resolve projectRoot (flag > env > PWD)
3. exec.LookPath("crush") — WARN if missing, continue
4. mcp.NewServer(...)
5. registerTools(server, projectRoot, defaultModel, noAdversarial)
6. server.Run(ctx, &mcp.StdioTransport{})   ← blocks until shutdown
```



---

## Concurrency Notes

- MCP Go SDK stdio transport processes tool handlers one at a time (serial).
- `WriteState` in `state.go` uses atomic write; `state.go` itself is only called from serial tool handlers (no concurrent access to state files).
