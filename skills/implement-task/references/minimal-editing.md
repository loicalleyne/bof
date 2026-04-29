# Minimal Editing

Source: https://nrehiew.github.io/blog/minimal_editing/

## Core Principle

Every line you change must be either:
- Directly required by a specific item in the task's In Scope list, **or**
- A necessary structural consequence of a required change.

If neither applies, revert the line.

**Necessary structural consequences** (always allowed):
- Import statements added or removed because the changed code needs them
- `gofmt`/`goimports`-required formatting (blank lines between declarations, etc.)
- Interface method changes cascading from a modified type
- Test helper signature updates required by a changed function signature
- YAML/Markdown table alignment changes caused by adding a required row

## Enforcement Procedure

Before every commit, run:

```sh
git diff           # see all unstaged changes
git diff --cached  # see staged changes
```

For every changed hunk, ask: "which In Scope item requires this, or is this a necessary structural consequence?"

- Yes to either: keep the hunk.
- No to both: revert it.

```sh
git restore --staged {file}    # unstage a file (working tree preserved)
git checkout -- {file}         # revert ALL unstaged changes in a file (only use when the
                               # file contains NOTHING required in unstaged changes)
```

**Mixed-unstaged files:** If a file has both required and unnecessary changes with nothing staged yet, do NOT use `git checkout -- {file}`. Manually edit the file to remove only the unnecessary hunks, then stage the required changes with `git add -p {file}`.

## Anti-Patterns (revert these unless they are necessary structural consequences)

| Pattern | Example | Why it's a bug |
|---------|---------|----------------|
| Style normalization | Fixing indentation in a function you didn't touch | Inflates diff, hides real changes |
| Variable rename | Renaming `err` to `taskErr` "for clarity" | Not required; adds noise to review |
| Dead code removal | Deleting a commented-out block | Out of scope unless the task says so |
| Incidental refactor | Extracting a helper "since I was here anyway" | Introduces risk; belongs in a dedicated task |
| Import reordering | Resorting imports in a file you modified | Not required; `goimports` handles this at CI |
| Blank line normalization | Adding/removing blank lines "for readability" | Out of scope unless gofmt requires it |
| Docstring rewrite | Expanding a comment you didn't need to touch | Out of scope |
| YAML/Markdown whitespace | Normalising table column widths in unrelated sections | Out of scope |
| Trailing comma / semicolon | Adding trailing commas "for consistency" | Out of scope |
| Line ending change | CRLF → LF in a file you modified | Out of scope (fix in a dedicated chore commit) |

## What Counts as In Scope

Only changes explicitly listed in the task's `## In Scope` bullets. If the
In Scope list says "implement `Foo()`", only lines inside or directly called
by `Foo()` are in scope. Adjacent functions, test helpers not named in
Acceptance Criteria, and surrounding comments are out of scope unless
explicitly listed.

## The Exception for Blocking Pre-Existing Bugs

If a pre-existing bug is directly blocking the acceptance criterion (compilation
error, panicking test), fix it minimally and note it as:

```
// ASSUMPTION: fixed pre-existing bug in {function} — required for {criterion}
```

This is a structural consequence exception, not a license to refactor.
