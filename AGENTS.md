# qcloud-cli

## Project overview

Go CLI for [Qdrant Cloud](https://cloud.qdrant.io), built with Cobra / Viper and gRPC.

- **Module:** `github.com/qdrant/qcloud-cli`
- **Binary:** `qcloud` (built to `build/qcloud`)
- **Go version:** 1.26+
- **Key dependencies:** `cobra`, `viper`, `google.golang.org/grpc`, `qdrant-cloud-public-api` (generated gRPC stubs)

## Project structure

```
cmd/qcloud/              # main entrypoint — creates State, builds root command, runs it
internal/
  cli/                   # root cobra command, global flags, subcommand registration
  cmd/                   # one sub-package per top-level subcommand
    cluster/             # cluster.go (parent) + list/describe/create/delete
    version/             # version subcommand
    output/              # shared output formatting helpers
    util/                # shared command utilities
  qcloudapi/             # gRPC client wrapper for the Qdrant Cloud API
  state/                 # State struct (shared deps: config, lazy gRPC client)
    config/              # Viper-based config (file, env vars, flags)
```

## Build & verification

**Always use Makefile targets — never raw `go build`, `go test`, or linter commands.**

| Target           | What it does                                  |
|------------------|-----------------------------------------------|
| `make build`     | Compile binary to `build/qcloud`              |
| `make test`      | Run all tests                                 |
| `make lint`      | Run golangci-lint (installs it if missing)    |
| `make format`    | Run golangci-lint with `--fix`                |
| `make bootstrap` | Install tool dependencies into `bin/`         |
| `make clean`     | Remove build artifacts                        |

## Conventions

### Subcommand pattern

Each subcommand group lives in `internal/cmd/<group>/`:

1. A public `NewCommand(s *state.State) *cobra.Command` creates the parent and registers sub-commands.
2. Leaf commands are unexported (`newListCommand`, `newDeleteCommand`, …).
3. All commands receive `*state.State` — use it to access config and the lazy gRPC client.

### State passing

`main` → `state.New(version)` → passed to every command constructor. Commands call `s.Client(ctx)` to get the gRPC client (created on first use) and `s.AccountID()` for the current account.

