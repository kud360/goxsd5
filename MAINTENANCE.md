# goxsd5 auto-maintenance

A self-running maintenance loop: agents discover work, you approve it, and a
build-and-review loop ships it — with the W3C conformance suite as the objective
gate. This file is the contract every (cold-start) agent shares.

## Roles

| Role | Defined in | Trigger | Mandate |
|---|---|---|---|
| **Planner** | `.claude/agents/planner.md` | scheduled, 08:00 ET daily | Discover gaps/features (incl. architecture) → write proposals to **Proposed**. Read-only; never codes. |
| *(you)* | — | manual | Triage **Proposed** → **Approved** (or reject). Add your own cards straight to **Approved**. |
| **Ship** | `.claude/skills/ship/SKILL.md` | scheduled, 02:00 / 14:00 / 20:00 ET | Pull one **Approved** card; run the Implementor ⇄ Evaluator loop to green; squash-merge. |
| **Implementor** | `.claude/agents/implementor.md` | spawned by Ship | Build the change on a `maint/*` branch, open a PR, fix on findings. Never merges. |
| **Evaluator** | `.claude/agents/evaluator.md` | spawned by Ship | Independently run the gate + review (correctness, conventions, **big-picture/pattern fit**), return GREEN/CHANGES. Never edits, never merges. |
| **XSD-expert** | `.claude/agents/xsd-expert.md` | spawned by Ship (spec-semantic changes) | Judge **spec fidelity** against `docs/` — normative rule, error-id clause, corner cases. Read-only consultant; returns SPEC-OK / SPEC-ISSUE. |

## Board — GitHub Project #1 (`kud360/goxsd5`)

State machine via the Status field:

```
Proposed ──(you approve)──► Approved ──(/ship picks up)──► In Progress
   ▲                                                            │
   │ (Planner adds)                                             ▼
 (reject = close)                                          In Review ──► Done
                                                               ▲   │
                                                               └───┘  (Eval ⇄ Impl loop)
```

- The **only human gate is `Proposed → Approved`.** Nothing reaches code until a
  human moves a card there. Your own ideas enter the same way.
- `maintenance` label on all auto-proposals; `needs-deep-review` on
  architecture/structural ones.

## The objective gate

A change is mergeable iff the gated conformance suites pass:

- `go test ./parser -run TestConformanceSuite` — schema validity, **baseline 5672**
- `go test ./parser -run TestInstanceConformance` — instance validity, **baseline 21429**

Both **fail on regression AND on unrecorded improvement** (the ratchet). Genuine
improvements require regenerating `testdata/*-expectations.txt` with the
`-update-*expectations` flags and committing them. A baseline is **never lowered**.
Plus `go test ./...` and `.claude/skills/run-goxsd5/smoke.sh`. The whole gate is
one command: **`tools/gate.sh`**.

On top of the mechanical gate, a PR must clear the Evaluator's judgment +
**big-picture/pattern** review, and — for spec-semantic changes — the
**XSD-expert**'s spec-fidelity verdict. GREEN = gate passes AND Evaluator GREEN
AND (if invoked) xsd-expert SPEC-OK.

## Guardrails

- **One card per Ship run** — at most 3 merges/day, easy to follow.
- **Round cap: 5.** If the Implementor ⇄ Evaluator loop doesn't reach GREEN in 5
  rounds, Ship stops, leaves the PR open with findings, and push-notifies you.
- **Separate contexts** for Implementor and Evaluator — the Evaluator must keep
  critical distance. They're continued with `SendMessage`, not re-spawned.
- **Squash-merge** keeps history revertable. The orchestrator merges, never the
  Evaluator.
- Notifications: **push notification** on plan-ready, on merge, and on stuck.

## Schedule (Phase 4 — wire only after the loop is proven by hand)

Timezone `America/New_York` (tracks the zone, not a fixed offset — DST-safe):

| Routine | Local ET | Cron |
|---|---|---|
| Planner | 08:00 daily | `0 8 * * *` |
| Ship | 02:00, 14:00, 20:00 daily | `0 2,14,20 * * *` |

## Rollout phases

- **Phase 0:** `gh auth refresh -s project` (grant the `project` scope).
- **Phase 1:** these contracts + the board columns exist. *Nothing fires.* ← you are here
- **Phase 2:** run the Planner once by hand → it populates **Proposed**; approve a few.
- **Phase 3:** run `/ship` on one Approved card; watch the loop merge it.
- **Phase 4:** enable the two cron routines above + Ship auto-merge.
