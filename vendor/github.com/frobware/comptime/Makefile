GOVERSION := $(shell go version)
COMMIT_HASH := $(shell git rev-parse HEAD)
VERSION := $(shell git describe --tags --abbrev=8 --dirty --always --long | sed 's/-0-g/-g/' | cut -c 2-)
DATE := $(shell git log -1 --format=%cd --date=format:'%Y-%m-%d' $(COMMIT_HASH))
PREFIX := main
LDFLAGS := -X '$(PREFIX).buildVersion=$(VERSION) $(DATE) $(GOVERSION)'

test:
	golangci-lint run ./...
	go test -cover ./...

coverage:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out
	$(RM) -f cover.out

lint:
	golangci-lint run ./...

benchmark:
	go test -bench=. -benchmem -count=1 -benchtime=1s

benchmark-profile:
	BENCHMARK_PROFILE_PORT=6060 go test -bench=. -benchmem -count=1 -benchtime=1s -cpuprofile=cpu.pprof
	go tool pprof cpu.pprof <<< "list github.com/frobware/comptime"

clean:
	$(RM) comptimeout comptimeout.test result

nix-build:
	nix build .#comptimeout && ./result/bin/comptimeout -v

.PHONY: test coverage lint benchmark benchmark-profile clean nix-build

