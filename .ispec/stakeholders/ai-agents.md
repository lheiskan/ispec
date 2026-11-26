---
type: stakeholder
stable_id: STK-2
slug: ai-agents
status: active
---

## Identity

AI coding agents (Claude Code, Cursor, etc.) operating on an ispec-using project. They read the canon to understand current state, propose changes through skills, and walk the link graph to assemble context.

They are also the consumers [[ADR-0006]] has in mind: an agent asked to write a user manual, answer why a design is the way it is, or reimplement a subsystem is working from the canon. ispec's job is to make that material worth reading; producing the artifact is outside ispec's loop.

## Goals

- Load only the canon relevant to the current task; avoid loading monolithic docs that exceed the context budget.
- Walk the card graph efficiently to assemble the right context.
- Receive precise lint feedback when their drafts are inconsistent.
- Have explicit confirmation points where user input is genuinely load-bearing, rather than guessing on ambiguous decisions.

## Constraints

- Bounded context window; cannot reliably load monolithic files.
- Stochastic outputs; need structured rules and confirmation points to avoid silent drift.
- No persistent memory across sessions beyond what is committed to the repo.

## Communication

Cards as Markdown with YAML frontmatter; lint output (errors and warnings); skill descriptions and confirmation dialogs surfaced through the host CLI.
