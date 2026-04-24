# GLOSSARY.md — bof-mcp

Domain vocabulary for `bof-mcp`. Use these exact terms in code, comments, and documentation.

---

## A

**adversarial pool**
The `[]string` of model IDs used for adversarial review. Built at handler construction by `buildModelPool()`, which reads `BOF_MODELS` (comma-separated `provider/model`) or falls back to `defaultModels` (5 entries). Rotation is family-interleaved across N rounds via `buildRotationOrder`.

**adversarial review**
A multi-round critique where an LLM is prompted to attack a plan using the 7-attack protocol. In bof-mcp, each round = one `crush run --model {model}` call with the review prompt. `rounds` (default 5, max 50) rounds are run per `adversarial_review` invocation. The implementing model is excluded from the pool via `exclude_model`. The worst verdict across all rounds is written to the state file.

**atomic write**
The pattern of writing data to a temp file in the same directory as the target, then calling `os.Rename` to replace the target. Prevents partial-write corruption if the process is killed mid-write. Used by `WriteState` (state.go).

---

## B

**bof-mcp**
The Go MCP stdio server in `bof/bof-mcp/`. Exposes 5 tools: `implementer_agent`, `spec_review`, `quality_review`, `adversarial_review`, `gate_review`. The only non-markdown component of bof.

---

## C

**crush**
The external CLI binary (`github.com/charmbracelet/crush`) that bof-mcp invokes as a subprocess. Used via `crush run --model {id} --quiet` for agent dispatch and adversarial review.

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
The `adversarial_review` param that removes a specific model from the pool for a single call. Exact case-insensitive equality match against full `provider/model` strings. If the value would empty the pool, it is silently ignored (fail-open). Applied via `excludeModelFilter`.

---

## F

**frontmatter stripping**
Removing the YAML frontmatter block (`---\n...\n---\n`) from agent `.agent.md` file content before passing to `crush run`. The frontmatter contains VS Code–specific keys (`target`, `user-invocable`, `tools`, `agents`) that are irrelevant to Crush.



---

## P

**plan slug**
The basename of a plan document file (without `.md`), used as the key for `.adversarial/{slug}.json` state files. Must not contain `/`, `\`, or `..`. Validated by `validateSlug`.

**`buildRotationOrder`**
Function in `models.go` that produces a `[]string` of model IDs for N rounds of adversarial review using family-interleaved shuffle. Ensures models from different provider families alternate rather than clustering consecutive rounds.

**`runOneRound`**
Function in `models.go` with signature `runOneRound(ctx, pool, targetModel, preamble, planContent string) (usedModel, output string, err error)`. Attempts the target model first; if `isModelUnavailable` (enterprise policy block), falls back through remaining pool models.

**`worstVerdict`**
Function in `models.go` that returns the most pessimistic verdict from a `[]string` of round verdicts. Order: `FAILED` > `CONDITIONAL` > `PASSED` > `""`.



---

## R

**`RunCrush`**
The function in `runner.go` with signature `RunCrush(ctx context.Context, model, promptContent string) (RunResult, error)`. Invokes `crush run --model {model} --quiet` with `cmd.Stdin = strings.NewReader(promptContent)`. Never interpolates model or prompt into a shell string.

**`RunResult`**
`struct { Output string; ExitCode int }` — the result of a `RunCrush` call.

**rotation slot**
The position in `buildRotationOrder`'s output used for a given round. Family-interleaved shuffle spreads model families across rounds; not a simple modulo-3 ring.

---

## S

**state file**
`.adversarial/{plan-slug}.json` — persists adversarial review state for a plan: `plan_slug`, `iteration`, `last_model`, `last_verdict`, `last_review_date`. Schema matches `gate-review.sh` field names exactly.

**`validateSlug`**
The function in `state.go` that rejects plan slugs containing `/`, `\`, `..`, or empty string. Allows single `.`. Used as a path traversal guard before constructing `.adversarial/` file paths.
