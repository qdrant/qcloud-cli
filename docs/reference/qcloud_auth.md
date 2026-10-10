## qcloud auth

Log in to Qdrant Cloud with a browser

### Synopsis

Log in to Qdrant Cloud with a browser (or a device code) instead of pasting an API key.

The login host is not a separate flag. It is derived from the gRPC
endpoint (flag --endpoint, env QDRANT_CLOUD_ENDPOINT, or the active context):

  grpc.cloud.qdrant.io:443  →  login.cloud.qdrant.io

When OAuth is enabled, the gateway also advertises the issuer at the
unauthenticated URL https://api.cloud.qdrant.io/.well-known/oauth-protected-resource.
That document wins over hostname inference.

After login, API calls send Authorization: Bearer <token>. Management API keys
keep working unchanged when --api-key / QDRANT_CLOUD_API_KEY is set.

### Examples

```
# Discover the login URL for the current endpoint
qcloud auth discover

# Browser login (authorization code + PKCE on 127.0.0.1)
qcloud auth login --client-id <native-app>

# Device code (SSH / no browser)
qcloud auth login --device --scope read-only
```

### Options

```
  -h, --help   help for auth
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

* [qcloud](qcloud.md)	 - Qdrant Cloud CLI
* [qcloud auth discover](qcloud_auth_discover.md)	 - Show the OAuth login URL inferred from the API endpoint
* [qcloud auth login](qcloud_auth_login.md)	 - Log in with a browser or device code
* [qcloud auth logout](qcloud_auth_logout.md)	 - Log out and drop stored OAuth tokens
* [qcloud auth status](qcloud_auth_status.md)	 - Show OAuth login status for the current endpoint
* [qcloud auth token](qcloud_auth_token.md)	 - Print a fresh OAuth access token (sensitive)

