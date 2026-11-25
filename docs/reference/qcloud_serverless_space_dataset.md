## qcloud serverless space dataset

Load sample datasets into a space

### Synopsis

Load sample datasets into a serverless space.

A sample dataset is a prebuilt dataset, with vectors already computed, that can
be loaded into a new collection of a space to try Qdrant without bringing your
own data. Only the datasets available in the region of the space are offered.
The collection size of a dataset counts against the collection size limits of
the space.

A load is asynchronous: it creates the collection and copies the data in the
background. Use "dataset load list" or "dataset load describe" to follow it, or
pass --wait to "dataset load start".

### Examples

```
# List the sample datasets that can be loaded into a space
qcloud serverless space dataset list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Show the vectors and size of a sample dataset
qcloud serverless space dataset describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles

# Load a sample dataset into a new collection and wait until it is done
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles --wait

# List the sample dataset loads of a space
qcloud serverless space dataset load list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60
```

### Options

```
  -h, --help[=true|false]   help for dataset
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

* [qcloud serverless space](qcloud_serverless_space.md)	 - Manage serverless spaces
* [qcloud serverless space dataset describe](qcloud_serverless_space_dataset_describe.md)	 - Describe a sample dataset
* [qcloud serverless space dataset list](qcloud_serverless_space_dataset_list.md)	 - List the sample datasets that can be loaded into a space
* [qcloud serverless space dataset load](qcloud_serverless_space_dataset_load.md)	 - Manage sample dataset loads of a space

