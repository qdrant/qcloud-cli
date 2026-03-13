# Contributing to the qcloud cli

This CLI wraps the [qdrant/qdrant-cloud-public-api](https://github.com/qdrant/qdrant-cloud-public-api) gRPC service. Familiarity with the API is recommended before contributing. Understanding the available services, request/response types, and field semantics makes it easier to implement or review commands correctly. The CLI is not a 1-to-1 mapping of the API, but knowledge of it helps in understanding the commands and flags.


## Bootstrapping

Install [mise](https://mise.jdx.dev/installing-mise.html). With it the CLI tools for the project are kept in sync with all contributors.

Run 
```bash
mise install
```

All necessary cli tools should be now installed and ready to use.

## Testing

The project is covered using mostly integration tests with a fake gRPC server. To run them, execute:

```bash
make test
```

## Building

The cli binary is built in the `./build` directory, this is mostly for building locally.

```bash
make build
```

## Releasing

The project uses [releaser-pleaser](https://apricote.github.io/releaser-pleaser/introduction.html) to create new releases.

It uses conventional commits to classify changes 


## Conventions

### Command naming

- Command nouns must be in singular (e.g. `cluster` instead of `clusters`).
- Most action commands should be verbs (e.g. `list`, `describe`, `create`).
- Multi-word command names must be **kebab-case** (e.g. `cloud-provider`, `cloud-region`). Never use camelCase or snake_case.

### Subcommand structure

Each subcommand group lives in `internal/cmd/<group>/`:

1. A public `NewCommand(s *state.State) *cobra.Command` creates the parent and registers sub-commands.
2. Leaf commands are unexported (`newListCommand`, `newDeleteCommand`, …).
3. All commands receive `*state.State` — use it to access config and the lazy gRPC client.

### Base command types

All leaf commands **must** be built using one of the five generic base types in `internal/cmd/base`. Never construct a raw `cobra.Command` for a leaf. Pick the type that matches the operation:

| Base type             | When to use                                                                    |
|-----------------------|--------------------------------------------------------------------------------|
| `base.ListCmd[T]`     | Listing a collection of resources                                              |
| `base.DescribeCmd[T]` | Fetching and displaying a single resource                                      |
| `base.CreateCmd[T]`   | Creating a resource                                                            |
| `base.UpdateCmd[T]`   | Updating an existing resource                                                  |
| `base.Cmd`            | Imperative/action commands that don't return a resource (delete, wait, use, …) |

Key rules that apply to all base types:

- JSON output is handled automatically for all base types except `base.Cmd`. Do not call `output.PrintJSON` in those commands.
- Use `util.ExactArgs(n, "description")` instead of `cobra.ExactArgs` for better error messages.

