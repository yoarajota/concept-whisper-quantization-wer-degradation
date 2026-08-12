.PHONY: quality test

quality:
	go vet ./...
	staticcheck ./...
	gocyclo -over 15 .
	golangci-lint run
	gitleaks detect --no-banner

test:
	go test ./...
