# Go Template

A minimal Go project template with a complete build, format, lint, test,
coverage, and security toolchain. The code does nothing useful — it’s a
starting point for new projects.

The toolchain lives in
[go-canon](https://github.com/SynthLuvr/go-canon): one meta-tool owning
every gate and preset, in the same spirit as ts-canon (TypeScript) and
canonist (Python). Go modules make the bundling half unnecessary — the
`tool` directives in `go.mod` pin the whole toolchain, compiled locally
by the Go toolchain itself. This repo keeps only `go-cmp` as a direct
dependency (the canonical diff library for table-driven tests).

## Tech Stack

| Tool | Purpose |
|----|----|
| [Go modules](https://go.dev/ref/mod) | Package manager (`go.sum` is the lockfile) |
| [go-canon](https://github.com/SynthLuvr/go-canon) | The shared lint/format/test/doctor/migrate toolchain |
| [golangci-lint](https://golangci-lint.run) | Meta-linter: errcheck, govet, staticcheck, unused, gosec, dupl, … |
| [gofumpt](https://github.com/mvdan/gofumpt) + [gci](https://github.com/daixiang0/gci) | Formatting and import order (via golangci-lint) |
| [modernize](https://go.dev/blog/gopls-modernize) | Idiom codemods (range-over-int, slices/maps/min/max) |
| [govulncheck](https://go.dev/blog/vuln) | SCA: symbol-level vulnerability reachability |
| [stdlib `testing`](https://pkg.go.dev/testing) | Test runner (table-driven, race detector, fuzzing built in) |
| [go-cmp](https://github.com/google/go-cmp) | Semantic diffs for test assertions |
| [Task](https://taskfile.dev) | Task runner for the one-command UX |
| [pandoc](https://pandoc.org) | Markdown formatter (GFM); system dependency |

Everything above except pandoc is pinned in the `go.mod` `tool` block
and `go.sum`, and compiled locally by `go tool`.

## Prerequisites

- [Go](https://go.dev) 1.27+ (`.go-version` pins 1.27.1 for
  asdf/gvm/mise; `GOTOOLCHAIN=auto` upgrades automatically)
- [pandoc](https://pandoc.org) ≥ 3.10 — required by the markdown gate;
  `task doctor` verifies it

The toolchain runs on Linux, macOS, and Windows; line endings are
normalized to LF via `.gitattributes`, and pandoc runs with `--eol=lf`
so its output stays LF on Windows too.

### Windows notes

No tool is ever a downloaded prebuilt binary: `go tool` compiles every
tool (task, go-canon, golangci-lint, govulncheck, modernize) from the
pinned module sources into the local build cache, and go-canon spawns
each one by absolute path — never through a shell, a `.cmd`/`.ps1` shim,
or a bare PATH binary. This structurally avoids the AppLocker
low-prevalence-`.exe` problem that shaped the ts-canon and canonist
Windows stories. Pandoc is the single system dependency.

## Quick Start

``` bash
go mod download
task build   # compile
task test    # tests + coverage gate
```

## Tasks

### Build

| Task         | Description          |
|--------------|----------------------|
| `task build` | Compile all packages |

### Lint

`task lint` runs `go-canon lint`, every check in order, failing fast on
the first non-zero step:

1.  `go build ./...` — the type check (Go has no separate type-check
    step; the compiler is it)
2.  golangci-lint — errcheck/govet/staticcheck/unused, revive,
    errorlint/wrapcheck, intrange/perfsprint/copyloopvar, gosec, dupl,
    and gofumpt/gci format *findings*, from the merged effective config
3.  modernize — idiom gate (non-zero on findings)
4.  `go mod tidy -diff` — module/lockfile freshness
5.  govulncheck — SCA, symbol-level (skipped with `--fast`)
6.  pandoc — markdown must be GFM-formatted (skipped with `--fast`)

`--fast` also disables the dupl gate inside golangci-lint.

### Format

`task format` runs `go-canon format`, every formatter in order (each
step’s output is the next step’s input):

1.  modernize `-fix` — idiom codemods
2.  golangci-lint fmt — gofumpt + gci
3.  `go mod tidy`
4.  pandoc — normalize markdown to canonical GFM

### Test

| Task            | Description                                |
|-----------------|--------------------------------------------|
| `task test`     | Tests + coverage (80% statement threshold) |
| `go test ./...` | Direct invocation, watch with `-run` etc.  |

`task test` runs `go test ./... -race -covermode=atomic -coverpkg=./...`
and enforces an 80% total statement coverage threshold. Tests are
excluded from measurement by construction (Go’s coverage measures
non-test statements). For a watch loop, use
`go test ./... -run TestName` or `gtester`-style tooling; a dedicated
`test:watch` task is intentionally absent.

### Doctor

`task doctor` verifies the environment: go version, each pinned tool
(`go tool -n`), version drift against go-canon’s known-good window,
pandoc ≥ 3.10, and OSV database reachability (advisory).

## Coding Conventions

These are **enforced** by the toolchain, not just preferences:

- **gofumpt formatting** — there is no style debate in Go
- **Import grouping** — std / external / local module (gci)
- **Modern idioms** — range-over-int, `slices`/`maps`/`min`/`max`, no
  `fmt.Sprintf` where `strconv` suffices (modernize + intrange +
  perfsprint)
- **Errors are values** — checked (errcheck), wrapped at package
  boundaries (wrapcheck), compared with `errors.Is/As` (errorlint)
- **Exported symbols are documented** (revive) — the docstring
  discipline analog
- **Table-driven tests**, `t.Parallel()` where possible
- **Markdown via pandoc** — all `.md` byte-identical to
  `pandoc --eol=lf -t gfm` output

## Project Structure

    ├── .github/workflows/     # CI (Ubuntu + Windows)
    ├── index.go               # Trivial module (replace with your code)
    ├── index_test.go          # Co-located table-driven test
    ├── go-canon.toml          # Local deltas over go-canon's presets
    ├── go.mod / go.sum        # Module + pinned tool block
    ├── Taskfile.yml           # build / lint / format / test / check / doctor
    └── .go-version            # Go version for asdf/gvm/mise

Add private packages under `internal/` as the project grows.

## Deliberate deviations from the ts/python templates

These layout differences are **by design** — Go idiom wins over
template-family symmetry. Don’t “fix” them:

1.  **No `src/` directory.** Go modules put packages at the module root
    and in subdirectories; `src/` is GOPATH-era baggage that fights the
    toolchain.
2.  **Tests are co-located** (`index_test.go` next to `index.go`), not
    in a `tests/` tree. `go test` discovers `_test.go` files by filename
    convention; co-location enables white-box tests and table-driven
    style, and `go test -cover` already excludes test files from
    measurement — a property both sibling templates configure manually.
3.  **No separate type-check command.** `go build` (compiler) plus
    govet/staticcheck via golangci-lint is the `tsc`/`pyright`
    equivalent; a dedicated typecheck step would be theater.
4.  **No style arguments.** gofumpt ended the formatter wars decades
    ago.

## Config

Rule/preset changes happen in
[go-canon](https://github.com/SynthLuvr/go-canon), not here: bump the
go-canon version in the `tool` block and the whole toolchain moves
together. Repo-specific deltas go in `go-canon.toml` (`[golangci]` deep
merges over the preset; `[dupl] tokens` and `[test] coverage-threshold`
are first-class knobs).
