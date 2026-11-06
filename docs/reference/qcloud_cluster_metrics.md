## qcloud cluster metrics

Show resource usage and request metrics of a cluster

### Synopsis

Show resource usage and request metrics of a cluster.

The summary shows the current CPU, RAM and disk usage of each node. The history
commands draw how the request rate and latency of the whole cluster, or the
resource usage of a single node, changed over a time range.

### Examples

```
# Show the current resource usage of every node of a cluster
qcloud cluster metrics summary 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Show the request rate and latency of a cluster over the last 7 days
qcloud cluster metrics history 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Show the resource usage of node 0 over the last 24 hours
qcloud cluster metrics node-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 0 --since 24h
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
* [qcloud cluster metrics history](qcloud_cluster_metrics_history.md)	 - Show the request rate and latency of a cluster over time
* [qcloud cluster metrics node-history](qcloud_cluster_metrics_node-history.md)	 - Show the CPU, RAM and disk usage of a cluster node over time
* [qcloud cluster metrics summary](qcloud_cluster_metrics_summary.md)	 - Show the current CPU, RAM and disk usage of each cluster node

