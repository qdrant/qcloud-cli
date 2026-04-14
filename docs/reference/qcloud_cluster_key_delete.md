## qcloud cluster key delete

Delete an API key from a cluster

```
qcloud cluster key delete <cluster-id> <key-id> [flags]
```

### Examples

```
# Delete an API key
qcloud cluster key delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

### Options

```
  -f, --force   Skip confirmation prompt
  -h, --help    help for delete
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

* [qcloud cluster key](qcloud_cluster_key.md)	 - Manage API keys for a cluster

