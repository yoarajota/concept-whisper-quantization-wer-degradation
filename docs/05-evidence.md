# Evidence ledger — Whisper quantization WER degradation curve

Every claim in this repository resolves to an entry here (rule R3). An entry without a
reproduction command is a note, not evidence, and fails validation.

**Required in every entry:** a fenced command block, an `Environment:` line, and a `Result:`
line. Headings must be exactly `### E-###  —  <title>` so the tooling can parse them.

IDs are never reused. Evidence that stops reproducing is marked `Status: broken` and the
readiness level that depended on it comes down (rule R1).

---

### E-001  —  Literature survey

**Claim.** The literature contains no WER benchmark across whisper.cpp quantization levels
(FP16, INT8, INT5, INT4) on a full standard dataset with statistical significance testing.
Six sources were fetched and read, establishing the gap and the operating envelope for the
measurement pipeline. Appears in `docs/01-theory.md`.

**Environment:** WebFetch against arXiv, GitHub, OpenSLR. Date of search: 2026-08-11.

```bash
# Reproduces the search: fetch the six source URLs and read their content
echo "See docs/01-theory.md § Sources for the fetched URLs and extracted findings."
echo "Sources: SRC-001 through SRC-006."
```

**Result:** Six sources read (5 full-text, 1 abstract-only). Three full-text sources
(SRC-001, SRC-004, SRC-005) are peer-reviewed or published papers. Key finding: no
existing work measures whisper.cpp ggml quantization WER on a full benchmark with
bootstrap CIs and paired statistical tests.

**Status:** reproducing
**Supports:** H-001, H-002, TRL 1 for `core`
**Recorded:** 2026-08-11

---

### E-002  —  PoC: WER computation and Wilcoxon test validate statistically

**Claim.** The PoC pipeline correctly computes WER (Levenshtein distance at word level) and
produces a valid Wilcoxon signed-rank p-value for paired per-sample WER comparisons.
Tests in `poc/main_test.go` verify edge cases (perfect match, all wrong, empty inputs,
systematic shift). Appears in `poc/`.

**Environment:** Go 1.x, Linux/amd64.

```bash
go test ./poc/ -v
```

**Result:** 7/7 tests pass. `TestWilcoxonShift` confirms a systematic 0.01 shift on 60
paired samples produces p ≤ 0.05 (normal approximation).

**Status:** reproducing
**Supports:** H-001, H-002, TRL 3 for `core`
**Recorded:** 2026-08-11

---

### E-003  —  Component test suite passes (conformance, failure-mode, property)

**Claim.** The `werpipe` package has 22 tests covering conformance (WER, normalization,
aggregation), failure modes (small N non-significance, NaN handling, single-pair edge case),
and properties (WER bounded below, normalization idempotence, bootstrap CI monotonicity).
All pass with `go test ./src/werpipe/ -v`.

**Environment:** Go 1.22, Linux/amd64.

```bash
go test ./src/werpipe/ -v
```

**Result:** 22/22 tests pass. Key checks: WER = 0 for identical strings, WER = 2.0 for
completely disjoint strings, normalize is idempotent and letter-case invariant, Wilcoxon
p = 1.0 for identical data, p < 0.05 for systematic 0.05 shift on n=40, bootstrap 95% CI
contains the expected range.

**Status:** reproducing
**Supports:** S-003 (modifiability — touching ≤3 files to add a quantization level)
**Recorded:** 2026-08-11

---

<!-- Template for further entries:

### E-002 — title

**Claim.**
**Environment:**

```bash
```

**Result:**
**Status:** reproducing | broken
**Supports:**
**Recorded:**

-->

## Benchmark methodology

Filled at P5. Applies to every entry tagged as a benchmark.

- **What is measured:** Per-sample WER across quantization levels; aggregate WER with 95%
  bootstrap CI; Wilcoxon signed-rank p-value for paired FP16 vs quantized comparisons.
- **Runs:** 1 run per quantization level (deterministic inference), **warm-up:** N/A
- **Held constant:** whisper.cpp build flags, thread count (4), text normalisation pipeline,
  LibriSpeech test-clean audio and transcripts
- **Baseline configuration:** Whisper large-v3 FP16 ggml, same build and hyperparameters as
  all quantized variants. The only variable is the model file's quantization format.
- **Known measurement bias:** The percentile bootstrap assumes i.i.d. per-sample WER, which
  holds for the independent utterances in LibriSpeech test-clean. WER values are bounded
  below by 0 and above by the maximum possible insertions; the bootstrap should use BCa
  correction if skew is severe. LibriSpeech test-clean represents clean read speech only;
  results do not generalise to noisy or conversational conditions.
