---
name: brainstorming
description: >
  You MUST use this before any creative work — creating features, building
  components, adding functionality, or modifying behavior. Explores user intent,
  requirements, and design before implementation. Triggers on: "I want to add",
  "help me design", "plan a feature", "brainstorm", "let's build".
---

# Brainstorming Ideas Into Designs

Help turn ideas into fully formed designs and specs through natural collaborative dialogue.

Start by understanding the current project context, then ask questions one at a time. Once you understand what you're building, present the design and get user approval.

<HARD-GATE>
Do NOT invoke any implementation skill, write any code, scaffold any project, or
take any implementation action until you have presented a design and the user
has approved it. This applies to EVERY project regardless of perceived simplicity.
</HARD-GATE>

---

## Anti-Pattern: "This Is Too Simple To Need A Design"

Every project goes through this process. A todo list, a single-function utility, a config change — all of them. "Simple" projects are where unexamined assumptions cause the most wasted work. The design can be short (a few sentences for truly simple projects), but you MUST present it and get approval.

---

## Checklist

Use `manage_todo_list` (VS Code) / `todos` (Crush) to create a task for each item and complete them in order:

1. **Explore project context** — read `AGENTS.md` (if present), check files, docs, recent commits
2. **Offer napkin visual companion** (if topic will involve visual questions) — this is its own message, not combined with a clarifying question
3. **Ask clarifying questions** — one at a time, understand purpose/constraints/success criteria
4. **Propose 2-3 approaches** — with trade-offs and your recommendation
5. **Present design** — in sections scaled to their complexity, get user approval after each section
6. **Write design doc** — save to `docs/specs/YYYY-MM-DD-<topic>-design.md` and commit
7. **Spec self-review** — quick inline check for placeholders, contradictions, ambiguity, scope
8. **User reviews written spec** — ask user to review the spec file before proceeding
9. **Transition to implementation** — invoke `bof:writing-plans` to create the implementation plan

---

## Process Flow

```
Explore project context
        ↓
Will visual questions arise? → YES → Offer napkin skill (own msg) → Ask clarifying questions
                            → NO  → Ask clarifying questions
        ↓
Propose 2-3 approaches
        ↓
Present design sections → User approves? → NO (revise) → re-present
                                        → YES
        ↓
Write design doc → Spec self-review (fix inline)
        ↓
User reviews spec? → Changes requested → revise → re-review
                  → Approved
        ↓
[INVOKE bof:writing-plans]
```

**The terminal state is invoking `bof:writing-plans`.** Do NOT invoke any implementation skill. The ONLY skill you invoke after brainstorming is `bof:writing-plans`.

---

## The Process

**Understanding the idea:**

- Check current project state first (read `AGENTS.md`, check `GLOSSARY.md`, look at `docs/`, recent commits via `run_in_terminal` (VS Code) / `bash` (Crush))
- Before asking detailed questions, assess scope: if the request describes multiple independent subsystems, flag this immediately and help decompose before proceeding
- For appropriately-scoped projects, ask questions one at a time to refine the idea
- Prefer multiple-choice questions when possible; use `vscode_askQuestions` (VS Code) / ask inline (Crush) with options for fixed choices
- Only one question per message

**Exploring approaches:**

- Propose 2-3 different approaches with trade-offs
- Lead with your recommended option and explain why

**Presenting the design:**

- Scale each section to its complexity: a few sentences if straightforward, up to 200-300 words if nuanced
- Ask after each section whether it looks right (use `vscode_askQuestions` (VS Code) / ask inline (Crush) for binary approve/revise)
- Cover: architecture, components, data flow, error handling, testing

---

## After the Design

**Documentation:**

- Write the validated design (spec) to `docs/specs/YYYY-MM-DD-<topic>-design.md`
  - If `docs/specs/` does not exist, create it
  - For Esquisse projects this aligns with the Esquisse directory structure
- Commit the design document via `run_in_terminal` (VS Code) / `bash` (Crush)

**Spec Self-Review:**
After writing the spec, check with fresh eyes:

1. **Placeholder scan:** Any "TBD", "TODO", incomplete sections, or vague requirements? Fix them.
2. **Internal consistency:** Do any sections contradict each other?
3. **Scope check:** Is this focused enough for a single implementation plan?
4. **Ambiguity check:** Could any requirement be interpreted two different ways? If so, pick one.

Fix issues inline. No need to re-review — just fix and move on.

**User Review Gate:**
After the spec self-review, ask the user to review:

> "Spec written and committed to `docs/specs/<filename>.md`. Please review it and let me know if you want to make any changes before we start writing the implementation plan."

Wait for the user's response. If they request changes, make them and re-run the spec review loop. Only proceed once the user approves.

---

## Visual Companion (napkin skill)

When you anticipate upcoming questions will involve visual content (mockups, layouts, diagrams), offer the napkin skill once:

> "Some of what we're working on might be easier to explain visually — mockups, diagrams, comparisons. I can use the napkin whiteboard skill to sketch these out in your browser. Want to try it?"

**This offer MUST be its own message.** Do not combine it with clarifying questions. Wait for the user's response before continuing.

Per-question decision: even after the user accepts, decide FOR EACH QUESTION whether to use the napkin skill or plain text. The test: **would the user understand this better by seeing it than reading it?**

- **Use napkin** for: mockups, wireframes, layout comparisons, architecture diagrams
- **Use plain text** for: requirements questions, conceptual choices, tradeoff lists, A/B/C options

---

## Key Principles

- **One question at a time** — Don't overwhelm with multiple questions
- **Multiple choice preferred** — Use `vscode_askQuestions` (VS Code) / ask inline (Crush) with options when choices are bounded
- **YAGNI ruthlessly** — Remove unnecessary features from all designs
- **Explore alternatives** — Always propose 2-3 approaches before settling
- **Incremental validation** — Present design, get approval before moving on
- **Read the project constitution** — `AGENTS.md` informs invariants; don't design against them
