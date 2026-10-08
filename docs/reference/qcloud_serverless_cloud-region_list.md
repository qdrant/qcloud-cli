## qcloud serverless cloud-region list

List cloud regions for serverless spaces

### Synopsis

List the cloud regions in which serverless spaces can be created.

The AVAILABLE column shows whether a region currently accepts new spaces.

```
qcloud serverless cloud-region list [flags]
```

### Examples

```
# List all serverless regions
qcloud serverless cloud-region list

# List the IDs of the available regions
qcloud serverless cloud-region list --json | jq -r '.items[] | select(.available) | .id'
```

### Options

```
  -h, --help         help for list
      --no-headers   Do not print column headers
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

* [qcloud serverless cloud-region](qcloud_serverless_cloud-region.md)	 - Explore cloud regions for serverless spaces

