.PHONY: test lint fuzz bench profile testdata

PATTERN ?= .

GO = GOAMD64=v4 GOEXPERIMENT=simd go

GOTEST = $(GO) test -run ^$$

test: FLAGS ?= -count 1
test:
	GOEXPERIMENT="" GOAMD64="" go test ./... $(FLAGS)
	GOEXPERIMENT=simd GOAMD64="" go test ./... $(FLAGS)
	GOEXPERIMENT=simd GOAMD64=v3 go test ./... $(FLAGS)
	GOEXPERIMENT=simd GOAMD64=v4 go test ./... $(FLAGS)

lint:
	golangci-lint run ./...

fuzz:
	$(GOTEST) -fuzz $(PATTERN) -fuzztime 30m

bench:
	$(GOTEST) -bench $(PATTERN) -benchtime 200ms

profile:
	$(GOTEST) -bench $(PATTERN) -benchtime 200ms -cpuprofile cpu.prof -memprofile mem.prof

benchmarks/index.bench:
	mkdir -p $(@D)
	$(GOTEST) -bench Index -count 6 -benchtime 180ms > $@

benchmarks/index.stats: benchmarks/index.bench
	benchstat -col /impl benchmarks/index.bench > $@

benchmarks/index.latency:
	mkdir -p $(@D)
	$(GO) test -tags latency -run IndexLatency -v > $@

testdata: testdata/enwik8 testdata/github testdata/source

testdata/enwik8:
	mkdir -p $(@D)
	wget -qO - http://mattmahoney.net/dc/enwik8.zip | zcat > $@

testdata/github:
	mkdir -p $(@D)
	wget -qO - https://data.gharchive.org/2015-01-01-15.json.gz | zcat > $@

testdata/source:
	mkdir -p $(@D)
	find $(shell go env GOROOT)/src -name "*.go" -type f -exec cat {} + > $@
