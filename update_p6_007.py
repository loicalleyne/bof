import re

with open('docs/tasks/P6-007-install-crush-sh.md', 'r') as f:
    content = f.read()

parts = content.split('## Session Notes')
session_notes = '## Session Notes' + parts[1] if len(parts) > 1 else ''

new_content = """# P6-007 — install_crush.sh: Install bof skills into Crush

**Phase:** P6 — Lab Skills + Install Infrastructure  
**Status:** Ready  
**Created:** 2026-04-16  
**Requires adversarial review:** minimum 3 rounds before implementation

---

## Goal

Create `scripts/install_crush.sh` — an idempotent shell script that copies
all bof skills into the Crush AI assistant's skill discovery directory
(`~/.config/crush/skills/`) and translates VS Code Copilot Chat tool names
into Crush-compatible tool names on the fly. This enables Crush to load and
execute bof workflow skills natively.

---

## Background

Crush discovers skills by walking directories listed in `Options.SkillsPaths`
(default: `${XDG_CONFIG_HOME:-$HOME/.config}/crush/skills/`). Crush only needs
`SKILL.md` files (skill name + description frontmatter + instructions body) —
it has no concept of `.agent.md` files or `.instructions.md` bootstrap injection.

Crucially, bof skills are written using VS Code Copilot Chat primitives
(`runSubagent`, `run_in_terminal`, `manage_todo_list`, `replace_string_in_file`,
`read_file`). Crush lacks these tools but possesses equivalents (`agent`, `bash`,
`todos`, `edit`/`multiedit`, `view`). To maintain bof's `AGENTS.md` invariant
(which mandates VS Code terminology as the source of truth), the installation
script must copy the skills and apply a cross-platform in-place translation 
(e.g., `sed`) to the `SKILL.md` files during installation.

---

## In Scope

- `scripts/install_crush.sh`:
  - **Script header** (first two lines after source guard): `#!/usr/bin/env bash` shebang, then `set -euo pipefail`. Script file must be executable (`chmod +x`).
  - **Source guard** (must appear before `set -euo pipefail` to prevent leaking set options into the parent shell when sourced): `[[ "${BASH_SOURCE[0]}" == "$0" ]] || { echo "ERROR: do not source this script — run it directly"; return 1; }`
  - **Argument parsing:** `--dry-run` is the only supported flag. Any unknown argument must print `USAGE: install_crush.sh [--dry-run]` and exit 1.
  - `--dry-run` flag: skip all actions (directory creation, copying, translation, and post-install validation). In dry-run mode: print `DRY-RUN: would copy and translate $src → $dst` for COPY cases; print `DRY-RUN: would update $dst` for UPDATE cases; print `DRY-RUN: would skip real file at $dst` for conflicts.
  - Resolve target dir using null-safe test `[ -n "${CRUSH_SKILLS_DIR:-}" ]` (required by `set -u` to avoid abort when `$CRUSH_SKILLS_DIR` is unset): if set and non-empty, strip trailing slashes; if result is empty after stripping (e.g. value was `/` or `///`), fall back to `/` if original value was root, otherwise XDG default with informational message; else use stripped value. If unset or empty, use `${XDG_CONFIG_HOME:-$HOME/.config}/crush/skills`.
  - Create target dir if absent (`mkdir -p`); on failure print `ERROR: could not create target directory $target — check permissions` and exit 1
  - **Environment guard:** `UNAME_O=$(uname -o 2>/dev/null || true)` (`|| true` makes the assignment always succeed so `set -e` does not abort on macOS where `uname -o` exits non-zero); use a `case` statement for portability across all bash versions:
    ```
    case "$UNAME_O" in
      [Mm][Ss][Yy][Ss]*|[Cc][Yy][Gg][Ww][Ii][Nn]*) echo "ERROR: ..."; exit 1 ;;
    esac
    ```
  - `shopt -s nullglob` before iterating skill dirs; print informational message if zero skills found. `shopt -u nullglob` must execute before any explicit early return or normal exit to restore default glob behavior.
  - Compute `REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)` once at startup. Extract skill name via `skillname=$(basename "$skilldir")` for each entry in the glob. `src="$REPO_ROOT/skills/$skillname"` — always an absolute path.
  - Iterate skill dirs with idempotency checks:
    - If `$dst` exists but is not a directory, print `WARN: real file at $dst — skipping`.
    - If `$dst` does not exist, copy directory `cp -R "$src" "$dst"`, print `COPIED:`, and apply translation.
    - If `$dst` exists as a directory, update the files `cp -R "$src/"* "$dst/"`, print `UPDATED:`, and apply translation.
  - **Translation Step:** After copying, run `sed` (or `perl`) safely for cross-platform in-place editing (e.g., `sed -i.bak ... && rm -f "$dst"/*.bak`) on `$dst/SKILL.md` to translate tool names:
    - `runSubagent` → `agent`
    - `run_in_terminal` → `bash`
    - `manage_todo_list` → `todos`
    - `read_file` → `view`
    - `replace_string_in_file` → `edit`
    - `grep_search` → `grep`
    - `list_dir` → `ls`
    - `vscode_askQuestions` → `agent` (or remove if inapplicable)
  - **Post-install validation** (skipped entirely in `--dry-run`; also skipped if zero skills were found via nullglob): for each skill name from the install loop, check `[ -d "$target/$skillname" ] && [ -f "$target/$skillname/SKILL.md" ]`; print `WARN: $target/$skillname does not contain SKILL.md` on failure.
  - Print `COPIED:`, `UPDATED:`, `EXISTS:`, `DRY-RUN:`, `WARN:` prefix per action; all path variables quoted in every command.
  - Exit 0 on success; exit 1 on unrecoverable error.
  - Final summary: print target dir; suggest `ls -la <target>` to verify.
- `docs/planning/ROADMAP.md`: add P6-007 row to P6 Tasks table.

---

## Out of Scope

- **Windows-native path resolution**
- **Installing agents** (`~/.config/crush/agents/` does not exist; Crush has no named subagent dispatch)
- **Installing instructions** (Crush has no `.instructions.md` bootstrap injection mechanism)
- **Validating SKILL.md frontmatter** (Crush's own `Discover()` already validates)
- **Uninstall logic** (covered by a future `uninstall_crush.sh` task)
- **macOS `~/Library/Application Support` path**
- **`--target` override flag**
- **hooks or instructions integration**

---

## Files

| Path | Action | What Changes |
|------|--------|-------------|
| `scripts/install_crush.sh` | Create | New idempotent skill installer for Crush |
| `docs/planning/ROADMAP.md` | Modify | Add P6-007 row to P6 Tasks table |

---

## Acceptance Criteria

1. `--dry-run` prints one `DRY-RUN:` line per skill dir; exits 0; no files/dirs created.
2. `bash scripts/install_crush.sh` creates the resolved target dir if absent, copies `skills/*/` into it, and executes the translation step on the copied `SKILL.md` files.
3. Running the script a second time (with no changes) updates the files, prints `UPDATED:`, and exits 0.
4. Validation failures → `WARN:` line; exit code unchanged (always 0 on success path).
5. Running the script from Git Bash or Cygwin exits 1 with a clear error message.
6. `ls -la "${XDG_CONFIG_HOME:-$HOME/.config}/crush/skills/"` shows copied directories for each bof skill.
7. Script is executable and starts with `#!/usr/bin/env bash` + `set -euo pipefail`.
8. Script produces no output about agents or instructions.
9. Translated `SKILL.md` files in the target directory contain Crush native tool names (`agent`, `bash`, `todos`, etc.) instead of VS Code primitives.

---

## Known Limitations

- Skills must be re-installed using this script any time the original `bof` skills are modified, as they are copied rather than symlinked.
- Cross-platform in-place editing requires safe `sed -i.bak` patterns to work across GNU (Linux/WSL) and BSD (macOS) `sed` implementations.

---

"""

# Add session note for iteration 14
new_session_note = """### Round 14 Adversarial Review (r0 / GPT-4.1 / iteration 13)

**Verdict:** CONDITIONAL — all valid issues addressed in post-review revision
**Issues accepted and resolved:**
1. **M1 (Context Blindness):** Plan updated to use "Build-on-install" Approach 1. The script now copies `SKILL.md` files and uses `sed` to translate VS Code Copilot Chat primitives into Crush native tools (`runSubagent` -> `agent`, etc.) ensuring skills are executable by Crush while preserving the `AGENTS.md` invariant.
2. **L1 (Trailing slash fallback):** Logic updated to handle root directory `/` cleanly.
3. **L2 (shopt -u nullglob contradiction):** Requirement softened to "execute before any explicit early return or normal exit" to account for implicit `set -e` exits.

"""

with open('docs/tasks/P6-007-install-crush-sh.md', 'w') as f:
    f.write(new_content + new_session_note + session_notes)

print("Plan successfully updated to use Approach 1 (Translation).")
