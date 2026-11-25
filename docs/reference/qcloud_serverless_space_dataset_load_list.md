## qcloud serverless space dataset load list

List the sample dataset loads of a space

### Synopsis

List the sample dataset loads of a serverless space.

Only loads created in the last 30 days are listed. PROGRESS shows how much data
was copied while a load is running. Use "dataset load describe" to see why a
load failed.

All pages are fetched by default. Use --page-size and --page-token for manual
pagination.

```
qcloud serverless space dataset load list <space-id> [flags]
```

### Examples

```
# List the sample dataset loads of a space
qcloud serverless space dataset load list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space dataset load list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json
```

### Options

```
  -h, --help[=true|false]         help for list
      --no-headers[=true|false]   Do not print column headers
      --page-size int32           Maximum number of loads to return per page (manual pagination mode)
      --page-token string         Page token from a previous response to resume from (manual pagination mode)
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

* [qcloud serverless space dataset load](qcloud_serverless_space_dataset_load.md)	 - Manage sample dataset loads of a space

