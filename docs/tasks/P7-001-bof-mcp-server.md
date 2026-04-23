# P7-001 — bof-mcp-server: bof MCP Server

**Phase:** P7 — Crush Compatibility  
**Status:** Done  
**Created:** 2026-04-19  
**Spec:** `docs/specs/2026-04-19-bof-crush-compatibility-design.md`

---

## Goal

Create `bof/bof-mcp/` — a standalone Go MCP server that gives Crush (and VS Code via MCP) the equivalent of VS Code's `runSubagent` and adversarial review agent dispatch.

---

## In Scope

- `bof-mcp/go.mod` — module `github.com/loicalleyne/bof/bof-mcp`
- `bof-mcp/main.go` — CLI entry point; flag parsing (`--project-root`, `--no-adversarial`, `--default-model`); env var overrides (`BOF_PROJECT_ROOT`, `BOF_NO_ADVERSARIAL`, `BOF_DEFAULT_MODEL`); MCP server start
- `bof-mcp/tools.go` — `registerTools()`: conditionally registers `adversarial_review` + `gate_review` (skipped when `--no-adversarial`); always registers `implementer_agent`, `spec_review`, `quality_review`
- `bof-mcp/runner.go` — `RunCrush(ctx, model, promptContent string) (RunResult, error)`: invokes `crush run --model {model} --quiet` with prompt piped via stdin (NOT passed as shell argument — security invariant); same pattern as `esquisse-mcp/runner.go`
- `bof-mcp/adversarial.go` — `adversarial_review` and `gate_review` handlers; `//go:embed ../skills/adversarial-review/references/*.md`; 5-slot model pool from `BOF_MODELS` env var (or defaults); reads/writes `.adversarial/{plan_slug}.json`; `exclude_model` filter
- `bof-mcp/dispatch.go` — `implementer_agent`, `spec_review`, `quality_review` handlers; `//go:embed ../agents/ImplementerAgent.agent.md ../agents/SpecReviewerAgent.agent.md ../agents/CodeQualityReviewerAgent.agent.md`; frontmatter stripping before prepending preamble to prompt; `model` param optional with `--default-model` fallback
- `bof-mcp/models.go` — Adversarial model pool management, `BOF_MODELS` env var parsing, `defaultModels` list.
- `bof-mcp/state.go` — `.adversarial/{slug}.json` read/write helpers; `validateSlug` rejects slugs containing `/`, `\`, or `..`
- `bof-mcp/README.md` — self-configuration docs (see spec Section 4); verbatim `crush.json` and `.vscode/mcp.json` snippets; `--no-adversarial` coexistence note; tool reference table.

---

## Out of Scope

- Windows-native path support (`\\` paths, `cmd.exe` invocation)
- `go.work` workspace file — bof-mcp is the only Go code; no workspace needed
- Streaming output from `crush run` — collect full output then return
- Parallel model probing — sequential only
- Any UI, web server, or HTTP transport — stdio MCP only
- Installing the binary — user builds from source per README

---

## Files

| Path | Action | What |
|---|---|---|
| `bof-mcp/go.mod` | Create | Standalone module `github.com/loicalleyne/bof/bof-mcp` |
| `bof-mcp/main.go` | Create | CLI flags, env vars, server startup |
| `bof-mcp/tools.go` | Create | `registerTools` — conditional + unconditional tool registration |
| `bof-mcp/runner.go` | Create | `RunCrush` subprocess invocation (stdin-only, no shell interpolation) |
| `bof-mcp/adversarial.go` | Create | `adversarial_review`, `gate_review` handlers + embedded references |
| `bof-mcp/dispatch.go` | Create | `implementer_agent`, `spec_review`, `quality_review` + frontmatter strip |
| `bof-mcp/models.go` | Create | Adversarial model pool management |
| `bof-mcp/state.go` | Create | `.adversarial/` state file read/write + slug validation |
| `bof-mcp/README.md` | Create | Self-config docs (verbatim JSON snippets, tool reference table) |

---

## Implementation Notes

- **Primary reference:** read `esquisse/esquisse-mcp/runner.go`, `adversarial.go`, `state.go`, `tools.go` before writing any file. The pattern is identical; adapt for bof's embed paths and the three new dispatch tools.
- **Frontmatter stripping:** find the first `---\n`, find the next `---\n` after it, discard both delimiters and everything between them. If no closing `---\n` is found, use the full content unchanged.
- **Security invariant for `RunCrush`:** the model ID and prompt content MUST NOT be interpolated into a shell command string. Pass model as a separate `exec.Command` argument; pass prompt via `cmd.Stdin`. Same invariant as `esquisse-mcp`.
- **Slug validation in `state.go`:** `validateSlug` must reject any value containing `/`, `\`, `.`, or being empty. `filepath.Join(projectRoot, ".adversarial", slug+".json")` is safe only after this guard.
- **MCP SDK:** use `github.com/modelcontextprotocol/go-sdk` — same dependency as `esquisse-mcp`. Check `esquisse-mcp/go.mod` for the pinned version and use the same.

---

## Acceptance Criteria

1. `cd bof-mcp && go build -o bof-mcp .` exits 0 and produces a binary.
2. `./bof-mcp --help` prints all three flags (`--project-root`, `--no-adversarial`, `--default-model`).
3. Server started without `--no-adversarial`: `mcp tools ./bof-mcp` lists all 5 tools (`adversarial_review`, `gate_review`, `implementer_agent`, `spec_review`, `quality_review`).
4. Server started with `--no-adversarial`: `mcp tools ./bof-mcp --no-adversarial` lists only 3 tools (`adversarial_review` and `gate_review` absent).
5. Frontmatter stripping: given input `"---\nname: X\n---\n# Body\ntext"`, the stripped output is `"# Body\ntext"`.
6. `RunCrush` uses `exec.Command(crushPath, "run", "--model", model, "--quiet")` with `cmd.Stdin = strings.NewReader(promptContent)` — model is a separate arg, NOT part of a shell string, prompt is NOT a temp file.
7. Starting the server when `crush` is not in PATH logs a WARN and does not abort.
8. `RunCrush` error when `crush` not in PATH includes the string "crush binary not found in PATH".
9. `validateSlug("../../evil")` returns an error; `validateSlug("my-plan")` and `validateSlug("my.plan")` both return nil.
10. Cache is written atomically: a temp file is created, written, then renamed to the final path; no partial-write state is possible.
11. `bof-mcp/README.md` contains a verbatim `crush.json` snippet with `"command"` and `"args"` keys and a separate `--no-adversarial` variant.
12. `go vet ./...` inside `bof-mcp/` reports no issues.

---

## Session Notes

- **`esquisse-mcp` is the reference but `RunCrush` intentionally differs:** esquisse-mcp passes a temp file path as prompt source; bof-mcp uses `strings.NewReader(promptContent)` assigned to `cmd.Stdin`. Add a code comment in `runner.go` noting this deviation.
- **`RunCrush` signature:** `RunCrush(ctx context.Context, model, promptContent string) (RunResult, error)` — accepts a string, not a file path.
- **Startup warning:** if `exec.LookPath("crush")` fails at startup, log `WARN: crush binary not found in PATH — dispatch tools will fail until crush is installed`. Do not abort. `RunCrush` error message must include "crush binary not found in PATH — install crush first".
- **Slug validation:** `validateSlug` rejects `..`, `/`, `\`, and empty string. Single `.` is allowed. Use `filepath.Clean` on the joined path and verify it stays within `.adversarial/` as a secondary guard.
- **Go version:** set `go 1.22` minimum in `go.mod` (matches esquisse-mcp).
- The `adversarial_review` tool uses a **5-slot** default pool, configurable via `BOF_MODELS`.
- The `.agent.md` frontmatter keys (`target`, `user-invocable`, `tools`, `agents`) are VS Code-only — strip with the frontmatter stripper before sending to `crush run`.

2026-04-19 — Completed. All 9 files created; Go embed limitation (no `..` paths) resolved by copying agent/.md files to `bof-mcp/embedded/`; binary builds clean with `go vet` passing and all unit tests green.
