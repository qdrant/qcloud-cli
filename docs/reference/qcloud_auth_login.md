## qcloud auth login

Log in with a browser or device code

### Synopsis

Log in and store a refreshable access token for the current endpoint.

The default flow is authorization code + PKCE with a loopback listener on
127.0.0.1 (RFC 8252). Use --device on SSH or when no browser is available.

The login URL comes from the gateway's unauthenticated protected-resource
metadata when OAuth is enabled, otherwise from rewriting grpc.cloud.qdrant.io
to login.cloud.qdrant.io. Override with --login-url only for debugging.

The native application client id is not hard-coded yet.
Pass --client-id or QDRANT_CLOUD_OAUTH_CLIENT_ID after that app exists.
Tokens are stored in credentials.yaml next to the config file (mode 0600).

```
qcloud auth login [flags]
```

### Examples

```
# Browser login (default production endpoint)
qcloud auth login --client-id <id>

# Read-only session
qcloud auth login --scope read-only --client-id <id>

# Device code
qcloud auth login --device --client-id <id>
```

### Options

```
      --client-id string   OAuth native app client id (env: QDRANT_CLOUD_OAUTH_CLIENT_ID)
      --device             Use the device authorization grant instead of a loopback browser flow
  -h, --help               help for login
      --login-url string   Override the login issuer URL (defaults to discovery / endpoint inference)
      --scope string       OAuth scope: read-only or manage (default "manage")
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

