# P7-001 — bof-mcp-server: bof MCP Server

**Phase:** P7 — Crush Compatibility  
**Status:** Done  
**Created:** 2026-04-19  
**Spec:** `docs/specs/2026-04-19-bof-crush-compatibility-design.md`

---

## Goal

Create `bof/bof-mcp/` — a standalone Go MCP server that gives Crush (and VS Code via MCP) the equivalent of VS Code's `runSubagent` and adversarial review agent dispatch, plus a `discover_models` tool that probes available Crush models.

---

## In Scope

- `bof-mcp/go.mod` — module `github.com/loicalleyne/bof/bof-mcp`
- `bof-mcp/main.go` — CLI entry point; flag parsing (`--project-root`, `--no-adversarial`, `--default-model`); env var overrides (`BOF_PROJECT_ROOT`, `BOF_NO_ADVERSARIAL`, `BOF_DEFAULT_MODEL`); background probe goroutine launch on startup if no fresh cache; MCP server start
- `bof-mcp/tools.go` — `registerTools()`: conditionally registers `adversarial_review` + `gate_review` (skipped when `--no-adversarial`); always registers `implementer_agent`, `spec_review`, `quality_review`, `discover_models`
- `bof-mcp/runner.go` — `RunCrush(ctx, model, promptContent string) (RunResult, error)`: invokes `crush run --model {model} --quiet` with prompt piped via stdin (NOT passed as shell argument — security invariant); same pattern as `esquisse-mcp/runner.go`
- `bof-mcp/adversarial.go` — `adversarial_review` and `gate_review` handlers; `//go:embed ../skills/adversarial-review/references/*.md`; 3-slot rotation pool (GPT-4.1 / Claude Opus 4 / GPT-4o); reads/writes `.adversarial/{plan_slug}.json`; `exclude_model` filter
- `bof-mcp/dispatch.go` — `implementer_agent`, `spec_review`, `quality_review` handlers; `//go:embed ../agents/ImplementerAgent.agent.md ../agents/SpecReviewerAgent.agent.md ../agents/CodeQualityReviewerAgent.agent.md`; frontmatter stripping before prepending preamble to prompt; `model` param optional with `--default-model` fallback
- `bof-mcp/models.go` — `discover_models` handler; `ModelCache` struct with `sync.RWMutex` (write lock during probe, read lock for cache reads); `probing bool`; cache at `~/.config/bof/model-cache.json` written atomically (temp file + `os.Rename`); TTL 1 day (`BOF_MODEL_CACHE_TTL_DAYS` override); `force_refresh` param; background probe via `crush models` then sequential `crush run --model {id} --quiet` with stdin `"1"`, 15s per-model timeout; model IDs parsed from `crush models` output are stripped of whitespace and empty lines before use
- `bof-mcp/state.go` — `.adversarial/{slug}.json` read/write helpers; `validateSlug` rejects slugs containing `/`, `\`, or `..`
- `bof-mcp/README.md` — self-configuration docs (see spec Section 4); verbatim `crush.json` and `.vscode/mcp.json` snippets; `--no-adversarial` coexistence note; tool reference table; "call `discover_models` first" note; background probe timing note

---

## Out of Scope

- Windows-native path support (`\\` paths, `cmd.exe` invocation)
- `go.work` workspace file — bof-mcp is the only Go code; no workspace needed
- A `discover_models`-equivalent that queries VS Code Copilot's model list directly (Crush CLI is the only probe mechanism)
- Streaming output from `crush run` — collect full output then return
- Parallel model probing — sequential only
- Any UI, web server, or HTTP transport — stdio MCP only
- Installing the binary — user builds from source per README

---

## Files

| Path | Action | What |
|---|---|---|
| `bof-mcp/go.mod` | Create | Standalone module `github.com/loicalleyne/bof/bof-mcp` |
| `bof-mcp/main.go` | Create | CLI flags, env vars, server startup, background probe |
| `bof-mcp/tools.go` | Create | `registerTools` — conditional + unconditional tool registration |
| `bof-mcp/runner.go` | Create | `RunCrush` subprocess invocation (stdin-only, no shell interpolation) |
| `bof-mcp/adversarial.go` | Create | `adversarial_review`, `gate_review` handlers + embedded references |
| `bof-mcp/dispatch.go` | Create | `implementer_agent`, `spec_review`, `quality_review` + frontmatter strip |
| `bof-mcp/models.go` | Create | `discover_models`, `ModelCache`, background probe goroutine |
| `bof-mcp/state.go` | Create | `.adversarial/` state file read/write + slug validation |
| `bof-mcp/README.md` | Create | Self-config docs (verbatim JSON snippets, tool reference table) |

---

## Implementation Notes

- **Primary reference:** read `esquisse/esquisse-mcp/runner.go`, `adversarial.go`, `state.go`, `tools.go` before writing any file. The pattern is identical; adapt for bof's embed paths and the three new dispatch tools.
- **Frontmatter stripping:** find the first `---\n`, find the next `---\n` after it, discard both delimiters and everything between them. If no closing `---\n` is found, use the full content unchanged.
- **Security invariant for `RunCrush`:** the model ID and prompt content MUST NOT be interpolated into a shell command string. Pass model as a separate `exec.Command` argument; pass prompt via `cmd.Stdin`. Same invariant as `esquisse-mcp`.
- **Slug validation in `state.go`:** `validateSlug` must reject any value containing `/`, `\`, `.`, or being empty. `filepath.Join(projectRoot, ".adversarial", slug+".json")` is safe only after this guard.
- **Non-blocking startup:** the MCP server MUST register all tools and be ready to accept connections before the background probe starts. Launch probe with `go func() { ... }()` after `server.Serve()` is set up, not before.
- **`discover_models` three states:** (1) no cache + probing → `probing:true, models:[]`; (2) cache stale + probing → `stale:true, probing:true, models:<stale list>`; (3) cache fresh → `probing:false, models:<list>`.
- **`force_refresh`:** cancel the current probe context if probing; set `probing=true`; launch new probe goroutine; return immediately with `probing:true`.
- **MCP SDK:** use `github.com/modelcontextprotocol/go-sdk` — same dependency as `esquisse-mcp`. Check `esquisse-mcp/go.mod` for the pinned version and use the same.

---

## Acceptance Criteria

1. `cd bof-mcp && go build -o bof-mcp .` exits 0 and produces a binary.
2. `./bof-mcp --help` prints all three flags (`--project-root`, `--no-adversarial`, `--default-model`).
3. Server started without `--no-adversarial`: `mcp tools ./bof-mcp` lists all 6 tools (`adversarial_review`, `gate_review`, `implementer_agent`, `spec_review`, `quality_review`, `discover_models`).
4. Server started with `--no-adversarial`: `mcp tools ./bof-mcp --no-adversarial` lists only 4 tools (`adversarial_review` and `gate_review` absent).
5. Frontmatter stripping: given input `"---\nname: X\n---\n# Body\ntext"`, the stripped output is `"# Body\ntext"`.
6. `RunCrush` uses `exec.Command(crushPath, "run", "--model", model, "--quiet")` with `cmd.Stdin = strings.NewReader(promptContent)` — model is a separate arg, NOT part of a shell string, prompt is NOT a temp file.
7. Starting the server when `crush` is not in PATH logs a WARN and does not abort.
8. `RunCrush` error when `crush` not in PATH includes the string "crush binary not found in PATH".
9. `validateSlug("../../evil")` returns an error; `validateSlug("my-plan")` and `validateSlug("my.plan")` both return nil.
10. `discover_models` called before probe completes returns `"probing": true` and `"models": []`.
11. After probe completes, `discover_models` returns a non-empty `models` array and `"probing": false`.
12. `discover_models` with `force_refresh: true` returns `"probing": true` immediately.
13. When all model probes fail, `discover_models` response includes a non-empty `"probe_errors"` array.
14. Cache is written atomically: a temp file is created, written, then renamed to the final path; no partial-write state is possible.
15. Concurrent calls to `discover_models` while probe is running return the cached/empty state immediately without blocking on a write lock.
16. `bof-mcp/README.md` contains a verbatim `crush.json` snippet with `"command"` and `"args"` keys and a separate `--no-adversarial` variant.
17. `go vet ./...` inside `bof-mcp/` reports no issues.

---

## Session Notes

- **Verify Crush flags first (do this before writing any Go):** run `crush run --help` and `crush models --help` (or read `internal/cmd/` in the crush repo) to confirm `--quiet` flag and `models` subcommand exist and their exact output format. Do not assume based on esquisse-mcp.
- **`esquisse-mcp` is the reference but `RunCrush` intentionally differs:** esquisse-mcp passes a temp file path as prompt source; bof-mcp uses `strings.NewReader(promptContent)` assigned to `cmd.Stdin`. Add a code comment in `runner.go` noting this deviation.
- **`RunCrush` signature:** `RunCrush(ctx context.Context, model, promptContent string) (RunResult, error)` — accepts a string, not a file path.
- **Startup warning:** if `exec.LookPath("crush")` fails at startup, log `WARN: crush binary not found in PATH — dispatch tools will fail until crush is installed`. Do not abort. `RunCrush` error message must include "crush binary not found in PATH — install crush first".
- **Probe goroutine safety:** wrap probe goroutine body in `defer func() { if r := recover(); r != nil { log.Printf("model probe panic: %v", r); /* set probing=false, set models=[] */ } }()` to prevent server crash.
- **`discover_models` error surface:** if all probes fail, include `"probe_errors": ["model X: reason", ...]` in response. If cache write fails, include `"cache_error": "..."` in response; do not fail the tool call.
- **Cache writes are atomic:** write to a temp file in the same directory, then `os.Rename` to the final path. This prevents partial-write corruption if the process is killed mid-write.
- **`ModelCache` uses `sync.RWMutex`:** `RLock` for reads (allow concurrent `discover_models` calls to return cached data without blocking each other); `Lock` only during probe state transitions and cache writes. This prevents mutex contention blocking all callers during a 150s sequential probe.
- **Model ID sanitization:** strip leading/trailing whitespace from each line of `crush models` output; skip empty lines and lines starting with `#`. Model IDs are passed as separate `exec.Command` arguments (not shell-interpolated) so there is no shell injection risk, but validating the format prevents surprises from malformed output.
- **Slug validation:** `validateSlug` rejects `..`, `/`, `\`, and empty string. Single `.` is allowed. Use `filepath.Clean` on the joined path and verify it stays within `.adversarial/` as a secondary guard.
- **Go version:** set `go 1.22` minimum in `go.mod` (matches esquisse-mcp).
- The `adversarial_review` tool uses a **3-slot** pool (not 5-slot) to match bof's `@Adversarial-r0/r1/r2` rotation.
- The `.agent.md` frontmatter keys (`target`, `user-invocable`, `tools`, `agents`) are VS Code-only — strip with the frontmatter stripper before sending to `crush run`.

2026-04-19 — Completed. All 9 files created; Go embed limitation (no `..` paths) resolved by copying agent/.md files to `bof-mcp/embedded/`; binary builds clean with `go vet` passing and all unit tests green.
