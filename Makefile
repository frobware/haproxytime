DATE		:= $(shell date --iso-8601=seconds)
GOVERSION	:= $(shell go version)
COMMIT		:= $(shell git describe --tags --match 'v*' --abbrev=0 | cut -c 2-)
PREFIX		:= main
LDFLAGS		:= -X '$(PREFIX).buildVersion=$(COMMIT) ($(DATE)) $(GOVERSION)'

build: test lint
	go build -ldflags "$(LDFLAGS)" -o ./haproxytimeout ./cmd/haproxytimeout

test:
	@t="/tmp/go-test-coverage.$$.tmp" && \
	    go test -coverprofile=$$t ./... && \
	    rm -f "$$t"

lint:
	golangci-lint run ./...

benchmark:
	go test -bench=. -benchmem -count=1 -benchtime=1s

benchmark-profile:
	BENCHMARK_PROFILE_PORT=6060 go test -bench=. -benchmem -count=1 -benchtime=1s -cpuprofile=cpu.pprof
	go tool pprof cpu.pprof <<< "list consumeNumber"
	go tool pprof cpu.pprof <<< "list consumeUnit"
	go tool pprof cpu.pprof <<< "list ParseDuration"

clean:
	$(RM) ./haproxytimeout ./haproxytimeout.test

nix-build:
	nix build .#haproxytimeout

new-release:
	echo "{ version = \"$(COMMIT)\"; }" > version.nix

.PHONY: build test clean benchmark lint nix-build version.nix
