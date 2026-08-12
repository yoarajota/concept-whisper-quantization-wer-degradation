#!/bin/bash
set -euo pipefail

MODEL_DIR="${MODEL_DIR:-/models}"
WHISPER_DIR="${WHISPER_DIR:-/whisper.cpp}"

echo "Downloading large-v3 (FP16) ..."
curl -sSL "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin" \
    -o "${MODEL_DIR}/ggml-large-v3.bin"

QUANTIZE="${WHISPER_DIR}/build/bin/quantize"
SRC="${MODEL_DIR}/ggml-large-v3.bin"

echo "Quantizing to Q8_0 ..."
"${QUANTIZE}" "${SRC}" "${MODEL_DIR}/ggml-large-v3-q8_0.bin" q8_0 2>/dev/null

echo "Quantizing to Q5_0 ..."
"${QUANTIZE}" "${SRC}" "${MODEL_DIR}/ggml-large-v3-q5_0.bin" q5_0 2>/dev/null

echo "Quantizing to Q4_0 ..."
"${QUANTIZE}" "${SRC}" "${MODEL_DIR}/ggml-large-v3-q4_0.bin" q4_0 2>/dev/null

echo "Done. Models in ${MODEL_DIR}:"
ls -lh "${MODEL_DIR}"
