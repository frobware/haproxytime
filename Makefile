DATE		:= $(shell date --iso-8601=seconds)
GOVERSION	:= $(shell go version)
COMMIT		:= $(shell git describe --tags --abbrev=8 --dirty --always --long)

PREFIX		:= main
LDFLAGS		:= -X '$(PREFIX).buildVersion=$(COMMIT) ($(DATE)) $(GOVERSION)'

build: test lint
	go build -ldflags "$(LDFLAGS)" -o ./haproxytimeout ./cmd/haproxytimeout

test:
	go test ./...

lint:
	golangci-lint run ./...

benchmark:
	go test -bench=. -benchmem -count=5 -benchtime=1s

clean:
	$(RM) ./haproxytimeout ./haproxytimeout.test

.PHONY: build test clean benchmark lint
