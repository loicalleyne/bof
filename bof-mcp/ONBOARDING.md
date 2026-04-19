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
    │       └─► slot = iteration % 3 → select from bofPool
    │       └─► crush run --model {model} --quiet (review prompt via stdin)
    │       └─► write .adversarial/{slug}.json (updated state)
    │
    ├─► gate_review
    │       └─► scan .adversarial/*.json → check last_verdict
    │
    └─► discover_models
            └─► background probe: crush models → sequential crush run probe
            └─► cache: ~/.config/bof/model-cache.json (atomic write, 1-day TTL)
            └─► returns snapshot (RLock, copy-before-return)
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

1. Handler built in `newAdversarialHandler` (in `adversarial.go`).
2. `validateSlug(input.PlanSlug)` — rejects path traversal.
3. `ReadState(projectRoot, input.PlanSlug)` — reads `.adversarial/{slug}.json` or zero value.
4. `excludeModelFromPool(bofPool, input.ExcludeModel)` — removes caller's model (exact match).
5. `slot = state.Iteration % 3` → `selectedModel = bofPool[slot]` (stable rotation); if filtered out, fall back to `pool[slot%len(pool)]`.
6. Embedded reference content (7-attack protocol) prepended to review preamble.
7. `RunCrush(ctx, selectedModel, reviewPrompt)`.
8. Verdict extracted via regex `(?m)^Verdict:\s*(PASSED|CONDITIONAL|FAILED)`.
9. State updated and written atomically via `WriteState`.
10. Verdict + output returned.

### `gate_review`

1. `filepath.Glob(.adversarial/*.json)` in project root.
2. Files filtered to those directly in the directory (not in `reports/` subdir).
3. Each file: unmarshal `last_verdict` field.
4. Any FAILED or missing verdict → `blocked=true`.
5. `strict=true`: CONDITIONAL also blocks.

### `discover_models`

1. `prober.currentState()` acquires `RLock`, copies `cache.Entries` slice, returns snapshot.
2. `force_refresh=true`: cancels current probe context, resets `probing=true`, launches new goroutine.
3. Returns JSON: `models`, `probing`, `stale`, `cached_at`, `probe_errors`, `cache_error`.

### Background probe goroutine

1. `crush models` subprocess → parse output: one `provider/modelID` per line (non-TTY format).
2. Strip whitespace, skip empty and `#`-prefixed lines.
3. For each model ID: `crush run --model {id} --quiet` with `"1"` as stdin, 15s timeout.
4. Record `ModelEntry{ID, Provider, Available, ProbedAt}`.
5. Write to temp file → `os.Rename` to `~/.config/bof/model-cache.json`.
6. Update `prober.cache` under `Lock`, set `probing=false`, close `done` channel.
7. Entire goroutine body wrapped in `recover()` — server never crashes on probe panic.

---

## Key Files

| File | What to read first |
|------|--------------------|
| `main.go` | Server startup order: flags → crush check → prober → server → register tools → launch probe → Run |
| `runner.go` | `RunCrush` — the single subprocess invocation function |
| `dispatch.go` | `stripFrontmatter`, `newImplementerHandler`, `resolveModel` |
| `adversarial.go` | `bofPool`, `newAdversarialHandler`, `excludeModelFromPool` |
| `models.go` | `modelProber`, `currentState` (RLock + copy), `saveCache` (atomic write) |
| `state.go` | `validateSlug`, `ReadState`, `WriteState` |
| `tools.go` | `registerTools` — where `--no-adversarial` gates tool registration |
| `embedded/` | Source files for `//go:embed` (must not use `..`) |

---

## Startup Order

```
1. Parse flags (--project-root, --no-adversarial, --default-model)
2. Resolve projectRoot (flag > env > PWD)
3. exec.LookPath("crush") — WARN if missing, continue
4. newModelProber(cachePath, TTL)
5. mcp.NewServer(...)
6. registerTools(server, projectRoot, noAdversarial, defaultModel, prober)
7. go prober.startProbeIfStale(ctx)   ← non-blocking; server already registered
8. server.Run(ctx, &mcp.StdioTransport{})   ← blocks until shutdown
```

The probe launches AFTER tool registration so the server can accept `discover_models` calls immediately, even while the probe is running.

---

## Concurrency Notes

- MCP Go SDK stdio transport processes tool handlers one at a time (serial).
- `modelProber` runs a separate goroutine — the only source of concurrent writes to shared state.
- `sync.RWMutex` on `modelProber`: `RLock` for `currentState()` reads; `Lock` for probe state transitions and `saveCache` writes.
- `currentState()` copies `cache.Entries` before releasing `RLock` — callers iterate the copy after unlock with no race.
- `WriteState` in `state.go` uses atomic write; `state.go` itself is only called from serial tool handlers (no concurrent access to state files).
