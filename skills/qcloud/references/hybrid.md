# qcloud hybrid

Manage hybrid cloud environments

Commands:

- `qcloud hybrid bootstrap`: Generate bootstrap commands for a hybrid cloud environment
- `qcloud hybrid create`: Create a new hybrid cloud environment
- `qcloud hybrid delete`: Delete a hybrid cloud environment
- `qcloud hybrid describe`: Describe a hybrid cloud environment
- `qcloud hybrid list`: List all hybrid cloud environments
- `qcloud hybrid update`: Update a hybrid cloud environment

## qcloud hybrid bootstrap

Generate the commands needed to bootstrap a Kubernetes cluster into a hybrid cloud environment.

Each command in the output is ready to copy-paste or pipe to a shell. The credentials
printed to stderr are sensitive and should be treated as secrets.

Note: each invocation creates new Qdrant Cloud access tokens and registry credentials.
Only run this if the Kubernetes cluster is not yet registered to the environment.

```
qcloud hybrid bootstrap <env-id>
```

Examples:

```sh
# Generate bootstrap commands for an environment
qcloud hybrid bootstrap 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Pipe directly to a shell
qcloud hybrid bootstrap 7b2ea926-724b-4de2-b73a-8675c42a6ebe | bash
```

## qcloud hybrid create

Create a new Hybrid Cloud Environment to deploy and manage Qdrant on your own
Kubernetes clusters (on-premises, cloud, or edge) with enterprise-grade
reliability.

Hybrid Cloud access must be enabled for your account by the Qdrant sales team.
If your account does not have access, you will be prompted to contact us.

```
qcloud hybrid create [flags]
```

Flags:

```
      --cluster-domain string           Kubernetes cluster domain for in-cluster services (defaults to cluster.local if omitted)
      --database-storage-class string   Default database storage class (uses cluster default if omitted)
      --log-level string                Log level for deployed components ("debug", "info", "warn", "error")
      --name string                     Name of the hybrid cloud environment (required)
      --namespace string                Kubernetes namespace where Qdrant components are deployed (read-only after bootstrapping)
      --snapshot-storage-class string   Default snapshot storage class (uses cluster default if omitted)
```

Examples:

```sh
# Create a hybrid cloud environment
qcloud hybrid create --name my-hybrid-env

# Create with a custom namespace
qcloud hybrid create --name my-hybrid-env --namespace qdrant-hybrid

# Create with storage classes
qcloud hybrid create --name my-hybrid-env \
  --database-storage-class premium-rwo --snapshot-storage-class standard
```

## qcloud hybrid delete

Delete a hybrid cloud environment

```
qcloud hybrid delete <env-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete a hybrid cloud environment (prompts for confirmation)
qcloud hybrid delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud hybrid delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force
```

## qcloud hybrid describe

Describe a hybrid cloud environment

```
qcloud hybrid describe <env-id>
```

## qcloud hybrid list

List all hybrid cloud environments

```
qcloud hybrid list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

## qcloud hybrid update

Update a hybrid cloud environment

```
qcloud hybrid update <env-id> [flags]
```

Flags:

```
      --cluster-domain string           Kubernetes cluster domain for in-cluster services (defaults to cluster.local if omitted)
      --database-storage-class string   Default database storage class (uses cluster default if omitted)
      --log-level string                Log level for deployed components ("debug", "info", "warn", "error")
      --name string                     New name for the hybrid cloud environment
      --namespace string                Kubernetes namespace where Qdrant components are deployed (read-only after bootstrapping)
      --snapshot-storage-class string   Default snapshot storage class (uses cluster default if omitted)
```

Examples:

```sh
# Rename a hybrid cloud environment
qcloud hybrid update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --name new-name

# Update the default storage classes
qcloud hybrid update 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --database-storage-class premium-rwo --snapshot-storage-class standard

# Change the log level
qcloud hybrid update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --log-level debug
```
