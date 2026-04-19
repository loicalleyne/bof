# GLOSSARY.md — bof-mcp

Domain vocabulary for `bof-mcp`. Use these exact terms in code, comments, and documentation.

---

## A

**adversarial pool** (`bofPool`)
The 3-element `[]string` in `adversarial.go` containing model IDs for adversarial review rotation: `copilot/gpt-4.1` (slot 0), `copilot/claude-opus-4` (slot 1), `copilot/gpt-4o` (slot 2).

**adversarial review**
A single-round critique where an LLM is prompted to attack a plan using the 7-attack protocol. In bof-mcp, one round = one `crush run --model {model}` call with the review prompt. The implementing model is excluded from the pool via `exclude_model`.

**atomic write**
The pattern of writing data to a temp file in the same directory as the target, then calling `os.Rename` to replace the target. Prevents partial-write corruption if the process is killed mid-write. Used in both `saveCache` (models.go) and `WriteState` (state.go).

---

## B

**bof-mcp**
The Go MCP stdio server in `bof/bof-mcp/`. Exposes 6 tools: `discover_models`, `implementer_agent`, `spec_review`, `quality_review`, `adversarial_review`, `gate_review`. The only non-markdown component of bof.

---

## C

**crush**
The external CLI binary (`github.com/charmbracelet/crush`) that bof-mcp invokes as a subprocess. Used via `crush run --model {id} --quiet` (agent dispatch) and `crush models` (model listing).

**Crush Mode**
A `## Crush Mode (bof-mcp)` section in bof SKILL.md files that describes how to replace VS Code `runSubagent(...)` calls with bof-mcp MCP tool calls.

---

## D

**default model**
The model ID used by dispatch tools (`implementer_agent`, `spec_review`, `quality_review`) when the caller does not supply a `model` param. Set via `--default-model` flag or `BOF_DEFAULT_MODEL` env var.

**dispatch tool**
Any of the three agent-dispatch MCP tools: `implementer_agent`, `spec_review`, `quality_review`. Each strips frontmatter from its embedded agent `.md` file, prepends instructions, and runs `crush run --model {model} --quiet`.

---

## E

**embedded files**
Source files copied to `bof-mcp/embedded/` to satisfy the Go embed package requirement that `//go:embed` patterns must not contain `..`. Includes agent `.agent.md` files and adversarial review reference `.md` files.

**`exclude_model`**
The `adversarial_review` param that removes a specific model from the pool for a single call. Exact case-insensitive equality match against full `provider/model` strings. If the value would empty the pool, it is silently ignored (fail-open).

---

## F

**frontmatter stripping**
Removing the YAML frontmatter block (`---\n...\n---\n`) from agent `.agent.md` file content before passing to `crush run`. The frontmatter contains VS Code–specific keys (`target`, `user-invocable`, `tools`, `agents`) that are irrelevant to Crush.

---

## M

**ModelCache**
The struct in `models.go` that holds probed `ModelEntry` records, `CachedAt` timestamp, and `ProbeCompleted` flag. Persisted to `~/.config/bof/model-cache.json`. Read via `modelProber.currentState()` (RLock); written atomically by the probe goroutine (Lock + temp file + rename).

**modelProber**
The struct in `models.go` that owns the background probe goroutine, the `ModelCache`, and the `sync.RWMutex`. The single source of truth for model availability state.

---

## P

**plan slug**
The basename of a plan document file (without `.md`), used as the key for `.adversarial/{slug}.json` state files. Must not contain `/`, `\`, or `..`. Validated by `validateSlug`.

**probe**
The background process that runs `crush models` to list model IDs, then sequentially tests each with `crush run --model {id} --quiet` and a 15s timeout, recording availability in `ModelCache`.

**probing state**
When `modelProber.probing == true`, a probe goroutine is running. `discover_models` returns `"probing": true` in this state. Concurrent calls return immediately (no write-lock contention) thanks to `sync.RWMutex`.

---

## R

**`RunCrush`**
The function in `runner.go` with signature `RunCrush(ctx context.Context, model, promptContent string) (RunResult, error)`. Invokes `crush run --model {model} --quiet` with `cmd.Stdin = strings.NewReader(promptContent)`. Never interpolates model or prompt into a shell string.

**`RunResult`**
`struct { Output string; ExitCode int }` — the result of a `RunCrush` call.

**rotation slot**
`slot = state.Iteration % 3` — determines which model from `bofPool` is used for the next adversarial review round.

---

## S

**state file**
`.adversarial/{plan-slug}.json` — persists adversarial review state for a plan: `plan_slug`, `iteration`, `last_model`, `last_verdict`, `last_review_date`. Schema matches `gate-review.sh` field names exactly.

**`validateSlug`**
The function in `state.go` that rejects plan slugs containing `/`, `\`, `..`, or empty string. Allows single `.`. Used as a path traversal guard before constructing `.adversarial/` file paths.
