---
name: ship
description: Orchestrate the goxsd5 build-and-review loop for one Approved backlog item. Spawns the Implementor and Evaluator as subagents and iterates them back-and-forth in ONE session until the conformance gates pass, then squash-merges. Use to run the auto-maintenance Ship phase (manually, or from the scheduled routine).
---

# /ship — Implementor ⇄ Evaluator loop

Drive one Approved card from goxsd5's backlog to a merged change in a single
session. See `MAINTENANCE.md` for the full system. Roles are defined in
`.claude/agents/{implementor,evaluator}.md`.

`$ARGUMENTS` may name an issue/card; if empty, **drain the Approved column** —
ship cards back-to-back until it's empty or the session quota runs out.

## Procedure

**Outer loop — drain the queue.** Repeat steps 1–5 for successive Approved cards.
Stop only when the Approved column is empty (report "nothing approved") **or the
session is cut off by quota/budget** — there is no fixed card limit per run. Don't
try to estimate remaining quota; just keep pulling the next card until either the
queue empties or the run is terminated mid-card. A single arg (`$ARGUMENTS`
naming one card) ships just that card, then stops. A mid-card cutoff is **not**
guarded by guessing quota — it is *recovered* by step 0 on the next run.

0. **Resume in-flight work first.** Before touching the Approved queue, look for a
   card interrupted by a prior cutoff: one **In Progress** or **In Review** whose
   issue has an **open, unmerged PR**. If found, read that PR's
   `🔄 ship-checkpoint` comments to recover where the loop left off (last round,
   outstanding findings) and **resume its review→fix loop (step 3) from there**
   before picking any new Approved card.

1. **Pick the card.** `gh project item-list 1 --owner kud360 --format json` →
   first item with status **Approved**. If the Approved column is empty, stop the
   outer loop and report "nothing approved." Move the chosen card to
   **In Progress**.

2. **Implement.** Spawn the `implementor` agent with the issue number/body.
   It branches, implements, runs the gates, and opens a PR. Capture the PR #.
   Move the card to **In Review**, and post the opening `🔄 ship-checkpoint`
   comment on the PR (see Rules).

3. **Review ⇄ fix loop** (max **5 rounds**):
   - Spawn the `evaluator` agent with the PR #.
   - **If the change touches spec-semantic packages** (`xsd`, `parser`,
     `xsdvalidate`, `xpath`, `xsdwalk`, `xsdtemporal`, `builtin/*`), also spawn
     the `xsd-expert` agent on the PR, in parallel. GREEN requires the Evaluator
     GREEN **and** xsd-expert SPEC-OK.
   - **All green** → break out of the loop.
   - **Otherwise** → merge the Evaluator's CHANGES with any xsd-expert
     SPEC-ISSUE items into one combined findings list and send it to the
     Implementor. **Prefer `SendMessage`** to the *same* `implementor` agent (its
     context is intact). **If `SendMessage` is unavailable** (some cloud sessions
     don't have it), **spawn a fresh `implementor`** instead, handing it full
     resume context: the issue #, the `maint/*` branch + PR #, what is already
     done, and the combined findings to fix — it picks up from the existing
     branch. Then re-review: `SendMessage` the *same* reviewer agent(s), or — if
     SendMessage is unavailable — spawn a fresh `evaluator` (+ `xsd-expert`) on the
     PR; a fresh reviewer only strengthens independence. Repeat.
   - **Checkpoint each round:** post a `🔄 ship-checkpoint` comment on the PR with
     the round #, the Evaluator (and xsd-expert) verdict, and the outstanding
     findings — so a cutoff mid-loop is resumable by step 0.

4. **On GREEN:** squash-merge the PR (`gh pr merge <#> --squash --delete-branch`),
   move the card to **Done**, and send a push notification:
   *"shipped #<issue>: <title> — baseline held <schema>/<instance>."*
   **Append a log entry** for this PR to the open `maint-log` issue (see Rules),
   then **return to step 1** for the next Approved card.

5. **On stuck** (5 rounds without GREEN, or the Implementor reports the work
   can't hold the baseline): do **not** merge. Leave the PR open with the
   Evaluator's outstanding findings, move the card to a `blocked` label (not back
   to Approved — it must not be re-picked), and push-notify:
   *"#<issue> stuck after N rounds — needs you."*
   **Log the stuck outcome** to the open `maint-log` issue (see Rules), then
   **return to step 1** for the next Approved card — one stuck card must not block
   the rest of the queue.

## Rules
- Keep Implementor and Evaluator in **separate agent contexts** — the Evaluator's
  independence is the point. Continue each with `SendMessage` when available; where
  it isn't (some cloud sessions lack it), re-spawn a fresh agent each round with
  full resume context (branch, PR #, combined findings). Re-spawning is
  functionally equivalent — the branch/PR carries the state — just less
  context-efficient.
- **Log every PR outcome.** After each merge or stuck, append a one-line entry to
  the *current open* `maint-log` issue (`gh issue list --label maint-log --state
  open` → `gh issue comment <#>`): `#<issue> "<title>" → PR #<pr>: MERGED|STUCK ·
  rounds n/5 · Evaluator <verdict> · xsd-expert <verdict|n/a> · baselines
  5672/21429 · <what changed> · notable: <…>`. This is the loop's durable history
  — the cloud transcripts are ephemeral.
- **Checkpoint on the PR as you go.** Post a `🔄 ship-checkpoint` comment on the PR
  at each transition — opened, each round's verdict + outstanding findings,
  xsd-expert verdict, and final merge/stuck. Use the literal `🔄 ship-checkpoint`
  marker so a resuming run (step 0) can find them. These are the per-step resume
  breadcrumbs; the `maint-log` entry is the single high-level per-PR summary.
- The **orchestrator merges**, never the Evaluator. Merge only on a GREEN verdict
  whose objective gate (5672 / 21429 conformance baselines) actually passed.
- **Unattended (scheduled) runs merge on GREEN without any further human
  confirmation** — the sole human gate is Proposed → Approved upstream. GREEN is
  defined entirely by the objective gate + Evaluator (+ xsd-expert) verdicts; if
  GREEN is not reached within the 5-round cap, leave the PR open and push-notify.
- Never push to `main` directly; everything goes through the PR.
