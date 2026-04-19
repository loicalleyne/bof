# bof-mcp

bof-mcp is an MCP server that gives Crush (and VS Code via MCP) Crush-compatible agent dispatch tools plus `discover_models` for model discovery.

It exposes the bof framework's agent workflow — adversarial review, implementer dispatch, spec review, and code quality review — as MCP tools that any MCP-compatible client can call.

---

## Build

```sh
cd bof-mcp && go build -o bof-mcp .
```

Requires Go 1.22+ and the `crush` binary in PATH at runtime.

---

## Configuration

### crush.json

Add bof-mcp as an MCP server in your `crush.json`:

```json
{
  "mcpServers": {
    "bof-mcp": {
      "command": "/path/to/bof-mcp",
      "args": ["--project-root", "/path/to/project"]
    }
  }
}
```

With `--no-adversarial` (if you also use `esquisse-mcp` and want to avoid tool name conflicts):

```json
{
  "mcpServers": {
    "bof-mcp": {
      "command": "/path/to/bof-mcp",
      "args": ["--project-root", "/path/to/project", "--no-adversarial"]
    }
  }
}
```

### .vscode/mcp.json

Add bof-mcp as an MCP server in VS Code:

```json
{
  "servers": {
    "bof-mcp": {
      "type": "stdio",
      "command": "/path/to/bof-mcp",
      "args": ["--project-root", "/path/to/project"]
    }
  }
}
```

---

## Flags

| Flag | Env Override | Default | Description |
|---|---|---|---|
| `--project-root` | `BOF_PROJECT_ROOT` | `$PWD` | Project root directory (used for `.adversarial/` state files) |
| `--no-adversarial` | `BOF_NO_ADVERSARIAL` | `false` | Disable `adversarial_review` and `gate_review` tools |
| `--default-model` | `BOF_DEFAULT_MODEL` | `""` | Default model ID for dispatch tools when not specified per-call |

---

## Tool Reference

> **Note:** Call `discover_models` first to see which models are available before using dispatch tools.

| Tool | Description | Required params |
|---|---|---|
| `adversarial_review` | One adversarial review round via 3-slot rotation pool. Pool: `copilot/gpt-4.1` / `copilot/claude-opus-4` / `copilot/gpt-4o`. Writes verdict to `.adversarial/{plan_slug}.json`. | `plan_slug`, `plan_content` |
| `gate_review` | Check whether all `.adversarial/` verdicts are PASSED or CONDITIONAL. | none (optional: `strict`) |
| `implementer_agent` | Dispatch an ImplementerAgent to implement a task document following TDD. | `task_content` |
| `spec_review` | Dispatch a SpecReviewerAgent to review a specification document. | `spec_content` |
| `quality_review` | Dispatch a CodeQualityReviewerAgent to review code or a diff. | `code_content` |
| `discover_models` | List available crush models and their availability status. | none (optional: `filter`, `force_refresh`) |

**`adversarial_review` and `gate_review` are hidden when `--no-adversarial` is set.**

---

## Model Discovery

bof-mcp probes model availability in the background on startup. The first call to `discover_models` may return `probing: true` while the probe is running (typically 30–90 seconds for ~5 models at 15s timeout each).

The model cache is stored at `~/.config/bof/model-cache.json` with a default TTL of 1 day. Override with `BOF_MODEL_CACHE_TTL_DAYS`.

`discover_models` returns a JSON object with:
- `models`: array of `{id, provider, available, probed_at}` objects
- `probing`: `true` while a probe is running
- `stale`: `true` if the cache is older than the TTL
- `cached_at`: ISO-8601 timestamp of last successful probe
- `probe_errors`: array of per-model error strings (present if all probes failed)
- `cache_error`: error string if the cache write failed (the tool still returns results)

---

## Coexistence with esquisse-mcp

If you also use `esquisse-mcp`, add `--no-adversarial` to bof-mcp's args to avoid tool name conflicts (`adversarial_review` and `gate_review` are registered by both).

```json
{
  "mcpServers": {
    "esquisse-mcp": {
      "command": "/path/to/esquisse-mcp",
      "args": ["--project-root", "/path/to/project"]
    },
    "bof-mcp": {
      "command": "/path/to/bof-mcp",
      "args": ["--project-root", "/path/to/project", "--no-adversarial"]
    }
  }
}
```

---

## Security

- Model IDs and prompt content are **never** interpolated into shell command strings.
- Model is passed as a separate `exec.Command` argument; prompt is passed via `cmd.Stdin`.
- Slug validation rejects any `.adversarial/` state file name containing `/`, `\`, or `..`.
