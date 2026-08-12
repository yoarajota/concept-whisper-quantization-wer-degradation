# Agent contract — Whisper quantization WER degradation curve

This repository is a concept project (`C-003`, archetype `implementation`).
It answers exactly one question:

> What INT quantization level produces the first statistically significant WER degradation in Whisper vs FP16 on clean speech?

## Content rules that bind every document here

- **No unevidenced claim.** Every capability, number, or comparison in `README.md` or `docs/`
  carries an `E-###` that resolves to an entry with a reproduction command, an environment, and
  an observed result. Adjectives without an ID get deleted, not softened.
- **No hand-written scores.** The `computed` block in `.sota/readiness.yaml` is machine-written
  and never hand-edited.
- **Readiness is descriptive.** Claim the highest level whose evidence reproduces *today*. If
  something broke, lower the level and say so.
- **Baseline or nothing.** "Faster" requires the named baseline, the metric, the conditions, and
  the `E-###`.
- **Complexity must be purchased.** Anything structural traces to a scenario `S-###` in
  `.sota/quality-gates.yaml` and is recorded in `complexity_ledger`. Otherwise remove it.
- **One question.** A second interesting idea goes to the shared backlog, not into this
  repository.
- **Real citations only.** Sources must have been fetched and read, not recalled.
- **Honest limitations.** `README.md § Limitations` names the conditions under which this
  degrades, what is untested, and what would move it up a readiness level.

## Session handoff

Before running low on context, append an `L-###` entry to `docs/03-log.md` with
`Disposition: open` describing what is half-finished. That file is the state that survives
a session boundary.

## Stable IDs used in this repository

| Prefix | Meaning | Lives in |
| :--- | :--- | :--- |
| `C-###` | Concept | `.sota/concept.yaml` |
| `H-###` | Hypothesis | `.sota/concept.yaml` |
| `E-###` | Evidence | `docs/05-evidence.md` |
| `L-###` | Working-log entry | `docs/03-log.md` (append-only) |
| `SRC-###` | Source, with an `Access:` level | `docs/01-theory.md` |
| `S-###` | Quality-attribute scenario | `.sota/quality-gates.yaml` |
| `D-###` | Architecture decision record | `docs/adr/D-###-*.md` |
| `SP-###` / `TP-###` | Sensitivity / tradeoff point | `docs/04-tradeoffs.md` |
| `R-###` / `NR-###` | Risk / non-risk | `docs/04-tradeoffs.md` |
| `X-###` | ISO 5055 exemption | `.sota/quality-gates.yaml` |

IDs are never reused or renumbered. Retired IDs stay in place marked `retired`.
