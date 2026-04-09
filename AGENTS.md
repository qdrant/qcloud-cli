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
    clusterutil/         # shared cluster helpers (e.g. wait-for-healthy)
    output/              # shared output formatting helpers
    util/                # shared command utilities
  qcloudapi/             # gRPC client wrapper for the Qdrant Cloud API
  state/                 # State struct (shared deps: config, lazy gRPC client)
    config/              # Viper-based config (file, env vars, flags
```

## Build & verification

**Always use Makefile targets — never raw `go build`, `go test`, or linter commands.**

| Target           | What it does                                  |
|------------------|-----------------------------------------------|
| `make build`     | Compile binary to `build/qcloud`              |
| `make test`      | Run all tests                                 |
| `make lint`      | Run golangci-lint (installs it if missing)    |
| `make format`    | Run golangci-lint with `--fix`                |
| `make bootstrap` | Install tool dependencies via `mise install`  |
| `make clean`     | Remove build artifacts                        |

To verify your changes, you should run the following makefile targets:
```bash
make lint
make build
make test
```

If make lint fails from formatting problems, use `make format` to fix them.

## Conventions

### Long descriptions and examples — mandatory

Every leaf command and group command **must** have a `Long` description and an `Example` block.

**`Long`:**
- First line expands the `Short` description into a full sentence.
- Blank line, then one or two paragraphs explaining behaviour, use cases, and important caveats.
- Use the proto service/message comments as the authoritative source of truth for what a resource or operation does.
- Do NOT describe individual flags — only document unusual or non-obvious flag interactions.

**`Example`:**
- One example per meaningful use case (basic call, common flag combinations, scripting).
- Prefix every line with `# ` comment explaining what the example does.
- Real command invocations with plausible IDs/values.

All five base types (`ListCmd`, `DescribeCmd`, `CreateCmd`, `UpdateCmd`, `Cmd`) expose `Long` and `Example` as top-level struct fields. Never set them inside `BaseCobraCommand()`.

### Tests — mandatory

Every new command package **must** ship tests. This is not optional.

- Place tests in `internal/cmd/<group>/` as `<file>_test.go` using `package <group>_test`.
- Use `testutil.NewTestEnv` + `testutil.Exec` — never call command functions directly.
- When adding a new gRPC service, also add a `fake_<service>.go` in `internal/testutil/` and register it in `server.go` / `TestEnv`.
- Cover: table output (assert header columns + key values), JSON output (unmarshal and assert), request fields sent to server, backend errors (use `Returns(nil, fmt.Errorf(...))` and assert `require.Error`), input errors (missing args, wrong flags).
- Run `make test` before declaring done.

### Subcommand pattern

Each subcommand group lives in `internal/cmd/<group>/`:

1. A public `NewCommand(s *state.State) *cobra.Command` creates the parent and registers sub-commands.
2. Leaf commands are unexported (`newListCommand`, `newDeleteCommand`, …).
3. All commands receive `*state.State` — use it to access config and the lazy gRPC client.

### State passing

`main` → `state.New(version)` → passed to every command constructor. Commands call `s.Client(ctx)` to get the gRPC client (created on first use) and `s.AccountID()` for the current account.

### Base command types (`internal/cmd/base`)

All leaf commands are built using one of five generic base types. Always prefer these over raw `cobra.Command`.

#### `base.ListCmd[T]`
For listing resources. No args, no flags needed (extend via `BaseCobraCommand` if flags are required).

Prefer `OutputTable` over `PrintText` for table output. When `OutputTable` is set, the base automatically registers `--no-headers` and handles header suppression. `PrintText` is used as a fallback when `OutputTable` is not set.

```go
base.ListCmd[*foov1.ListFoosResponse]{
    Use:   "list",
    Short: "List all foos",
    Fetch: func(s *state.State, cmd *cobra.Command) (*foov1.ListFoosResponse, error) {
        // call gRPC, return response
    },
    OutputTable: func(_ *cobra.Command, w io.Writer, resp *foov1.ListFoosResponse) output.Renderable {
        t := output.NewTable[*foov1.Foo](w)
        t.AddField("ID", func(v *foov1.Foo) string { return v.GetId() })
        t.SetItems(resp.Items)
        return t
    },
}.CobraCommand(s)
```

#### `base.DescribeCmd[T]`
For fetching and displaying a single resource (typically by ID positional arg).
```go
base.DescribeCmd[*foov1.Foo]{
    Use:   "describe <foo-id>",
    Short: "Describe a foo",
    Args:  util.ExactArgs(1, "a foo ID"),
    Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*foov1.Foo, error) {
        // call gRPC using args[0]
    },
    PrintText: func(_ *cobra.Command, w io.Writer, resource *foov1.Foo) error {
        // print fields
        return nil
    },
}.CobraCommand(s)
```

#### `base.CreateCmd[T]`
For creating a resource. Define flags in `BaseCobraCommand`; read them in `Run` via `cmd.Flags().GetString()` — do NOT use bound vars.
```go
base.CreateCmd[*foov1.Foo]{
    BaseCobraCommand: func() *cobra.Command {
        cmd := &cobra.Command{Use: "create", Short: "Create a foo", Args: cobra.NoArgs}
        cmd.Flags().String("name", "", "Name of the foo")
        return cmd
    },
    Run: func(s *state.State, cmd *cobra.Command, args []string) (*foov1.Foo, error) {
        name, _ := cmd.Flags().GetString("name")
        // call gRPC, return created resource
    },
    PrintResource: func(_ *cobra.Command, out io.Writer, resource *foov1.Foo) {
        fmt.Fprintf(out, "Foo %s created.\n", resource.GetId())
    },
}.CobraCommand(s)
```

#### `base.UpdateCmd[T]`
For updating a resource. Fetches first, then applies changes.
```go
base.UpdateCmd[*foov1.Foo]{
    BaseCobraCommand: func() *cobra.Command {
        cmd := &cobra.Command{Use: "update <foo-id>", Short: "Update a foo", Args: cobra.ExactArgs(1)}
        cmd.Flags().String("name", "", "New name")
        return cmd
    },
    Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*foov1.Foo, error) {
        // fetch existing resource by args[0]
    },
    Update: func(s *state.State, cmd *cobra.Command, resource *foov1.Foo) (*foov1.Foo, error) {
        name, _ := cmd.Flags().GetString("name")
        resource.Name = name
        // call gRPC update, return updated resource
    },
    PrintResource: func(_ *cobra.Command, out io.Writer, updated *foov1.Foo) {
        fmt.Fprintf(out, "Foo %s updated.\n", updated.GetId())
    },
}.CobraCommand(s)
```

#### `base.Cmd`
For imperative/action commands that don't return a resource (delete, wait, use, set, …).
```go
base.Cmd{
    BaseCobraCommand: func() *cobra.Command {
        cmd := &cobra.Command{Use: "delete <foo-id>", Short: "Delete a foo", Args: util.ExactArgs(1, "a foo ID")}
        cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
        return cmd
    },
    Run: func(s *state.State, cmd *cobra.Command, args []string) error {
        force, _ := cmd.Flags().GetBool("force")
        if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Delete foo %s?", args[0])) {
            fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
            return nil
        }
        // call gRPC delete
        fmt.Fprintf(cmd.OutOrStdout(), "Foo %s deleted.\n", args[0])
        return nil
    },
}.CobraCommand(s)
```

**Key rules:**
- JSON output is handled automatically by all base types — never call `output.PrintJSON` yourself.
- Always read flags via `cmd.Flags().GetString()` etc. in `Run`/`Update`; do not use cobra bound variables.
- Use `util.ExactArgs(n, "description")` instead of `cobra.ExactArgs` for better error messages.

### Proto enum pretty printing (`internal/cmd/output/`)

All TrimPrefix-based enum formatters live in `internal/cmd/output/`, grouped by proto package:

| File | Functions |
|------|-----------|
| `output/cluster.go` | `ClusterPhase`, `ClusterNodeState`, `TolerationOperator`, `TolerationEffect` |
| `output/booking.go` | `PackageTier` |
| `output/hybrid.go` | `HybridEnvironmentPhase`, `ClusterCreationStatus`, `HybridComponentPhase` |
| `output/backup.go` | `BackupStatus`, `BackupScheduleStatus`, `BackupRestoreStatus` |

Each function strips the proto enum prefix via `strings.TrimPrefix(x.String(), "PREFIX_")`. Functions are named after the type they format, without a redundant `String` suffix, since the package qualifier already provides context (`output.ClusterPhase(...)`).

**Rules:**
- Never inline `strings.TrimPrefix(x.String(), "PREFIX_")` in a cmd package. Add a function to the appropriate `output/*.go` file instead.
- Never define a private `phaseString` / `statusString` / etc. helper in a cmd package for TrimPrefix formatting. These belong in `output`.
- Switch-based format/parse pairs (`storageTierString`, `restartPolicyString`, etc.) encode semantic mappings paired with parse functions and belong with their cmd package, not in `output`.

### Output helpers (`internal/cmd/output/`)

General-purpose output formatting helpers belong in the `output` package, not in individual cmd packages.

Examples: `BoolYesNo` (formats a bool as `"yes"` / `"no"`), `BoolMark`, `HumanTime`, `OptionalValue`, etc.

**Rules:**
- If a helper formats a value for display and could be reused across more than one cmd package, add it to `output/`.
- Never define a private display-formatting helper in a cmd package when it belongs in `output`.

### Inline pointer literals

Go 1.26 allows passing a literal directly to `new`, which returns a pointer to it. Use this wherever a pointer to a constant value is needed inline:

```go
req.MultiAz = new(true)
req.Gpu    = new(false)
cfg.Version = new("1.13.0")
```

Never add helper functions (`boolPtr`, `stringPtr`, `intPtr`, etc.) for this purpose — they are redundant.

