---
name: coin2html-bundle
description: >-
  Use this skill when syncing changes between the JS/TS viewer subproject
  (cmd/coin2html/js) and the Go coin2html binary, updating the embedded HTML
  bundle (dist/head.html and dist/body.html), changing the JSON data contract,
  or regenerating the examples/yearly/viewer/index.html demo output.
---

# Integrating `cmd/coin2html/js` with Go `coin2html`

The `cmd/coin2html/js` subproject and the Go project are developed independently, but join together when building the `coin2html` binary or regenerating the example viewer.

## Workflow Steps

1. **Verify the JS/TS subproject in isolation**:
   ```sh
   npm --prefix cmd/coin2html/js test
   npm --prefix cmd/coin2html/js run lint
   ```

2. **Bundle the HTML fragments (`dist/head.html` and `dist/body.html`)**:
   ```sh
   npm --prefix cmd/coin2html/js run build
   ```
   *(Alternatively, running `go generate ./cmd/coin2html` from the repository root invokes the same build script.)*

3. **Verify Go compilation and tests (which embed `cmd/coin2html/js/dist/*`)**:
   ```sh
   CGO_ENABLED=0 go test ./...
   ```

4. **Rebuild the `coin2html` binary and regenerate the example viewer (if requested)**:
   ```sh
   make coin2html
   make examples
   ```
   This runs `coin2html` against `COINDB=./examples/yearly` and writes the standalone HTML file to `examples/yearly/viewer/index.html`.

