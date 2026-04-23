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

---

## Tool Reference


| Tool | Description | Required params |
|---|---|---|
| `adversarial_review` | Dispatch one or more adversarial review rounds for the given plan using a family-interleaved model pool (default: 5 models from copilot, gemini). Reads .adversarial/{plan_slug}.json for iteration state, runs the requested rounds, and writes the worst verdict back to the state file. Configure the pool with BOF_MODELS (comma-separated provider/model). Optional: pass rounds (default 5, max 50) and exclude_model to skip the caller's model. | `plan_slug`, `plan_content` |
| `gate_review` | Check whether all `.adversarial/` verdicts are PASSED or CONDITIONAL. | none (optional: `strict`) |
| `implementer_agent` | Dispatch an ImplementerAgent to implement a task document following TDD. | `task_content` |
| `spec_review` | Dispatch a SpecReviewerAgent to review a specification document. | `spec_content` |
| `quality_review` | Dispatch a CodeQualityReviewerAgent to review code or a diff. | `code_content` |

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
