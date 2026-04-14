## qcloud account

Manage Qdrant Cloud accounts

### Synopsis

Manage Qdrant Cloud accounts and their members.

Use these commands to list, inspect, and update accounts that the current
management key has access to. Account member commands show who belongs to the
current account and whether they are the owner.

### Options

```
  -h, --help   help for account
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

* [qcloud](qcloud.md)	 - Qdrant Cloud CLI
* [qcloud account describe](qcloud_account_describe.md)	 - Describe an account
* [qcloud account list](qcloud_account_list.md)	 - List accounts
* [qcloud account member](qcloud_account_member.md)	 - Manage account members
* [qcloud account update](qcloud_account_update.md)	 - Update an account

