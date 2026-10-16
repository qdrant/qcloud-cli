## qcloud auth token

Print a fresh OAuth access token (sensitive)

### Synopsis

Print a fresh OAuth access token for the current endpoint.

The token is sensitive. Use this for local Terraform (auth = "cli") or scripts
that need a Bearer token. Expired access tokens are refreshed when a refresh
token is stored.

```
qcloud auth token [flags]
```

### Examples

```
qcloud auth token
qcloud auth token --json
```

### Options

```
  -h, --help   help for token
```

### Options inherited from parent commands

```
      --account-id string    Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)
      --api-key string       Management API Key (env: QDRANT_CLOUD_API_KEY)
  -c, --config string        Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)
      --console-url string   Qdrant Cloud web console base URL (env: QDRANT_CLOUD_CONSOLE_URL, default https://cloud.qdrant.io)
      --context string       Override the active context (env: QDRANT_CLOUD_CONTEXT)
      --debug                Enable debug logging to stderr
      --endpoint string      gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)
      --json                 Output as JSON
```

### SEE ALSO

* [qcloud auth](qcloud_auth.md)	 - Log in to Qdrant Cloud with a browser

