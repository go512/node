# Repository Guide

## Overview

This repository is primarily a Go playground and notes site. It contains two
independent Go modules:

- `server/`: the runnable application (`module node`)
- `notefile/Golang/`: standalone Go examples and tests

Run Go commands from the module directory they apply to. There is no root
`go.mod` or `go.work`.

## Main Application

The `server` module has two entry points:

- `server/main.go`: HTTP notes browser. It embeds `server/tpl` and
  `server/static`, reads content from `notefile`, and listens on port `1024`.
- `server/cmd/main.go`: CLI application with commands defined under
  `server/cmd/cli`.

Shared packages live in `server/pkg`, including channel/concurrency examples,
Kafka helpers, MySQL/GORM clients, logging, AI integration, and utilities.
Configuration is loaded from `server/config.toml` and `server/conf`.

## Common Commands

From `server/`:

```sh
go run .
go run ./cmd kafka_consumer -c ./config.toml
go run ./cmd log_cli -c ./config.toml
make build
go test ./pkg/chann ./pkg/utils
```

From `notefile/Golang/`:

```sh
go test ./...
```

Use `gofmt` on changed Go files. Prefer targeted tests for the packages being
edited before attempting a module-wide test.

In restricted environments, set `GOCACHE` to a writable temporary directory,
for example:

```sh
GOCACHE=/tmp/node-go-build go test ./pkg/chann
```

## Current Test Caveats

- `go test ./...` under `server/` currently does not compile because
  `pkg/mysqlPkg2/client.go` contains an unfinished `Open` method.
- `server/pkg/kafka_test.go` contains non-terminating demo tests.
- `server/pkg/ai/ai_test.go` may call an external service and require
  credentials/network access.
- `server/pkg/utils` currently panics in `TestIncludeFunc` when the test's
  callback calls `reflect.Value.Field` on a string value.

Do not report the full suite as passing unless these conditions have been
addressed. Avoid running integration/demo tests by default when a targeted unit
test is sufficient.

## Change Guidelines

- Preserve the separation between the two Go modules.
- Keep embedded asset paths compatible with the `//go:embed tpl` and
  `//go:embed static` directives.
- Treat Kafka, database, and AI code as integration code: do not assume local
  services or credentials are available.
- Do not commit secrets or real service credentials to TOML configuration.
- Many files are learning examples rather than production components; keep
  changes scoped to the requested area instead of broadly rewriting examples.
