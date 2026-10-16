## qcloud skills install

Install the qcloud Agent Skill for AI coding agents

### Synopsis

Install the qcloud Agent Skill so AI coding agents can discover it.

By default the skill is installed into the current project: always into
.agents/skills (read by Codex, Cursor, Gemini CLI, GitHub Copilot and OpenCode),
and into .claude/skills when a .claude directory exists. Use --agent to pick
targets explicitly, --global to install into your home directory for every
project, or --dir to write to any skills directory.

Installing replaces any previous qcloud skill in the target, so re-run this
command after upgrading qcloud to keep the skill in sync with the binary.

```
qcloud skills install [flags]
```

### Examples

```
# Install into the current project for detected agents
qcloud skills install

# Install for Claude Code only, for all projects
qcloud skills install --agent claude --global

# Install into a custom skills directory
qcloud skills install --dir ./tools/skills
```

### Options

```
      --agent stringArray   Agent to target (agents, claude); can be specified multiple times
      --dir string          Skills directory to target; the skill lives in <dir>/qcloud (ignores --agent and --global)
      --global              Target the user-level skills directories in your home directory instead of the current project
  -h, --help                help for install
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

