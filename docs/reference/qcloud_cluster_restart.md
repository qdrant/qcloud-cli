## qcloud cluster restart

Restart a cluster

```
qcloud cluster restart <cluster-id> [flags]
```

### Examples

```
# Restart a cluster (prompts for confirmation)
qcloud cluster restart 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Restart without confirmation and wait for healthy status
qcloud cluster restart 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force --wait
```

### Options

```
  -f, --force                   Skip confirmation prompt
  -h, --help                    help for restart
      --wait                    Wait for the cluster to restart to a healthy status
      --wait-timeout duration   Maximum time to wait for cluster the cluster to restart to healthy status (default 10m0s)
```

### Options inherited from parent commands

```
      --account-id string   Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)
      --api-key string      Management API Key (env: QDRANT_CLOUD_API_KEY)
  -c, --config string       Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)
      --context string      Override the active context (env: QDRANT_CLOUD_CONTEXT)
      --debug               Enable debug logging to stderr
      --endpoint string     gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)
      --json                Output as JSON
```

### SEE ALSO

* [qcloud cluster](qcloud_cluster.md)	 - Manage Qdrant Cloud clusters

