## qcloud serverless space dataset list

List the sample datasets that can be loaded into a space

### Synopsis

List the sample datasets that can be loaded into a serverless space.

Only datasets available in the region of the space are listed. SIZE and
INDEXED SIZE are the storage the loaded collection uses; they count against the
collection size limits of the space. The VECTORS column lists the vector names
of the collection a load creates; "default" is the unnamed vector. Use
"dataset describe" to see the vector sizes, distances and embedding models.

```
qcloud serverless space dataset list <space-id> [flags]
```

### Examples

```
# List the sample datasets that can be loaded into a space
qcloud serverless space dataset list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space dataset list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json
```

### Options

```
  -h, --help[=true|false]         help for list
      --no-headers[=true|false]   Do not print column headers
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

