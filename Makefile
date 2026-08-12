.PHONY: quality test tools build docker-build docker-run docker-models docker-up docker-down

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
	MODEL_DIR=./data/models docker compose -f docker/compose.yaml run --rm whisper \
		sh -c "apk add curl bash && /usr/local/bin/download-models"

docker-up:
	docker compose -f docker/compose.yaml up --build

docker-down:
	docker compose -f docker/compose.yaml down -v

bin/werpipe: cmd/werpipe/main.go src/werpipe/*.go
	go build -o bin/werpipe ./cmd/werpipe/
