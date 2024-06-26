.PHONY: build install test coverage lint clean nix-build

GOVERSION := $(shell go version)
COMMIT_HASH := $(shell git rev-parse HEAD)
VERSION := $(shell git describe --tags --abbrev=8 --dirty --always --long | sed 's/-0-g/-g/' | cut -c 2-)
DATE := $(shell git log -1 --format=%cd --date=format:'%Y-%m-%d' $(COMMIT_HASH))
PREFIX := main
LDFLAGS := -X '$(PREFIX).buildVersion=$(VERSION) $(DATE) $(GOVERSION)'

build: test
	go build -mod=vendor -ldflags "$(LDFLAGS)"

install: test
	go install -ldflags "$(LDFLAGS)"

test:
	golangci-lint run ./...
	go test -cover ./...

coverage:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out
	$(RM) -f cover.out

lint:
	golangci-lint run ./...

clean:
	$(RM) haproxytime haproxytime.test result

nix-build:
	nix build .#haproxytime && ./result/bin/haproxytime -v
