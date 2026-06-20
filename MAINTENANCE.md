# goxsd5 auto-maintenance

A self-running maintenance loop: agents discover work, you approve it, and a
build-and-review loop ships it — with the W3C conformance suite as the objective
gate. This file is the contract every (cold-start) agent shares.

## Roles

| Role | Defined in | Trigger | Mandate |
|---|---|---|---|
| **Planner** | `.claude/agents/planner.md` | scheduled, 08:00 ET daily | Discover gaps/features (incl. architecture) → write proposals to **Proposed**; refine open proposals from reviewer comments. Read-only on code; never codes. |
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

- **Ship drains the Approved queue** — after each merge it pulls the next
  Approved card and repeats until the queue is empty or the session quota runs
  out. No fixed card limit per run; throughput is bounded by quota, not a count.
- **Round cap: 5 (per card).** If the Implementor ⇄ Evaluator loop doesn't reach
  GREEN in 5 rounds, Ship leaves that PR open with findings, labels it `blocked`,
  push-notifies you, and moves on to the next card. This is a per-card
  convergence guard, not a per-run throttle.
- **Separate contexts** for Implementor and Evaluator — the Evaluator must keep
  critical distance. Continue each with `SendMessage` when available; where it
  isn't (some cloud sessions lack it), re-spawn a fresh agent per round with full
  resume context (branch, PR #, findings) — equivalent, just less context-efficient.
- **Squash-merge** keeps history revertable. The orchestrator merges, never the
  Evaluator.
- Notifications: **push notification** on plan-ready, on merge, and on stuck.

## Schedule (Phase 4 — LIVE)

Implemented as **Claude Code cloud routines** (`/schedule` → RemoteTrigger), not
OS cron. The cloud scheduler is **UTC-only**, so the `America/New_York` target
times are converted at the EDT offset (UTC−4). This is *not* DST-safe: after ET
returns to EST (UTC−5) each routine fires one hour earlier in local terms —
acceptable for an overnight loop; re-convert the crons at the DST boundary if the
local time matters.

| Routine | Target ET | UTC cron (live) |
|---|---|---|
| Planner | 08:00 daily | `0 12 * * *` |
| Ship | 02:00, 14:00, 20:00 daily | `0 0,6,18 * * *` |

## Rollout phases

- **Phase 0:** `gh auth refresh -s project` (grant the `project` scope).
- **Phase 1:** these contracts + the board columns exist. *Nothing fires.* ✅ done
- **Phase 2:** run the Planner once by hand → it populates **Proposed**; approve a few. ✅ done
- **Phase 3:** run `/ship` on one Approved card; watch the loop merge it. ✅ done (PR #4)
- **Phase 4:** the two cloud routines above are live and Ship merges unattended on
  GREEN. ✅ **active** — the only human gate that remains is **Proposed → Approved**.
