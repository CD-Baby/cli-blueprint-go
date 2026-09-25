BINARY      := mycli
MODULE      := github.com/example/mycli
CLI_PKG     := $(MODULE)/internal/cli
GO          ?= go

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE  ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w \
	-X $(CLI_PKG).Version=$(VERSION) \
	-X $(CLI_PKG).Commit=$(COMMIT) \
	-X $(CLI_PKG).BuildDate=$(BUILD_DATE)

COVER_FILE  := coverage.out
COVER_MIN   := 85

.PHONY: build
build: ## build bin/$(BINARY) for the host
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY) ./cmd/$(BINARY)

.PHONY: test
test: ## run the unit and script tests
	$(GO) test ./...

.PHONY: race
race: ## run tests with the race detector
	$(GO) test -race ./...

.PHONY: cover
cover: ## run tests and enforce the coverage gate (library code only)
	$(GO) test -coverprofile=$(COVER_FILE) ./internal/...
	@total=$$($(GO) tool cover -func=$(COVER_FILE) | awk '/^total:/ {gsub(/%/,"",$$3); print $$3}'); \
	echo "total coverage: $$total% (min $(COVER_MIN)%)"; \
	awk "BEGIN{exit !($$total >= $(COVER_MIN))}" || { echo "coverage below gate"; exit 1; }

.PHONY: update-golden
update-golden: ## regenerate golden files (-update)
	$(GO) test ./... -update

.PHONY: vet
vet: ## go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## golangci-lint
	golangci-lint run

.PHONY: tidy
tidy: ## go mod tidy
	$(GO) mod tidy

.PHONY: verify
verify: vet lint test ## the local pre-push gate

# Cross-compile the release targets (static, CGO disabled).
.PHONY: cross
cross: cross-linux-amd64 cross-linux-arm64 cross-darwin-arm64 ## build all release targets

cross-%:
	@os=$$(echo $* | cut -d- -f1); arch=$$(echo $* | cut -d- -f2); \
	echo "building $$os/$$arch"; \
	CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build -trimpath -ldflags '$(LDFLAGS)' \
		-o bin/$(BINARY)-$$os-$$arch ./cmd/$(BINARY)

.PHONY: clean
clean: ## remove build and coverage artifacts
	rm -rf bin $(COVER_FILE)

.PHONY: help
help: ## list targets
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'
