# bof-mcp

bof-mcp is an MCP server that gives Crush (and VS Code via MCP) Crush-compatible agent dispatch tools.

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

### Model Pool (`adversarial_review`)

The adversarial review pool defaults to 5 models with family-interleaved rotation:

```
copilot/claude-sonnet-4.6
copilot/gpt-4.1
copilot/claude-opus-4
copilot/gpt-4o
gemini/gemini-2.5-pro-preview-05-06
```

Override with `BOF_MODELS` (comma-separated `provider/model` entries):

```sh
export BOF_MODELS="copilot/gpt-4.1,copilot/claude-opus-4,gemini/gemini-2.5-pro-preview-05-06"
```

---

## Tool Reference


| Tool | Description | Required params | Optional params |
|---|---|---|---|
| `adversarial_review` | Run N adversarial review rounds using a family-interleaved model pool. Reads `.adversarial/{plan_slug}.json` for state, writes worst verdict on completion. Pool configured via `BOF_MODELS`. | `plan_slug`, `plan_content` | `rounds` (default 5, max 50), `exclude_model` |
| `gate_review` | Check whether all `.adversarial/` verdicts are PASSED or CONDITIONAL. | — | `strict` |
| `implementer_agent` | Dispatch an ImplementerAgent to implement a task document following TDD. | `task_content` | `model` |
| `spec_review` | Dispatch a SpecReviewerAgent to review a specification document. | `spec_content` | `model` |
| `quality_review` | Dispatch a CodeQualityReviewerAgent to review code or a diff. | `code_content` | `model` |

**`adversarial_review` and `gate_review` are hidden when `--no-adversarial` is set.**

---


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
