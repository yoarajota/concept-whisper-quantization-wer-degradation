.PHONY: quality test tools

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
