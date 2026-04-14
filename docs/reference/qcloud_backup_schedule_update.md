## qcloud backup schedule update

Update a backup schedule

### Synopsis

Update a backup schedule.

The --cluster-id flag is required because the API requires the cluster ID to look up a schedule by ID.

```
qcloud backup schedule update <schedule-id> [flags]
```

### Options

```
      --cluster-id string       Cluster ID (required)
  -h, --help                    help for update
      --retention-days uint32   New retention period in days (1-365)
      --schedule string         New cron schedule expression in UTC, e.g. '0 2 * * *'
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

* [qcloud backup schedule](qcloud_backup_schedule.md)	 - Manage backup schedules

