## qcloud cluster alert

Show alerts of clusters

### Synopsis

Show alerts of clusters.

Alerts are raised when a cluster needs attention, for example when it is close
to running out of disk or memory, is out of disk or memory, has nodes in
recovery mode, is unhealthy or CPU throttled, has too many collections, has a
resource hotspot, runs a version not covered by the SLA, or has a database API
key that is about to expire. An alert is either firing or resolved.

### Examples

```
# List the alerts of a cluster that are currently firing
qcloud cluster alert list 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# List both firing and resolved alerts
qcloud cluster alert list 7b2ea926-724b-4de2-b73a-8675c42a6ebe --show-resolved
```

### Options

```
  -h, --help   help for alert
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
* [qcloud cluster alert list](qcloud_cluster_alert_list.md)	 - List alerts of a cluster

