## qcloud iam role assign-permission

Add permissions to a role

### Synopsis

Add permissions to a custom role.

Fetches the role's current permissions, merges the new ones (deduplicating),
and updates the role. Use "qcloud iam permission list" to see available
permissions.

```
qcloud iam role assign-permission <role-id> [flags]
```

### Examples

```
# Add a single permission
qcloud iam role assign-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe --permission read:clusters

# Add multiple permissions
qcloud iam role assign-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --permission read:clusters --permission read:backups
```

### Options

```
  -h, --help                 help for assign-permission
      --permission strings   Permission to add (repeatable)
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

