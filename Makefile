BUILT := $(shell date -u '+%Y-%m-%d %I:%M:%S')
COMMIT := $(shell git rev-parse HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GO_VERSION := $(shell go version)
GOPATH1 := $(shell echo $(GOPATH) | cut -f 1 -d:)

LDFLAGS += -X "github.com/mkobetic/coin.Built=$(BUILT)"
LDFLAGS += -X "github.com/mkobetic/coin.Commit=$(COMMIT)"
LDFLAGS += -X "github.com/mkobetic/coin.Branch=$(BRANCH)"
LDFLAGS += -X "github.com/mkobetic/coin.GoVersion=$(GO_VERSION)"

BUILD := go install 

build: coin gc2coin ofx2coin

coin: *.go cmd/coin/*.go
	go install -ldflags '$(LDFLAGS)' ./cmd/coin

gc2coin: *.go cmd/gc2coin/*.go
	go install -ldflags '$(LDFLAGS)' ./cmd/gc2coin

ofx2coin: *.go cmd/ofx2coin/*.go
	go install -ldflags '$(LDFLAGS)' ./cmd/ofx2coin

dfa: dfa.bash
	cp ./dfa.bash $(GOPATH1)/bin/

test: test-go test-fixtures

test-go:
	go test ./...

test-fixtures:
	find tests -name '*.test' -exec coin test '{}' \;

fmt:
	gofmt -s -l -w .

lint:
	golangci-lint run ./...

cover:
	go test -cover ./...

browse-coverage:
	go test -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

.PHONY: test test-fixtures test-go fmt lint cover browse-coverage