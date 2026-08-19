# Working log — Whisper quantization WER degradation curve

Append-only. **Never edit an existing entry**; add a new one that supersedes it. This is the
only artefact in the repository that records work *in progress*, and it is deliberately narrow.

Open entries are printed as handoff state to the next session, so this file is how one session
tells the next what was in flight. That is its primary job — write to it before you run out of
context, not after.

## What belongs here — and what does not

Most work-in-progress knowledge already has a home. Use the table before writing an entry:

| What you have | Where it goes |
| :--- | :--- |
| A measurement, even a disappointing one — it has a command and a result | `docs/05-evidence.md` as an `E-###` with the claim it refuted |
| A choice between real alternatives | `docs/adr/D-###` — the rejected options table exists for this |
| A bound on when the concept degrades | `README.md § Limitations` |
| A second concept worth its own repository | the shared backlog |
| **No decision made and no reproducible command** — a surprise, a blind alley, an abandoned attempt | **here** |
| **Work half-finished right now** | **here, as `Disposition: open`** |

If it fits a row above this file, put it there. A log entry that should have been an ADR
weakens both.

Throwaway scripts and intermediate data are not log entries — those belong in a temp directory,
not the repository. This file records findings, not files.

## Entry format

Headings must be exactly `### L-###  —  YYYY-MM-DD  —  <short title>`, and every entry needs a
`Disposition:` line. Both are parsed by the gate checks.

| Disposition | Means |
| :--- | :--- |
| `open` | Still in flight. Printed by `next` as handoff state. **Blocks the phase gate.** |
| `dead-end` | Tried, did not work, deliberately not pursued. Terminal and legitimate. |
| `promoted: D-###` | Became an architecture decision. |
| `promoted: E-###` | Became evidence. |
| `promoted: README` | Became a stated limitation. |

Every entry must reach a terminal disposition before its phase gate passes. That is the rule
that stops this file becoming a pile — and unlike most promotion rules, it is checked.

---

### L-001  —  2026-08-15  —  100-sample benchmark raw data lost in temp-dir wipe

**Context:** The 100-sample large-v3 benchmark (originally E-005) ran on the local machine
with results written to `/tmp/whisper-bench/`. A system temp cleanup wiped the directory
before the data could be committed anywhere.

**Found:** Raw per-sample data unrecoverable; only the summary table survived inside the
evidence entry. FD-011 (raw data must be committed with a checksum) cannot be met for that
entry, so it was demoted: its story is folded into the superseding full-dataset entry
E-006, which has its data committed at `evidence-data/E-006-final.json`.

Update 2026-08-19: the n=100 analysis is now derivable from the committed data —
`bench/slice.py` reproduces it (`evidence-data/E-006-n100.json`, sha256-verified).
A true re-run on fresh samples remains cheap (~30 min on a GPU) if a genuinely
new 100-sample measurement is wanted.

**Disposition:** promoted: E-006
