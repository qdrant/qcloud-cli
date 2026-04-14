## qcloud backup schedule create

Create a backup schedule for a cluster

```
qcloud backup schedule create [flags]
```

### Options

```
      --cluster-id string       Cluster ID (required)
  -h, --help                    help for create
      --retention-days uint32   Retention period in days (1-365) (required)
      --schedule string         Cron schedule expression in UTC (required), e.g. '0 2 * * *'
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

