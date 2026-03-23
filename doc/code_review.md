# Code Review: qcloud-cli

**Date:** 2026-03-23
**Scope:** Full codebase review (~12,500 lines of Go across 120+ files)
**Version reviewed:** v0.11.0 (commit `5a61dab`)

---

## Executive Summary

`qcloud-cli` is a well-engineered CLI for the Qdrant Cloud platform. The codebase demonstrates strong architectural discipline: a generic base-command abstraction eliminates boilerplate, a clean dependency-injection pattern (`*state.State`) enables thorough testing, and a `bufconn`-backed test infrastructure exercises the full command stack (argument parsing through gRPC serialization) without network I/O.

Two confirmed bugs were found (swapped format arguments in the scale prompt, and incorrect version stripping in self-upgrade), along with a floating-point precision issue in millicore parsing, a double-suffix problem in Makefile-built versions, and several lower-priority design and test-coverage observations.

Overall quality is high. The issues identified are straightforward to fix.

---

## Architecture Overview

```
cmd/qcloud/main.go          Entry point: builds version string, creates State, runs root command
internal/cli/root.go         Root cobra command: persistent flags, config loading, subcommand wiring
internal/state/              State (DI container) + Config (viper-backed, multi-source)
internal/qcloudapi/          gRPC client wrapper with auth interceptor
internal/cmd/base/           Generic command types: Cmd, CreateCmd, ListCmd, DescribeCmd, UpdateCmd
internal/cmd/cluster/        Cluster CRUD, scaling, keys, wait, completion (~3,800 lines)
internal/cmd/backup/         Backup CRUD, schedules, restores (~1,600 lines)
internal/cmd/context/        Local context management (~700 lines)
internal/cmd/output/         Table rendering + JSON output
internal/cmd/util/           Shared helpers: ConfirmAction, ExactArgs, IP/label parsing
internal/resource/           ByteQuantity and Millicores value types (pflag.Value)
internal/selfupgrade/        GitHub-based binary self-update
internal/testutil/           Test infrastructure: fake gRPC server, MethodSpy, config helpers
```

Key design decisions:
- **Generic base commands** (`base.ListCmd[T]`, `base.CreateCmd[T]`, etc.) enforce a consistent UX pattern (JSON output, error wrapping, completion) across all commands.
- **`*state.State`** flows through all constructors as a lightweight DI container. `SetClient` / `SetUpdater` methods enable test injection.
- **Config precedence:** flag > env var > config-file context > config-file defaults > built-in defaults, implemented via Viper layers.
- **Test infrastructure:** `testutil.NewTestEnv` sets up an in-process gRPC server via `bufconn` with `MethodSpy[Req, Resp]` for recording and dispatching per-call responses.

---

## Strengths

1. **Consistent command structure.** Every command follows the same pattern: construct base cobra command, wire flags, implement business logic callback, register completions. This makes the codebase highly navigable and predictable.

2. **Excellent test coverage.** ~50 test files covering happy paths, error conditions, edge cases, and shell completion. The `MethodSpy` generics pattern eliminates repetitive test-double boilerplate. Tests exercise the full CLI stack in-process.

3. **Good UX design.** Auto-pagination on list commands, context-aware shell completion (e.g., filtering CPU/RAM/GPU completions based on already-selected flags), confirmation prompts for destructive operations, `--wait` support with progress polling, human-readable diff formatting (`old => new`), and sensible defaults throughout.

4. **Clean error handling.** Errors are consistently wrapped with `fmt.Errorf("failed to ...: %w", err)`, preserving the chain while adding context. Custom `ExactArgs` provides usage hints on wrong argument counts.

5. **Security practices.** Config files written with `0600`, config dirs with `0700`. API keys sent via gRPC metadata. `context show` intentionally excludes the API key from output (verified by tests).

6. **Well-designed config system.** Multi-source config with named contexts, env-var overrides, and flag precedence. The context `set` command inherits unspecified values from the current config, reducing user friction.

---

## Bugs

### 1. Swapped Disk / Multi-AZ in scale confirmation prompt

**File:** `internal/cmd/cluster/scale.go`, lines 331-338
**Severity:** Medium — users see incorrect information in the confirmation prompt

The format string expects `Disk` then `Multi AZ`, but the arguments pass them in reverse order:

```go
prompt := fmt.Sprintf(
    "...Disk:    %s\n  Multi AZ: %s",
    // ...
    output.DiffValue(boolToYesNo(oldPkg.GetMultiAz()), boolToYesNo(newPkg.GetMultiAz())), // ← goes to Disk
    diskLine, // ← goes to Multi AZ
)
```

**Impact:** The scale confirmation prompt shows the Multi-AZ diff value in the "Disk" line and the disk diff value in the "Multi AZ" line, misleading users before they confirm a scaling operation.

**Fix:** Swap the two arguments so `diskLine` precedes the Multi-AZ diff:

```go
prompt := fmt.Sprintf(
    "...Disk:    %s\n  Multi AZ: %s",
    // ...
    diskLine,
    output.DiffValue(boolToYesNo(oldPkg.GetMultiAz()), boolToYesNo(newPkg.GetMultiAz())),
)
```

### 2. Incorrect version stripping in self-upgrade

**File:** `internal/cmd/selfupgrade/selfupgrade.go`, line 37
**Severity:** Low — mitigated by the `isDev` guard that skips the equality check

```go
currentVersion = strings.SplitN(currentVersion, "-", 2)[0]
```

The intent is to strip the `-dev` suffix to get the base semver version. But `SplitN` splits on the *first* hyphen in the string. For a version like `"0.11.0-dev"`, this produces `"0"` instead of `"0.11.0"`.

**Impact:** For dev builds, `UpdateSelf` receives version `"0"` instead of `"0.11.0"`. The upgrade still works because `go-selfupdate` will always find a newer release, but display messages show the wrong current version.

**Fix:** Use `strings.TrimSuffix`:

```go
currentVersion = strings.TrimSuffix(currentVersion, "-dev")
```

Or, more robustly for snapshot builds (version like `"0.11.0-dev+abc123"`):

```go
currentVersion, _, _ = strings.Cut(currentVersion, "-")
// This is wrong too for semver — use a proper suffix strip instead.
currentVersion = strings.TrimSuffix(currentVersion, "-"+versionPrerelease)
```

---

## Issues by Priority

### High

| # | Issue | Location | Description |
|---|-------|----------|-------------|
| 1 | Swapped scale prompt args | `cluster/scale.go:331-338` | Disk and Multi-AZ values transposed in confirmation prompt (see Bugs section) |

### Medium

| # | Issue | Location | Description |
|---|-------|----------|-------------|
| 2 | Float-to-int truncation in `ParseMillicores` | `resource/millicores.go:27` | `int64(v * 1000)` truncates instead of rounding. `ParseMillicores("0.3")` → 299 instead of 300. Fix: use `math.Round(v * 1000)` |
| 3 | Double `-dev` suffix in Makefile builds | `Makefile:6` | ldflags set `main.version=$(VERSION)-dev` but `versionPrerelease` stays `"dev"`, producing e.g. `v0.11.0-dev-dev`. Fix: also set `-X main.versionPrerelease=` in ldflags |
| 4 | Version stripping in self-upgrade | `selfupgrade.go:37` | `SplitN` on `-` breaks semver versions (see Bugs section) |
| 5 | Nil-pointer risk in `scale.go` `PrintResource` | `cluster/scale.go:296` | `updated.State.Phase` accessed without nil-checking `State`. Use `updated.GetState().GetPhase()` |
| 6 | `ConfirmAction` reads from `os.Stdin` directly | `util/util.go:19` | Makes interactive confirmation path untestable. Accept `io.Reader` or use `cmd.InOrStdin()` |
| 7 | `actions/checkout@v4` in release workflow | `.github/workflows/release.yml:15` | CI and build use `@v6`; release still uses `@v4`. Update for consistency and security fixes |

### Low

| # | Issue | Location | Description |
|---|-------|----------|-------------|
| 8 | Regex recompiled on every call | `cluster/helpers.go:19` | `regexp.MatchString` recompiles the UUID regex each invocation. Use `regexp.MustCompile` at package level |
| 9 | No overflow check in `ParseByteQuantity` | `resource/bytes.go:63` | `ByteQuantity(n) * e.mult` can silently overflow `int64` for very large values |
| 10 | Auth interceptor is unary-only | `qcloudapi/client.go:87` | If streaming RPCs are added, they will not get authenticated. Consider also adding a `StreamClientInterceptor` |
| 11 | gRPC connection never closed | `state/state.go` | `Client()` creates a connection lazily but `State` has no `Close()`. Harmless for a CLI (OS reclaims), but a code-hygiene concern |
| 12 | No signal handling on root context | `cmd/qcloud/main.go:25` | `context.Background()` has no cancellation. Consider `signal.NotifyContext` for clean abort on Ctrl+C |
| 13 | `PrintText` nil-dereference risk in base commands | `base/describe.go`, `base/list.go` | `DescribeCmd` and `ListCmd` call `PrintText` without nil-checking, unlike `CreateCmd` which nil-checks `PrintResource` |
| 14 | No client-side CIDR validation | `util/ips.go` | `--allowed-ip` values are passed through without format validation; server-side validation is relied upon |
| 15 | No client-side cron validation | `backup/schedule_create.go` | `--schedule` accepts arbitrary strings. The `cron` package is already a dependency (used in `schedule_describe.go`) and could validate client-side |
| 16 | Missing upper-bound for `--retention-days` | `backup/create.go` | Help text says `(1-365)` but only the lower bound (`< 1`) is validated |
| 17 | `OptionalValue` uses reflection | `output/format.go:41` | Could be replaced with a generic function `func OptionalValue[T any](v *T, fallback string) string` for type safety |
| 18 | First poll delayed in `waitForHealthy` | `cluster/wait_helpers.go:57-70` | The `for/select` waits for the first ticker tick before polling. An immediate first poll would avoid a 5-second delay when the cluster is already healthy |
| 19 | Completion functions report no error message | `completion/completion.go` | All completion functions return `ShellCompDirectiveError` without any user-visible message explaining why completion failed |

---

## Package-by-Package Findings

### `cmd/qcloud` (main)

Clean entry point. The `versionPrerelease` build variable and `versionString()` logic is straightforward. The `SilenceErrors: true` / manual stderr print pattern in root.go is correct.

**Suggestion:** Wire `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` for graceful cancellation of long-running operations (e.g., `cluster wait`, `cluster create --wait`).

### `internal/state`

Well-designed DI container. Lazy initialization of `Client` and `Updater` avoids startup cost for commands that don't need them. `SetClient`/`SetUpdater` injection methods enable testing.

**Note:** `Client()` is not goroutine-safe — concurrent callers could race on `s.client`. Acceptable for a single-threaded CLI, but worth documenting.

### `internal/state/config`

Robust Viper-backed config with context support. Security-conscious file permissions (0600/0700). `errors.AsType` usage is clean.

**Note:** `MergeConfigMap` error is silently ignored on line 120. Consider debug-level logging.

### `internal/qcloudapi`

Clean gRPC client wrapper. Auth interceptor correctly injects the API key via metadata.

**Note:** `New()` accepts a `ctx` parameter but doesn't use it (since `grpc.NewClient` doesn't take a context). The parameter is vestigial but harmless.

### `internal/cmd/cluster`

Largest package (~3,800 lines). Well-structured with resolution logic (`create_resolve.go`), helpers (`helpers.go`), completions (`completion.go`), and wait logic (`wait_helpers.go`) properly separated.

**Key findings:**
- Scale confirmation prompt bug (see Bugs section).
- `helpers_test.go` is empty — no unit tests for `isUUID`, `parseDiskPerformance`, `parseRestartMode`, `parseRebalanceStrategy`, or `formatMillicents`. These are covered indirectly through integration tests but would benefit from direct unit tests.
- Duplicated "cluster is ready" message logic across `create.go`, `restart.go`, and `wait.go` — could be extracted to a shared helper.
- Duplicated package-filtering logic between `completion.go` and `create_resolve.go`.

### `internal/cmd/backup`

Consistent and well-tested. Schedule CRUD is clean. Restore operations are straightforward.

**Notes:**
- `--cluster-id` flag is registered after `CobraCommand()` in some files (`list.go`, `restore_list.go`), unlike other commands where flags are defined inside `BaseCobraCommand`. Works but splits flag definitions across two locations.
- `schedule_create.go` does not validate cron expressions client-side despite the `cron` library already being a dependency.
- Retention days has no upper-bound check (`--retention-days=99999` would be accepted).

### `internal/cmd/context`

Clean local-only commands. Good security practice excluding API keys from `context show` output.

**Notes:**
- `list.go` doesn't use `base.ListCmd` and manually handles `--json`, inconsistent with other commands.
- API keys are stored in plaintext in the config file. The file permissions (0600) mitigate this, but a warning or credential-helper support would be a stronger approach.

### `internal/cmd/base`

Elegant generic abstraction. `ListCmd[T]`, `CreateCmd[T]`, `DescribeCmd[T]`, and `UpdateCmd[T]` enforce consistent UX while reducing per-command boilerplate to ~50-100 lines.

**Note:** `CreateCmd` nil-checks `PrintResource` but `DescribeCmd` and `ListCmd` do not nil-check `PrintText`. If a developer forgets to set `PrintText`, a nil-pointer panic occurs at runtime. Consider either nil-checking consistently or documenting the requirement.

### `internal/cmd/util`

`ParseIPs` and `ParseLabels` have well-designed "last wins" semantics, thoroughly tested. `ExactArgs` provides good error UX.

**Note:** `labels.go` line 56 has a minor comment describing the wrong operation ("if a label is set after it's set" should read "if a label is removed after it's set").

### `internal/resource`

`ByteQuantity` and `Millicores` are well-designed value types implementing `pflag.Value`.

**`Millicores` precision issue:** `int64(v * 1000)` on line 27 of `millicores.go` truncates instead of rounding. For `v = 0.3`, floating-point representation yields `299.99999999999997`, which truncates to `299` instead of `300`. Fix with `math.Round`:

```go
return Millicores(int64(math.Round(v * 1000))), nil
```

### `internal/selfupgrade`

Clean wrapper around `go-selfupdate`. The `Updater` interface enables test mocking.

**Note:** `NewGitHubUpdater()` creates an unauthenticated GitHub source. Users hitting rate limits could benefit from `GITHUB_TOKEN` support.

### `internal/testutil`

Excellent test infrastructure:
- `MethodSpy[Req, Resp]` eliminates repetitive mock code with type-safe generics.
- `bufconn`-backed server avoids network overhead.
- `RequestCapture` interceptor enables metadata/header assertions.
- `sync.Once` cleanup prevents double-close panics.

No issues found.

---

## Test Coverage Assessment

**Strengths:**
- ~50 test files with well-structured, readable tests
- Integration-style tests exercise the full CLI stack (argument parsing → gRPC serialization → response formatting)
- Edge cases consistently covered: missing flags, API errors, empty responses, invalid inputs
- JSON output verified via unmarshaling alongside table output
- Shell completion thoroughly tested (cluster package has 332 lines of completion tests)

**Gaps:**
- `helpers_test.go` is empty — no direct unit tests for helper functions
- Interactive confirmation path (`ConfirmAction` without `--force`) is untested everywhere, as `os.Stdin` is hardcoded
- Cloud provider/region tests are minimal (1-2 tests each)
- `schedule_completion.go` and `context/completion.go` have no dedicated tests
- No test for the self-upgrade "successful update" path
- No fuzz testing for `ParseIPs`, `ParseLabels`, `ParseMillicores`, or `ParseByteQuantity`
- No test for `ParseMillicores("0.3")` which would expose the truncation bug

---

## CI/CD Review

**GitHub Actions configuration** (4 workflows: `ci.yml`, `build.yml`, `release.yml`, `releaser-pleaser.yml`):

- CI runs lint (golangci-lint), tests (`-race -coverpkg`), and `go mod tidy` checks. Good.
- Build creates snapshot artifacts via GoReleaser on every push/PR. Good for verification.
- Release triggered on version tags via GoReleaser. Clean and correct.
- Releaser-pleaser automates release PRs via conventional commits.

**Issues:**
1. `release.yml` uses `actions/checkout@v4` while `ci.yml` and `build.yml` use `@v6`. Should be updated for consistency.
2. No coverage threshold enforcement or upload to a coverage service (Codecov, Coveralls). Coverage is generated (`-coverprofile`) but not consumed.
3. `releaser-pleaser.yml` uses `pull_request_target` trigger (runs with base repo secrets). The action is pinned to a version tag, but a SHA pin would be more secure.

**Makefile:**
- `VERSION` fallback on line 1: `|| "undefined"` outputs the literal string `"undefined"` (with quotes) due to shell quoting inside `$(shell ...)`. Should be `|| echo undefined`.
- Build target appends `-dev` to `main.version` via ldflags but does not clear `main.versionPrerelease`, producing double `-dev` suffixes.

---

## Recommendations

### Quick Wins (easy to fix, high value)

1. **Fix the swapped Disk/Multi-AZ arguments** in `scaleConfirmPrompt` — swap lines 337 and 338.
2. **Fix `ParseMillicores` truncation** — use `math.Round(v * 1000)` instead of `int64(v * 1000)`.
3. **Fix Makefile double-dev** — add `-X main.versionPrerelease=` to the ldflags.
4. **Fix self-upgrade version stripping** — replace `SplitN` with `TrimSuffix`.
5. **Compile UUID regex once** — `var uuidRegex = regexp.MustCompile(...)` at package level.
6. **Update `release.yml`** to use `actions/checkout@v6`.

### Medium-Term Improvements

7. **Add unit tests for `helpers.go`** — `isUUID`, parse/format functions, `formatMillicents`.
8. **Make `ConfirmAction` testable** — accept an `io.Reader` parameter instead of reading from `os.Stdin`.
9. **Add client-side validation** for cron expressions and CIDR notation.
10. **Add `nil` checks in `DescribeCmd.PrintText`** and `ListCmd.PrintText` to match `CreateCmd`.
11. **Use `updated.GetState().GetPhase()`** instead of `updated.State.Phase` in `scale.go:296`.
12. **Extract shared "cluster is ready" message** used in `create.go`, `restart.go`, and `wait.go`.

### Longer-Term Considerations

13. **Signal handling** — use `signal.NotifyContext` for graceful Ctrl+C during long operations.
14. **Credential helper / keyring integration** — API keys are currently stored in plaintext in the config file. While file permissions mitigate the risk, a credential helper would be more secure.
15. **Coverage reporting** — upload coverage from CI to track trends and enforce minimums.
16. **Streaming auth interceptor** — add `StreamClientInterceptor` alongside the existing unary interceptor for future-proofing.
17. **Immediate first poll in `waitForHealthy`** — avoid the initial 5-second delay when the cluster is already healthy.
