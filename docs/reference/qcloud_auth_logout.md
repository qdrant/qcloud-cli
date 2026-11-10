## qcloud auth logout

Log out and drop stored OAuth tokens

### Synopsis

Log out of the current endpoint: remove its stored session and the contexts that use it.

The refresh token is revoked at the login server first. Logging out still
succeeds when that is not possible (the keychain entry is gone, or the login
server cannot be reached); qcloud then warns and removes the session anyway.
The access you granted stays listed under "Connected apps" in the console, where
you revoke it. Contexts with an API key are not changed.

```
qcloud auth logout [flags]
```

### Examples

```
# Log out of the current endpoint
qcloud auth logout

# Log out of another endpoint
qcloud auth logout --endpoint grpc.cloud.qdrant.io:443
```

### Options

```
  -h, --help[=true|false]   help for logout
```

### Options inherited from parent commands

```
      --account-id string    Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)
      --api-key string       Management API Key (env: QDRANT_CLOUD_API_KEY)
  -c, --config string        Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)
      --console-url string   Qdrant Cloud web console base URL (env: QDRANT_CLOUD_CONSOLE_URL, default https://cloud.qdrant.io)
      --context string       Override the active context (env: QDRANT_CLOUD_CONTEXT)
      --debug[=true|false]   Enable debug logging to stderr
      --endpoint string      gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)
      --json[=true|false]    Output as JSON
```

### SEE ALSO

* [qcloud auth](qcloud_auth.md)	 - Log in to Qdrant Cloud with a browser

