## qcloud iam role list

List all roles

### Synopsis

List all roles for the account, including both system and custom roles.

System roles are managed by Qdrant and cannot be modified. Custom roles are
created and managed by the account administrator.

```
qcloud iam role list [flags]
```

### Examples

```
# List all roles
qcloud iam role list

# Output as JSON
qcloud iam role list --json
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

* [qcloud iam role](qcloud_iam_role.md)	 - Manage roles in Qdrant Cloud

