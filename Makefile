BUILD_VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "N/A")
BUILD_DATE=$(shell date -u +%Y-%m-%d_%H:%M:%S || echo "N/A")
BUILD_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "N/A")

.PHONY: all
all: build

build:
	go build -o ./cmd/shortener ./cmd/shortener

test:
	go clean -testcache
	go test -count 1 -v ./...

profiles_dir = profiles

bench-mem:
	go test ./internal/handler -run="^$$" -bench="^BenchmarkMainHandler$$" -benchmem \
		-benchtime=30s -memprofile=$(profiles_dir)/new.pprof


pprof-mem:
	go tool pprof -http=:8080 $(profiles_dir)/new.pprof

build-flags:
	go build -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)" -o ./cmd/shortener ./cmd/shortener

style:
	gofmt -s -w .
	goimports -l -w .