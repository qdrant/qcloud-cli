## qcloud serverless space key delete

Delete an API key from a space

### Synopsis

Delete an API key from a serverless space.

Requests authenticated with the key are rejected once the deletion has been
propagated to the space. Deletion cannot be undone.

```
qcloud serverless space key delete <space-id> <key-id> [flags]
```

### Examples

```
# Delete an API key (prompts for confirmation)
qcloud serverless space key delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Delete an API key without confirmation
qcloud serverless space key delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 a1b2c3d4-e5f6-7890-abcd-ef1234567890 --force
```

### Options

```
  -f, --force   Skip confirmation prompt
  -h, --help    help for delete
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

* [qcloud serverless space key](qcloud_serverless_space_key.md)	 - Manage API keys for a space

