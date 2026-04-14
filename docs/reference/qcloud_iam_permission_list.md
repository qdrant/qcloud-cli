## qcloud iam permission list

List all available permissions

### Synopsis

List all permissions known in the system for the account.

Permissions are the individual access rights that can be assigned to roles.
Each permission has a value (e.g. "read:clusters") and a category
(e.g. "Cluster").

```
qcloud iam permission list [flags]
```

### Examples

```
# List all available permissions
qcloud iam permission list

# Output as JSON
qcloud iam permission list --json
```

### Options

```
  -h, --help         help for list
      --no-headers   Do not print column headers
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

* [qcloud iam permission](qcloud_iam_permission.md)	 - Manage permissions in Qdrant Cloud

