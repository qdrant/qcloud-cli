## qcloud account member

Manage account members

### Synopsis

Manage members of the current Qdrant Cloud account.

Members are users who have been added to the account. Each member has an
associated user record and an ownership flag indicating whether they are the
account owner.

### Options

```
  -h, --help   help for member
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

* [qcloud account](qcloud_account.md)	 - Manage Qdrant Cloud accounts
* [qcloud account member describe](qcloud_account_member_describe.md)	 - Describe an account member
* [qcloud account member list](qcloud_account_member_list.md)	 - List account members

