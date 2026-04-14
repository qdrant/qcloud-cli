## qcloud iam role

Manage roles in Qdrant Cloud

### Synopsis

Manage roles for the Qdrant Cloud account.

Roles define sets of permissions that control access to resources. There are two
types of roles: system roles (immutable, managed by Qdrant) and custom roles
(created and managed by the account). Use these commands to list, inspect, create,
update, and delete custom roles, as well as manage their permissions.

### Options

```
  -h, --help   help for role
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
* [qcloud iam role assign-permission](qcloud_iam_role_assign-permission.md)	 - Add permissions to a role
* [qcloud iam role create](qcloud_iam_role_create.md)	 - Create a custom role
* [qcloud iam role delete](qcloud_iam_role_delete.md)	 - Delete a custom role
* [qcloud iam role describe](qcloud_iam_role_describe.md)	 - Describe a role
* [qcloud iam role list](qcloud_iam_role_list.md)	 - List all roles
* [qcloud iam role remove-permission](qcloud_iam_role_remove-permission.md)	 - Remove permissions from a role
* [qcloud iam role update](qcloud_iam_role_update.md)	 - Update a custom role

