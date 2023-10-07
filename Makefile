.PHONY: build test test-html lint benchmark benchmark-profile clean nix-build

build: test lint
	go build -o haproxytimeout ./cmd/haproxytimeout

test:
	go test -cover ./...

test-html:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out
	$(RM) -f cover.out

lint:
	golangci-lint run ./...

benchmark:
	go test -bench=. -benchmem -count=1 -benchtime=1s

benchmark-profile:
	BENCHMARK_PROFILE_PORT=6060 go test -bench=. -benchmem -count=1 -benchtime=1s -cpuprofile=cpu.pprof
	go tool pprof cpu.pprof <<< "list github.com/frobware/haproxytime"

clean:
	$(RM) haproxytimeout haproxytimeout.test result

nix-build:
	nix build .#haproxytimeout && ./result/bin/haproxytimeout -v
