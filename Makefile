include version.mk

ARCH := $(shell arch)

ifeq ($(ARCH),x86_64)
	ARCH = amd64
else
	ifeq ($(ARCH),aarch64)
		ARCH = arm64
	endif
endif
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/target
GOENVVARS := GOBIN=$(GOBIN) CGO_ENABLED=1 GOARCH=$(ARCH)
GOBINARY := cdk-node
GOCMD := $(GOBASE)/cmd

LDFLAGS += -X 'github.com/0xPolygon/cdk.Version=$(VERSION)'
LDFLAGS += -X 'github.com/0xPolygon/cdk.GitRev=$(GITREV)'
LDFLAGS += -X 'github.com/0xPolygon/cdk.GitBranch=$(GITBRANCH)'
LDFLAGS += -X 'github.com/0xPolygon/cdk.BuildDate=$(DATE)'

# Check dependencies
# Check for Go
.PHONY: check-go
check-go:
	@which go > /dev/null || (echo "Error: Go is not installed" && exit 1)

# Check for Docker
.PHONY: check-docker
check-docker:
	@which docker > /dev/null || (echo "Error: docker is not installed" && exit 1)

# Check for Docker-compose
.PHONY: check-docker-compose
check-docker-compose:
	@which docker-compose > /dev/null || (echo "Error: docker-compose is not installed" && exit 1)

# Check for Protoc
.PHONY: check-protoc
check-protoc:
	@which protoc > /dev/null || (echo "Error: Protoc is not installed" && exit 1)

# Check for Curl
.PHONY: check-curl
check-curl:
	@which curl > /dev/null || (echo "Error: curl is not installed" && exit 1)

# Targets that require the checks
build: check-go
lint: check-go
build-docker: check-docker
build-docker-nc: check-docker
stop: check-docker check-docker-compose
install-linter: check-go check-curl
generate-code-from-proto: check-protoc

.PHONY: build
build: build-rust build-go ## Builds the binaries locally into ./target

.PHONY: build-rust
build-rust:
	export BUILD_SCRIPT_DISABLED=1 && cargo build --release

.PHONY: build-go
build-go:
	$(GOENVVARS) go build -ldflags "all=$(LDFLAGS)" -o $(GOBIN)/$(GOBINARY) $(GOCMD)

.PHONY: build-docker
build-docker: ## Builds a docker image with the cdk binary
	docker build -t cdk -f ./Dockerfile .

.PHONY: build-docker-nc
build-docker-nc: ## Builds a docker image with the cdk binary - but without build cache
	docker build --no-cache=true -t cdk -f ./Dockerfile .

.PHONY: stop
stop: ## Stops all services
	docker-compose down

.PHONY: test-unit
test-unit:
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -count=1 -short -race -p 1 -covermode=atomic -coverprofile=coverage.out  -coverpkg ./... -timeout 200s ./...

.PHONY: test-seq_sender
test-seq_sender:
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -count=1 -short -race -p 1  -covermode=atomic -coverprofile=../coverage.out   -timeout 200s ./sequencesender/...

.PHONY: lint
lint: ## Runs the linter
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/golangci-lint run --timeout 5m

.PHONY: generate-code-from-proto
generate-code-from-proto: ## Generates code from proto files
	cd proto/src/proto/aggregator/v1 && protoc --proto_path=. --proto_path=../../../../include --go_out=../../../../../aggregator/prover --go-grpc_out=../../../../../aggregator/prover --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative aggregator.proto
	cd proto/src/proto/datastream/v1 && protoc --proto_path=. --proto_path=../../../../include --go_out=../../../../../state/datastream --go-grpc_out=../../../../../state/datastream --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative datastream.proto


## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sort \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
