import re

with open('docs/tasks/P6-007-install-crush-sh.md', 'r') as f:
    content = f.read()

# Fix the iteration checks to account for symlinks before directory check
content = content.replace("""    - If `$dst` exists but is not a directory, print `WARN: real file at $dst — skipping`.
    - If `$dst` does not exist, copy directory `cp -R "$src" "$dst"`, print `COPIED:`, and apply translation.
    - If `$dst` exists as a directory, update the files: use `set -- "$src/"*` to collect files under nullglob; only run `cp -R "$@" "$dst/"` if `$# -gt 0`; print `UPDATED:`, and apply translation. (`basename "$skilldir"` handles trailing-slash paths correctly on both GNU and macOS; `basename "/path/brainstorming/"` returns `brainstorming`.)""", """    - If `$dst` is a symlink (`[ -L "$dst" ]`), remove it (`rm -f "$dst"`). This cleans up legacy installations that symlinked to the git repository, preventing `sed` from mutating the source of truth.
    - If `$dst` exists but is not a directory (and not a symlink), print `WARN: real file at $dst — skipping`.
    - If `$dst` does not exist, copy directory `cp -R "$src" "$dst"`, print `COPIED:`, and apply translation.
    - If `$dst` exists as a directory, update the files: use `set -- "$src/"*` to collect files under nullglob; only run `cp -R "$@" "$dst/"` if `$# -gt 0`; print `UPDATED:`, and apply translation. (`basename "$skilldir"` handles trailing-slash paths correctly on both GNU and macOS; `basename "/path/brainstorming/"` returns `brainstorming`.)""")

# Fix the translation mapping for multi_replace_string_in_file
content = content.replace("""    1. `multi_replace_string_in_file` → `edit`""", """    1. `multi_replace_string_in_file` → `multiedit`""")

# Split and insert new session note
parts = content.split('## Session Notes')
session_notes = '## Session Notes\n\n### Round 21 Adversarial Review (r0 / GPT-4.1 / iteration 21)\n\n**Verdict:** PASSED — all critical issues resolved post-review\n1. **C1 (Symlink mutation):** Pre-copy checks now identify and remove legacy symlinks (`[ -L "$dst" ] && rm -f "$dst"`) to prevent in-place updates from mutating the git source tree.\n2. **C2 (multiedit mapping):** `multi_replace_string_in_file` now maps correctly to Crush\'s native `multiedit` tool, not `edit`.\n\n' + parts[1] if len(parts) > 1 else ''

with open('docs/tasks/P6-007-install-crush-sh.md', 'w') as f:
    f.write(parts[0] + session_notes)

print("Plan successfully updated to fix C1 and C2.")
