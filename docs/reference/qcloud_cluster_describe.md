## qcloud cluster describe

Describe a cluster

```
qcloud cluster describe <cluster-id> [flags]
```

### Examples

```
# Describe a cluster
qcloud cluster describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Output as JSON
qcloud cluster describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json
```

### Options

```
  -h, --help   help for describe
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

