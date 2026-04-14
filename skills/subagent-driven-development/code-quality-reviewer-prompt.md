# Code Quality Reviewer Prompt Template

Use this as the structure for `runSubagent("CodeQualityReviewerAgent", prompt)` calls.
Fill in the bracketed sections from your current context.

---

```
You are reviewing the code quality for [TASK TITLE] in [PROJECT NAME].

The implementation has already passed spec compliance review.

## Changed Files

The following files were created or modified:
[LIST FILES from ImplementerAgent's report]

## Project Standards

From AGENTS.md, the relevant code conventions are:
[PASTE relevant sections: error handling, naming, godoc requirements, test requirements,
invariants, common mistakes to avoid]

## Git Range

[OUTPUT OF: git log --oneline -5]

## Review Instructions

Check ONLY: code quality, idioms, correctness, and conformance to project conventions.

Do NOT re-check spec compliance — that is already confirmed.

Severity guide:
- **Critical**: Must fix before merge (security, data loss, panic, broken invariant)
- **Important**: Should fix before merge (no tests, no godoc on exported, misleading name, violation of stated convention)
- **Minor**: May defer (style, nitpick, optional improvement)

## Output Format

If approved:
✅ APPROVED
[List any Minor issues as FYI, labeled Minor]

If issues found:
❌ ISSUES FOUND
Critical:
- [issue description with file:line reference]
Important:
- [issue description]
Minor:
- [issue description]
```
