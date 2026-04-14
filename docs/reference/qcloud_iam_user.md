## qcloud iam user

Manage users in Qdrant Cloud

### Synopsis

Manage users in the Qdrant Cloud account.

Provides commands to list users, view user details and assigned roles, and
manage role assignments.

### Options

```
  -h, --help   help for user
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
* [qcloud iam user assign-role](qcloud_iam_user_assign-role.md)	 - Assign one or more roles to a user
* [qcloud iam user describe](qcloud_iam_user_describe.md)	 - Describe a user and their assigned roles
* [qcloud iam user list](qcloud_iam_user_list.md)	 - List users in the account
* [qcloud iam user remove-role](qcloud_iam_user_remove-role.md)	 - Remove one or more roles from a user

