## qcloud serverless space update

Update an existing space

### Synopsis

Update an existing serverless space.

Use this command to rename a space or change its labels, network restrictions,
per-collection size limits, size warning threshold, search-worker settings, and
its storage and monthly spend limits. The cloud region cannot be changed after
creation.

Only the settings passed as flags are changed. Set --storage-limit or
--spend-limit to 'unlimited' to remove a limit. A limit's type or alert
threshold can be changed on its own once the limit is set. Lowering a limit
below current usage, switching a reached hard limit to soft, or removing it
applies to requests within about a minute.

Labels are merged with existing labels. Use 'key=value' to add or overwrite a
label, and 'key-' (with a trailing dash) to remove one. Allowed IPs and allowed
origins are merged the same way: specify a value to add it, or append '-' to
remove it (e.g. '10.0.0.0/8-').

```
qcloud serverless space update <space-id> [flags]
```

### Examples

```
# Rename a space
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --name my-renamed-space

# Add a label and remove another one
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --label env=prod --label team-

# Allow a new origin and remove an old IP range
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --allowed-origin https://app.example.com --allowed-ip 10.0.0.0/8-

# Tune the collection limits and search workers
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --max-collection-indexed-size 2GiB --collection-size-warning-percent 90 \
  --searcher-max-workers 4

# Cap monthly spend at $49.99 and alert at 90%
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --spend-limit 49.99 --spend-limit-alert-percent 90

# Switch the storage limit to a soft limit, keeping its size
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --storage-limit-type soft

# Remove the monthly spend limit
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --spend-limit unlimited
```

### Options

```
      --allowed-ip stringArray                   Allowed client IP CIDR range (e.g. "10.0.0.0/8"); append '-' to remove; IPv4 only
      --allowed-origin stringArray               Allowed browser origin for CORS (e.g. "https://app.example.com"); append '-' to remove; max 10
      --collection-size-warning-percent uint32   Percentage of the maximum collection size at which a size warning is sent (40 to 100)
      --cost-allocation-label string             Label for billing reports
  -h, --help[=true|false]                        help for update
      --label stringArray                        Label ('key=value') to add/overwrite; append '-' to remove ('key-'), can be specified multiple times
      --max-collection-indexed-size string       Maximum indexed size per collection (e.g. 1GiB, 512MiB); must not exceed the platform limit or the maximum collection size
      --max-collection-size string               Maximum size per collection (e.g. 10GiB, 512MiB); must not exceed the platform limit
      --name string                              New name of the space
      --searcher-max-workers uint                Maximum number of search workers per collection
      --spend-limit string                       Maximum monthly spend on searches and writes, in USD (e.g. 50 or 49.99), or 'unlimited' to remove the limit
      --spend-limit-alert-percent uint32         Percentage of the spend limit at which an alert is sent (1 to 99, default: 80)
      --spend-limit-type string                  What happens at the spend limit: 'hard' rejects all requests until next month, 'soft' only alerts (default: hard)
      --storage-limit string                     Maximum storage of the space (e.g. 10GiB, 512MiB), or 'unlimited' to remove the limit
      --storage-limit-alert-percent uint32       Percentage of the storage limit at which an alert is sent (1 to 99, default: 80)
      --storage-limit-type string                What happens at the storage limit: 'hard' rejects upserts, 'soft' only alerts (default: hard)
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

* [qcloud serverless space](qcloud_serverless_space.md)	 - Manage serverless spaces

