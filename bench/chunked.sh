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

START_OFFSET="${START_OFFSET:-0}"
MAX_CHUNKS="${MAX_CHUNKS:-999999}"

TOTAL=$(find "$AUDIO_DIR" -name '*.flac' | wc -l)
echo "total samples: $TOTAL, chunk size: $CHUNK_SIZE, cpus: $CPUS, levels: $LEVELS"
echo "slice: start offset $START_OFFSET, max chunks $MAX_CHUNKS"

FIRST_AUDIO=$(find "$AUDIO_DIR" -name '*.flac' | sort | head -1)
if [ -z "$FIRST_AUDIO" ]; then
  echo "no audio files found under $AUDIO_DIR" >&2
  exit 1
fi

echo "preflight: loading each model once"
for LVL in $(echo "$LEVELS" | tr ',' ' '); do
  MODEL="$LVL"
  case "$LVL" in
    f16)  MODEL="ggml-large-v3.bin" ;;
    q8_0) MODEL="ggml-large-v3-q8_0.bin" ;;
    q5_0) MODEL="ggml-large-v3-q5_0.bin" ;;
    q4_0) MODEL="ggml-large-v3-q4_0.bin" ;;
  esac
  if ! "$WHISPER_CLI" -m "$MODEL_DIR/$MODEL" -f "$FIRST_AUDIO" --no-timestamps -t 2 -l en > /dev/null 2>&1; then
    echo "preflight FAILED for $MODEL — model file is missing or corrupt" >&2
    echo "run: docker-models target, or quantize manually" >&2
    exit 1
  fi
  echo "  $MODEL OK"
done

OFFSET=$START_OFFSET
CHUNK=1
while [ "$OFFSET" -lt "$TOTAL" ] && [ "$CHUNK" -le "$MAX_CHUNKS" ]; do
  if [ "$START_OFFSET" -gt 0 ]; then
    OUT="$OUT_DIR/chunk-$OFFSET.json"
  else
    OUT="$OUT_DIR/chunk-$CHUNK.json"
  fi
  if [ -s "$OUT" ]; then
    echo "=== offset $OFFSET: already done, skipping ==="
  else
    echo "=== offset $OFFSET, limit $CHUNK_SIZE ==="
    werpipe \
      -audio "$AUDIO_DIR" -transcripts "$TRANS_DIR" \
      -model-dir "$MODEL_DIR" -whisper-cli "$WHISPER_CLI" \
      -threads "$CPUS" \
      -offset "$OFFSET" -limit "$CHUNK_SIZE" \
      -levels "$LEVELS" \
      2>&1 1> "$OUT" | tee -a "$OUT.log"
  fi
  OFFSET=$((OFFSET + CHUNK_SIZE))
  CHUNK=$((CHUNK + 1))
done

echo "=== merging $CHUNK chunks ==="
werpipe merge "$OUT_DIR"/chunk-*.json | tee "$OUT_DIR/final.json"
