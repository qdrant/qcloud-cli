## qcloud cluster node-metrics-history

Show the CPU, RAM and disk usage of a cluster node over time

### Synopsis

Show the CPU, RAM and disk usage of a cluster node over time.

The node is given by its number, such as 0, as shown by "qcloud cluster metrics",
or by its full node ID. Each resource is drawn as a stacked area chart, with a
red line marking the capacity of the node: CPU shows the resources reserved for
the system with the usage of the database on top, and RAM shows the reserved
memory, then the page cache, then the memory used by the database. Without
color, the areas are drawn as stacked lines instead. For a cluster with GPUs,
the GPU usage in percent and the GPU RAM in use are shown as well.

By default, the usage over the last 7 days is shown. --since and --until accept
an RFC3339 timestamp, a YYYY-MM-DD date in UTC, or a duration ago such as 6h or
7d. When only --until is given, the range covers the 7 days before it. Times on
the charts are in UTC.

```
qcloud cluster node-metrics-history <cluster-id> <node> [flags]
```

### Examples

```
# Show the usage of node 0 over the last 7 days
qcloud cluster node-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 0

# Show the usage of node 1 over the last 24 hours
qcloud cluster node-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 1 --since 24h

# Show the usage in a specific time range
qcloud cluster node-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 0 --since 2026-10-01 --until 2026-10-07

# Output the time series as JSON
qcloud cluster node-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 0 --json
```

### Options

```
  -h, --help           help for node-metrics-history
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

