## qcloud skills print

Print the qcloud Agent Skill to stdout

### Synopsis

Print the qcloud Agent Skill to stdout without installing it.

Without flags, prints SKILL.md, the entry point an agent reads first. Use
--reference to print the generated reference for one command group, or
"global-flags" for the flags shared by every command. This is useful for
agents that do not support skill directories, or to inspect what
"qcloud skills install" would write.

```
qcloud skills print [flags]
```

### Examples

```
# Print SKILL.md
qcloud skills print

# Print the reference for the cluster command group
qcloud skills print --reference cluster

# Print the global flags reference
qcloud skills print --reference global-flags
```

### Options

```
  -h, --help               help for print
      --reference string   Print the reference for a command group instead of SKILL.md
```

### Options inherited from parent commands

```
      --account-id string    Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)
      --api-key string       Management API Key (env: QDRANT_CLOUD_API_KEY)
  -c, --config string        Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)
      --console-url string   Qdrant Cloud web console base URL (env: QDRANT_CLOUD_CONSOLE_URL, default https://cloud.qdrant.io)
      --context string       Override the active context (env: QDRANT_CLOUD_CONTEXT)
      --debug                Enable debug logging to stderr
      --endpoint string      gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)
      --json                 Output as JSON
```

### SEE ALSO

* [qcloud skills](qcloud_skills.md)	 - Manage the qcloud Agent Skill for AI coding agents

