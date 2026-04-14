# ADR-0003: AST Cache Copy-at-Creation + Incremental Update Strategy for Git Worktrees

**Date:** 2026-04-13
**Status:** Accepted

---

## Context

bof's `using-git-worktrees` skill creates isolated feature branches in `.worktrees/`.
Projects using the `duckdb-code` skill maintain an AST cache (`code_ast.duckdb`) at
the project root (gitignored). The question is: how should a new worktree gain access
to the AST cache, and how should it update the cache after implementation tasks are done?

### Key facts about `code_ast.duckdb`:
- Built by the `sitting_duck` DuckDB extension via `read_ast()`.
- `read_ast()` stores **relative** file paths (e.g., `internal/cache/store.go`).
- Relative paths remain valid from any directory rooted at the project root.
- The file is gitignored — not shared via branches.
- A git worktree is a separate checkout under `.worktrees/feature-name/`.
- The main checkout's cache is valid from the worktree root because paths are relative.

### Constraints:
- VS Code Copilot Chat has no PostToolUse hook that is cheap enough for per-tool-call cache updates.
- `sitting_duck`'s `--incremental` flag (via `scripts/rebuild-ast.sh --incremental`) re-parses only files changed since the last rebuild. Requires a pre-existing database.
- A full `scripts/rebuild-ast.sh` (no `--incremental`) builds from scratch; requires no prior database.

---

## Decision

**Three-phase AST cache strategy for worktrees:**

### Phase 1: Worktree creation (copy-at-creation)
When `using-git-worktrees` creates a new worktree:
1. If `code_ast.duckdb` exists at the main repo root (`git rev-parse --show-toplevel`), **copy** it to the worktree root:
   ```sh
   cp "$(git rev-parse --show-toplevel)/code_ast.duckdb" ./code_ast.duckdb
   ```
   This is fast (file copy), immediately valid (relative paths), and represents the codebase at branch point.
2. If no cache exists and `scripts/rebuild-ast.sh` is present (Esquisse project), build it from scratch:
   ```sh
   bash scripts/rebuild-ast.sh
   ```
   (Full build, no `--incremental` — there is no prior database to diff against.)
3. If neither condition is met, skip. The agent uses file reads as fallback.

### Phase 2: During implementation
- No automatic cache updates. The cache may become stale as files are modified.
- Agents are instructed to use the potentially stale cache for structure/navigation, and `read_file` for the specific functions being changed.

### Phase 3: Task completion (incremental update)
When `ImplementerAgent` completes a task and tests pass, before reporting status:
- If `scripts/rebuild-ast.sh` exists: `bash scripts/rebuild-ast.sh --incremental`
  (re-parses only files modified since the last build — git-aware, fast)
- If no `rebuild-ast.sh` (non-Esquisse project): inline fallback:
  ```sh
  duckdb code_ast.duckdb "LOAD sitting_duck;
  DELETE FROM ast WHERE file_path IN ('file1', 'file2');
  INSERT INTO ast SELECT * FROM read_ast(['file1', 'file2'], ignore_errors:=true, peek:=200);"
  ```
- Skip entirely if `code_ast.duckdb` does not exist in the worktree.

---

## Alternatives Considered

| Alternative | Reason Rejected |
|---|---|
| **PostToolUse hook for per-tool cache updates** | Fires on every tool call (read_file, grep, etc.); adds latency to the entire implementation loop; makes the hook essentially a background daemon |
| **Full rebuild at each task completion** | Unnecessary given the 10-file task limit; `--incremental` is 3-5× faster |
| **ATTACH-only on main cache (no copy)** | `ATTACH`ing the main's `.duckdb` from the worktree requires the main path to be encoded. More importantly: cross-worktree writes corrupt both databases. Read-only ATTACH is possible but then the worktree agent cannot update the cache. |
| **No AST in worktrees** | Loses the structural navigation benefit for worktree-based development, which is exactly when large codebases need it most |
| **Inline SQL over script** | `rebuild-ast.sh` handles DuckDB binary discovery (`~/.duckdb/cli/latest/duckdb`), gitignore exclusions, and changed-file detection via git automatically. Inline SQL requires the agent to do all of this manually. |

---

## Consequences

**Positive:**
- Cache is valid immediately on worktree creation (copy is fast, relative paths work).
- No hook overhead during implementation loop.
- Incremental updates are fast (git-aware, changed-files-only).
- Non-Esquisse projects have an inline SQL fallback.

**Negative:**
- Cache becomes stale during the implementation loop (between Phase 1 copy and Phase 3 update).
- Agents must use file reads for functions being actively modified during Phase 2.
- Cross-worktree structural diff (comparing two branches' AST) requires `ATTACH` + FULL OUTER JOIN or `structural_diff` macro from `duck_tails`.

**Schema:** `read_ast()` schema unchanged. No new columns or tables required.

---

## References

- sitting_duck docs: https://sitting-duck.readthedocs.io/en/latest/
- `scripts/rebuild-ast.sh --incremental` flag: Esquisse scripts
- bof:using-git-worktrees skill: `skills/using-git-worktrees/SKILL.md`
- bof ImplementerAgent: `agents/ImplementerAgent.agent.md`
