# qcloud package

Manage packages

## qcloud package list

List available packages for cluster creation

```
qcloud package list [flags]
```

Flags:

```
      --cloud-provider string   Cloud provider ID (required)
      --cloud-region string     Cloud provider region ID (required for non-hybrid providers)
      --no-headers              Do not print column headers
```

Examples:

```sh
# List packages for a cloud provider and region
qcloud package list --cloud-provider aws --cloud-region eu-central-1

# List packages for a hybrid cloud provider (no region required)
qcloud package list --cloud-provider hybrid
```
