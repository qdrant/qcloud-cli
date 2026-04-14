## qcloud account list

List accounts

### Synopsis

List all accounts associated with the authenticated management key.

Returns every account the current API key has access to. No account ID is
required because the server resolves accounts from the caller's credentials.

```
qcloud account list [flags]
```

### Examples

```
# List all accessible accounts
qcloud account list

# Output as JSON
qcloud account list --json
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

* [qcloud account](qcloud_account.md)	 - Manage Qdrant Cloud accounts

