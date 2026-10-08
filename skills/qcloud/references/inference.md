# qcloud inference

Inspect the inference models offered by Qdrant Cloud.

Inference models turn text or images into dense, sparse or multi vectors directly
inside Qdrant Cloud, so a cluster can embed documents and queries without a
separate embedding service. Which models are offered depends on the cloud
provider and region a cluster runs in.

## qcloud inference models list

List the inference models globally available for a cloud provider and region.

The listing is global, not account-specific: it shows every model Qdrant Cloud
offers in that region, together with the vector type and modality it produces,
its output dimensionality and per-request token limit, and its price per one
million processed tokens. External models are served by a third-party vendor and
require that vendor's API key to be configured on the cluster.

Inference is only offered on managed cloud, so unlike "qcloud package
list" this command cannot be used with hybrid cloud.

```
qcloud inference models list [flags]
```

Flags:

```
      --cloud-provider string   Cloud provider ID (required)
      --cloud-region string     Cloud provider region ID (required)
      --no-headers              Do not print column headers
```

Examples:

```sh
# List inference models available on AWS in eu-central-1
qcloud inference models list --cloud-provider aws --cloud-region eu-central-1

# List inference models as JSON
qcloud inference models list --cloud-provider gcp --cloud-region us-east4 --json
```
