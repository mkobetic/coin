BUILT := $(shell date -u '+%Y-%m-%d %I:%M:%S')
COMMIT := $(shell git rev-parse HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GO_VERSION := $(shell go version)
GOPATH1 := $(shell echo $(GOPATH) | cut -f 1 -d:)

LDFLAGS += -X "github.com/mkobetic/coin.Built=$(BUILT)"
LDFLAGS += -X "github.com/mkobetic/coin.Commit=$(COMMIT)"
LDFLAGS += -X "github.com/mkobetic/coin.Branch=$(BRANCH)"
LDFLAGS += -X "github.com/mkobetic/coin.GoVersion=$(GO_VERSION)"

BUILD := CGO_ENABLED=0 go install
TEST := CGO_ENABLED=0 go test

build: coin gc2coin ofx2coin csv2coin gen2coin

cmd/coin/charts.go: cmd/coin/charts/*.js cmd/coin/charts/*.css
	go generate ./cmd/coin

coin: *.go cmd/coin/*.go cmd/coin/charts.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/coin

gc2coin: *.go cmd/gc2coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/gc2coin

ofx2coin: *.go cmd/ofx2coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/ofx2coin

csv2coin: *.go cmd/csv2coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/csv2coin

gen2coin: *.go cmd/gen2coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/gen2coin

dfa: dfa.bash
	cp ./dfa.bash $(GOPATH1)/bin/

test: test-go test-fixtures

test-go:
	$(TEST) ./...

test-fixtures:
	find tests -name '*.test' -exec coin test '{}' \;

fmt:
	gofmt -s -l -w .

lint:
	golangci-lint run ./...

cover:
	$(TEST) -cover ./...

browse-coverage:
	$(TEST) -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

.PHONY: test test-fixtures test-go fmt lint cover browse-coverage