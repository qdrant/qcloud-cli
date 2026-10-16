# qcloud cluster

Manage Qdrant Cloud clusters

Commands:

- `qcloud cluster create`: Create a new cluster
- `qcloud cluster create-from-backup`: Create a new cluster from a backup
- `qcloud cluster dashboard`: Open a cluster's dashboard in your browser
- `qcloud cluster delete`: Delete a cluster
- `qcloud cluster describe`: Describe a cluster
- `qcloud cluster key create`: Create an API key for a cluster
- `qcloud cluster key delete`: Delete an API key from a cluster
- `qcloud cluster key list`: List API keys for a cluster
- `qcloud cluster list`: List all clusters
- `qcloud cluster logs`: Retrieve logs for a cluster
- `qcloud cluster restart`: Restart a cluster
- `qcloud cluster scale`: Scales the resources of a cluster
- `qcloud cluster suspend`: Suspend a cluster
- `qcloud cluster unsuspend`: Unsuspend a cluster
- `qcloud cluster update`: Update an existing cluster
- `qcloud cluster version list`: List available Qdrant versions
- `qcloud cluster wait`: Wait for a cluster to become healthy

## qcloud cluster create

Create a new cluster

```
qcloud cluster create [flags]
```

Flags:

```
      --allowed-ip stringArray                   Allowed client IP CIDR range (e.g. "10.0.0.0/8"); append '-' to remove; max 20
      --annotation stringArray                   (cloud-provider: hybrid) Pod annotation ('key=value'); append '-' to remove, can be specified multiple times
      --api-key-secret string                    (cloud-provider: hybrid) API key Kubernetes secret ('secretName:key')
      --async-scorer                             Enable async scorer (uses io_uring on Linux)
      --audit-log-max-files uint32               Maximum number of audit log files (1-1000)
      --audit-log-rotation string                Audit log rotation ("daily", "hourly")
      --audit-log-trust-forwarded-headers        Trust forwarded headers in audit logs
      --audit-logging                            Enable audit logging
      --cloud-provider string                    Cloud provider ID (required, see 'cloud-provider list)
      --cloud-region string                      Cloud provider region ID (required, see 'cloud-region list --cloud-provider <provider_id>)
      --cost-allocation-label string             Label for billing reports
      --cpu millicores                           CPU to select a package (e.g. "1", "0.5", or "1000m")
      --database-storage-class string            (cloud-provider: hybrid) Kubernetes storage class for database volumes
      --db-log-level string                      Database log level ("trace", "debug", "info", "warn", "error", "off")
      --disk bytes                               Total disk size (e.g. "200GiB"); if larger than the package's included disk, the difference is provisioned as additional storage
      --disk-performance string                  Disk performance tier ("balanced", "cost-optimised", "performance")
      --enable-tls                               (cloud-provider: hybrid) Enable TLS for the database service
      --gpu millicores                           Number of GPUs to select a package (e.g. "1", "2", or "1000m")
      --label stringArray                        Label ('key=value') to add/overwrite; append '-' to remove ('key-'), can be specified multiple times
      --multi-az                                 Require a multi-AZ package
      --name string                              Cluster name (auto-generated if not provided)
      --node-selector stringArray                (cloud-provider: hybrid) Node selector label ('key=value'); append '-' to remove, can be specified multiple times
      --nodes uint32                             Number of nodes (default 1)
      --optimizer-cpu-budget int32               CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)
      --package string                           Booking package name or ID (see 'cluster package list')
      --pod-label stringArray                    (cloud-provider: hybrid) Pod label ('key=value'); append '-' to remove, can be specified multiple times
      --ram bytes                                RAM to select a package (e.g. "8", "8G", "8Gi", or "8GiB")
      --read-only-api-key-secret string          (cloud-provider: hybrid) Read-only API key Kubernetes secret ('secretName:key')
      --rebalance-strategy string                Shard rebalance strategy ("by-count", "by-size", "by-count-and-size")
      --replication-factor uint32                Default replication factor for new collections
      --reserved-cpu-percentage uint32           (cloud-provider: hybrid) Percentage of CPU reserved for system components, 1-80 (default 20)
      --reserved-memory-percentage uint32        (cloud-provider: hybrid) Percentage of memory reserved for system components, 1-80 (default 20)
      --restart-mode string                      Restart policy ("rolling", "parallel", "automatic")
      --service-annotation stringArray           (cloud-provider: hybrid) Service annotation ('key=value'); append '-' to remove, can be specified multiple times
      --service-type string                      (cloud-provider: hybrid) Kubernetes service type ("cluster-ip", "node-port", "load-balancer")
      --snapshot-storage-class string            (cloud-provider: hybrid) Kubernetes storage class for snapshot volumes
      --tls-cert-secret string                   (cloud-provider: hybrid) TLS certificate Kubernetes secret ('secretName:key')
      --tls-key-secret string                    (cloud-provider: hybrid) TLS private key Kubernetes secret ('secretName:key')
      --toleration stringArray                   (cloud-provider: hybrid) Toleration ('key=value:Effect' or 'key:Exists:Effect'); use 'key-' to remove by key, can be specified multiple times
      --topology-spread-constraint stringArray   (cloud-provider: hybrid) Topology spread constraint ('topologyKey[:maxSkew[:whenUnsatisfiable]]'); use 'topologyKey-' to remove, can be specified multiple times
      --vectors-on-disk                          Store vectors in memmap storage for new collections
      --version string                           Qdrant version (e.g. "v1.17.0" or "latest")
      --volume-attributes-class string           (cloud-provider: hybrid) Kubernetes volume attributes class
      --volume-snapshot-class string             (cloud-provider: hybrid) Kubernetes volume snapshot class
      --wait                                     Wait for the cluster to become healthy
      --wait-timeout duration                    Maximum time to wait for the cluster to become healthy (default 10m0s)
      --write-consistency-factor int32           Default write consistency factor for new collections
```

Examples:

```sh
# Create a free-tier cluster
qcloud cluster create --cloud-provider aws --cloud-region eu-central-1 --package free

# Create a cluster with specific resources
qcloud cluster create --cloud-provider aws --cloud-region eu-central-1 --cpu 0.5 --ram 4Gi

# Create a cluster and wait for it to become healthy
qcloud cluster create --cloud-provider aws --cloud-region eu-central-1 --cpu 2 --ram 8Gi --wait

# Create with labels and extra disk
qcloud cluster create --cloud-provider aws --cloud-region eu-central-1 --cpu 4 --ram 32Gi \
  --disk 200Gi --label env=production --label team=search

# Create a hybrid cloud cluster with a load balancer service type
qcloud cluster create --cloud-provider hybrid --cloud-region my-env --cpu 2 --ram 8Gi \
  --service-type load-balancer

# Create a hybrid cluster with node selectors and tolerations
qcloud cluster create --cloud-provider hybrid --cloud-region my-env --cpu 2 --ram 8Gi \
  --node-selector disktype=ssd --toleration "dedicated=qdrant:NoSchedule"

# Create a hybrid cluster with custom storage classes
qcloud cluster create --cloud-provider hybrid --cloud-region my-env --cpu 4 --ram 16Gi \
  --database-storage-class fast-ssd --snapshot-storage-class standard
```

## qcloud cluster create-from-backup

Create a new Qdrant Cloud cluster seeded with the data from an existing backup.

The new cluster is provisioned using the same configuration as the original cluster
at the time the backup was taken. The backup must belong to the current account.

```
qcloud cluster create-from-backup [flags]
```

Flags:

```
      --backup-id string        ID of the backup to restore from (required)
      --name string             Name for the new cluster (required)
      --wait                    Wait for the cluster to become healthy
      --wait-timeout duration   Maximum time to wait for the cluster to become healthy (default 10m0s)
```

Examples:

```sh
# Create a cluster from a backup
qcloud cluster create-from-backup --backup-id <backup-id> --name my-restored-cluster

# Create a cluster from a backup and wait until it is healthy
qcloud cluster create-from-backup --backup-id <backup-id> --name my-restored-cluster --wait

# Create with a custom wait timeout
qcloud cluster create-from-backup --backup-id <backup-id> --name my-restored-cluster --wait --wait-timeout 20m
```

## qcloud cluster dashboard

Open a cluster's dashboard in your default browser.

The command builds the Cloud UI dashboard URL and opens it. The Cloud UI page
handles authentication using your existing browser session and redirects to the
cluster's dashboard.

```
qcloud cluster dashboard <cluster-id> [flags]
```

Flags:

```
      --print-url   Print the dashboard URL instead of opening a browser
```

Examples:

```sh
# Open a cluster's dashboard in your default browser
qcloud cluster dashboard 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Print the dashboard URL instead of opening a browser (headless/SSH)
qcloud cluster dashboard 7b2ea926-724b-4de2-b73a-8675c42a6ebe --print-url
```

## qcloud cluster delete

Delete a cluster

```
qcloud cluster delete <cluster-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete a cluster (prompts for confirmation)
qcloud cluster delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud cluster delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force
```

## qcloud cluster describe

Describe a cluster

```
qcloud cluster describe <cluster-id>
```

Examples:

```sh
# Describe a cluster
qcloud cluster describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Output as JSON
qcloud cluster describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json
```

## qcloud cluster key create

Create an API key for a cluster

```
qcloud cluster key create <cluster-id> [flags]
```

Flags:

```
      --access-type string      Access type: manage or read-only (default: server assigns manage)
      --expires string          Expiration date in YYYY-MM-DD format; the key is valid until the end of that day (UTC)
      --name string             Name of the API key (required)
      --wait                    Wait for the API key to become active on the cluster
      --wait-timeout duration   Maximum time to wait for the API key to become active on the cluster (default 1m0s)
```

Examples:

```sh
# Create an API key
qcloud cluster key create 7b2ea926-724b-4de2-b73a-8675c42a6ebe --name my-key

# Create a read-only key with expiration
qcloud cluster key create 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --name read-key --access-type read-only --expires 2025-12-31

# Create a key and wait for it to become active on the cluster
qcloud cluster key create 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --name my-key --wait
```

## qcloud cluster key delete

Delete an API key from a cluster

```
qcloud cluster key delete <cluster-id> <key-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete an API key
qcloud cluster key delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

## qcloud cluster key list

List API keys for a cluster

```
qcloud cluster key list <cluster-id> [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List API keys for a cluster
qcloud cluster key list 7b2ea926-724b-4de2-b73a-8675c42a6ebe
```

## qcloud cluster list

List all clusters in the current account.

By default, all clusters are fetched automatically across multiple pages.

Use --page-size and --page-token for manual pagination:
  --page-size limits how many clusters are returned per call.
  --page-token resumes from a specific page (token is printed when more pages exist).
  If --page-token is omitted, listing starts from the beginning.

Use --cloud-provider and --cloud-region to filter results server-side:
  --cloud-provider filters clusters by cloud provider ID (e.g. aws, gcp).
  --cloud-region filters clusters by cloud provider region ID (e.g. us-east-1).

```
qcloud cluster list [flags]
```

Flags:

```
      --cloud-provider string   Filter by cloud provider ID
      --cloud-region string     Filter by cloud provider region ID
      --no-headers              Do not print column headers
      --page-size int32         Maximum number of clusters to return per page (manual pagination mode)
      --page-token string       Page token from a previous response to resume from (manual pagination mode)
```

Examples:

```sh
# List all clusters
qcloud cluster list

# List clusters in JSON format
qcloud cluster list --json

# Filter by cloud provider and region
qcloud cluster list --cloud-provider aws --cloud-region eu-central-1

# Manual pagination
qcloud cluster list --page-size 10
```

## qcloud cluster logs

Retrieve logs for a cluster.

By default, logs from the last 3 days up to now are returned. --since and --until
accept an RFC3339 timestamp or a YYYY-MM-DD date in UTC. A date passed to
--since starts at the beginning of that day, and a date passed to --until
includes the whole day.

```
qcloud cluster logs <cluster-id> [flags]
```

Flags:

```
  -s, --since string   Start time for logs (RFC3339 or YYYY-MM-DD, default: 3 days ago)
  -t, --timestamps     Prepend each log line with its timestamp
  -u, --until string   End time for logs (RFC3339 or YYYY-MM-DD, default: now)
```

Examples:

```sh
# Get logs for a cluster
qcloud cluster logs abc-123

# Get logs since a specific date
qcloud cluster logs abc-123 --since 2024-01-01

# Get logs in a specific time range
qcloud cluster logs abc-123 --since 2024-01-01T00:00:00Z --until 2024-01-02T00:00:00Z

# Get logs in JSON format
qcloud cluster logs abc-123 --json
```

## qcloud cluster restart

Restart a cluster

```
qcloud cluster restart <cluster-id> [flags]
```

Flags:

```
  -f, --force                   Skip confirmation prompt
      --wait                    Wait for the cluster to restart and become healthy
      --wait-timeout duration   Maximum time to wait for the cluster to restart and become healthy (default 10m0s)
```

Examples:

```sh
# Restart a cluster (prompts for confirmation)
qcloud cluster restart 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Restart without confirmation and wait for healthy status
qcloud cluster restart 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force --wait
```

## qcloud cluster scale

Scales the resources of a cluster.

Use this command to change the resource package for the nodes of a Qdrant cluster, change
the number of nodes in the cluster, and allocate more disk space per node. You can select
one of the pre-defined resource packages to apply for all nodes in the cluster. A package
defines CPU, RAM, GPU, and minimum disk size.

The --cpu, --ram, and --gpu flags specify per-node resources and are used to match a
package. If none of these flags are provided, the current package is preserved. Available
packages can be listed with 'package list' using the --cloud-provider and
--cloud-region flags.

Each package includes a baseline disk size. Requesting more disk than the baseline with
--disk provisions the difference as additional storage. Disk cannot be downscaled. If a
new package has a larger baseline disk than the current total, the disk size increases to
match.

Reducing RAM is always allowed, but a cluster whose memory usage does not fit the smaller
package can run out of memory and become unavailable. A warning is printed before the
scale when that is the case, or when the usage could not be determined; --force skips the
confirmation but not the warning.

```
qcloud cluster scale <cluster-id> [flags]
```

Flags:

```
      --cpu millicores            CPU per node (e.g. "1", "0.5", or "1000m")
      --disk bytes                Total disk size per node (e.g. "200GiB"); if larger than the node's included disk, the difference is provisioned as additional storage
      --disk-performance string   Disk performance tier ("balanced", "cost-optimised", "performance")
  -f, --force                     Skip confirmation prompts
      --gpu millicores            Number of GPUs per node (e.g. "1", "2", or "1000m")
      --nodes uint32              Number of nodes
      --ram bytes                 RAM per node (e.g. "8", "8G", "8Gi", or "8GiB")
      --wait                      Wait for the cluster to become healthy
      --wait-timeout duration     Maximum time to wait for the cluster to become healthy (default 10m0s)
```

Examples:

```sh
# Scale up CPU and RAM
qcloud cluster scale 7b2ea926-724b-4de2-b73a-8675c42a6ebe --cpu 4 --ram 16Gi

# Add more nodes
qcloud cluster scale 7b2ea926-724b-4de2-b73a-8675c42a6ebe --nodes 3

# Increase disk and wait for completion
qcloud cluster scale 7b2ea926-724b-4de2-b73a-8675c42a6ebe --disk 500Gi --wait
```

## qcloud cluster suspend

Suspend a cluster

```
qcloud cluster suspend <cluster-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Suspend a cluster
qcloud cluster suspend 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force
```

## qcloud cluster unsuspend

Unsuspend a cluster

```
qcloud cluster unsuspend <cluster-id>
```

Examples:

```sh
# Unsuspend a cluster
qcloud cluster unsuspend 7b2ea926-724b-4de2-b73a-8675c42a6ebe
```

## qcloud cluster update

Updates the configuration of a cluster.

Use this command to modify cluster settings such as the Qdrant version, labels,
database defaults, IP restrictions, restart mode, rebalance strategy, and hybrid
cluster configuration.

Version upgrades (--version) will trigger a rolling restart of the cluster.

Database configuration changes (--replication-factor, --write-consistency-factor,
--async-scorer, --optimizer-cpu-budget, --vectors-on-disk, --db-log-level,
--audit-logging and related flags, --enable-tls, --api-key-secret,
--read-only-api-key-secret, --tls-cert-secret, --tls-key-secret) will trigger a
rolling restart of the cluster. The cluster remains available during the restart,
but individual nodes will be briefly unavailable as they cycle.

Hybrid cluster configuration changes (--service-type, --node-selector,
--toleration, --topology-spread-constraint, --annotation, --pod-label,
--service-annotation, --reserved-cpu-percentage, --reserved-memory-percentage,
and storage class flags) will also trigger a rolling restart.

Cluster configuration changes (--allowed-ip, --restart-mode, --rebalance-strategy,
--disk-performance, --cost-allocation-label) and label changes take effect without
a restart.

Labels are merged with existing labels by default. Use 'key=value' to add or
overwrite a label, and 'key-' (with a trailing dash) to remove one.

Allowed IPs are merged with existing IPs by default. Specify an IP CIDR to add
it, or append '-' (e.g. '10.0.0.0/8-') to remove one.

Node selectors, annotations, pod labels, and service annotations support the same
'key=value' / 'key-' merge syntax as labels.

Tolerations are merged with existing tolerations. Use 'key-' to remove all
tolerations matching that key.

Topology spread constraints are merged by topologyKey. Use 'topologyKey-' to
remove a constraint.

```
qcloud cluster update <cluster-id> [flags]
```

Flags:

```
      --allowed-ip stringArray                   Allowed client IP CIDR range (e.g. "10.0.0.0/8"); append '-' to remove; max 20
      --annotation stringArray                   (cloud-provider: hybrid) Pod annotation ('key=value'); append '-' to remove, can be specified multiple times
      --api-key-secret string                    (cloud-provider: hybrid) API key Kubernetes secret ('secretName:key')
      --async-scorer                             Enable async scorer (uses io_uring on Linux)
      --audit-log-max-files uint32               Maximum number of audit log files (1-1000)
      --audit-log-rotation string                Audit log rotation ("daily", "hourly")
      --audit-log-trust-forwarded-headers        Trust forwarded headers in audit logs
      --audit-logging                            Enable audit logging
      --cost-allocation-label string             Label for billing reports
      --database-storage-class string            (cloud-provider: hybrid) Kubernetes storage class for database volumes
      --db-log-level string                      Database log level ("trace", "debug", "info", "warn", "error", "off")
      --disk-performance string                  Disk performance tier ("balanced", "cost-optimised", "performance")
      --enable-tls                               (cloud-provider: hybrid) Enable TLS for the database service
  -f, --force                                    Skip confirmation prompt
      --label stringArray                        Label ('key=value') to add/overwrite; append '-' to remove ('key-'), can be specified multiple times
      --node-selector stringArray                (cloud-provider: hybrid) Node selector label ('key=value'); append '-' to remove, can be specified multiple times
      --optimizer-cpu-budget int32               CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)
      --pod-label stringArray                    (cloud-provider: hybrid) Pod label ('key=value'); append '-' to remove, can be specified multiple times
      --read-only-api-key-secret string          (cloud-provider: hybrid) Read-only API key Kubernetes secret ('secretName:key')
      --rebalance-strategy string                Shard rebalance strategy ("by-count", "by-size", "by-count-and-size")
      --replication-factor uint32                Default replication factor for new collections
      --reserved-cpu-percentage uint32           (cloud-provider: hybrid) Percentage of CPU reserved for system components, 1-80 (default 20)
      --reserved-memory-percentage uint32        (cloud-provider: hybrid) Percentage of memory reserved for system components, 1-80 (default 20)
      --restart-mode string                      Restart policy ("rolling", "parallel", "automatic")
      --service-annotation stringArray           (cloud-provider: hybrid) Service annotation ('key=value'); append '-' to remove, can be specified multiple times
      --service-type string                      (cloud-provider: hybrid) Kubernetes service type ("cluster-ip", "node-port", "load-balancer")
      --snapshot-storage-class string            (cloud-provider: hybrid) Kubernetes storage class for snapshot volumes
      --tls-cert-secret string                   (cloud-provider: hybrid) TLS certificate Kubernetes secret ('secretName:key')
      --tls-key-secret string                    (cloud-provider: hybrid) TLS private key Kubernetes secret ('secretName:key')
      --toleration stringArray                   (cloud-provider: hybrid) Toleration ('key=value:Effect' or 'key:Exists:Effect'); use 'key-' to remove by key, can be specified multiple times
      --topology-spread-constraint stringArray   (cloud-provider: hybrid) Topology spread constraint ('topologyKey[:maxSkew[:whenUnsatisfiable]]'); use 'topologyKey-' to remove, can be specified multiple times
      --vectors-on-disk                          Store vectors in memmap storage for new collections
      --version string                           Qdrant version (e.g. "v1.17.0" or "latest")
      --volume-attributes-class string           (cloud-provider: hybrid) Kubernetes volume attributes class
      --volume-snapshot-class string             (cloud-provider: hybrid) Kubernetes volume snapshot class
      --write-consistency-factor int32           Default write consistency factor for new collections
```

Examples:

```sh
# Add a label to a cluster
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --label env=staging

# Remove a label
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --label env-

# Restrict access to specific IPs
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --allowed-ip 10.0.0.0/8

# Upgrade the Qdrant version
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --version v1.17.0

# Change replication factor (triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --replication-factor 3 --force

# Set service type to load balancer (hybrid only, triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --service-type load-balancer

# Add a node selector and toleration (hybrid only, triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --node-selector disktype=ssd --toleration "dedicated=qdrant:NoSchedule"

# Remove a node selector (hybrid only)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --node-selector disktype-

# Change database storage class (hybrid only, triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --database-storage-class fast-ssd
```

## qcloud cluster version list

List available Qdrant versions

```
qcloud cluster version list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List available Qdrant versions
qcloud cluster version list
```

## qcloud cluster wait

Wait for a cluster to become healthy

```
qcloud cluster wait <cluster-id> [flags]
```

Flags:

```
      --timeout duration   Maximum time to wait for the cluster to become healthy (default 10m0s)
```

Examples:

```sh
# Wait for a cluster to become healthy
qcloud cluster wait 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Wait with a custom timeout
qcloud cluster wait 7b2ea926-724b-4de2-b73a-8675c42a6ebe --timeout 30m
```
