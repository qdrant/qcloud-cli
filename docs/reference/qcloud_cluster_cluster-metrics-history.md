## qcloud cluster cluster-metrics-history

Show the request rate and latency of a cluster over time

### Synopsis

Show the request rate and latency of a cluster over time.

The requests per second and the request latency in milliseconds across all
nodes of the cluster are drawn as separate line charts. Request metrics are
measured at the ingress of Qdrant Cloud, so they are not available for Hybrid
Cloud clusters.

By default, the last 7 days are shown. --since and --until accept an RFC3339
timestamp, a YYYY-MM-DD date in UTC, or a duration ago such as 6h or 7d. When
only --until is given, the range covers the 7 days before it. Times on the
charts are in UTC.

```
qcloud cluster cluster-metrics-history <cluster-id> [flags]
```

### Examples

```
# Show the request rate and latency over the last 7 days
qcloud cluster cluster-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Show the request rate and latency over the last 6 hours
qcloud cluster cluster-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe --since 6h

# Show the request rate and latency in a specific time range
qcloud cluster cluster-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe --since 2026-10-01 --until 2026-10-07

# Output the time series as JSON
qcloud cluster cluster-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json
```

### Options

```
  -h, --help           help for cluster-metrics-history
  -s, --since string   Start of the time range (RFC3339, YYYY-MM-DD, or a duration ago such as 24h or 7d; default: 7 days before --until)
  -u, --until string   End of the time range (RFC3339, YYYY-MM-DD, or a duration ago such as 1h; default: now)
```

### Options inherited from parent commands

```
      --account-id string    Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)
      --api-key string       Management API Key (env: QDRANT_CLOUD_API_KEY)
  -c, --config string        Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)
      --console-url string   Qdrant Cloud web console base URL (env: QDRANT_CLOUD_CONSOLE_URL, default https://cloud.qdrant.io)
      --context string       Override the active context (env: QDRANT_CLOUD_CONTEXT)
      --debug                Enable debug logging to stderr
      --endpoint string      gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)
      --json                 Output as JSON
```

### SEE ALSO

* [qcloud cluster](qcloud_cluster.md)	 - Manage Qdrant Cloud clusters

