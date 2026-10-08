# qcloud backup

Manage Qdrant Cloud backups

Commands:

- `qcloud backup create`: Create a backup for a cluster
- `qcloud backup delete`: Delete a backup
- `qcloud backup describe`: Describe a backup
- `qcloud backup list`: List backups
- `qcloud backup restore list`: List backup restores
- `qcloud backup restore trigger`: Trigger a restore from a backup
- `qcloud backup schedule create`: Create a backup schedule for a cluster
- `qcloud backup schedule delete`: Delete a backup schedule
- `qcloud backup schedule describe`: Describe a backup schedule
- `qcloud backup schedule list`: List backup schedules
- `qcloud backup schedule update`: Update a backup schedule

## qcloud backup create

Create a backup for a cluster

```
qcloud backup create [flags]
```

Flags:

```
      --cluster-id string       Cluster ID to back up (required)
      --retention-days uint32   Retention period in days (1-365) (required)
```

## qcloud backup delete

Delete a backup

```
qcloud backup delete <backup-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

## qcloud backup describe

Describe a backup

```
qcloud backup describe <backup-id>
```

## qcloud backup list

List backups in the current account.

Backups can be filtered by cluster.

By default, all backups are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.

```
qcloud backup list [flags]
```

Flags:

```
      --cluster-id string   Filter by cluster ID
      --no-headers          Do not print column headers
      --page-size int32     Maximum number of backups to return per page (manual pagination mode)
      --page-token string   Page token from a previous response to resume from (manual pagination mode)
```

Examples:

```sh
# List all backups in the account
qcloud backup list

# List backups of a cluster
qcloud backup list --cluster-id 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Manual pagination
qcloud backup list --page-size 10 --json
```

## qcloud backup restore list

List backup restore operations in the current account.

Restores can be filtered by cluster.

By default, all restores are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.

```
qcloud backup restore list [flags]
```

Flags:

```
      --cluster-id string   Filter by cluster ID
      --no-headers          Do not print column headers
      --page-size int32     Maximum number of restores to return per page (manual pagination mode)
      --page-token string   Page token from a previous response to resume from (manual pagination mode)
```

Examples:

```sh
# List all backup restores in the account
qcloud backup restore list

# List restores of a cluster
qcloud backup restore list --cluster-id 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Manual pagination
qcloud backup restore list --page-size 10 --json
```

## qcloud backup restore trigger

Trigger a restore from a backup

```
qcloud backup restore trigger <backup-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

## qcloud backup schedule create

Create a backup schedule for a cluster

```
qcloud backup schedule create [flags]
```

Flags:

```
      --cluster-id string       Cluster ID (required)
      --retention-days uint32   Retention period in days (1-365) (required)
      --schedule string         Cron schedule expression in UTC (required), e.g. '0 2 * * *'
```

## qcloud backup schedule delete

Delete a backup schedule

```
qcloud backup schedule delete <schedule-id> [flags]
```

Flags:

```
      --delete-backups   Also delete all backups created by this schedule
  -f, --force            Skip confirmation prompt
```

## qcloud backup schedule describe

Describe a backup schedule.

The --cluster-id flag is required because the API requires the cluster ID to look up a schedule by ID.

```
qcloud backup schedule describe <schedule-id> [flags]
```

Flags:

```
      --cluster-id string   Cluster ID (required)
```

## qcloud backup schedule list

List backup schedules in the current account.

Schedules can be filtered by cluster. The NEXT RUN column is computed locally
from the cron expression.

By default, all schedules are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.

```
qcloud backup schedule list [flags]
```

Flags:

```
      --cluster-id string   Filter by cluster ID
      --no-headers          Do not print column headers
      --page-size int32     Maximum number of schedules to return per page (manual pagination mode)
      --page-token string   Page token from a previous response to resume from (manual pagination mode)
```

Examples:

```sh
# List all backup schedules in the account
qcloud backup schedule list

# List schedules of a cluster
qcloud backup schedule list --cluster-id 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Manual pagination
qcloud backup schedule list --page-size 10 --json
```

## qcloud backup schedule update

Update a backup schedule.

The --cluster-id flag is required because the API requires the cluster ID to look up a schedule by ID.

```
qcloud backup schedule update <schedule-id> [flags]
```

Flags:

```
      --cluster-id string       Cluster ID (required)
      --retention-days uint32   New retention period in days (1-365)
      --schedule string         New cron schedule expression in UTC, e.g. '0 2 * * *'
```
