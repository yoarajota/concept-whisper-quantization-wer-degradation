#!/bin/sh
# Chunked benchmark runner. Splits the full dataset into N chunks, runs each with
# the werpipe CLI (at low CPU priority), and merges the per-chunk JSON reports.
#
# Usage:
#   bench/chunked.sh [CHUNK_SIZE] [CPUS] [LEVELS]
#
# Examples:
#   bench/chunked.sh 200 2                       # 200-sample chunks, 2 CPUs
#   bench/chunked.sh 100 1 "f16,q4_0"            # probe: f16 + q4_0 only
set -eu

CHUNK_SIZE="${1:-200}"
CPUS="${2:-2}"
LEVELS="${3:-f16,q8_0,q5_0,q4_0}"

AUDIO_DIR="${AUDIO_DIR:-/data}"
TRANS_DIR="${TRANS_DIR:-/data}"
MODEL_DIR="${MODEL_DIR:-/models}"
WHISPER_CLI="${WHISPER_CLI:-/whisper.cpp/build/bin/whisper-cli}"
OUT_DIR="${OUT_DIR:-/out}"

mkdir -p "$OUT_DIR"

TOTAL=$(find "$AUDIO_DIR" -name '*.flac' | wc -l)
echo "total samples: $TOTAL, chunk size: $CHUNK_SIZE, cpus: $CPUS, levels: $LEVELS"

OFFSET=0
CHUNK=0
while [ "$OFFSET" -lt "$TOTAL" ]; do
  CHUNK=$((CHUNK + 1))
  OUT="$OUT_DIR/chunk-$CHUNK.json"
  if [ -s "$OUT" ]; then
    echo "=== chunk $CHUNK: already done, skipping (offset $OFFSET) ==="
  else
    echo "=== chunk $CHUNK: offset $OFFSET, limit $CHUNK_SIZE ==="
    werpipe \
      -audio "$AUDIO_DIR" -transcripts "$TRANS_DIR" \
      -model-dir "$MODEL_DIR" -whisper-cli "$WHISPER_CLI" \
      -threads "$CPUS" \
      -offset "$OFFSET" -limit "$CHUNK_SIZE" \
      -levels "$LEVELS" \
      > "$OUT" 2> "$OUT.log"
  fi
  OFFSET=$((OFFSET + CHUNK_SIZE))
done

echo "=== merging $CHUNK chunks ==="
werpipe merge "$OUT_DIR"/chunk-*.json | tee "$OUT_DIR/final.json"
