---
description: "Review qdrant-cloud-public-api for module updates, new RPCs, and CLI coverage gaps. Use when asked to check API changes, find missing commands, or detect new proto fields."
allowed-tools: [Read, Glob, Grep, Bash, LSP]
---

# Review Public API Coverage

Analyze the qdrant-cloud-public-api module for updates, missing RPCs, and unused proto fields in existing commands. Follow all phases in order.

## Phase 1 -- Auto-update the module

1. Read `go.mod` and extract the current version of `github.com/qdrant/qdrant-cloud-public-api`.
2. Run in the CLI repo root:
   ```
   go get github.com/qdrant/qdrant-cloud-public-api@latest
   go mod tidy
   ```
3. Re-read `go.mod` and note whether the version changed. Report old and new version (or "already at latest").

## Phase 2 -- Locate proto files from Go module cache

1. Run `go env GOMODCACHE` to get the module cache path.
2. Read the updated version from `go.mod`.
3. Set `PROTO_ROOT=$GOMODCACHE/github.com/qdrant/qdrant-cloud-public-api@v<VERSION>/proto/qdrant/cloud`
4. Set `GEN_ROOT=$GOMODCACHE/github.com/qdrant/qdrant-cloud-public-api@v<VERSION>/gen/go/qdrant/cloud`
5. Verify `PROTO_ROOT` exists. If not, run `go mod download` and retry.

## Phase 3 -- Build RPC inventory for wired services

Dynamically discover all wired services and their proto files. Do NOT use a hardcoded list.

### Discover wired services from code

1. Read `internal/qcloudapi/client.go`.
2. Extract every import whose path starts with `github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/`. Each import gives:
   - The **import alias** (e.g., `clusterv1`, `backupv1`)
   - The **subpath** after `.../gen/go/qdrant/cloud/` (e.g., `cluster/v1`, `cluster/backup/v1`)
3. For each import, derive:
   - **Proto dir**: `PROTO_ROOT/<subpath>/` -- glob for `*.proto` files there
   - **Generated gRPC file**: `GEN_ROOT/<subpath>/` -- glob for `*_grpc.pb.go` files there
4. From the `Client` struct fields, match each field's type to an import alias to confirm which accessor method goes with which import path.

### For each discovered proto file

1. Extract every `rpc MethodName(RequestType) returns (ResponseType)` line.
2. Check whether the RPC has `supported_actor_types` set to `ACTOR_TYPE_USER` within its option block. If so, mark it **user-only** (not implementable -- the CLI authenticates with management API keys).
3. Classify the RPC by its verb prefix:
   - **read**: `List*`, `Get*` 
   - **write**: `Create*`, `Update*`
   - **action**: `Delete*`, `Restart*`, `Enable*`, `Disable*`, and anything else
4. Record the request and response message type names.

## Phase 4 -- Map CLI commands to RPCs using LSP

For each wired service, use gopls to find which RPCs are actually called by the CLI:

1. Use **LSP `documentSymbol`** on the service's `_grpc.pb.go` file (at `GEN_ROOT/<path>`) to list all methods on the client interface (the type ending in `Client`).
2. For each RPC method on the client interface, use **LSP `findReferences`** to locate call sites.
3. Filter references to files under `internal/cmd/` and `internal/qcloudapi/` (exclude `*_test.go`).
4. If an RPC has no references, it is **missing** (not implemented).
5. For each implemented RPC, record the command file(s) that call it.
6. In each command file, identify the base command type used:
   - `base.ListCmd` or `base.DescribeCmd` -> confirms **read** RPC
   - `base.CreateCmd` or `base.UpdateCmd` -> confirms **write** RPC
   - `base.Cmd` -> confirms **action** RPC

## Phase 5 -- Field analysis for implemented RPCs

For each implemented RPC, analyze whether the CLI uses all available proto fields. Limit analysis to **top-level fields plus one level of nested messages**.

### Read RPCs (List/Get)

Focus on **response** fields (what could be shown in output):

1. In the proto file, find the response message type. For List RPCs, the response typically contains a repeated field of the resource message (e.g., `ListClustersResponse` has `repeated Cluster items`). Analyze the **resource message** fields.
2. In the command file, read the `OutputTable` function (for ListCmd) or `PrintText` function (for DescribeCmd). Identify all `.Get<Field>()` calls and direct field accesses.
3. Report fields present in the proto resource message but not accessed in the output function. These are candidates for new table columns or display lines.

### Write RPCs (Create/Update)

Focus on **request** fields (what could become flags):

1. In the proto file, find the request message type. For Create/Update RPCs, the request typically wraps the resource message (e.g., `CreateClusterRequest` has a `Cluster cluster` field). Analyze the **resource message** fields.
2. In the command file, read:
   - Flag definitions in `BaseCobraCommand` (the flags registered on the cobra command)
   - Field assignments in the `Run` or `Update` function (where request fields are populated from flags)
3. Report resource message fields not populated from any flag. These are candidates for new CLI flags.
4. Also check `PrintResource` for response fields shown after creation/update.

### Action RPCs (Delete/Restart/etc.)

Lighter analysis:
1. Check if the request message has fields beyond the resource ID and account ID.
2. If the command defines flags, check which request fields they map to.
3. Report any request fields not exposed.

### What to skip

- Pagination fields (`page_token`, `page_size`) in request messages -- handled by the framework
- `account_id` fields -- always set from config, not a user flag
- `update_mask` fields in update requests -- handled by the framework
- Fields on wrapper messages that just hold the resource (e.g., `CreateClusterRequest.cluster` itself is not a "missing field")

## Phase 6 -- Generate report

### 1. Module Version Status
```
**Module**: github.com/qdrant/qdrant-cloud-public-api
**Previous version**: v0.X.Y
**Current version**: v0.X.Z (updated / already at latest)
```

### 2. RPC Coverage Summary

Markdown table sorted by coverage ascending (worst gaps first):

| Service | Total | Implemented | User-Only | Missing | Coverage |
|---------|-------|-------------|-----------|---------|----------|

### 3. Missing RPCs

For each service with missing RPCs, list them grouped as:
- **Implementable**: RPCs that could be added to the CLI
- **User-only**: RPCs restricted to `ACTOR_TYPE_USER` (cannot implement with management API keys)

### 4. Field Gaps in Existing Commands

For each implemented RPC where unused fields were found:

```
### <CLI command path> -> <Service>.<RPC> (<read|write|action>)

**Response fields not in output:**
- `field_name` (<type>) -- <proto comment if available>

**Request fields not exposed as flags:**
- `field_name` (<type>) -- <proto comment if available>
```

Only include sections with actual gaps. Skip RPCs where all relevant fields are used.

### 5. Recommendations

Prioritized list:
1. **Quick wins** -- new fields on existing commands (just add a flag or table column)
2. **New RPCs** -- missing RPCs in services that are mostly covered
3. **User-only RPCs** -- listed for awareness, flagged as not implementable
