.PHONY: quality test tools build docker-build docker-run docker-models docker-up docker-down bench-chunk bench-probe overnight

quality:
	go vet ./...
	staticcheck ./...
	gocyclo -over 15 src/
	golangci-lint run
	gitleaks detect --no-banner

test:
	go test ./...

tools:
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

build:
	go build -o bin/werpipe ./cmd/werpipe/

docker-build:
	docker build -f docker/Dockerfile -t concept-whisper-wer .

docker-models:
	docker run --rm -v $(MODELS_DIR):/models --entrypoint sh concept-whisper-wer \
	  -c "MODEL_DIR=/models WHISPER_DIR=/whisper.cpp sh /models/download.sh"

docker-up:
	docker compose -f docker/compose.yaml up --build

docker-down:
	docker compose -f docker/compose.yaml down -v

# Chunked benchmark at low priority — won't disturb other work.
# Set LIBRISPEECH= path to the flat dataset dir (flac+txt pairs).
# CHUNK_SIZE=200 CPUS=2 by default; override as needed.
bench-chunk:
	docker run --rm --cpus="$(CPUS)" \
	  -v $(MODELS_DIR):/models \
	  -v $(LIBRISPEECH):/data:ro \
	  -v $(OUT_DIR):/out \
	  --entrypoint sh concept-whisper-wer \
	  bench/chunked.sh $(CHUNK_SIZE) $(CPUS) "$(LEVELS)"

# Quick probe: 100 samples, f16 + q4_0 only, ~70 min at 2 CPUs.
bench-probe:
	docker run --rm --cpus="2" \
	  -v $(MODELS_DIR):/models \
	  -v $(LIBRISPEECH):/data:ro \
	  --entrypoint werpipe concept-whisper-wer \
	  -audio /data -transcripts /data -model-dir /models \
	  -whisper-cli /whisper.cpp/build/bin/whisper-cli -threads 2 \
	  -limit 100 -levels "ggml-large-v3.bin,ggml-large-v3-q4_0.bin"

# Detached resumable full-dataset run (see bench/overnight.sh for details).
overnight:
	sh bench/overnight.sh

bin/werpipe: cmd/werpipe/main.go src/werpipe/*.go
	go build -o bin/werpipe ./cmd/werpipe/
