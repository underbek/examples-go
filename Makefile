.PHONY: enum
enum:
	go install github.com/abice/go-enum@latest
	go-enum -f ./errors/types.go

.PHONY: lint
lint:
##	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.0
	golangci-lint run -c .golangci.yml

.PHONY: test
test: ## run unit-tests
	go test ./... -buildvcs=false
