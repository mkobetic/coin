# Coin — Agent Guidelines

`coin` is a plain-text accounting (PTA) system inspired by `ledger-cli`, designed around multi-file ledgers (`$COINDB`) and strict upfront commodity and account declarations.

## Dual-Project Architecture

This repository contains two distinct projects that are developed somewhat independently and coupled only through a narrow build and JSON data boundary:

1. **Main Go Project (Repository Root & `cmd/*`)**
   - Module `github.com/mkobetic/coin`: core accounting domain, parser, and ledger engine at the repository root, plus internal packages (`assert/`, `check/`, `gnucash/`, `rex/`).
   - CLI binaries under `cmd/`:
     - `cmd/coin`: primary CLI (`balance`, `register`, `accounts`, `commodities`, `stats`, `format`, `test`, etc.)
     - `cmd/gc2coin`: GnuCash XML v2 importer
     - `cmd/ofx2coin`: OFX/QFX importer
     - `cmd/csv2coin`: CSV importer
     - `cmd/gen2coin`: sample ledger generator
     - `cmd/coin2html`: single-file HTML viewer generator (Go wrapper that serializes ledger data as JSON and embeds the built JS viewer)
2. **JS/TS Viewer Subproject (`cmd/coin2html/js/`)**
   - Standalone TypeScript + D3.js + Webpack single-page application that builds `dist/head.html` and `dist/body.html`.
   - Has its own dependencies (`package.json`), test suite (`jest`), linter (`prettier` + `eslint`), and local agent rules in [`cmd/coin2html/js/AGENTS.md`](cmd/coin2html/js/AGENTS.md).

### Boundary & Independent Development Guidelines

- **Go-only changes**: Work directly in the root package and `cmd/*`. Do not modify or rebuild `cmd/coin2html/js/` unless the JSON data contract emitted by `cmd/coin2html/main.go` changes. Note that `go test ./...` requires `cmd/coin2html/js/dist/body.html` and `head.html` to exist for `//go:embed` in `cmd/coin2html`.
- **JS/TS-only changes**: Work inside `cmd/coin2html/js/` following [`cmd/coin2html/js/AGENTS.md`](cmd/coin2html/js/AGENTS.md). Run Jest tests and linters directly via `npm --prefix cmd/coin2html/js` without rebuilding Go binaries during iteration.
- **Cross-boundary / Integration changes**:
  - The contract between `cmd/coin2html/main.go` and `cmd/coin2html/js/` consists of:
    1. Four `<script type="application/json">` elements (`#importedCommodities`, `#importedPrices`, `#importedAccounts`, `#importedTransactions`) serialized by Go and parsed by `cmd/coin2html/js/src/ui.ts`.
    2. Two bundled HTML fragments (`cmd/coin2html/js/dist/head.html` and `cmd/coin2html/js/dist/body.html`) produced by `npm --prefix cmd/coin2html/js run build` (or `go generate ./cmd/coin2html`) and embedded via `//go:embed` in `cmd/coin2html/main.go`.
  - When a task spans both Go and JS/TS, keep changes modular or delegate the `cmd/coin2html/js` work to a subagent scoped to `cmd/coin2html/js/`.

---

## Go Project Conventions & Domain Invariants

- **Minimal dependencies & static builds**: All Go builds and tests run with `CGO_ENABLED=0`. Avoid introducing new external Go dependencies unless explicitly requested.
- **Fixed-point `Amount` arithmetic (`amount.go`)**:
  - `Amount` is implemented as `big.Int` paired with a `Commodity`.
  - Precision (decimal places) is strictly dictated by the `Commodity` (`commodity.Decimals`), never inferred from amount literals.
  - Computations truncate to the commodity's decimal places at every step.
- **Strict upfront definitions (`coin.go`, `account.go`, `commodity.go`)**:
  - Accounts and commodities are never inferred; `accounts.coin` and `commodities.coin` are loaded first from `$COINDB`.
  - Accounts are single-commodity.
  - Commodity names have strict character rules (no whitespace, no prefix symbols like `$10`) so quoting is never needed.
- **Source locations (`Location()`)**:
  - All parsed ledger entities (`Account`, `Commodity`, `Transaction`, `Posting`, `Price`) record their `file:line` origin via `Location()` to support editor navigation and diagnostics. Preserve location tracking when modifying parsers or AST structs.

## Go Commands (from Repository Root)

- **Run Go unit tests**:
  ```sh
  make test-go
  # or directly (once cmd/coin2html/js/dist/* exists):
  CGO_ENABLED=0 go test ./...
  ```
- **Run CLI fixture tests** (`tests/*.test`):
  ```sh
  make coin && make test-fixtures
  ```
- **Build all binaries**:
  ```sh
  make build
  ```
- **Format & lint Go code**:
  ```sh
  make fmt
  make lint
  ```

