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
