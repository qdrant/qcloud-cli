## qcloud

Qdrant Cloud CLI

### Synopsis

The command-line interface for Qdrant Cloud.

Get started:
  qcloud context set default --api-key <KEY> --account-id <ID>
  qcloud cluster list

Documentation: https://github.com/qdrant/qcloud-cli

### Options

```
      --account-id string   Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)
      --api-key string      Management API Key (env: QDRANT_CLOUD_API_KEY)
  -c, --config string       Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)
      --context string      Override the active context (env: QDRANT_CLOUD_CONTEXT)
      --debug               Enable debug logging to stderr
      --endpoint string     gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)
  -h, --help                help for qcloud
      --json                Output as JSON
```

### SEE ALSO

* [qcloud account](qcloud_account.md)	 - Manage Qdrant Cloud accounts
* [qcloud backup](qcloud_backup.md)	 - Manage Qdrant Cloud backups
* [qcloud cloud-provider](qcloud_cloud-provider.md)	 - Manage cloud providers
* [qcloud cloud-region](qcloud_cloud-region.md)	 - Manage cloud regions
* [qcloud cluster](qcloud_cluster.md)	 - Manage Qdrant Cloud clusters
* [qcloud context](qcloud_context.md)	 - Manage named configuration contexts
* [qcloud hybrid](qcloud_hybrid.md)	 - Manage hybrid cloud environments
* [qcloud iam](qcloud_iam.md)	 - Manage IAM resources in Qdrant Cloud
* [qcloud package](qcloud_package.md)	 - Manage packages
* [qcloud self-upgrade](qcloud_self-upgrade.md)	 - Upgrade qcloud to the latest version
* [qcloud version](qcloud_version.md)	 - Print the qcloud CLI version

