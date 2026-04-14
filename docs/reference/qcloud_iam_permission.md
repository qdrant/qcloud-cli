## qcloud iam permission

Manage permissions in Qdrant Cloud

### Synopsis

Manage permissions for the Qdrant Cloud account.

Permissions represent individual access rights that can be assigned to roles.
Use these commands to discover which permissions are available in the system.

### Options

```
  -h, --help   help for permission
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

* [qcloud iam](qcloud_iam.md)	 - Manage IAM resources in Qdrant Cloud
* [qcloud iam permission list](qcloud_iam_permission_list.md)	 - List all available permissions

