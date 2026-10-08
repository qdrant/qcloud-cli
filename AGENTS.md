# qcloud-cli

## Project overview

Go CLI for [Qdrant Cloud](https://cloud.qdrant.io), built with Cobra / Viper and gRPC.

- **Module:** `github.com/qdrant/qcloud-cli`
- **Binary:** `qcloud` (built to `build/qcloud`)
- **Go version:** 1.27+
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
    util/                # shared command helpers (wait, pagination, dates, args, …)
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
| `make bootstrap` | Install tool dependencies via `mise install`  |
| `make clean`     | Remove build artifacts                        |

To verify your changes, you should run the following makefile targets:
```bash
make lint
make build
make test
```

If make lint fails from formatting problems, use `make format` to fix them.

## Commits and pull requests

Releases are cut by [releaser-pleaser](https://apricote.github.io/releaser-pleaser/) from the commit messages on `main`, so the version bump and changelog depend on them. PRs are squash-merged: a single-commit PR lands with its commit subject, a multi-commit PR with its PR title. **Both the commit subject and the PR title must follow [Conventional Commits](https://www.conventionalcommits.org/).**

Format: `<type>(<optional scope>): <lowercase imperative summary>`

| Type                                               | Use for                                                | Release effect |
|----------------------------------------------------|--------------------------------------------------------|----------------|
| `feat`                                             | New command, flag, or user-visible behaviour           | minor          |
| `fix`                                              | Bug fix in user-visible behaviour                      | patch          |
| `feat!` / `fix!`, or a `BREAKING CHANGE:` footer   | Maintainers only (see below)                           | major          |
| `docs`, `refactor`, `test`, `chore`, `ci`, `build` | Everything else                                        | none           |

- Scope is the command group or area: `cluster`, `hybrid`, `backup`, `serverless`, `inference`, `cd`, `deps`.
- Type is lowercase: `feat:`, not `Feat:`. Summary is lowercase and has no trailing period.
- Pick the type by user impact, not by effort: a new flag is `feat`, not `fix`; an internal refactor is `refactor`, even if large.
- **Never mark a change as breaking yourself.** Do not use `!` or a `BREAKING CHANGE:` footer in commits, PR titles, or `rp-commits` blocks; a major release is a human decision. If a change looks breaking, use `feat`/`fix` and tell the user so they can decide.
- When a PR mixes changes that need separate changelog entries, add an `rp-commits` block to the PR description, one conventional message per line:

  ````markdown
  ```rp-commits
  feat(inference): add inference root command with model list
  fix: use the global endpoint for package listing
  ```
  ````

Examples:
- `feat(cluster): add dashboard command`
- `fix(hybrid): resolve named packages with multi-az flag`
- `chore(deps): bump google.golang.org/grpc to 1.83.1`

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
For listing resources. `OutputTable` must be set. The base automatically registers `--no-headers` and handles header suppression. By default the command takes no positional args; set `Args` to accept them.

```go
base.ListCmd[*foov1.ListFoosResponse]{
    Use:   "list",
    Short: "List all foos",
    Fetch: func(s *state.State, cmd *cobra.Command) (*foov1.ListFoosResponse, error) {
        // call gRPC, return response
    },
    OutputTable: func(_ *cobra.Command, w io.Writer, resp *foov1.ListFoosResponse) (output.TableRenderer, error) {
        t := output.NewTable[*foov1.Foo](w)
        t.AddField("ID", func(v *foov1.Foo) string { return v.GetId() })
        t.SetItems(resp.GetItems())
        return t, nil
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
- JSON output is handled automatically by all base types — never call `output.PrintJSON` yourself (except on the create-with-wait failure path, see below).
- Always read flags via `cmd.Flags().GetString()` etc. in `Run`/`Update`; do not use cobra bound variables.
- Use `util.ExactArgs(n, "description")` instead of `cobra.ExactArgs` for better error messages.

### Shared helpers (`internal/cmd/util`, `internal/cmd/output`)

Before writing a polling loop, a flag-registration block, a parser or a display formatter, read `internal/cmd/util` and `internal/cmd/output` (e.g. `output.BoolYesNo`, `output.HumanTime`, `output.Duration`): it most likely exists already. If you write a helper that a second cmd package could use, put it there, not in your cmd package.

The concerns below come up in most new commands. Each has a reference implementation; copy its shape.

**Waiting (`--wait`).** Never hand-roll a ticker loop or the wait flags.
```go
util.AddWaitFlags(cmd, "the space to become ready", 10*time.Minute, 5*time.Second)
// ...
return util.PollUntilDone(ctx, timeout, pollInterval, "space to become ready",
    func(ctx context.Context) (*spacev1.Space, bool, error) { /* get, report phase, return done */ })
```
Reference: `internal/cmd/serverless/space/wait_helpers.go`, `internal/cmd/clusterutil/wait.go`.

**Create with `--wait` that fails.** The resource exists even though the wait failed. Print it (JSON or text) before returning an error that says it was created and gives its ID. For API keys this is the only time the secret is ever shown.
```go
if err := waitForKeyReady(ctx, cmd.ErrOrStderr(), probe, waitTimeout, pollInterval); err != nil {
    if s.Config.JSONOutput() {
        _ = output.PrintJSON(cmd.OutOrStdout(), created)
    } else {
        output.CreatedAPIKey(cmd.OutOrStdout(), created)
    }
    return nil, fmt.Errorf("API key %s was created but is not active on the cluster: %w", created.GetId(), err)
}
```
Reference: `internal/cmd/cluster/key_create.go`, `internal/cmd/cluster/create.go`.

**Pagination.** A list RPC with page tokens must paginate. By default the command fetches every page; `--page-size` / `--page-token` switch to manual paging, and the next token is returned in the JSON output.
```go
items, next, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*clusterv1.Cluster, string, error) {
    resp, err := client.Cluster().ListClusters(ctx, &clusterv1.ListClustersRequest{AccountId: accountID, PageSize: pageSize, PageToken: pageToken})
    // ...
    return resp.GetItems(), resp.GetNextPageToken(), nil
})
// ...
return &clusterv1.ListClustersResponse{Items: items, NextPageToken: next}, nil
```
Register the flags with `util.AddPaginationFlags(cmd, "clusters")`. Reference: `internal/cmd/cluster/list.go`.

**Date-only upper bounds.** `--expires 2026-12-31` or `--until 2026-12-31` includes that whole day: parse with `util.ParseDateEndOfDay` (23:59:59 UTC), not `time.Parse(time.DateOnly, ...)`. Reference: `internal/cmd/cluster/key_create.go`.

**Time-range flags.** `--since` / `--until` accept RFC3339, `YYYY-MM-DD`, or a duration ago (`30m`, `24h`, `7d`): read both with `util.ReadTimeRange`, which treats a `--until` date as the end of that day and rejects a `--since` that is not before `--until`. Reference: `internal/cmd/cluster/logs.go`, `internal/cmd/serverless/space/metrics_usage.go`.

**Unset values.** An unset optional proto message or field renders as `not set` or as what it means, never as a misleading zero value. A nil retention period is `indefinite`, not `0s` (`output.RetentionPeriod`); a nil optional field goes through `output.OptionalValue(v, "not set")`, not `fmt.Sprint(v.GetX())`. Reference: `internal/cmd/output/backup.go`, `internal/cmd/output/format.go`.

### Proto enum pretty printing (`internal/cmd/output/`)

All TrimPrefix-based enum formatters live in `internal/cmd/output/<proto package>.go` (e.g. `output/cluster.go`, `output/serverless.go`). Check there before adding one.

Each function strips the proto enum prefix via `strings.TrimPrefix(x.String(), "PREFIX_")`. Functions are named after the type they format, without a redundant `String` suffix, since the package qualifier already provides context (`output.ClusterPhase(...)`).

**Rules:**
- Never inline `strings.TrimPrefix(x.String(), "PREFIX_")` in a cmd package. Add a function to the appropriate `output/*.go` file instead.
- Never define a private `phaseString` / `statusString` / etc. helper in a cmd package for TrimPrefix formatting. These belong in `output`.
- Switch-based format/parse pairs (`storageTierString`, `restartPolicyString`, etc.) encode semantic mappings paired with parse functions and belong with their cmd package, not in `output`.

### Inline pointer literals

Go 1.26 allows passing a literal directly to `new`, which returns a pointer to it. Use this wherever a pointer to a constant value is needed inline:

```go
req.MultiAz = new(true)
req.Gpu    = new(false)
cfg.Version = new("1.13.0")
```

Never add helper functions (`boolPtr`, `stringPtr`, `intPtr`, etc.) for this purpose — they are redundant.

