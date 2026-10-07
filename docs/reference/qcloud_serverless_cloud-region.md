## qcloud serverless cloud-region

Explore cloud regions for serverless spaces

### Synopsis

Explore the cloud regions in which serverless spaces can be created.

Serverless regions are identified by provider-agnostic IDs. Pass a region ID to
'qcloud serverless space create --cloud-region' to host a new space there. Only
regions marked as available accept new spaces.

### Examples

```
# List all serverless regions
qcloud serverless cloud-region list

# Show the details of a region
qcloud serverless cloud-region describe eu-central-1
```

### Options

```
  -h, --help   help for cloud-region
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

* [qcloud serverless](qcloud_serverless.md)	 - Manage Qdrant Cloud Serverless resources
* [qcloud serverless cloud-region describe](qcloud_serverless_cloud-region_describe.md)	 - Describe a cloud region for serverless spaces
* [qcloud serverless cloud-region list](qcloud_serverless_cloud-region_list.md)	 - List cloud regions for serverless spaces

