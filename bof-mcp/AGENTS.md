# AGENTS.md — bof-mcp

## The Most Important Rules

1. **Security invariant — no shell interpolation.** Model IDs and prompt content must never be interpolated into a shell command string. Pass model as a separate `exec.Command` argument; pass prompt via `cmd.Stdin = strings.NewReader(promptContent)`.

2. **Tests define correctness.** Fix code to match tests, never the reverse.

3. **All errors returned, never panicked.** `mcpErr` wraps errors as MCP `IsError=true` responses.


5. **Atomic cache writes.** Cache files are always written via temp file + `os.Rename`. No direct overwrite.

6. **`//go:embed` prohibits `..`.** All embed sources must live under `bof-mcp/embedded/`. Copying source files to `embedded/` is required.

---

## Project Overview

`bof-mcp` is a Go MCP stdio server that gives Crush (and VS Code via MCP) the equivalent of VS Code's `runSubagent` agent dispatch.

It exposes 6 tools:
- **`implementer_agent`** — dispatches ImplementerAgent role via `crush run`
- **`spec_review`** — dispatches SpecReviewerAgent role via `crush run`
- **`quality_review`** — dispatches CodeQualityReviewerAgent role via `crush run`
- **`adversarial_review`** — runs one adversarial review round via 3-slot model pool
- **`gate_review`** — checks `.adversarial/` verdict status

`adversarial_review` and `gate_review` are disabled when `--no-adversarial` is set (for coexistence with `esquisse-mcp`).

**Module:** `github.com/loicalleyne/bof/bof-mcp`
**Go version:** 1.26.2+
**Runtime dependency:** `crush` binary in PATH

---

## Repository Layout

```
bof-mcp/
├── AGENTS.md           ← This file
├── GLOSSARY.md         ← Domain vocabulary
├── ONBOARDING.md       ← Agent orientation
├── llms.txt            ← Concise API index
├── README.md           ← Build, config, tool reference
│
├── main.go             ← CLI flags, env vars, server startup, background probe launch
├── tools.go            ← registerTools() — conditional + unconditional tool registration
├── runner.go           ← RunCrush() — subprocess invocation via strings.NewReader stdin
├── adversarial.go      ← adversarial_review + gate_review handlers; 3-slot pool
├── dispatch.go         ← implementer_agent, spec_review, quality_review + frontmatter strip
├── models.go           ← Adversarial model pool management
├── state.go            ← .adversarial/ state r/w, validateSlug
│
├── embedded/           ← Files copied here for //go:embed (no .. allowed)
│   ├── agents/         ← Copies of ImplementerAgent, SpecReviewerAgent, CodeQualityReviewerAgent .agent.md
│   └── references/     ← Copies of adversarial-review/references/*.md
│
├── go.mod
└── go.sum
```

---

## Build Commands

```sh
cd bof-mcp
go build -o bof-mcp .
go vet ./...
```

---

## Test Commands

```sh
cd bof-mcp
go test -count=1 ./...
go test -count=1 -race ./...
```

---

## Key Dependencies

| Dependency | Role |
|-----------|------|
| `github.com/modelcontextprotocol/go-sdk v1.5.0` | MCP stdio server SDK |
| `crush` (runtime) | LLM dispatcher — `crush run --model {id} --quiet` |

---

## Code Conventions

- **`RunCrush(ctx, model, promptContent string) (RunResult, error)`** — accepts prompt as a string, assigns `strings.NewReader(promptContent)` to `cmd.Stdin`. This differs intentionally from `esquisse-mcp` (which uses a temp file). Both patterns are secure.
- **Frontmatter stripping:** discard `---\n...---\n` block from agent `.md` files before prepending to prompt. If no closing `---\n` is found, use full content.
- **3-slot adversarial pool:** slot 0 = `copilot/gpt-4.1`, slot 1 = `copilot/claude-opus-4`, slot 2 = `copilot/gpt-4o`. Rotation: `slot = state.Iteration % 3` against the full `bofPool` (not the filtered pool).
- **`validateSlug`** rejects `/`, `\`, `..`, and empty string. Allows single `.`. Secondary guard: `filepath.Clean`.
- **Startup WARN** if `crush` not in PATH — log and continue, do not abort.

---

## Security Boundaries

- Model IDs and prompt content are NEVER interpolated into shell strings. `exec.Command` receives model as a discrete argument; prompt arrives via `cmd.Stdin`.
- `validateSlug` prevents `.adversarial/` path traversal.
- No credentials, tokens, or API keys are handled by bof-mcp (crush manages LLM auth).

---

## Common Mistakes to Avoid

1. **`//go:embed` with `..` paths** — fails at compile time. Copy files to `embedded/` and use relative paths within `bof-mcp/`.
2. **`entries = p.cache.Entries` without copying** — data race: probe goroutine modifies elements in-place. Use `make` + `copy` under the `RLock`.
3. **`state.Iteration % len(filteredPool)` for rotation** — breaks when `exclude_model` shrinks the pool. Always use `state.Iteration % 3` against `bofPool` first.
4. **Direct cache file overwrite** — atomic write via temp file + `os.Rename` is required.

---

## References

- [`README.md`](README.md) — configuration, JSON snippets, tool reference
- [`ONBOARDING.md`](ONBOARDING.md) — request lifecycle, data flow
- [`GLOSSARY.md`](GLOSSARY.md) — domain vocabulary
- [`llms.txt`](llms.txt) — concise API index
- Parent: [`../AGENTS.md`](../AGENTS.md) — bof project constitution
- Reference implementation: `github.com/loicalleyne/esquisse/esquisse-mcp/`
