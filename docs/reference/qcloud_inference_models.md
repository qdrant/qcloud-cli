## qcloud inference models

Manage inference models

### Synopsis

Inspect the individual inference models offered by Qdrant Cloud.

A model is identified by its name (for example "cohere/*"), and describes the
vectors it produces: the vector type, the modality of the input it accepts, and
the dimensionality of its output.

### Examples

```
# List the inference models available in a region
qcloud inference models list --cloud-provider aws --cloud-region eu-central-1
```

### Options

```
  -h, --help   help for models
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

* [qcloud inference](qcloud_inference.md)	 - Manage inference resources
* [qcloud inference models list](qcloud_inference_models_list.md)	 - List available inference models

