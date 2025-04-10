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

BINARIES := coin gc2coin ofx2coin csv2coin gen2coin coin2html

build: $(BINARIES)

coin: *.go cmd/coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/coin

gc2coin: *.go cmd/gc2coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/gc2coin

ofx2coin: *.go cmd/ofx2coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/ofx2coin

csv2coin: *.go cmd/csv2coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/csv2coin

gen2coin: *.go cmd/gen2coin/*.go
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/gen2coin

coin2html: *.go cmd/coin2html/*.go cmd/coin2html/js/src/*.ts cmd/coin2html/js/*.html cmd/coin2html/js/*.css
	go generate ./cmd/coin2html
	$(BUILD) -ldflags '$(LDFLAGS)' ./cmd/coin2html

examples/yearly/viewer/index.html: export COINDB=./examples/yearly
examples/yearly/viewer/index.html: coin2html
	coin2html >$(COINDB)/viewer/index.html

examples: examples/yearly/viewer/index.html

dfa: dfa.bash
	cp ./dfa.bash $(GOPATH1)/bin/

test: test-go test-fixtures test-ts

cmd/coin2html/js/dist/body.html:
	go generate ./cmd/coin2html

test-go: cmd/coin2html/js/dist/body.html
	$(TEST) ./...

test-fixtures: export COIN_TESTS=./tests
test-fixtures:
	find tests -name '*.test' -print0 | xargs -0 -n1 $(shell go env GOPATH)/bin/coin test

test-ts:
	npm --prefix cmd/coin2html/js test

fmt:
	gofmt -s -l -w .

lint:
	golangci-lint run ./...

cover:
	$(TEST) -cover ./...

browse-coverage:
	$(TEST) -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

setup-ts:
	npm --prefix cmd/coin2html/js install

setup: setup-ts
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest


dist-linux-arm64: BUILD := GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/coin-linux-arm64/
dist-linux-arm64: build
	tar zcf dist/coin-linux-arm64.tgz -C dist/coin-linux-arm64/ .

dist-linux-amd64: BUILD := GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/coin-linux-amd64/
dist-linux-amd64: build
	tar zcf dist/coin-linux-amd64.tgz -C dist/coin-linux-amd64/ .

dist-osx-arm64: BUILD := GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o dist/coin-osx-arm64/
dist-osx-arm64: build
	tar zcf dist/coin-osx-arm64.tgz -C dist/coin-osx-arm64/ .

dist-osx-amd64: BUILD := GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/coin-osx-amd64/
dist-osx-amd64: build
	tar zcf dist/coin-osx-amd64.tgz -C dist/coin-osx-amd64/ .

dist-windows-amd64: BUILD := GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/coin-windows-amd64/
dist-windows-amd64: build
	tar zcf dist/coin-windows-amd64.tgz -C dist/coin-windows-amd64 .

dist: setup-ts
	$(MAKE) dist-linux-arm64
	$(MAKE) dist-linux-amd64
	$(MAKE) dist-osx-arm64
	$(MAKE) dist-osx-amd64
	$(MAKE) dist-windows-amd64

dist-clean:
	rm -rf dist/*

.PHONY: build coin test test-fixtures test-go fmt lint cover browse-coverage dist dist-clean
