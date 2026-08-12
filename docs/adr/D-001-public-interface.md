# D-001 — Public interface: `werpipe` package

**Date:** 2026-08-11
**Status:** accepted

## Context

The critical function from P0 is: measure WER of Whisper across whisper.cpp quantization
levels and compute statistical significance of differences. P2 proved the mechanism with
a monolithic poc/main.go. P3 requires a proper Go package with tests.

## Decision

Design the package as `src/werpipe` with a single `Pipeline` struct and pure functions:

```
Pipeline.Run(level, modelName, samples) → []SampleResult
Aggregate([]SampleResult)                → LevelResults
Compare(baseline, quant []SampleResult)  → Comparison
ComputeWER(reference, hypothesis)        → float64
Normalize(text)                          → string
```

**`Pipeline` contains the whisper.cpp dependency** (exec path, model directory, thread
count). All statistical functions (`wilcoxonP`, `bootstrapCI95`, `Aggregate`, `Compare`)
are pure and operate on `[]SampleResult` — they need no external process.

**`Sample` and `SampleResult` carry sample identity through the pipeline**, so errors
on individual samples can be traced and reported without stopping the run.

## Rejected options

| Option | Reason rejected |
| :--- | :--- |
| Streaming / channel-based pipeline | Adds concurrency complexity with no scenario requiring it. LibriSpeech test-clean is ~5.4 hours; serial processing completes within the S-001 time budget. |
| Embed whisper.cpp via CGo | Pin a C++ build dependency that varies by OS/arch; exec wrapper isolates the tool boundary. |
| Config file for quantization levels | The levels list is the concept's input; adding a config layer between it and the pipeline violates P2 simplicity. |
| External stats library (gonum, etc.) | Adds a ~30 MB dependency for 3 functions totalling < 100 lines; a concept repo should own its arithmetic so readers can verify it directly. |
