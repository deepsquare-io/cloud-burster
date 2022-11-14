.PHONY: all
all: build

.PHONY: build
build: cloud-burster

.PHONY: test
test: unit

.PHONY: cloud-burster
cloud-burster:
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/$@ ./cmd

.PHONY: unit
unit:
	go test -race -covermode=atomic -tags=unit -timeout=30s ./...

.PHONY: coverage
coverage:
	go test -race -covermode=atomic -tags=unit -timeout=30s -coverprofile=coverage.out ./...
	go tool cover -html coverage.out -o coverage.html

.PHONY: integration
integration:
	go test -race -covermode=atomic -tags=integration -timeout=300s ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: mocks
mocks:
	mockery --all
