## qcloud serverless space describe

Describe a space

### Synopsis

Describe a serverless space.

Shows the space's phase, region and endpoint together with its configuration:
network restrictions, per-collection size limits, search-worker settings, and
the storage and monthly spend limits of the space. Limits prefixed with
"platform" are derived from the account's quota and cannot be changed directly.

When usage has been measured, the Usage section shows storage and this month's
spend against each limit, and which operations the space accepts: FULL,
WRITES_BLOCKED (a hard storage limit is reached: upserts are rejected, searches
and deletes work) or BLOCKED (a hard spend limit is reached: all requests are
rejected until the next month). Usage is refreshed about every 5 minutes.

```
qcloud serverless space describe <space-id> [flags]
```

### Examples

```
# Describe a space
qcloud serverless space describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json
```

### Options

```
  -h, --help[=true|false]   help for describe
```

### Options inherited from parent commands

```
      --account-id string    Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)
      --api-key string       Management API Key (env: QDRANT_CLOUD_API_KEY)
  -c, --config string        Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)
      --console-url string   Qdrant Cloud web console base URL (env: QDRANT_CLOUD_CONSOLE_URL, default https://cloud.qdrant.io)
      --context string       Override the active context (env: QDRANT_CLOUD_CONTEXT)
      --debug[=true|false]   Enable debug logging to stderr
      --endpoint string      gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)
      --json[=true|false]    Output as JSON
```

### SEE ALSO

* [qcloud serverless space](qcloud_serverless_space.md)	 - Manage serverless spaces

