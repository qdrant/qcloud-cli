## qcloud iam role create

Create a custom role

### Synopsis

Create a new custom role for the account.

Custom roles allow fine-grained access control by combining specific permissions.
Use "qcloud iam permission list" to see available permissions.

```
qcloud iam role create [flags]
```

### Examples

```
# Create a role with specific permissions
qcloud iam role create --name "Cluster Viewer" --permission read:clusters --permission read:cluster-endpoints

# Create a role with a description
qcloud iam role create --name "Backup Manager" --description "Can manage backups" \
  --permission read:clusters --permission read:backups --permission write:backups
```

### Options

```
      --description string   Description of the role
  -h, --help                 help for create
      --name string          Name of the role (4-64 characters)
      --permission strings   Permission to assign (repeatable)
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

