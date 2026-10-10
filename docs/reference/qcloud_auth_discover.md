## qcloud auth discover

Show the OAuth login URL inferred from the API endpoint

### Synopsis

Print the OAuth API resource and authorization-server issuer for the current endpoint.

The gRPC endpoint (default grpc.cloud.qdrant.io:443) is turned into
https://api.cloud.qdrant.io. The CLI then GETs
/.well-known/oauth-protected-resource with no credentials. If the gateway
returns authorization_servers, that issuer is used; otherwise
login.cloud.qdrant.io is inferred. No client id is required.

```
qcloud auth discover [flags]
```

### Examples

```
qcloud auth discover
```

### Options

```
  -h, --help   help for discover
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

