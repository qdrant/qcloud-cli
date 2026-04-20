# qcloud e2e tests

End-to-end tests that run the real `qcloud` binary against a real Qdrant Cloud
backend. These tests **create paid resources** on whatever account you point
them at — never run them against a production account without first
double-checking the credentials.

## Running

```
make e2e
```

Or directly:

```
QCLOUD_E2E=1 \
QDRANT_CLOUD_API_KEY=... \
QDRANT_CLOUD_ACCOUNT_ID=... \
go test -timeout 20m -v -count=1 ./test/e2e/...
```

Without `QCLOUD_E2E=1` every test in this tree is skipped, so `go test ./...`
from the repo root stays network-free.

## Environment variables

| Variable                   | Required | Purpose                                                                 |
| -------------------------- | -------- | ----------------------------------------------------------------------- |
| `QCLOUD_E2E`               | yes      | Enables the suite. Any non-empty value works.                           |
| `QDRANT_CLOUD_API_KEY`     | yes      | Management API key used by every invocation.                            |
| `QDRANT_CLOUD_ACCOUNT_ID`  | yes      | Account ID the test resources are created under.                        |
| `QDRANT_CLOUD_ENDPOINT`    | no       | Override the gRPC endpoint (defaults to production).                    |
| `QCLOUD_E2E_BINARY`        | no       | Absolute path to a pre-built `qcloud` binary. Skips the download.       |
| `QCLOUD_E2E_RELEASE`       | no       | GitHub release tag to download (default `latest`).                      |

## Binary acquisition

The binary is resolved exactly once per `go test` invocation in `TestMain`:

1. If `QCLOUD_E2E_BINARY` is set, it is used as-is.
2. Otherwise the `qcloud-<os>-<arch>.tar.gz` archive from the release given by
   `QCLOUD_E2E_RELEASE` (default `latest`) is downloaded, its sha256 is checked
   against the release's `checksums.txt`, and the binary is extracted into
   `$XDG_CACHE_HOME/qcloud-e2e/<sha>/qcloud` (or the platform equivalent).

Cache entries are keyed by the archive's sha256, so repeated runs on the same
host reuse the same extracted binary — and automatically pick up a fresh copy
when the upstream release changes.

To test a locally-built binary:

```
make build
QCLOUD_E2E_BINARY=$PWD/build/qcloud make e2e
```

## Writing new tests

Every test starts with `framework.NewEnv(t)`; it returns an `*Env` that wraps
the binary path, an isolated empty config file, and generic CLI wrappers:

- `env.Run(t, args...)` — runs `qcloud`, fails the test on non-zero exit,
  streams stdout/stderr to `t.Log`.
- `env.RunAllowFail(t, args...)` — same, but returns the result instead of
  failing. Use it for negative tests.
- `env.RunJSON(t, &v, args...)` — appends `--json` and decodes stdout into `v`.

Resource-specific helpers live next to the tests that use them — they're
ordinary `_test.go` files in `package e2e_test`, free to grow without
bloating the framework. For clusters, see `cluster_helpers_test.go`:

- `createCluster(t, env, opts)` — creates a cluster and registers cleanup.
  The cluster name defaults to `e2e-<random>`, which makes leak sweeps safe.
- `waitCluster(t, env, id, timeout)` — blocks until the cluster is healthy.
- `sweepLeakedClusters(t, env, maxAge)` — best-effort cleanup of stale
  `e2e-*` clusters; call it from a dedicated test if you want automatic
  housekeeping in CI.

Prefer `RunJSON` over scraping human output — the JSON shape is stable,
human output is not.

## Safety notes

- Every cluster created through `createCluster` is scheduled for deletion via
  `t.Cleanup`. If a runner is killed mid-test, orphans remain; use
  `sweepLeakedClusters` to catch them.
- Tests are **not** safe to run with `t.Parallel()` today — they share a
  single account and a small quota. Don't add `t.Parallel()` without also
  isolating each test's account or region.
