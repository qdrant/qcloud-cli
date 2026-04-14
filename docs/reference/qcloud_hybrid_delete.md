## qcloud hybrid delete

Delete a hybrid cloud environment

```
qcloud hybrid delete <env-id> [flags]
```

### Examples

```
# Delete a hybrid cloud environment (prompts for confirmation)
qcloud hybrid delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud hybrid delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force
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

* [qcloud hybrid](qcloud_hybrid.md)	 - Manage hybrid cloud environments

