## qcloud serverless usage

Show the billed serverless usage per space

### Synopsis

Show the billed serverless usage of the current account per space and billable item.

Each row is one billable item of one space in one calendar month, as named on
the invoice (for example "Serverless Updated Data (GiB)"), with the quantity
consumed, the charge, the discount and the net charge. The net total of all
rows is shown below the table. Prepaid credits drawn for the period are not
subtracted.

By default, the period is the current calendar month (UTC) up to now. Use
--month for a whole calendar month, or --since and --until for any other
period. When only --until is given, the period starts on the first day of its
month. Usage is reported per calendar month, so a period that starts or ends
within a month only includes the usage metered in that part of the month.

For a parent billing account, the usage of all child accounts is included and
an ACCOUNT column is shown. Cluster usage and the prepaid credits drawn are
left out, also from --json. The API key needs the read:payment_information
permission.

```
qcloud serverless usage [flags]
```

### Examples

```
# Show the serverless usage of the current month so far
qcloud serverless usage

# Show the usage of October 2026
qcloud serverless usage --month 2026-10

# Show the usage of one space during the third quarter of 2026
qcloud serverless usage --space 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 2026-07-01 --until 2026-09-30

# Export the usage as JSON
qcloud serverless usage --month 2026-10 --json
```

### Options

```
  -h, --help[=true|false]   help for usage
      --month string        Calendar month in YYYY-MM format (UTC); cannot be combined with --since or --until
  -s, --since string        Start of the period (RFC3339, YYYY-MM-DD, or a duration ago such as 24h or 7d; default: first day of the month of --until)
      --space string        Only include the space with this ID
  -u, --until string        End of the period, exclusive (RFC3339, YYYY-MM-DD for the end of that day, or a duration ago such as 1h; default: now)
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

* [qcloud serverless](qcloud_serverless.md)	 - Manage Qdrant Cloud Serverless resources

