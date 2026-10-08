## qcloud skills

Manage the qcloud Agent Skill for AI coding agents

### Synopsis

Manage the qcloud Agent Skill, which teaches AI coding agents how to use qcloud.

The skill follows the Agent Skills format (https://agentskills.io) and is read
by Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, OpenCode and other
agents. It contains a guide to authentication, conventions and common
workflows, plus a reference for every command group generated from this
binary's help text, so it always matches the installed qcloud version.
Re-run "qcloud skills install" after upgrading qcloud to refresh it.

### Examples

```
# Install the skill into the current project for detected agents
qcloud skills install

# Install the skill for all your projects
qcloud skills install --global

# Print the skill to stdout
qcloud skills print
```

### Options

```
  -h, --help   help for skills
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

* [qcloud](qcloud.md)	 - Qdrant Cloud CLI
* [qcloud skills install](qcloud_skills_install.md)	 - Install the qcloud Agent Skill for AI coding agents
* [qcloud skills print](qcloud_skills_print.md)	 - Print the qcloud Agent Skill to stdout
* [qcloud skills uninstall](qcloud_skills_uninstall.md)	 - Remove the qcloud Agent Skill

