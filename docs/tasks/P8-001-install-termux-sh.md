# P8-001 — install_termux.sh: Install bof for Crush on Termux/Android

**Phase:** P8 — Termux / Android Support
**Status:** Ready
**Created:** 2026-04-23
**Requires adversarial review:** minimum 1 round before implementation

---

## Goal

Create `scripts/install_termux.sh` — an idempotent install script that copies bof
skills into Crush's skills directory, then uses `jq` to merge `options.skills_paths`
(and optionally a `mcpServers.bof-mcp` entry) into `crush.json` without disturbing
any existing settings. Designed for Termux on Android but works on any Linux/macOS
with Crush installed.

---

## Background

On Termux, VS Code is unavailable, so `runSubagent` and `manage_todo_list` don't
exist. Crush with bof-mcp provides equivalent agent dispatch. The install script must:

1. **Auto-detect the Crush config directory** — `crush dirs config` (if crush is
   installed), with `CRUSH_GLOBAL_CONFIG` env override, falling back to
   `${XDG_CONFIG_HOME:-$HOME/.config}/crush`.
2. **Auto-detect or default the skills target directory** — `$CRUSH_CONFIG_DIR/skills/`
   (or `$CRUSH_SKILLS_DIR` env override).
3. **Copy and translate bof skills** — same sed translation table as `install_crush.sh`
   (VS Code → Crush tool names), applied longest-pattern-first.
4. **Merge crush.json** — use `jq` to add the skills path to `options.skills_paths`
   (deduplicated) and optionally add/update `mcpServers.bof-mcp`. Atomic write via
   temp file + `mv`.
5. **Optionally build and register bof-mcp** — if `go` is available and
   `bof-mcp/main.go` exists, offer to build and register in crush.json.

---

## In Scope

- `scripts/install_termux.sh`
  - All specification below
  - Executable (`chmod +x`)
- `docs/planning/ROADMAP.md`: P8 section + P8-001 row added

---

## Out of Scope

- Modifying `scripts/install.sh` (VS Code/WSL installer, separate concern)
- Modifying `scripts/install_crush.sh` (generic Crush installer, separate concern)
- Automating Crush installation itself (user's responsibility)
- Cross-compiling bof-mcp for non-native architectures
- Modifying any source `skills/*/SKILL.md` files

---

## Files

| Path | Action | What changes |
|------|--------|-------------|
| `scripts/install_termux.sh` | Create | New idempotent install script |
| `docs/planning/ROADMAP.md` | Modify | Add P8 section and P8-001 row |

---

## Specification

### Script Header and Guards

```bash
#!/usr/bin/env bash
# install_termux.sh — Install bof skills + crush.json config for Crush on Termux/Android
# Idempotent: safe to run multiple times.
#
# Usage:
#   bash scripts/install_termux.sh [--dry-run] [--no-mcp]
#
# Environment overrides:
#   CRUSH_GLOBAL_CONFIG   Override Crush config directory (same as Crush env var)
#   CRUSH_SKILLS_DIR      Override skills install target (default: $crush_config/skills)
#   BOF_MCP_BIN           Path to pre-built bof-mcp binary (skips build step)
#
# Dependencies:
#   required: jq
#   optional: crush (for crush dirs config discovery), go (to build bof-mcp)
```

Source guard (before `set -euo pipefail`):
```bash
[[ "${BASH_SOURCE[0]}" == "$0" ]] || { echo "ERROR: do not source this script — run it directly"; return 1; }
```

Then `set -euo pipefail`.

### Argument Parsing

Parse `--dry-run` and `--no-mcp` flags into `DRY_RUN=false` and `NO_MCP=false`
variables immediately after the source guard and `set`. Unknown arguments print
`USAGE: install_termux.sh [--dry-run] [--no-mcp]` and exit 1.

### Termux Detection

```bash
UNAME_O=$(uname -o 2>/dev/null || true)
IS_TERMUX=false
if [ -n "${TERMUX_VERSION:-}" ] || [[ "$UNAME_O" == "Android" ]]; then
  IS_TERMUX=true
fi
```

Print `INFO: Termux detected (Android)` or `INFO: Non-Termux Linux/macOS — script will
still work`. Do **not** abort if not Termux.

### Dependency Check

Check for `jq` before any mutations: `command -v jq >/dev/null 2>&1 || { echo "ERROR: jq is required — install with: pkg install jq"; exit 1; }`.

Check for `crush`: if absent, print `WARN: crush not found — using XDG fallback for
config dir discovery` and continue.

### Config Directory Resolution

```bash
resolve_crush_config_dir() {
  # 1. CRUSH_GLOBAL_CONFIG env (same override that Crush itself respects)
  if [ -n "${CRUSH_GLOBAL_CONFIG:-}" ]; then
    CRUSH_CONFIG_DIR="${CRUSH_GLOBAL_CONFIG%/}"
    return
  fi
  # 2. crush dirs config (respects CRUSH_GLOBAL_CONFIG transparently)
  if command -v crush >/dev/null 2>&1; then
    local discovered
    discovered=$(crush dirs config 2>/dev/null | tr -d '\r\n')
    # Reject non-absolute-path output (e.g. usage text written to stdout)
    [[ "$discovered" = /* ]] || discovered=""
    if [ -n "$discovered" ]; then
      CRUSH_CONFIG_DIR="${discovered%/}"
      return
    fi
  fi
  # 3. XDG fallback — guard against unset HOME
  if [ -z "${HOME:-}" ]; then
    echo "ERROR: \$HOME is not set and \$CRUSH_GLOBAL_CONFIG is unset — cannot determine config directory"
    exit 1
  fi
  CRUSH_CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/crush"
}
```

`CRUSH_SKILLS_DIR` default assignment (immediately after calling `resolve_crush_config_dir`):
```bash
CRUSH_SKILLS_DIR="${CRUSH_SKILLS_DIR:-$CRUSH_CONFIG_DIR/skills}"
```
Strip trailing slashes: `CRUSH_SKILLS_DIR="${CRUSH_SKILLS_DIR%/}"`.

Print:
```
Crush config dir: $CRUSH_CONFIG_DIR
Skills target:    $CRUSH_SKILLS_DIR
```

### Skills Installation

`CRUSH_CONFIG_JSON="$CRUSH_CONFIG_DIR/crush.json"`

**Symlink resolution** — wrap in a named function to allow `local`; call it after setting `CRUSH_CONFIG_JSON`:
```bash
resolve_crush_config_json() {
  if [ -L "$CRUSH_CONFIG_JSON" ]; then
    local resolved
    if readlink -f "$CRUSH_CONFIG_JSON" >/dev/null 2>&1; then
      resolved=$(readlink -f "$CRUSH_CONFIG_JSON")
    elif command -v realpath >/dev/null 2>&1; then
      resolved=$(realpath "$CRUSH_CONFIG_JSON")
    else
      resolved=$(python3 -c "import os,sys; print(os.path.realpath(sys.argv[1]))" "$CRUSH_CONFIG_JSON" 2>/dev/null || echo "$CRUSH_CONFIG_JSON")
    fi
    CRUSH_CONFIG_JSON="$resolved"
    # Update CRUSH_CONFIG_DIR so mktemp stays on the same filesystem
    CRUSH_CONFIG_DIR="$(dirname "$CRUSH_CONFIG_JSON")"
  fi
}
resolve_crush_config_json
```
This chain handles: GNU `readlink -f` (Linux/Termux), `realpath` (GNU coreutils), macOS BSD (no `-f`), and Python fallback.

Create `$CRUSH_CONFIG_DIR` and `$CRUSH_SKILLS_DIR` if absent (`mkdir -p`); skip if
`DRY_RUN=true` — print `DRY-RUN: would mkdir -p $CRUSH_SKILLS_DIR`.

Verify `$REPO_ROOT/skills/` exists and is a non-empty directory before the loop:
- If absent or not a directory: print `ERROR: skills directory not found: $REPO_ROOT/skills/` and exit 1.
- If the glob expands to zero entries (empty `skills/` dir): print `INFO: no skill directories found in $REPO_ROOT/skills/ — nothing to install` and exit 0.

`shopt -s nullglob`. Declare `installed_skills=()` before the glob. Iterate `"$REPO_ROOT/skills/"*/`:

For each `skilldir`:
- `skillname=$(basename "$skilldir")`
- `src="$REPO_ROOT/skills/$skillname"`
- `dst="$CRUSH_SKILLS_DIR/$skillname"`

Idempotency checks (in order):
1. `[ -L "$dst" ]` — symlink (legacy): dry-run prints `DRY-RUN: would remove legacy symlink at $dst and copy $src → $dst`; live: `rm -f "$dst"`, then copy+translate+accumulate.
2. `[ -e "$dst" ] && [ ! -d "$dst" ]` — non-directory file conflict: print `WARN: real file at $dst — skipping`; `continue` without accumulating.
3. `[ ! -e "$dst" ]` — new install: dry-run prints `DRY-RUN: would copy and translate $src → $dst`; live: `cp -R "$src" "$dst"`, translate, accumulate.
4. `[ -d "$dst" ]` — update: dry-run check **first** (`DRY-RUN: would update $dst`); live: `rm -rf "$dst"` then `cp -R "$src" "$dst"` (exit 1 on cp failure), translate, accumulate.

`shopt -u nullglob` once after the loop.

### Translation Step

Apply after `cp`, not in dry-run. If `[ -f "$dst/SKILL.md" ]`, apply in-place substitutions with a
single `sed -i.bak` invocation (GNU sed on Linux/Termux; macOS uses the same flag), then
`rm -f "$dst/SKILL.md.bak"`.

**Source:** `docs/specs/tool-mapping.md` (VS Code → Crush column). There is no `install_crush.sh` in this repo — the table is defined in the tool-mapping spec and must be inlined here.

Apply longest-pattern-first to prevent partial matches. The 9 substitutions (longest first):

```bash
sed -i.bak \
  -e 's/multi_replace_string_in_file/multiedit/g' \
  -e 's/replace_string_in_file/edit/g' \
  -e 's/run_in_terminal/bash/g' \
  -e 's/manage_todo_list/todos/g' \
  -e 's/grep_search/grep/g' \
  -e 's/file_search/glob/g' \
  -e 's/create_file/write/g' \
  -e 's/runSubagent/agent/g' \
  -e 's/read_file/view/g' \
  -e 's/list_dir/ls/g' \
  "$dst/SKILL.md"
rm -f "$dst/SKILL.md.bak"
```

The ordering guarantees `multi_replace_string_in_file` is matched before `replace_string_in_file` (which is a substring of it), preventing a double-substitution bug.

### crush.json Merge (jq)

Skipped entirely in `--dry-run`; print `DRY-RUN: would merge skills path and mcpServers
into $CRUSH_CONFIG_JSON`.

If `crush.json` does not exist, create it with `echo '{"$schema":"https://charm.land/crush.json"}' > "$CRUSH_CONFIG_JSON"` before the jq step.

**Backup before merge:** `cp "$CRUSH_CONFIG_JSON" "$CRUSH_CONFIG_JSON.bak"` before any write.

Atomic merge:
```bash
tmp=$(mktemp "${CRUSH_CONFIG_DIR}/.crush_json_XXXXXX")
jq --arg skillsPath "$CRUSH_SKILLS_DIR" \
   '.options //= {} |
    .options.skills_paths = ((.options.skills_paths // []) + [$skillsPath] | unique)' \
   "$CRUSH_CONFIG_JSON" > "$tmp" \
  && mv "$tmp" "$CRUSH_CONFIG_JSON" \
  || { rm -f "$tmp"; mv "$CRUSH_CONFIG_JSON.bak" "$CRUSH_CONFIG_JSON"; echo "ERROR: jq merge failed — crush.json restored"; exit 1; }
rm -f "$CRUSH_CONFIG_JSON.bak"
echo "MERGED:  skills_paths += $CRUSH_SKILLS_DIR"
```

The `mktemp` temp file is created inside `$CRUSH_CONFIG_DIR` (the resolved real path after symlink expansion) to ensure `mv` is atomic (same filesystem). The `||` guard restores from backup if jq or mv fails.

### bof-mcp Build and Registration

Skip if `--no-mcp` is set.

Check for `BOF_MCP_BIN` env override — if set and the file is executable, skip build
and go directly to registration.

If `go` is available and `"$REPO_ROOT/bof-mcp/main.go"` exists:
- Default install path: `${BOF_MCP_INSTALL:-$HOME/.local/bin/bof-mcp}` (on Termux,
  `$PREFIX/bin/bof-mcp` if `$PREFIX` is set; prefer `$PREFIX/bin` since it's on PATH)
- Termux install path logic:
  ```bash
  if $IS_TERMUX && [ -n "${PREFIX:-}" ]; then
    BOF_MCP_INSTALL="${BOF_MCP_INSTALL:-$PREFIX/bin/bof-mcp}"
  else
    BOF_MCP_INSTALL="${BOF_MCP_INSTALL:-$HOME/.local/bin/bof-mcp}"
  fi
  ```
- Build: `(cd "$REPO_ROOT/bof-mcp" && go build -o "$BOF_MCP_INSTALL" .)`. Print
  `BUILT:   bof-mcp → $BOF_MCP_INSTALL`.
- On build failure: print `WARN: bof-mcp build failed — skipping MCP registration`
  and skip registration (do not exit 1; bof still works without bof-mcp).

If `BOF_MCP_BIN` is set OR build succeeded, register in crush.json (atomic merge):
```bash
tmp=$(mktemp "${CRUSH_CONFIG_DIR}/.crush_json_XXXXXX")
jq --arg cmd "$BOF_MCP_BIN_OR_BUILT" \
   --arg root "$REPO_ROOT" \
   '.mcpServers //= {} |
    .mcpServers["bof-mcp"] = {"command": $cmd, "args": ["--project-root", $root]}' \
   "$CRUSH_CONFIG_JSON" > "$tmp" \
  && mv "$tmp" "$CRUSH_CONFIG_JSON" \
  || { rm -f "$tmp"; echo "WARN: mcpServers merge failed — bof-mcp not registered"; }
echo "MERGED:  mcpServers.bof-mcp → $BOF_MCP_BIN_OR_BUILT"
```

### Post-Install Validation

Skipped in `--dry-run` and if zero skills were installed via nullglob.

For each name in `installed_skills[@]`:
- If `[ -f "$REPO_ROOT/skills/$skillname/SKILL.md" ]`: check `[ -f "$CRUSH_SKILLS_DIR/$skillname/SKILL.md" ]`; print `WARN: $CRUSH_SKILLS_DIR/$skillname does not contain SKILL.md` on failure.

### Summary

Print:
```
Done. $CRUSH_SKILLS_DIR contains bof skills.
Verify with: ls -la "$CRUSH_SKILLS_DIR"
Crush config: $CRUSH_CONFIG_JSON
```

Exit 0 on success; exit 1 on unrecoverable error (jq not found, skills dir missing).

### Environment Guards

Must run before any mutations. Case statement (same pattern as install_crush.sh):
```bash
case "$UNAME_O" in
  [Mm][Ss][Yy][Ss]*|[Cc][Yy][Gg][Ww][Ii][Nn]*)
    echo "ERROR: install_termux.sh does not support Git Bash or Cygwin — use WSL or install.sh instead"
    exit 1
    ;;
esac
```

---

## Acceptance Criteria

| # | Test | Pass condition |
|---|------|---------------|
| AC-1 | `test -x scripts/install_termux.sh` | Exit 0 |
| AC-2 | `bash scripts/install_termux.sh --dry-run` | Exit 0; prints DRY-RUN lines; no files created or modified |
| AC-3 | `bash scripts/install_termux.sh --foo` | Exit 1; prints USAGE line |
| AC-4 | jq absent (`PATH` stripped to exclude jq) | Exit 1; prints `ERROR: jq is required` |
| AC-5 | `HOME` unset and `CRUSH_GLOBAL_CONFIG` unset | Exit 1; prints `ERROR: $HOME is not set` |
| AC-6 | After real run: `test -f "$CRUSH_SKILLS_DIR/brainstorming/SKILL.md"` | Pass |
| AC-7 | No VS Code tool names in installed SKILL.md files | `grep -rE 'runSubagent|create_file|replace_string_in_file|run_in_terminal|manage_todo_list|read_file|grep_search|file_search|list_dir' "$CRUSH_SKILLS_DIR"` → no output |
| AC-8 | `jq -e '.options.skills_paths | contains([env.CRUSH_SKILLS_DIR])' "$CRUSH_CONFIG_JSON"` | Exit 0 |
| AC-9 | Idempotency: run twice, path appears exactly once in `options.skills_paths` | `jq '.options.skills_paths | map(select(. == env.CRUSH_SKILLS_DIR)) | length'` → `1` |
| AC-10 | crush not installed: prints WARN and continues with XDG path | Does not exit 1 |
| AC-11 | `--no-mcp`: bof-mcp not built; `mcpServers.bof-mcp` absent | `jq '.mcpServers // {} | has("bof-mcp")'` → `false` |
| AC-12 | crush.json is a symlink before run: symlink unchanged; symlink target updated | `readlink "$CRUSH_CONFIG_JSON"` same before and after; target contains merged content |
| AC-13 | Corrupted crush.json (invalid JSON) fed to jq: exits 1; backup restored | crush.json has original content; script exits 1 |

---

## Session Notes

- `crush dirs config` is the canonical discovery mechanism; `CRUSH_GLOBAL_CONFIG`
  env var is the Crush-native override for non-standard config locations.
- Termux ships busybox by default; `sed -i.bak` requires GNU sed installed via
  `pkg install sed`. Script should note this dependency if Termux is detected.
  Actually, Termux's default `sed` (busybox) does support `sed -i.bak` as of
  BusyBox 1.36 — verify before relying on it; fall back to `sed -i '' ` (BSD style)
  is NOT needed since Termux is Linux (not macOS). Use `-i.bak` (GNU style).
- `mktemp` inside `$CRUSH_CONFIG_DIR` (not `/tmp`) ensures `mv` is atomic — both
  source and dest are on the same filesystem. On Termux internal storage,
  `/tmp` and `$HOME/.config/crush` may be different mount points.
- bof-mcp build may fail if `CGO_ENABLED=1` is required for DuckDB deps — but
  bof-mcp itself has no DuckDB dependency, so `CGO_ENABLED=0` or default is fine.
- The script is intentionally platform-agnostic (Linux + macOS) with extra
  Termux-aware paths. It does not need to run on Windows — `install.sh` handles that.
- `ASSUMPTION:` skills_paths deduplication via `| unique` in jq sorts the array
  alphabetically as a side effect. This is acceptable since Crush treats the array
  as an unordered set of paths to scan.
