## qcloud auth status

Show OAuth login status for the current endpoint

### Synopsis

Show whether an OAuth session exists for the current endpoint.

Prints the inferred login URL, token expiry, and scope. Does not print the
access token.

```
qcloud auth status [flags]
```

### Examples

```
qcloud auth status
qcloud auth status --endpoint grpc.staging-cloud.qdrant.io:443
```

### Options

```
  -h, --help   help for status
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

