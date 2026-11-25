## qcloud serverless space metrics inference

Show inference token usage of a space's collections

### Synopsis

Show the inference token usage of the collections in a serverless space over a period.

The server reports, per collection and inference model, the number of tokens
processed by inference, aggregated into buckets of one day, week or month. The
table shows one row per collection and model with the total number of tokens
over the period and the largest bucket. The total over all rows is shown above
the table. Use --json to get the individual buckets.

By default, the period covers the last 7 days and buckets are aggregated per
day. When only --until is given, the period covers the 7 days before it.

By default, all collections are fetched automatically across multiple pages.
Use --page-size and --page-token for manual pagination.

```
qcloud serverless space metrics inference <space-id> [flags]
```

### Examples

```
# Show inference token usage over the last 7 days
qcloud serverless space metrics inference 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Show weekly token usage over the last 90 days
qcloud serverless space metrics inference 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 90d --interval week

# Show token usage of a single model in a single collection during October
qcloud serverless space metrics inference 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --collection products --model sentence-transformers/all-minilm-l6-v2 \
  --since 2026-10-01 --until 2026-10-31

# Export the daily buckets as JSON
qcloud serverless space metrics inference 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json
```

### Options

```
      --collection string            Only include the collection with this exact name
      --collection-contains string   Only include collections whose name contains this substring
  -h, --help[=true|false]            help for inference
      --interval string              Size of the aggregation buckets (day, week, month; default day)
      --model string                 Only include the inference model with this ID
      --page-size int32              Maximum number of collections to return per page (manual pagination mode)
      --page-token string            Page token from a previous response to resume from (manual pagination mode)
  -s, --since string                 Start of the time range (RFC3339, YYYY-MM-DD, or a duration ago such as 24h or 7d; default: 7 days before --until)
  -u, --until string                 End of the time range (RFC3339, YYYY-MM-DD, or a duration ago such as 1h; default: now)
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

* [qcloud serverless space metrics](qcloud_serverless_space_metrics.md)	 - Show metrics of a space's collections

