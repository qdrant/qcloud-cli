## qcloud iam key create

Create a cloud management key

### Synopsis

Create a new cloud management key for the account.

Management keys grant access to the Qdrant Cloud API. The full key value is returned
only once at creation time — store it securely, as it cannot be retrieved again. If a
key is lost, delete it and create a new one.

```
qcloud iam key create [flags]
```

### Examples

```
# Create a new management key
qcloud iam key create

# Create and capture the key value in a script
qcloud iam key create --json | jq -r '.key'
```

### Options

```
  -h, --help   help for create
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

* [qcloud iam key](qcloud_iam_key.md)	 - Manage cloud management keys

