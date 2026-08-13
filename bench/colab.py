#!/usr/bin/env python3
"""
Google Colab runner for the Whisper quantization WER benchmark (C-003).

Paste ONE cell in a Colab notebook with a GPU runtime (T4 is enough):

    !wget -q https://raw.githubusercontent.com/yoarajota/concept-whisper-quantization-wer-degradation/master/bench/colab.py
    %run colab.py

The script is resumable: chunk files are stored on Google Drive, so if the
session dies (free tier: ~12h cap), re-run the cell in a new session and it
continues from the last completed chunk.

Environment note for evidence: this is a GPU environment (CUDA). Results from
here are a separate environment declaration from the CPU run — do not merge
chunk files from the two environments into one comparison.
"""

import os
import pathlib
import shutil
import subprocess
import sys
import tarfile
import time

WORK = pathlib.Path("/content/whisper-bench")
OUT = None  # set after Drive mount


def sh(cmd: str, **kw) -> subprocess.CompletedProcess:
    print(f"\n$ {cmd}", flush=True)
    p = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                         text=True, bufsize=1, **kw)
    assert p.stdout is not None
    for line in p.stdout:
        print(line.rstrip(), flush=True)
    p.wait()
    if p.returncode != 0:
        print(f"\nFAILED ({p.returncode}): {cmd}", flush=True)
        sys.exit(p.returncode)
    return subprocess.CompletedProcess(cmd, p.returncode, "", "")


def drive_mount():
    global OUT
    try:
        from google.colab import drive  # type: ignore
        drive.mount("/content/drive")
        OUT = pathlib.Path("/content/drive/MyDrive/whisper-bench-out")
        OUT.mkdir(parents=True, exist_ok=True)
        print(f"Drive mounted, outputs go to {OUT}", flush=True)
    except ImportError:
        OUT = WORK / "out"
        OUT.mkdir(parents=True, exist_ok=True)
        print(f"No Drive (not Colab?). Outputs go to {OUT}", flush=True)


def install_go():
    if shutil.which("go") and b"1.22" not in subprocess.run(
            ["go", "version"], capture_output=True).stdout:
        sh("rm -rf /usr/local/go")
    if not shutil.which("go"):
        os.environ["PATH"] = "/usr/local/go/bin:" + os.environ["PATH"]
        if not (pathlib.Path("/usr/local/go/bin/go").exists()):
            sh("wget -q https://go.dev/dl/go1.22.12.linux-amd64.tar.gz -O /tmp/go.tar.gz")
            sh("tar -C /usr/local -xzf /tmp/go.tar.gz")
    print(subprocess.run(["go", "version"], capture_output=True, text=True).stdout, flush=True)


def build_whisper():
    wdir = WORK / "whisper.cpp"
    has_nvcc = shutil.which("nvcc") is not None
    cuda_flag = "-DGGML_CUDA=1 -DCMAKE_CUDA_COMPILER=/usr/local/cuda/bin/nvcc" if has_nvcc else ""
    print(f"CUDA build: {'yes' if has_nvcc else 'no — CPU build'} (GPU runtime needs "
          "Runtime > Change runtime type > T4/GPU)", flush=True)
    if not (wdir / "build" / "bin" / "whisper-cli").exists():
        sh(f"rm -rf {wdir} && git clone --depth 1 --branch v1.9.2 "
           "https://github.com/ggml-org/whisper.cpp " + str(wdir))
        sh(f"cmake -B {wdir}/build -S {wdir} {cuda_flag} -DCMAKE_BUILD_TYPE=Release")
        sh(f"cmake --build {wdir}/build -j4 --config Release")


def fetch_models():
    mdir = WORK / "models"
    mdir.mkdir(exist_ok=True)
    src = mdir / "ggml-large-v3.bin"
    if not src.exists():
        sh(f"wget -q https://huggingface.co/ggerganov/whisper.cpp/resolve/main/"
           f"ggml-large-v3.bin -O {src}")
    q = WORK / "whisper.cpp" / "build" / "bin" / "whisper-quantize"
    for lvl in ("q8_0", "q5_0", "q4_0"):
        dst = mdir / f"ggml-large-v3-{lvl}.bin"
        if not dst.exists():
            sh(f"{q} {src} {dst} {lvl} > /dev/null 2>&1")


def fetch_dataset():
    data = WORK / "data" / "flat"
    if data.exists() and len(list(data.glob("*.flac"))) == 2620:
        return
    data.parent.mkdir(parents=True, exist_ok=True)
    tarball = WORK / "test-clean.tar.gz"
    if not tarball.exists():
        sh(f"wget -q https://openslr.trmal.net/resources/12/test-clean.tar.gz -O {tarball}")
    with tarfile.open(tarball) as tf:
        tf.extractall(WORK / "data")
    flatten(data)


def flatten(flat: pathlib.Path):
    """LibriSpeech packs transcripts per speaker dir; werpipe wants one .txt per .flac."""
    import glob
    src = WORK / "data" / "LibriSpeech" / "test-clean"
    flat.mkdir(parents=True, exist_ok=True)
    trans = {}
    for tf in glob.glob(str(src) + "/**/*.trans.txt", recursive=True):
        key = os.path.relpath(os.path.dirname(tf), str(src))
        with open(tf) as fh:
            for line in fh:
                parts = line.split(" ", 1)
                if len(parts) == 2:
                    trans[os.path.join(key, parts[0])] = parts[1].strip().lower()
    count = 0
    for audio in glob.glob(str(src) + "/**/*.flac", recursive=True):
        rel = os.path.relpath(audio, str(src))[:-5]
        if rel in trans:
            out = rel.replace("/", "_")
            shutil.copy2(audio, flat / f"{out}.flac")
            (flat / f"{out}.txt").write_text(trans[rel])
            count += 1
    print(f"flattened {count} samples into {flat}", flush=True)
    assert count == 2620, f"expected 2620, got {count}"


def build_werpipe():
    repo = WORK / "concept"
    if not (repo / "cmd" / "werpipe" / "main.go").exists():
        sh(f"git clone --depth 1 "
           "https://github.com/yoarajota/concept-whisper-quantization-wer-degradation "
           + str(repo))
    if not shutil.which("werpipe"):
        sh(f"cd {repo} && go build -o /usr/local/bin/werpipe ./cmd/werpipe/")


def run_chunks():
    env = dict(os.environ)
    env.update({
        "AUDIO_DIR": str(WORK / "data" / "flat"),
        "TRANS_DIR": str(WORK / "data" / "flat"),
        "MODEL_DIR": str(WORK / "models"),
        "WHISPER_CLI": str(WORK / "whisper.cpp" / "build" / "bin" / "whisper-cli"),
        "OUT_DIR": str(OUT),
        "PATH": "/usr/local/go/bin:" + os.environ["PATH"],
    })
    script = str(WORK / "concept" / "bench" / "chunked.sh")
    print("running chunked benchmark (resumable)", flush=True)
    r = subprocess.run(
        f"sh {script} 200 2 'f16,q8_0,q5_0,q4_0'",
        shell=True, env=env,
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
        text=True, bufsize=1,
    )
    for line in r.stdout.splitlines():
        print(line, flush=True)
    if r.returncode != 0:
        sys.exit(r.returncode)
    done = len(list(OUT.glob("chunk-*.json"))) - sum(
        1 for p in OUT.glob("chunk-*.json") if p.stat().st_size == 0)
    print(f"\n{done} chunks complete", flush=True)


def merge():
    chunks = sorted(str(p) for p in OUT.glob("chunk-*.json") if p.stat().st_size > 0)
    if len(chunks) < 2:
        print("not enough chunks to merge yet", flush=True)
        return
    r = subprocess.run(["werpipe", "merge"] + chunks, capture_output=True, text=True)
    (OUT / "final.json").write_text(r.stdout)
    print(r.stderr, flush=True)
    print(f"final.json written to {OUT / 'final.json'}", flush=True)


if __name__ == "__main__":
    gpu = shutil.which("nvidia-smi")
    if gpu:
        sh("nvidia-smi --query-gpu=name,memory.total --format=csv,noheader")
    else:
        print("no GPU detected — running on CPU (works, but ~50h total).", flush=True)
        print("For ~10h: Runtime > Change runtime type > T4 GPU, then re-run.", flush=True)
    WORK.mkdir(parents=True, exist_ok=True)
    drive_mount()
    install_go()
    build_whisper()
    fetch_models()
    fetch_dataset()
    build_werpipe()
    run_chunks()
    merge()
