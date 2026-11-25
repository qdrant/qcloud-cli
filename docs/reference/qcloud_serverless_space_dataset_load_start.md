## qcloud serverless space dataset load start

Load a sample dataset into a new collection

### Synopsis

Load a sample dataset into a new collection of a serverless space.

The collection is created by the load, so no collection with that name may
exist in the space. Without --collection, the collection gets the ID of the
dataset as its name.

The load is asynchronous: the command returns as soon as the load is accepted.
With --wait it follows the load until it succeeds, and fails when the load
fails (for example because the space reached its collection limit). The load
is printed before the error, so its ID is not lost.

```
qcloud serverless space dataset load start <space-id> <dataset-id> [flags]
```

### Examples

```
# Load a sample dataset into a collection named after the dataset
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles

# Load a sample dataset into a collection named "styles" and wait until it is done
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles \
  --collection styles --wait

# Start a load and keep its ID for scripting
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles --json | jq -r .id
```

### Options

```
      --collection string       Name of the new collection (default: the dataset ID)
  -h, --help[=true|false]       help for start
      --wait[=true|false]       Wait for the sample dataset load to finish
      --wait-timeout duration   Maximum time to wait for the sample dataset load to finish (default 30m0s)
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

* [qcloud serverless space dataset load](qcloud_serverless_space_dataset_load.md)	 - Manage sample dataset loads of a space

