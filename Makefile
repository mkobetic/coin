build: coin gc2coin ofx2coin

coin: *.go cmd/coin/*.go
	go build -o ${HOME}/bin/coin ./cmd/coin

gc2coin: *.go cmd/gc2coin/*.go
	go build -o ${HOME}/bin/gc2coin ./cmd/gc2coin

ofx2coin: *.go cmd/ofx2coin/*.go
	go build -o ${HOME}/bin/ofx2coin ./cmd/ofx2coin

dfa: dfa.bash
	cp ./dfa.bash ${HOME}/bin/dfa.bash

test-all: test test-fixtures

test:
	go test ./...

test-fixtures:
	find tests -name '*.test' -exec coin test '{}' \;

.PHONY: test test-fixtures test-all