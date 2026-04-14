## qcloud package list

List available packages for cluster creation

```
qcloud package list [flags]
```

### Examples

```
# List packages for a cloud provider and region
qcloud package list --cloud-provider aws --cloud-region eu-central-1

# List packages for a hybrid cloud provider (no region required)
qcloud package list --cloud-provider hybrid
```

### Options

```
      --cloud-provider string   Cloud provider ID (required)
      --cloud-region string     Cloud provider region ID (required for non-hybrid providers)
  -h, --help                    help for list
      --no-headers              Do not print column headers
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

* [qcloud package](qcloud_package.md)	 - Manage packages

