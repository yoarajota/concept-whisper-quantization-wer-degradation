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
