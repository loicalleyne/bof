# P6-007 — install_crush.sh: Install bof skills into Crush

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
  - **Argument parsing:** `--dry-run` is the only supported flag. Any unknown argument must print `USAGE: install_crush.sh [--dry-run]` and exit 1. **The dry-run flag MUST be stored in a named variable (e.g. `DRY_RUN=false`/`true`) immediately after parsing, before the skill loop. All dry-run checks inside the loop MUST reference this variable, never `$1` or `$@` — the loop uses `set --` internally which clobbers positional parameters.**
  - `--dry-run` flag: skip all actions (directory creation, copying, translation, and post-install validation). In dry-run mode: print `DRY-RUN: would copy and translate $src → $dst` for COPY cases; print `DRY-RUN: would update $dst` for UPDATE cases; print `DRY-RUN: would skip real file at $dst` for conflicts.
  - Resolve target dir using null-safe test `[ -n "${CRUSH_SKILLS_DIR:-}" ]` (required by `set -u` to avoid abort when `$CRUSH_SKILLS_DIR` is unset): if set and non-empty, strip trailing slashes; if result is empty after stripping (e.g. value was `/` or `///`), fall back to `/` if original value was root, otherwise XDG default with informational message; else use stripped value. If unset or empty, use `${XDG_CONFIG_HOME:-$HOME/.config}/crush/skills`.
  - Create target dir if absent (`mkdir -p`); **skipped if `DRY_RUN=true`** (dry-run must not create directories); on failure print `ERROR: could not create target directory $target — check permissions` and exit 1
  - **Environment guard:** `UNAME_O=$(uname -o 2>/dev/null || true)` (`|| true` makes the assignment always succeed so `set -e` does not abort on macOS where `uname -o` exits non-zero); use a `case` statement for portability across all bash versions:
    ```
    case "$UNAME_O" in
      [Mm][Ss][Yy][Ss]*|[Cc][Yy][Gg][Ww][Ii][Nn]*) echo "ERROR: install_crush.sh does not support Git Bash or Cygwin — run from Linux, macOS, or WSL instead"; exit 1 ;;
    esac
    ```
  - `shopt -s nullglob` before iterating skill dirs; print informational message if zero skills found. `shopt -u nullglob` is called exactly once, after the skill loop ends (or on any explicit early return before the loop) — it is **not** called inside the loop body, so nullglob remains active throughout the entire loop including the `set -- "$src/"*` empty-dir guard.
  - Compute `REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)` once at startup. Extract skill name via `skillname=$(basename "$skilldir")` for each entry in the glob. `src="$REPO_ROOT/skills/$skillname"` — always an absolute path.
  - Iterate skill dirs with idempotency checks:
    - If `$dst` is a symlink (`[ -L "$dst" ]`), remove it (`rm -f "$dst"`). This cleans up legacy installations that symlinked to the git repository, preventing `sed` from mutating the source of truth.
    - If `$dst` exists but is not a directory (and not a symlink), print `WARN: real file at $dst — skipping`.
    - If `$dst` does not exist, copy directory `cp -R "$src" "$dst"`, print `COPIED:`, and apply translation.
    - If `$dst` exists as a directory, update the files: use `set -- "$src/"*` to collect files under nullglob; only run `cp -R "$@" "$dst/"` if `$# -gt 0`; print `UPDATED:`, and apply translation. (`basename "$skilldir"` handles trailing-slash paths correctly on both GNU and macOS; `basename "/path/brainstorming/"` returns `brainstorming`.)
  - **Translation Step:** After copying or updating, apply translation only if `[ -f "$dst/SKILL.md" ]`. Run `sed -i.bak` for cross-platform in-place editing on `$dst/SKILL.md`; remove the backup with `rm -f "$dst/SKILL.md.bak"` (explicit filename, not glob) after successful sed. **Substitutions must be applied longest-pattern-first to prevent partial matches** (e.g. `multi_replace_string_in_file` must be substituted before `replace_string_in_file`, otherwise the latter matches inside the former and produces `multi_edit` instead of `edit`). Translation table in required application order (VS Code → Crush native):
    1. `multi_replace_string_in_file` → `multiedit`
    2. `replace_string_in_file` → `edit`
    3. `runSubagent` → `agent`
    4. `run_in_terminal` → `bash`
    5. `manage_todo_list` → `todos`
    6. `read_file` → `view`
    7. `grep_search` → `grep`
    8. `list_dir` → `ls`
    9. `file_search` → `grep` (best available Crush equivalent)
    10. `vscode_askQuestions` → *(no equivalent — leave as-is; no sed expression)*
    - Untranslated (leave as-is; no sed expression): `semantic_search`, `get_errors`, `create_file`, `view_image`, `vscode_listCodeUsages`
  - **Post-install validation** (skipped entirely in `--dry-run`; also skipped if zero skills were found via nullglob): for each skill name from the install loop, if `[ -f "$src/SKILL.md" ]` (source had a SKILL.md), then check `[ -f "$target/$skillname/SKILL.md" ]`; print `WARN: $target/$skillname does not contain SKILL.md` on failure. Does not iterate pre-existing non-bof entries in the target dir.
  - Print `COPIED:`, `UPDATED:`, `DRY-RUN:`, `WARN:` prefix per action; all path variables quoted in every command.
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
3. Running the script a second time always prints `UPDATED:` for every entry that was previously installed as a directory, and exits 0 — no content diff is performed; files are always overwritten and re-translated on subsequent runs.
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
- `sed` substitutions use unanchored patterns (e.g. `s/read_file/view/g`). A tool name that appears as a prefix in a longer identifier (e.g. `read_file_path`) would be incorrectly translated to `view_path`. Adding word-boundary anchors is not cross-platform (`\b` = GNU, `[[:<:]]` = BSD), so this is accepted as-is; bof skill SKILL.md files do not currently contain such identifiers.
- Several VS Code Copilot Chat tool names used in bof skills have no direct Crush equivalent (`semantic_search`, `get_errors`, `create_file`, `view_image`, `vscode_listCodeUsages`). These are left untranslated; Crush will encounter them in skill prose but will not error — it simply won't recognise them as tool calls.
- `vscode_askQuestions` has no safe Crush equivalent (Crush's `agent` dispatches sub-agents, not user prompts). Occurrences in skill text are left as-is; no sed expression is applied.

---

### Round 19 Adversarial Review (r1 / Auto / iteration 19)

**Verdict:** CONDITIONAL — one major and two minors resolved post-review
1. **M1 (`mkdir -p` in dry-run violates AC1):** `mkdir -p` now explicitly skipped when `DRY_RUN=true`
2. **L1 (AC3 wording):** AC3 now says "every entry that was previously installed as a directory" (conflict case prints WARN:, not UPDATED:)
3. **L2 (`shopt -u nullglob` timing):** Clarified as called exactly once after the loop ends; NOT inside the loop body (nullglob must stay active for the `set -- "$src/"*` guard)

### Round 20 Adversarial Review (r2 / GPT-4o / iteration 20)

**Verdict:** PASSED — no critical or major issues identified. Plan cleared for implementation.

### Round 18 Adversarial Review (r0 / GPT-4.1 / iteration 18)

**Verdict:** PASSED — one minor addressed post-review
1. **L1 (error message underspecified):** Environment guard `case` now specifies exact error message: `"ERROR: install_crush.sh does not support Git Bash or Cygwin — run from Linux, macOS, or WSL instead"`

### Round 17 Adversarial Review (r1 / Auto / iteration 16)

**Verdict:** CONDITIONAL — two majors and two minors resolved post-review
1. **M1 (`set --` clobbers `$1`):** Added explicit requirement: `--dry-run` flag MUST be stored in a named variable (e.g. `DRY_RUN`) before the skill loop; loop body MUST check this variable, never `$1`/`$@`
2. **M2 (`vscode_askQuestions` three incompatible descriptions):** Canonicalized to "leave as-is; no sed expression" in both In Scope table and Known Limitations
3. **L1 (sed word-boundary anchors):** Added to Known Limitations as accepted limitation; word-boundary anchors are not cross-platform; bof skills do not currently have compound identifiers
4. **L2 (`EXISTS:` leftover from symlink design):** Removed from output-prefix list

### Round 16 Adversarial Review (r0 / GPT-4.1 / iteration 15)

**Verdict:** CONDITIONAL — five major issues resolved post-review
1. **M1 (sed order):** Translation table now numbered with `multi_replace_string_in_file` first (step 1) before `replace_string_in_file` (step 2) — prevents substring match corruption
2. **M2 (post-install validation source check):** Validation now checks `[ -f "$src/SKILL.md" ]` first; only warns if source had a SKILL.md but destination doesn't
3. **M3 (`basename` trailing slash):** Explicit note added: `basename "/path/brainstorming/"` returns `brainstorming` on both GNU and macOS
4. **M4 (cp guard mechanism):** Specified as `set -- "$src/"*; [ $# -gt 0 ] && cp -R "$@" "$dst/"` — explicit positional arg count check
5. **M5 (AC3 UPDATED: semantics):** AC3 now explicitly states files are always overwritten on re-run; no diff check is performed

### Round 15 Adversarial Review (r2 / GPT-4o / iteration 14)

**Verdict:** CONDITIONAL — five valid issues resolved post-review
1. **C1 (empty-src `cp` guard):** UPDATED: path now guards against empty source dir with nullglob — only calls `cp -R "$src/"*` if glob expands to at least one item
2. **C2 (`vscode_askQuestions → agent` wrong):** Removed the mapping; `vscode_askQuestions` has no safe Crush equivalent — left untranslated or removed, documented in Known Limitations
3. **C3 (`.bak` glob unsafe):** Changed `rm -f "$dst"/*.bak` to `rm -f "$dst/SKILL.md.bak"` — explicit filename, not a glob, so nullglob state is irrelevant
4. **C4 (SKILL.md existence before sed):** Translation step now guarded by `[ -f "$dst/SKILL.md" ]`
5. **C5 (untranslated tools):** Added `multi_replace_string_in_file → edit`, `file_search → grep`; remaining untranslated tools (`semantic_search`, `get_errors`, `create_file`, `view_image`, `vscode_listCodeUsages`) documented in Known Limitations
**Dismissed:**
- Translation idempotency concern: substitutions are one-directional (long VS Code names → short Crush names); re-running sed on already-translated content cannot produce false matches

### Round 14 Adversarial Review (r0 / GPT-4.1 / iteration 13)

**Verdict:** CONDITIONAL — all valid issues addressed in post-review revision
**Issues accepted and resolved:**
1. **M1 (Context Blindness):** Plan updated to use "Build-on-install" Approach 1. The script now copies `SKILL.md` files and uses `sed` to translate VS Code Copilot Chat primitives into Crush native tools (`runSubagent` -> `agent`, etc.) ensuring skills are executable by Crush while preserving the `AGENTS.md` invariant.
2. **L1 (Trailing slash fallback):** Logic updated to handle root directory `/` cleanly.
3. **L2 (shopt -u nullglob contradiction):** Requirement softened to "execute before any explicit early return or normal exit" to account for implicit `set -e` exits.

## Session Notes

### Round 21 Adversarial Review (r0 / GPT-4.1 / iteration 21)

**Verdict:** PASSED — all critical issues resolved post-review
1. **C1 (Symlink mutation):** Pre-copy checks now identify and remove legacy symlinks (`[ -L "$dst" ] && rm -f "$dst"`) to prevent in-place updates from mutating the git source tree.
2. **C2 (multiedit mapping):** `multi_replace_string_in_file` now maps correctly to Crush's native `multiedit` tool, not `edit`.



### Round 13 Adversarial Review (r0 / GPT-4.1 / iteration 12)

**Verdict:** PASSED — zero issues found across all 7 attacks
- All 12 rounds of hardening held. Plan confirmed implementation-ready.

### Round 12 Adversarial Review (r2 / GPT-4o / iteration 11)

**Verdict:** CONDITIONAL — four issues resolved, one dismissed post-review
1. **C1 (shebang missing from In Scope):** Added explicit "Script header" bullet: `#!/usr/bin/env bash`, `set -euo pipefail`, file must be executable
2. **C2 (`shopt -u` unreachable on `set -e`):** **Dismissed** — the script runs in its own subprocess; when `set -e` exits the script, the entire subprocess terminates and all shell settings (including `nullglob`) die with it. The source guard prevents sourcing, so `shopt` settings can never leak into the parent shell.
3. **C3 (`${UNAME_O,,}` bash 3.2 incompatible):** Replaced with a `case` statement matching `[Mm][Ss][Yy][Ss]*|[Cc][Yy][Gg][Ww][Ii][Nn]*` — portable across all bash versions including macOS default bash 3.2
4. **C4 (`readlink` comparison requires absolute `$src`):** Added note that `ln -s` always uses absolute `$src`; relative symlinks would never match conditions 3/4 and would be re-linked on every run
5. **M1 (`|| true` in UNAME_O):** Added inline explanation to env guard spec

### Round 11 Adversarial Review (r1 / Auto / iteration 10)

**Verdict:** PASSED — four minors addressed post-review
1. **L1 (skillname extraction):** Plan now specifies `skillname=$(basename "$skilldir")` — prevents empty string from `${d##*/}` on trailing-slash glob paths
2. **L2 (source guard ordering):** Source guard now explicitly described as "must appear before `set -euo pipefail`" to prevent leaking set options into parent shell
3. **L3 (`uname -o` case sensitivity):** Environment guard now specifies `[[ "${UNAME_O,,}" == *msys* ]]` case-insensitive glob match to handle MSYS/msys/MSYS2 variants
4. **L4 (unknown flag behavior):** Plan now requires unknown arguments to print `USAGE: install_crush.sh [--dry-run]` and exit 1

### Round 10 Adversarial Review (r0 / GPT-4.1 / iteration 9)

**Verdict:** PASSED — two minors addressed post-review
1. **L1 (`$entry` ambiguity):** Validation now explicitly uses `$target/$skillname` path
2. **L2 (`shopt -u` on all exit paths):** Spec now requires `shopt -u nullglob` to run on all paths (early return, end of main loop, any `exit` call)

### Round 9 Adversarial Review (r2 / GPT-4o / iteration 8)

**Verdict:** FAILED — all issues resolved in post-review revision
1. **C1 (`rm -f` race with `set -e`):** Existing TOCTOU Known Limitation now explicitly covers this: `rm -f` exits non-zero on a directory (without `-r`); `set -e` exits the script — correct behavior
2. **C2 (`shopt -s nullglob` scope):** `shopt -u nullglob` added after main loop to restore default glob behavior
3. **M1 (condition 1 `ln -sf` inconsistency):** Changed `ln -sf` to `ln -s` in condition 1 for consistency
4. **M2 (`REPO_ROOT` with script-as-symlink):** Changed to `$(dirname "${BASH_SOURCE[0]}")` + `pwd -P`; file-level symlink documented as unsupported

### Planning Assumptions

- **ASSUMPTION:** `${XDG_CONFIG_HOME:-$HOME/.config}` is correct fallback for Linux/macOS/WSL. When `$CRUSH_SKILLS_DIR` is set, the script installs there instead, matching `GlobalSkillsDirs()` priority. Trailing slashes are stripped; empty string after stripping falls back to default.
- **ASSUMPTION:** Standard `ln -sf` from WSL pointing to `/mnt/c/...` paths may NOT be reliably followed by Crush's fastwalk across the 9P filesystem boundary. The script warns the user in this case but does not attempt Windows-native links — that is explicitly out of scope.
- **ASSUMPTION:** All bof skill directory names already match their `name:` frontmatter field.
- **ASSUMPTION:** Skills are the only bof artifact relevant to Crush.
- **ASSUMPTION:** `GlobalSkillsDirs()` is the actual function at `internal/config/load.go:849` — verified from Crush source.
- **ASSUMPTION:** `fastwalk` with `Follow: true` is configured in `internal/skills/skills.go:133-137` — verified from Crush source.
- **Manual test (not automatable):** Re-running after bof repo move/deletion produces `RELINKED:` for all entries (not `EXISTS:`), confirming broken/stale symlinks are repaired.

### Round 1 Adversarial Review (r0 / GPT-4.1 / iteration 0)

**Verdict:** CONDITIONAL  
**Issues addressed in this revision:**
1. Added WSL `/mnt/` cross-filesystem warning
2. Added post-install symlink validation with `WARN:` output
3. Added environment guard for Git Bash/Cygwin on native Windows

### Round 3 Adversarial Review (r2 / GPT-4o / iteration 2)

**Verdict:** FAILED  
**Issues addressed in this revision:**
1. **C1 (TOCTOU):** Documented as known limitation; explicit three-way branch for exists-as-symlink vs exists-as-real-dir vs does-not-exist
2. **C2 (unenforceable guard):** Specified `uname -o` for `Msys`/`Cygwin` detection — portable and enforceable on both Git Bash and Cygwin
3. **M1 ($CRUSH_SKILLS_DIR special chars):** All path variables quoted; trailing slashes stripped; empty-after-strip falls back to default
4. **M2 (dangling symlink validation):** Post-install check now uses `[ -d "$dst" ] && [ -f "$dst/SKILL.md" ]` — `[ -d ]` fails for dangling symlinks
5. **L1 (zero-entry glob):** `shopt -s nullglob` added before skill dir iteration; informational message if zero skills found
6. **L2 (trailing slash):** Explicit strip applied to `$CRUSH_SKILLS_DIR` before use

### Round 8 Adversarial Review (r1 / Auto / iteration 7)

**Verdict:** FAILED — all issues resolved in post-review revision  
**Issues addressed:**
1. **C1 (`ln -sf` macOS/BSD follow bug):** Conditions 2 and 3 now specify `rm -f "$dst" && ln -s "$src" "$dst"` with an explicit macOS/BSD note; `ln -sf` alone is insufficient because BSD `ln` follows valid directory symlinks
2. **M1 (`set -u` + unset `$CRUSH_SKILLS_DIR`):** Resolve target dir now uses null-safe `[ -n "${CRUSH_SKILLS_DIR:-}" ]` idiom throughout
3. **L1 (WARN message):** Changed from "real directory" to "real file or directory"
4. **L2 (AC1 precision):** AC1 now specifies both DRY-RUN output forms
5. **L3 (`readlink` syntax):** Conditions now use `[ "$(readlink "$dst")" != "$src" ]` command-substitution syntax
6. **L4 (source guard):** Source guard added: `[[ "${BASH_SOURCE[0]}" == "$0" ]] || { echo "ERROR: ..."; return 1; }`

### Round 7 Adversarial Review (r0 / GPT-4.1 / iteration 6)

**Verdict:** CONDITIONAL — three minor gaps resolved in post-review revision  
1. **M1 (`$src` relative path):** Plan now specifies `REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)` computed once at startup; `$src` always absolute, making `readlink` comparisons unambiguous
2. **M2 (`///` → empty fallback):** Documented explicitly — stripping `///` yields empty string; script falls back to XDG default with informational message
3. **M3 (`mkdir -p` error message):** Plan now requires `ERROR: could not create target directory $target — check permissions` before exit 1

### Round 6 Adversarial Review (r2 / GPT-4o / iteration 5)

**Verdict:** CONDITIONAL — all valid issues addressed in post-review revision  
**Issues accepted and resolved:**
1. **C2 (dry-run + WARN: ambiguity):** Clarified: in dry-run mode real-dir WARN: is replaced with `DRY-RUN: would skip real directory at $dst`; no `WARN:` is printed during dry-run
2. **C3 (condition-1 wording):** Made explicit: condition 1 is `! [ -e "$dst" ] && ! [ -L "$dst" ]`; all five conditions now stated with explicit test expressions
3. **C4 (nullglob + validation coupling):** Stated explicitly: post-install validation also skipped if zero skills found
4. **C5 (AC3b untestable):** Reframed as manual test note; removed from formal AC list
**Issues rejected as invalid:**
- **C1 (uname -o fallback):** Already addressed since round 3 via `|| true` guard; reviewer did not read plan
- **Attack 3 (symlink attack):** Out of scope; single-user developer tool, not setuid; TOCTOU acceptance unchanged

### Round 5 Adversarial Review (r1 / Auto / iteration 4)

**Verdict:** FAILED  
**Issues addressed in this revision:**
1. **C1 (`uname -o` macOS crash):** Guarded with `UNAME_O=$(uname -o 2>/dev/null || true)` to prevent `set -euo pipefail` abort on macOS where `uname -o` is not supported
2. **M1 (broken symlinks silently skipped):** Added explicit branch for `[ -L ] && [ ! -e ]` (broken symlink) → `RELINKED:`
3. **M2 (stale symlinks silently skipped):** Added `readlink` comparison; stale symlinks re-linked → `RELINKED:`
4. **M3 (post-install validation scope too broad):** Validation now iterates only bof skill names, not all entries in target dir
5. **L1/L2 (reviewer claims):** `GlobalSkillsDirs()` and `fastwalk Follow: true` are both verified from Crush source code; these findings were incorrect

### Round 4 Adversarial Review (r0 / GPT-4.1 / iteration 3)

**Verdict:** CONDITIONAL (r0 reviewer incorrectly demanded implementation code; invalid for plan review)
**Self-identified issues addressed in this revision:**
1. `--dry-run` scope made explicit: all three idempotency branches AND post-install validation are skipped in dry-run mode
2. Exit code when `WARN:` conditions exist is now specified: always exits 0
3. AC2 variable name corrected to "resolved target dir" for consistency with In Scope

### Adversarial Review Requirement

Minimum 3 rounds completed and cleared. Current state: 4 rounds complete (r0, r1, r2, r0). Proceeding with requested rounds 5 and 6 per user instruction.

State file: `.adversarial/P6-007-install-crush-sh.json`
