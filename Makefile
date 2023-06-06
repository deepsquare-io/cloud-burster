GO_SRCS := $(shell find . -type f -name '*.go' -a ! \( -name 'zz_generated*' -o -name '*_test.go' \))
GO_TESTS := $(shell find . -type f -name '*_test.go')
TAG_NAME = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
TAG_NAME_DEV = $(shell git describe --tags --abbrev=0 2>/dev/null)
VERSION_CORE = $(shell echo $(TAG_NAME) | sed 's/^\(v[0-9]\+\.[0-9]\+\.[0-9]\+\)\(+\([0-9]\+\)\)\?$$/\1/')
VERSION_CORE_DEV = $(shell echo $(TAG_NAME_DEV) | sed 's/^\(v[0-9]\+\.[0-9]\+\.[0-9]\+\)\(+\([0-9]\+\)\)\?$$/\1/')
GIT_COMMIT = $(shell git rev-parse --short=7 HEAD)
VERSION = $(or $(and $(TAG_NAME),$(VERSION_CORE)),$(and $(TAG_NAME_DEV),$(VERSION_CORE_DEV)-dev),$(GIT_COMMIT))

bin/cloud-burster: $(GO_SRCS) set-version
	go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-darwin-amd64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-darwin-arm64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-freebsd-amd64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-freebsd-arm64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-linux-amd64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-linux-arm64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-linux-mips64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-linux-mips64le: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64le go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-linux-ppc64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-linux-ppc64le: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-linux-riscv64: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=linux GOARCH=riscv64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-linux-s390x: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=linux GOARCH=s390x go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bin/cloud-burster-windows-amd64.exe: $(GO_SRCS) set-version
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$@" ./cmd/main.go

bins := cloud-burster-darwin-amd64 cloud-burster-darwin-arm64 cloud-burster-freebsd-arm64 cloud-burster-freebsd-arm64 cloud-burster-linux-amd64 cloud-burster-linux-arm64 cloud-burster-linux-mips64 cloud-burster-linux-mips64le cloud-burster-linux-ppc64 cloud-burster-linux-ppc64le cloud-burster-linux-riscv64 cloud-burster-linux-s390x cloud-burster-windows-amd64.exe

bin/checksums.txt: $(addprefix bin/,$(bins))
	sha256sum -b $(addprefix bin/,$(bins)) | sed 's/bin\///' > $@

bin/checksums.md: bin/checksums.txt
	@echo "### SHA256 Checksums" > $@
	@echo >> $@
	@echo "\`\`\`" >> $@
	@cat $< >> $@
	@echo "\`\`\`" >> $@

.PHONY:
set-version:
	@sed -Ei 's/Version:(\s+)".*",/Version:\1"$(VERSION)",/g' cmd/main.go

.PHONY: build-all
build-all: $(addprefix bin/,$(bins)) bin/checksums.md

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

$(golint):
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: lint
lint: $(golint)
	$(golint) run ./...

.PHONY: mocks
mocks:
	mockery --all

.PHONY: clean
clean:
	rm -rf bin/
