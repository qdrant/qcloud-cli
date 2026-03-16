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


## Disclaimer

`qcloud` currently is under heavy development. The output and command shape can heavily change from version to version.


## Quick Start

Before using `qcloud`, create a management API key and note your account ID from the [Qdrant Cloud UI](https://cloud.qdrant.io).

```sh
# 1. Create a context with your credentials
qcloud context set my-cloud \
  --api-key <YOUR_API_KEY> \
  --account-id <YOUR_ACCOUNT_ID>

# 2. List available cloud providers and regions
qcloud cluster cloud-provider list
qcloud cluster cloud-region list --cloud-provider aws

# 3. Create a cluster by specifying resources (waits until healthy)
#    Use --cpu, --ram, --disk (and optionally --gpu / --multi-az) to select
#    a matching package automatically.
qcloud cluster create \
  --cloud-provider aws \
  --cloud-region us-east-1 \
  --cpu 2000m \
  --ram 8GiB \
  --disk 50GiB \
  --name my-cluster \
  --wait

#    Alternatively, pick a package explicitly with --package:
#      qcloud cluster package list --cloud-provider aws --cloud-region us-east-1
#      qcloud cluster create --cloud-provider aws --cloud-region us-east-1 \
#        --package <PACKAGE_ID> --name my-cluster --wait

# 4. Describe your new cluster
qcloud cluster describe <CLUSTER_ID>

# 5. Create an API key for it
qcloud cluster key create <CLUSTER_ID> --name my-key
```


## Configuration

`qcloud` can be configured in three ways, listed here from lowest to highest precedence:

**Config file** at `~/.config/qcloud/config.yaml` (override with `--config`). Stores named contexts so you can switch between accounts and environments.

**Named contexts** allow you to save and switch between sets of credentials:

```sh
qcloud context set my-cloud --api-key <KEY> --account-id <ID>
qcloud context use my-cloud
qcloud context show
```

**Environment variables** override the active context:

| Variable                  | Description                                 |
|---------------------------|---------------------------------------------|
| `QDRANT_CLOUD_API_KEY`    | API key for authentication                  |
| `QDRANT_CLOUD_ACCOUNT_ID` | Account ID to operate against               |
| `QDRANT_CLOUD_ENDPOINT`   | API endpoint URL (defaults to Qdrant Cloud) |
| `QDRANT_CLOUD_CONTEXT`    | Name of the context to use                  |

Pass `--json` to any command for machine-readable output.


## Acknowledgements

`qcloud` is inspired by and partially based on [hetznercloud/cli](https://github.com/hetznercloud/cli). Thank you to the Hetzner Cloud team and contributors for building such a well-designed CLI that we could learn from.


## Contributing

If you are interested in contributing follow the instructions in [CONTRIBUTING.md](./CONTRIBUTING.md)
