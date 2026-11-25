## qcloud serverless space dataset load describe

Describe a sample dataset load

### Synopsis

Describe a sample dataset load of a serverless space.

Shows the status of the load, its progress while it is running and how long it
took once it finished. When a load failed, Reason explains why, for example
because the space reached its collection limit.

```
qcloud serverless space dataset load describe <space-id> <load-id> [flags]
```

### Examples

```
# Describe a sample dataset load
qcloud serverless space dataset load describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 7f6e5d4c-3b2a-4190-8f7e-6d5c4b3a2918

# Output as JSON
qcloud serverless space dataset load describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 7f6e5d4c-3b2a-4190-8f7e-6d5c4b3a2918 --json
```

### Options

```
  -h, --help[=true|false]   help for describe
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

* [qcloud serverless space dataset load](qcloud_serverless_space_dataset_load.md)	 - Manage sample dataset loads of a space

