## qcloud serverless space dataset describe

Describe a sample dataset

### Synopsis

Describe a sample dataset that can be loaded into a serverless space.

Shows the size of the dataset and the configuration of the collection a load
creates: its dense and sparse vectors, and the embedding model that made each
of them. Use the same model to make query vectors.

```
qcloud serverless space dataset describe <space-id> <dataset-id> [flags]
```

### Examples

```
# Describe a sample dataset
qcloud serverless space dataset describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles

# Output as JSON
qcloud serverless space dataset describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles --json
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

* [qcloud serverless space dataset](qcloud_serverless_space_dataset.md)	 - Load sample datasets into a space

