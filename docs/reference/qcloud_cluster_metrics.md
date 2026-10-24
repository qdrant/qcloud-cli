## qcloud cluster metrics

Show the current CPU, RAM and disk usage of each cluster node

### Synopsis

Show the current CPU, RAM and disk usage of each node in a cluster.

Each resource is drawn as a bar relative to the capacity of the node, followed
by the usage in percent and in absolute values. CPU and RAM include the
resources reserved for the system, which the database cannot use. RAM is further
split into the page cache and the memory used by the database itself.

The values are short-term averages reported by the cluster's monitoring, so a
node that has just started may show no data yet.

```
qcloud cluster metrics <cluster-id> [flags]
```

### Examples

```
# Show resource usage for every node of a cluster
qcloud cluster metrics 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Output the usage as JSON
qcloud cluster metrics 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json

# List the RAM usage in percent per node
qcloud cluster metrics 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json | jq -r '.nodes[] | "\(.node) \(.ram.usedPercent)"'
```

### Options

```
  -h, --help   help for metrics
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

