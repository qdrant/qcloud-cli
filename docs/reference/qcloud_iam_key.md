## qcloud iam key

Manage cloud management keys

### Synopsis

Manage cloud management keys for the account.

Management keys authenticate requests to the Qdrant Cloud API. Use them to authorize
the CLI, automation scripts, or any other tooling that calls the Qdrant Cloud API.

### Options

```
  -h, --help   help for key
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

* [qcloud iam](qcloud_iam.md)	 - Manage IAM resources in Qdrant Cloud
* [qcloud iam key create](qcloud_iam_key_create.md)	 - Create a cloud management key
* [qcloud iam key delete](qcloud_iam_key_delete.md)	 - Delete a cloud management key
* [qcloud iam key list](qcloud_iam_key_list.md)	 - List cloud management keys

