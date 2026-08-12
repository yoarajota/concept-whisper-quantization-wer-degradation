FROM golang:1.22-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download
COPY src/ src/
COPY cmd/ cmd/
RUN CGO_ENABLED=0 go build -o /build/werpipe ./cmd/werpipe/

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake build-essential git wget ca-certificates curl \
    python3 python3-pip python3-venv \
    libsdl2-dev \
    && rm -rf /var/lib/apt/lists/*

ENV WHISPER_VERSION=v1.9.2

RUN git clone --depth 1 --branch ${WHISPER_VERSION} \
    https://github.com/ggml-org/whisper.cpp /whisper.cpp

WORKDIR /whisper.cpp
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release \
    && cmake --build build -j$(nproc) --config Release
ENV PATH="/whisper.cpp/build/bin:${PATH}"

RUN mkdir -p /models /data
WORKDIR /home/wer

COPY --from=builder /build/werpipe /usr/local/bin/werpipe

ENTRYPOINT ["werpipe"]
