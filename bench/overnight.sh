#!/bin/sh
# Overnight benchmark launcher (host-side). Starts a detached, resumable
# container that runs the full dataset in chunks. Safe to re-run — finished
# chunks are skipped.
#
# Usage:
#   sh bench/overnight.sh
#
# Env overrides:
#   CPUS=6          CPU cap for the container (default: all cores)
#   CHUNK_SIZE=200  samples per chunk
#   LEVELS          default: f16,q8_0,q5_0,q4_0
#   CACHE_DIR       default: ~/.cache/whisper-bench
#
# Status:
#   docker logs -f werpipe-overnight     # live progress
#   ls ~/.cache/whisper-bench/out/       # chunk files done so far
set -eu

CPUS="${CPUS:-6}"
CHUNK_SIZE="${CHUNK_SIZE:-200}"
LEVELS="${LEVELS:-f16,q8_0,q5_0,q4_0}"
CACHE_DIR="${CACHE_DIR:-$HOME/.cache/whisper-bench}"

MODELS_DIR="$CACHE_DIR/models"
DATA_DIR="$CACHE_DIR/data/flat"
OUT_DIR="$CACHE_DIR/out"

if [ ! -d "$DATA_DIR" ] || [ ! -d "$MODELS_DIR" ]; then
  echo "missing data or models under $CACHE_DIR" >&2
  echo "expected: $MODELS_DIR and $DATA_DIR" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

docker rm -f werpipe-overnight > /dev/null 2>&1 || true

echo "launching overnight run: $LEVELS, $CHUNK_SIZE-sample chunks, $CPUS CPUs"
docker run -d --name werpipe-overnight \
  --cpus="$CPUS" \
  -v "$MODELS_DIR":/models \
  -v "$DATA_DIR":/data:ro \
  -v "$OUT_DIR":/out \
  --entrypoint sh concept-whisper-wer \
  -c "AUDIO_DIR=/data TRANS_DIR=/data MODEL_DIR=/models OUT_DIR=/out \
      WHISPER_CLI=/whisper.cpp/build/bin/whisper-cli \
      sh /bench/chunked.sh $CHUNK_SIZE $CPUS '$LEVELS'"

echo "started. watch with: docker logs -f werpipe-overnight"
echo "when done: docker run --rm -v $OUT_DIR:/out --entrypoint werpipe concept-whisper-wer merge /out/chunk-*.json > $OUT_DIR/final.json"
