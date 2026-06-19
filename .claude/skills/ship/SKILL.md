---
name: ship
description: Orchestrate the goxsd5 build-and-review loop for one Approved backlog item. Spawns the Implementor and Evaluator as subagents and iterates them back-and-forth in ONE session until the conformance gates pass, then squash-merges. Use to run the auto-maintenance Ship phase (manually, or from the scheduled routine).
---

# /ship — Implementor ⇄ Evaluator loop

Drive one Approved card from goxsd5's backlog to a merged change in a single
session. See `MAINTENANCE.md` for the full system. Roles are defined in
`.claude/agents/{implementor,evaluator}.md`.

`$ARGUMENTS` may name an issue/card; if empty, pick the top item in the
**Approved** column.

## Procedure

1. **Pick the card.** If no arg: `gh project item-list 1 --owner kud360
   --format json` → first item with status **Approved**. If the Approved column
   is empty, stop and report "nothing approved." Move the chosen card to
   **In Progress**. **One card per run.**

2. **Implement.** Spawn the `implementor` agent with the issue number/body.
   It branches, implements, runs the gates, and opens a PR. Capture the PR #.
   Move the card to **In Review**.

3. **Review ⇄ fix loop** (max **5 rounds**):
   - Spawn the `evaluator` agent with the PR #.
   - **If the change touches spec-semantic packages** (`xsd`, `parser`,
     `xsdvalidate`, `xpath`, `xsdwalk`, `xsdtemporal`, `builtin/*`), also spawn
     the `xsd-expert` agent on the PR, in parallel. GREEN requires the Evaluator
     GREEN **and** xsd-expert SPEC-OK.
   - **All green** → break out of the loop.
   - **Otherwise** → merge the Evaluator's CHANGES with any xsd-expert
     SPEC-ISSUE items and `SendMessage` the *same* `implementor` agent with the
     combined list (its context is intact — do NOT re-spawn). When it reports
     back, `SendMessage` the *same* reviewer agent(s) to re-check. Repeat.

4. **On GREEN:** squash-merge the PR (`gh pr merge <#> --squash --delete-branch`),
   move the card to **Done**, and send a push notification:
   *"shipped #<issue>: <title> — baseline held <schema>/<instance>."*

5. **On stuck** (5 rounds without GREEN, or the Implementor reports the work
   can't hold the baseline): do **not** merge. Leave the PR open with the
   Evaluator's outstanding findings, move the card to **In Progress** (or a
   `blocked` label), and push-notify:
   *"#<issue> stuck after N rounds — needs you."*

## Rules
- Keep Implementor and Evaluator in **separate agent contexts** — the Evaluator's
  independence is the point. Continue each with `SendMessage`, don't re-spawn.
- The **orchestrator merges**, never the Evaluator. Merge only on a GREEN verdict
  whose objective gate (5672 / 21429 conformance baselines) actually passed.
- Never push to `main` directly; everything goes through the PR.
