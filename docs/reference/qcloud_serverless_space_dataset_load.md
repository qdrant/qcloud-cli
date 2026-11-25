## qcloud serverless space dataset load

Manage sample dataset loads of a space

### Synopsis

Manage sample dataset loads of a serverless space.

A load copies a sample dataset into a new collection of the space. It runs in
the background and ends as SUCCEEDED or FAILED; the reason of a failure (for
example the collection limit of the space) is shown by "load describe". Only
loads created in the last 30 days are listed.

### Examples

```
# Load a sample dataset into a collection with the same name as the dataset
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles

# Load a sample dataset into a collection with a custom name
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles --collection styles

# List the loads of a space
qcloud serverless space dataset load list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Follow a single load
qcloud serverless space dataset load describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 7f6e5d4c-3b2a-4190-8f7e-6d5c4b3a2918
```

### Options

```
  -h, --help[=true|false]   help for load
```

### Options inherited from parent commands

```
      --account-id string    Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)
      --api-key string       Management API Key (env: QDRANT_CLOUD_API_KEY)
  -c, --config string        Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)
      --console-url string   Qdrant Cloud web console base URL (env: QDRANT_CLOUD_CONSOLE_URL, default https://cloud.qdrant.io)
      --context string       Override the active context (env: QDRANT_CLOUD_CONTEXT)
      --debug[=true|false]   Enable debug logging to stderr
      --endpoint string      gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)
      --json[=true|false]    Output as JSON
```

### SEE ALSO

* [qcloud serverless space dataset](qcloud_serverless_space_dataset.md)	 - Load sample datasets into a space
* [qcloud serverless space dataset load describe](qcloud_serverless_space_dataset_load_describe.md)	 - Describe a sample dataset load
* [qcloud serverless space dataset load list](qcloud_serverless_space_dataset_load_list.md)	 - List the sample dataset loads of a space
* [qcloud serverless space dataset load start](qcloud_serverless_space_dataset_load_start.md)	 - Load a sample dataset into a new collection

