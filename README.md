# qcloud - The CLI for Qdrant Cloud

<p align="center">
  <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://github.com/qdrant/qdrant/raw/master/docs/logo-dark.svg">
      <source media="(prefers-color-scheme: light)" srcset="https://github.com/qdrant/qdrant/raw/master/docs/logo-light.svg">
      <img height="100" alt="Qdrant" src="https://github.com/qdrant/qdrant/raw/master/docs/logo.svg">
  </picture>
</p>

<p align="center">
    <b>The official command-line interface for managing Qdrant Cloud</b>
</p>

`qcloud` is the CLI for [Qdrant Cloud](https://qdrant.tech/). It lets you manage clusters, authentication and anything the API has to offer with a terminal experience.

![demo](docs/demo.gif)

## Installation

### From Homebrew

On macOS and Linux:

```sh
brew install qdrant/tap/qcloud
```

### From the install script

For systems without Homebrew, such as CI runners, containers, and sandboxed agents:

```sh
curl -fsSL https://raw.githubusercontent.com/qdrant/qcloud-cli/main/install.sh | sh
```

The script installs the latest release for your OS and CPU architecture to `~/.local/bin`, after checking the archive against the release's `checksums.txt`. Set `QCLOUD_VERSION` (for example `v0.32.0`) to pin a release, or `QCLOUD_INSTALL_DIR` to install somewhere else. It needs `curl` or `wget`, `sha256sum` or `shasum`, and `tar`.

### From GitHub Releases

Download the latest release from [GitHub Releases](https://github.com/qdrant/qcloud-cli/releases).

Select the archive that matches your OS and CPU architecture, extract it, and place the `qcloud` binary somewhere in your `PATH` (e.g. `~/.local/bin` or `/usr/local/bin`).

> If `~/.local/bin` is not in your `PATH`, you can use `/usr/local/bin` instead (requires `sudo`).

### From source

With Go installed, you can build and install directly from source:

```sh
go install github.com/qdrant/qcloud-cli/cmd/qcloud@latest
```

### Verify

```sh
qcloud version
```

## Quick Start

```sh
# 1. Authenticate
qcloud auth login

# 2. If you authenticated multiple accounts, pick the righ context for the account you want to use
qcloud context list
qcloud context use <CONTEXT_NAME>

# 3. List available cloud providers and regions
qcloud cloud-provider list
qcloud cloud-region list --cloud-provider aws

# 4. Create a cluster by specifying resources (waits until healthy)
#    Use --cpu, --ram, --disk to select a matching package automatically.
qcloud cluster create \
  --cloud-provider aws \
  --cloud-region us-east-1 \
  --cpu 2000m \
  --ram 8GiB \
  --disk 50GiB \
  --name my-cluster \
  --wait

# 5. Describe your new cluster
qcloud cluster describe <CLUSTER_ID>

# 6. Create an API key for it
qcloud cluster key create <CLUSTER_ID> --name my-key
```

## Command Groups

| Command | Manages |
|---|---|
| `qcloud cluster` | Clusters: create, describe, list, scale, update, restart, suspend, delete, logs, alerts, API keys, and Qdrant versions |
| `qcloud backup` | Backups, backup schedules, and restores |
| `qcloud account` | Accounts |
| `qcloud iam` | Identity and access management resources |
| `qcloud inference` | Inference resources |
| `qcloud hybrid` | Hybrid Cloud environments |
| `qcloud cloud-provider`, `qcloud cloud-region`, `qcloud package` | The providers, regions, and node packages you can choose from |
| `qcloud context` | Named configuration contexts |

Run `qcloud <command> --help` for the flags of each command.

## Common Tasks

```sh
# Scale: more RAM and nodes, plus extra disk
qcloud cluster scale <CLUSTER_ID> --ram 16Gi --nodes 3 --disk 100Gi

# Upgrade the Qdrant version (a rolling restart)
qcloud cluster version list
qcloud cluster update <CLUSTER_ID> --version <VERSION>

# Restrict client IP ranges (takes effect without a restart)
qcloud cluster update <CLUSTER_ID> --allowed-ip 10.0.0.0/8

# Create a read-only key that expires at the end of a day (UTC)
qcloud cluster key create <CLUSTER_ID> --name read-key \
  --access-type read-only --expires 2027-06-30

# Create a new cluster from a backup, and wait until it is healthy
qcloud cluster create-from-backup --backup-id <BACKUP_ID> --name restored --wait

# Check what needs attention
qcloud cluster alert list <CLUSTER_ID>
qcloud cluster logs <CLUSTER_ID> --since 6h
```

Reducing RAM is allowed, but `qcloud cluster scale` warns you first if the cluster's memory usage does not fit the smaller package. For the same workflow in every tool, see [Automate Qdrant Cloud](/documentation/cloud-automation/).

## Configuration

`qcloud` can be configured in three ways, listed here from lowest to highest precedence:

### OAuth Authentication

Run `qcloud auth login`, authenticate and authorize the CLI to access on or more of your Qdrant Cloud accounts. Each account will get a named context, to make it easy to switch between them.

```sh
qcloud context list
qcloud context use my-account
qcloud context show
```

### Named Contexts

Named contexts allow you to also use Cloud Management API Keys for an account. Create a [management API](/documentation/cloud-api/) key and note your account ID from the [Qdrant Cloud Console](https://cloud.qdrant.io).

```sh
qcloud context set my-cloud --api-key <KEY> --account-id <ID>
qcloud context use my-cloud
qcloud context show
```

### Config File

The config file is stored at `~/.config/qcloud/config.yaml` by default. You can override the location with the `--config` flag. It stores named contexts so you can switch between accounts and environments.


### Environment Variables

Environment variables override the active context and take the highest precedence:

| Variable | Description |
|---|---|
| `QDRANT_CLOUD_API_KEY` | API key for authentication |
| `QDRANT_CLOUD_ACCOUNT_ID` | Account ID to operate against |
| `QDRANT_CLOUD_ENDPOINT` | gRPC API endpoint (defaults to `grpc.cloud.qdrant.io:443`) |
| `QDRANT_CLOUD_CONTEXT` | Name of the context to use |

Pass `--json` to any command for machine-readable output.


## Getting Help

For a full list of available commands, run:

```sh
qcloud --help
```

Found a bug or something not working as expected? [Open an issue](https://github.com/qdrant/qcloud-cli/issues/new) on GitHub and include:

- The `qcloud` version (`qcloud version`)
- The command you ran and the output you got
- Your OS and architecture


## Acknowledgements

`qcloud` is inspired by and partially based on [hetznercloud/cli](https://github.com/hetznercloud/cli). Thank you to the Hetzner Cloud team and contributors for building such a well-designed CLI that we could learn from.


## Contributing

If you are interested in contributing follow the instructions in [CONTRIBUTING.md](./CONTRIBUTING.md)
