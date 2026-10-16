## qcloud skills uninstall

Remove the qcloud Agent Skill

### Synopsis

Remove the qcloud Agent Skill from agent skills directories.

By default the skill is removed from every supported agent directory in the
current project. Use --agent, --global and --dir to choose targets the same
way as "qcloud skills install". Only directories containing the qcloud skill
are removed.

```
qcloud skills uninstall [flags]
```

### Examples

```
# Remove the skill from the current project
qcloud skills uninstall

# Remove the user-level skill without confirmation
qcloud skills uninstall --global --force
```

### Options

```
      --agent stringArray   Agent to target (agents, claude); can be specified multiple times
      --dir string          Skills directory to target; the skill lives in <dir>/qcloud (ignores --agent and --global)
  -f, --force               Skip confirmation prompt
      --global              Target the user-level skills directories in your home directory instead of the current project
  -h, --help                help for uninstall
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

