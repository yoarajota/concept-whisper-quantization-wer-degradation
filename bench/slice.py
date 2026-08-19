#!/usr/bin/env python3
"""Slice the first N samples per level from a merged werpipe report.

Reproduces the n=100 sub-analysis of the full dataset (E-006) to show what a
100-sample measurement can and cannot see. Deterministic: samples are sorted
by ID, so the slice is identical on every run.

Usage:
    python3 bench/slice.py <input.json> <output.json> [N]
"""

import json
import sys


def main() -> int:
    if len(sys.argv) < 3:
        print(__doc__)
        return 1
    src, dst = sys.argv[1], sys.argv[2]
    n = int(sys.argv[3]) if len(sys.argv) > 3 else 100

    report = json.load(open(src))
    sliced = []
    for entry in report:
        samples = sorted(entry["results"]["Samples"], key=lambda s: s["SampleID"])
        sliced.append({
            **{k: v for k, v in entry.items() if k != "results"},
            "results": {**entry["results"], "Samples": samples[:n]},
        })
    json.dump(sliced, open(dst, "w"), indent=2)
    print(f"sliced {n} samples per level from {src} into {dst}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
