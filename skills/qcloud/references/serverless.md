# qcloud serverless

Manage Qdrant Cloud Serverless resources.

Qdrant Cloud Serverless runs collections in spaces instead of dedicated clusters.
A space is hosted in a single cloud region and scales its search workers
automatically, so there are no nodes, packages or disks to size. Use the
commands in this group to manage spaces together with their API keys and
backups, to explore the regions spaces can be created in, and to inspect the
account's serverless quota.

Commands:

- `qcloud serverless cloud-region describe`: Describe a cloud region for serverless spaces
- `qcloud serverless cloud-region list`: List cloud regions for serverless spaces
- `qcloud serverless quota`: Show the serverless quota of the account
- `qcloud serverless space alert list`: List alerts of a space
- `qcloud serverless space backup create`: Create a backup of a serverless space
- `qcloud serverless space backup delete`: Delete a backup of a serverless space
- `qcloud serverless space backup describe`: Describe a backup of a serverless space
- `qcloud serverless space backup list`: List backups of serverless spaces
- `qcloud serverless space backup restore list`: List restores of serverless space backups
- `qcloud serverless space backup restore trigger`: Restore a backup into its original space
- `qcloud serverless space backup schedule create`: Create a backup schedule for a serverless space
- `qcloud serverless space backup schedule delete`: Delete a backup schedule of a serverless space
- `qcloud serverless space backup schedule describe`: Describe a backup schedule of a serverless space
- `qcloud serverless space backup schedule list`: List backup schedules of serverless spaces
- `qcloud serverless space backup schedule update`: Update a backup schedule of a serverless space
- `qcloud serverless space create`: Create a new space
- `qcloud serverless space create-from-backup`: Create a new space from a backup
- `qcloud serverless space delete`: Delete a space
- `qcloud serverless space describe`: Describe a space
- `qcloud serverless space key create`: Create an API key for a space
- `qcloud serverless space key delete`: Delete an API key from a space
- `qcloud serverless space key list`: List API keys for a space
- `qcloud serverless space list`: List all spaces
- `qcloud serverless space metrics summary`: Show a metrics overview of a space's collections
- `qcloud serverless space metrics usage`: Show usage metrics of a space's collections over time
- `qcloud serverless space suggest-name`: Suggest a name for a new space
- `qcloud serverless space update`: Update an existing space
- `qcloud serverless space wait`: Wait for a space to become ready

## qcloud serverless cloud-region describe

Describe a cloud region in which serverless spaces can be created.

Shows the region's display name, whether it currently accepts new spaces, and
its geographical location.

```
qcloud serverless cloud-region describe <region-id>
```

Examples:

```sh
# Describe a region
qcloud serverless cloud-region describe eu-central-1

# Output as JSON
qcloud serverless cloud-region describe eu-central-1 --json
```

## qcloud serverless cloud-region list

List the cloud regions in which serverless spaces can be created.

The AVAILABLE column shows whether a region currently accepts new spaces.

```
qcloud serverless cloud-region list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List all serverless regions
qcloud serverless cloud-region list

# List the IDs of the available regions
qcloud serverless cloud-region list --json | jq -r '.items[] | select(.available) | .id'
```

## qcloud serverless quota

Show the serverless quota of the current account.

The quota caps how many spaces the account can create and sets the platform
limits that are copied into every space it owns: the maximum number of
collections per space, the maximum size of a collection and the maximum number
of search workers per collection. A limit of 0 is shown as "unlimited".

```
qcloud serverless quota
```

Examples:

```sh
# Show the serverless quota
qcloud serverless quota

# Output as JSON
qcloud serverless quota --json
```

## qcloud serverless space alert list

List the alerts of a serverless space, most recently firing first.

Alerts that are tied to the space as a whole rather than to a single collection
are shown with "(space)" in the COLLECTION column. Use --space-only to list only
those alerts, or --collection / --collection-contains to list only the alerts of
matching collections.

By default, all alerts are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.

```
qcloud serverless space alert list <space-id> [flags]
```

Flags:

```
      --collection string            Only include the collection with this exact name
      --collection-contains string   Only include collections whose name contains this substring
      --no-headers                   Do not print column headers
      --page-size int32              Maximum number of alerts to return per page (manual pagination mode)
      --page-token string            Page token from a previous response to resume from (manual pagination mode)
      --space-only                   Only include entries that are not tied to a collection
      --state string                 Only list alerts in this state (firing, resolved)
```

Examples:

```sh
# List all alerts of a space
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# List only the alerts that are currently firing
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --state firing

# List the alerts of a single collection
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --collection products

# List alerts that concern the space as a whole
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --space-only

# Show the alert descriptions as JSON
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json
```

## qcloud serverless space backup create

Create an on-demand backup of a serverless space.

By default the whole space is backed up; use --collection to back up a single
collection. Without --retention-days the backup is kept until it is deleted.
The backup runs asynchronously; use "qcloud serverless space backup describe" to
follow its status and progress.

```
qcloud serverless space backup create [flags]
```

Flags:

```
      --collection string       Only back up this collection (default: the whole space)
      --retention-days uint32   Retention period in days (1-365) (default: keep indefinitely)
      --space-id string         ID of the space to back up (required)
```

Examples:

```sh
# Back up a whole space
qcloud serverless space backup create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Back up a single collection and keep the backup for 7 days
qcloud serverless space backup create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --collection products --retention-days 7
```

## qcloud serverless space backup delete

Delete a backup of a serverless space.

Deletion cannot be undone; the backup can no longer be restored or used to
create a new space.

```
qcloud serverless space backup delete <backup-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete a backup (prompts for confirmation)
qcloud serverless space backup delete 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d

# Delete a backup without confirmation
qcloud serverless space backup delete 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d --force
```

## qcloud serverless space backup describe

Describe a backup of a serverless space.

Shows the backup status and statistics together with a snapshot of the space
(name, region and configuration) taken when the backup was created. That snapshot
is used when a new space is created from the backup.

```
qcloud serverless space backup describe <backup-id>
```

Examples:

```sh
# Describe a backup
qcloud serverless space backup describe 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d

# Output as JSON
qcloud serverless space backup describe 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d --json
```

## qcloud serverless space backup list

List backups of serverless spaces in the current account.

Backups can be filtered by space, by the schedule that created them, and by
collection. The COLLECTION column shows "(all)" for backups of a whole space.

By default, all backups are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.

```
qcloud serverless space backup list [flags]
```

Flags:

```
      --collection string    Filter by collection name
      --no-headers           Do not print column headers
      --page-size int32      Maximum number of backups to return per page (manual pagination mode)
      --page-token string    Page token from a previous response to resume from (manual pagination mode)
      --schedule-id string   Filter by the ID of the backup schedule that created the backups
      --space-id string      Filter by space ID
```

Examples:

```sh
# List all backups in the account
qcloud serverless space backup list

# List backups of a space
qcloud serverless space backup list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# List backups of a single collection created by a schedule
qcloud serverless space backup list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --schedule-id 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c --collection products
```

## qcloud serverless space backup restore list

List restores of serverless space backups in the current account.

The PROGRESS column is reported while a restore is running and may remain after
it completes. By default, all restores are fetched automatically across multiple
pages. Use --page-size and --page-token for manual pagination.

```
qcloud serverless space backup restore list [flags]
```

Flags:

```
      --no-headers          Do not print column headers
      --page-size int32     Maximum number of restores to return per page (manual pagination mode)
      --page-token string   Page token from a previous response to resume from (manual pagination mode)
      --space-id string     Filter by space ID
```

Examples:

```sh
# List all restores in the account
qcloud serverless space backup restore list

# List restores of a space
qcloud serverless space backup restore list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60
```

## qcloud serverless space backup restore trigger

Restore a backup into its original serverless space.

The restore runs asynchronously and replaces the current data of the backed-up
collections in the space. Use "qcloud serverless space backup restore list" to
follow its progress.

```
qcloud serverless space backup restore trigger <backup-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Restore a backup (prompts for confirmation)
qcloud serverless space backup restore trigger 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d

# Restore a backup without confirmation
qcloud serverless space backup restore trigger 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d --force
```

## qcloud serverless space backup schedule create

Create a backup schedule for a serverless space.

The --schedule flag takes a standard cron expression evaluated in UTC (for example
"0 2 * * *" for every day at 02:00), or a descriptor such as "@daily". By default
the whole space is backed up; use --collection to back up a single collection.
Without --retention-days, created backups are kept until they are deleted.

```
qcloud serverless space backup schedule create [flags]
```

Flags:

```
      --collection string       Only back up this collection (default: the whole space)
      --name string             Name of the schedule (required)
      --retention-days uint32   Retention period of created backups in days (1-365) (default: keep indefinitely)
      --schedule string         Cron schedule expression in UTC, e.g. '0 2 * * *' (required)
      --space-id string         ID of the space to back up (required)
```

Examples:

```sh
# Back up a space every night at 02:00 UTC
qcloud serverless space backup schedule create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name nightly --schedule "0 2 * * *"

# Back up a single collection every hour and keep backups for 7 days
qcloud serverless space backup schedule create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name products-hourly --schedule @hourly --collection products --retention-days 7
```

## qcloud serverless space backup schedule delete

Delete a backup schedule of a serverless space.

The schedule stops creating backups. Backups it already created are kept by
default; pass --delete-backups to remove them as well.

```
qcloud serverless space backup schedule delete <schedule-id> [flags]
```

Flags:

```
      --delete-backups   Also delete all backups created by this schedule
  -f, --force            Skip confirmation prompt
```

Examples:

```sh
# Delete a schedule (prompts for confirmation)
qcloud serverless space backup schedule delete 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c

# Delete a schedule and all backups it created, without confirmation
qcloud serverless space backup schedule delete 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c --delete-backups --force
```

## qcloud serverless space backup schedule describe

Describe a backup schedule of a serverless space.

Shows the cron expression together with the next time it fires, the retention
period applied to created backups, and whether the schedule is paused. The
--space-id flag is required because the API looks up schedules within a space.

```
qcloud serverless space backup schedule describe <schedule-id> [flags]
```

Flags:

```
      --space-id string   ID of the space the schedule belongs to (required)
```

Examples:

```sh
# Describe a backup schedule
qcloud serverless space backup schedule describe 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60
```

## qcloud serverless space backup schedule list

List backup schedules of serverless spaces in the current account.

The PAUSED column shows "scheduled" for schedules with a pause set in the future.
By default, all schedules are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination.

```
qcloud serverless space backup schedule list [flags]
```

Flags:

```
      --no-headers          Do not print column headers
      --page-size int32     Maximum number of schedules to return per page (manual pagination mode)
      --page-token string   Page token from a previous response to resume from (manual pagination mode)
      --space-id string     Filter by space ID
```

Examples:

```sh
# List all backup schedules in the account
qcloud serverless space backup schedule list

# List the backup schedules of a space
qcloud serverless space backup schedule list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60
```

## qcloud serverless space backup schedule update

Update a backup schedule of a serverless space.

Only the fields whose flags are given are changed. --pause stops the schedule from
creating new backups immediately and --resume starts it again; the schedule and
its existing backups are kept while paused. Pausing an already paused schedule
keeps its original pause time, while a pause scheduled for the future is brought
forward to now. The --space-id flag is required
because the API looks up schedules within a space.

```
qcloud serverless space backup schedule update <schedule-id> [flags]
```

Flags:

```
      --name string             New name of the schedule
      --pause                   Pause the schedule
      --resume                  Resume a paused schedule
      --retention-days uint32   New retention period of created backups in days (1-365)
      --schedule string         New cron schedule expression in UTC, e.g. '0 2 * * *'
      --space-id string         ID of the space the schedule belongs to (required)
```

Examples:

```sh
# Change the schedule to run every 6 hours
qcloud serverless space backup schedule update 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --schedule "0 */6 * * *"

# Pause a schedule
qcloud serverless space backup schedule update 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --pause

# Resume a schedule and change its retention
qcloud serverless space backup schedule update 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --resume --retention-days 30
```

## qcloud serverless space create

Create a new serverless space in a cloud region.

The region is fixed for the lifetime of the space. When --name is omitted, a
unique, human-friendly name is generated by the server. Collection size limits
and search-worker settings default to the platform values and can be tightened
later with "qcloud serverless space update".

The space is provisioned asynchronously. Use --wait to block until it is ready
and its endpoint is available.

```
qcloud serverless space create [flags]
```

Flags:

```
      --allowed-ip stringArray           Allowed client IP CIDR range (e.g. "10.0.0.0/8"); append '-' to remove; IPv4 only
      --allowed-origin stringArray       Allowed browser origin for CORS (e.g. "https://app.example.com"); append '-' to remove; max 10
      --cloud-region string              Cloud region ID to host the space in (required)
      --cost-allocation-label string     Label for billing reports
      --label stringArray                Label ('key=value') to add/overwrite; append '-' to remove ('key-'), can be specified multiple times
      --max-collection-size string       Maximum size per collection (e.g. 10GiB, 512MiB); must not exceed the platform limit
      --name string                      Name of the space (default: generated by the server)
      --searcher-idle-timeout duration   Idle timeout after which search workers may be scaled down (1m to 15m)
      --searcher-max-workers uint        Maximum number of search workers per collection
      --wait                             Wait for the space to become ready
      --wait-timeout duration            Maximum time to wait for the space to become ready (default 10m0s)
```

Examples:

```sh
# Create a space with a generated name
qcloud serverless space create --cloud-region eu-central-1

# Create a named space and wait until it is ready
qcloud serverless space create --cloud-region eu-central-1 --name my-space --wait

# Create a space restricted to an office network and a web app origin
qcloud serverless space create --cloud-region eu-central-1 --name my-space \
  --allowed-ip 203.0.113.0/24 --allowed-origin https://app.example.com

# Create a space with labels and per-collection limits
qcloud serverless space create --cloud-region eu-central-1 --name my-space \
  --label env=staging --max-collection-size 10GiB --searcher-max-workers 2
```

## qcloud serverless space create-from-backup

Create a new serverless space seeded with the data from an existing backup.

The new space is provisioned in the same region and with the same configuration
as the original space at the time the backup was taken. The backup must belong
to the current account. To restore a backup into its original space instead, use
"qcloud serverless space backup restore trigger".

```
qcloud serverless space create-from-backup [flags]
```

Flags:

```
      --backup-id string        ID of the backup to restore from (required)
      --name string             Name for the new space (required)
      --wait                    Wait for the space to become ready
      --wait-timeout duration   Maximum time to wait for the space to become ready (default 10m0s)
```

Examples:

```sh
# Create a space from a backup
qcloud serverless space create-from-backup --backup-id 9d8c7b6a-5e4f-4a3b-8c2d-1e0f9a8b7c6d --name my-restored-space

# Create a space from a backup and wait until it is ready
qcloud serverless space create-from-backup --backup-id 9d8c7b6a-5e4f-4a3b-8c2d-1e0f9a8b7c6d --name my-restored-space --wait
```

## qcloud serverless space delete

Delete a serverless space and all of its collections.

Deletion cannot be undone. Backups of the space are kept by default so that the
data can still be restored into a new space with "create-from-backup"; pass
--delete-backups to remove them as well.

```
qcloud serverless space delete <space-id> [flags]
```

Flags:

```
      --delete-backups   Also delete all backups of the space
  -f, --force            Skip confirmation prompt
```

Examples:

```sh
# Delete a space (prompts for confirmation)
qcloud serverless space delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Delete a space and its backups without confirmation
qcloud serverless space delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --delete-backups --force
```

## qcloud serverless space describe

Describe a serverless space.

Shows the space's phase, region and endpoint together with its configuration:
network restrictions, per-collection size limits and search-worker settings.
Limits prefixed with "platform" are derived from the account's quota and cannot
be changed directly.

```
qcloud serverless space describe <space-id>
```

Examples:

```sh
# Describe a space
qcloud serverless space describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json
```

## qcloud serverless space key create

Create an API key for a serverless space.

A key grants either global access to the whole space (--access-type) or access to
individual collections (--collection, repeatable). The two kinds of rules cannot be
combined in one key. When neither flag is given, the server assigns global manage
access.

An expiration date given with --expires is inclusive: the key stays valid until
the end of that day (23:59:59 UTC).

The secret key value is printed only once. Store it securely; it cannot be
retrieved later. If --wait fails after the key was created, the secret is still
printed before the error is returned.

```
qcloud serverless space key create <space-id> [flags]
```

Flags:

```
      --access-type string       Global access type: manage, read-only or metrics-read-only (default: server assigns manage)
      --collection stringArray   Collection access rule as 'name=read-only|read-write'; can be specified multiple times
      --expires string           Expiration date in YYYY-MM-DD format; the key is valid until the end of that day (UTC)
      --name string              Name of the API key (required)
      --wait                     Wait for the API key to become ready
      --wait-timeout duration    Maximum time to wait for the API key to become ready (default 1m0s)
```

Examples:

```sh
# Create an API key with manage access
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --name my-key

# Create a read-only key that expires at the end of the year
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name read-key --access-type read-only --expires 2026-12-31

# Create a key scoped to two collections
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name app-key --collection products=read-write --collection reviews=read-only

# Create a key and wait for it to become ready
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --name my-key --wait
```

## qcloud serverless space key delete

Delete an API key from a serverless space.

Requests authenticated with the key are rejected once the deletion has been
propagated to the space. Deletion cannot be undone.

```
qcloud serverless space key delete <space-id> <key-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete an API key (prompts for confirmation)
qcloud serverless space key delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Delete an API key without confirmation
qcloud serverless space key delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 a1b2c3d4-e5f6-7890-abcd-ef1234567890 --force
```

## qcloud serverless space key list

List the API keys of a serverless space.

The ACCESS column shows the global access type of a key, or the per-collection
access rules as "collection:ACCESS" pairs. Secret key values are never listed;
use the POSTFIX column to identify a key.

```
qcloud serverless space key list <space-id> [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List API keys for a space
qcloud serverless space key list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space key list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json
```

## qcloud serverless space list

List all serverless spaces in the current account.

By default, all spaces are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.

Use --cloud-region to only list the spaces hosted in a specific region.

```
qcloud serverless space list [flags]
```

Flags:

```
      --cloud-region string   Filter by cloud region ID
      --no-headers            Do not print column headers
      --page-size int32       Maximum number of spaces to return per page (manual pagination mode)
      --page-token string     Page token from a previous response to resume from (manual pagination mode)
```

Examples:

```sh
# List all spaces
qcloud serverless space list

# List spaces in a specific region
qcloud serverless space list --cloud-region eu-central-1

# List spaces in JSON format
qcloud serverless space list --json

# Manual pagination
qcloud serverless space list --page-size 10
```

## qcloud serverless space metrics summary

Show a metrics overview of the collections in a serverless space.

For every collection, the overview reports the current point count, storage
usage and number of search workers, together with the search rate, write rate
and average search latency. The server averages the rates and latency over
several intervals; the table shows the shortest one, which is named above the
table. Use --json to get the averages for all intervals.

The space's quota is shown above the table so that storage and worker usage
can be compared against the per-collection limits.

By default, all collections are fetched automatically across multiple pages.
Use --page-size and --page-token for manual pagination.

```
qcloud serverless space metrics summary <space-id> [flags]
```

Flags:

```
      --collection string            Only include the collection with this exact name
      --collection-contains string   Only include collections whose name contains this substring
      --page-size int32              Maximum number of collections to return per page (manual pagination mode)
      --page-token string            Page token from a previous response to resume from (manual pagination mode)
```

Examples:

```sh
# Show an overview of all collections in a space
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Only show a single collection
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --collection products

# Only show collections whose name contains "staging"
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --collection-contains staging

# Output all interval averages as JSON
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json
```

## qcloud serverless space metrics usage

Show usage metrics of the collections in a serverless space over a period.

The server returns a time series per collection for the search rate, write
rate, search latency, point count, storage usage and number of search workers.
The table condenses each series: point count, storage and workers show the
latest value, while request rates and latency show the average and maximum over
the period. Use --json to get the full time series.

When --since is omitted, the period covers the last hour.

By default, all collections are fetched automatically across multiple pages.
Use --page-size and --page-token for manual pagination.

```
qcloud serverless space metrics usage <space-id> [flags]
```

Flags:

```
      --aggregator string            Aggregation function applied to the time series (sum, avg, max, min; default sum)
      --collection string            Only include the collection with this exact name
      --collection-contains string   Only include collections whose name contains this substring
      --page-size int32              Maximum number of collections to return per page (manual pagination mode)
      --page-token string            Page token from a previous response to resume from (manual pagination mode)
      --since string                 Start of the period (RFC3339, YYYY-MM-DD, or a duration ago such as 6h or 7d; default 1h ago)
      --until string                 End of the period (RFC3339, YYYY-MM-DD, or a duration ago such as 1h; default now)
```

Examples:

```sh
# Show usage over the last hour
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Show usage over the last 24 hours
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 24h

# Show usage of a single collection on a specific day
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --collection products --since 2026-10-01 --until 2026-10-01

# Export the full time series as JSON
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 7d --json
```

## qcloud serverless space suggest-name

Suggest a unique, human-friendly name for a new space.

The suggested name is not reserved: it is only guaranteed to be unused in the
current account at the time of the call. "qcloud serverless space create" uses
the same suggestion automatically when --name is omitted.

```
qcloud serverless space suggest-name
```

Examples:

```sh
# Print a suggested space name
qcloud serverless space suggest-name

# Use the suggestion in a script
qcloud serverless space create --cloud-region eu-central-1 --name "$(qcloud serverless space suggest-name)"
```

## qcloud serverless space update

Update an existing serverless space.

Use this command to rename a space or change its labels, network restrictions,
per-collection size limit and search-worker settings. The cloud region cannot be
changed after creation.

Labels are merged with existing labels. Use 'key=value' to add or overwrite a
label, and 'key-' (with a trailing dash) to remove one. Allowed IPs and allowed
origins are merged the same way: specify a value to add it, or append '-' to
remove it (e.g. '10.0.0.0/8-').

```
qcloud serverless space update <space-id> [flags]
```

Flags:

```
      --allowed-ip stringArray           Allowed client IP CIDR range (e.g. "10.0.0.0/8"); append '-' to remove; IPv4 only
      --allowed-origin stringArray       Allowed browser origin for CORS (e.g. "https://app.example.com"); append '-' to remove; max 10
      --cost-allocation-label string     Label for billing reports
      --label stringArray                Label ('key=value') to add/overwrite; append '-' to remove ('key-'), can be specified multiple times
      --max-collection-size string       Maximum size per collection (e.g. 10GiB, 512MiB); must not exceed the platform limit
      --name string                      New name of the space
      --searcher-idle-timeout duration   Idle timeout after which search workers may be scaled down (1m to 15m)
      --searcher-max-workers uint        Maximum number of search workers per collection
```

Examples:

```sh
# Rename a space
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --name my-renamed-space

# Add a label and remove another one
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --label env=prod --label team-

# Allow a new origin and remove an old IP range
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --allowed-origin https://app.example.com --allowed-ip 10.0.0.0/8-

# Tune the search workers
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --searcher-idle-timeout 10m --searcher-max-workers 4
```

## qcloud serverless space wait

Wait for a serverless space to become ready.

Polls the space until its phase is READY and prints the endpoint. The command
fails as soon as the space is DISABLED or DELETING, printing the reason reported
by the server, or when the timeout expires.

```
qcloud serverless space wait <space-id> [flags]
```

Flags:

```
      --timeout duration   Maximum time to wait for the space to become ready (default 10m0s)
```

Examples:

```sh
# Wait for a space to become ready
qcloud serverless space wait 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Wait with a custom timeout
qcloud serverless space wait 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --timeout 20m
```
