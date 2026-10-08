# qcloud context

Manage named configuration contexts

Commands:

- `qcloud context delete`: Delete a context
- `qcloud context list`: List all contexts
- `qcloud context set`: Create or update a context
- `qcloud context show`: Show the active context configuration
- `qcloud context use`: Set the active context

## qcloud context delete

Delete a context

```
qcloud context delete <name> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete a context
qcloud context delete staging

# Delete without confirmation
qcloud context delete staging --force
```

## qcloud context list

List all contexts

```
qcloud context list
```

Examples:

```sh
# List all contexts
qcloud context list
```

## qcloud context set

Create or update a named context in the configuration file.

A context stores connection settings (endpoint, account ID) and API key
credentials under a name, so you can switch between environments with
"qcloud context use <name>".

There are three ways to provide the API key:

  --api-key              Store the key directly in the config file (plaintext).
  --api-key-command      Store a shell command that is executed at runtime to
                         retrieve the key. The command is run via "sh -c" and
                         its stdout is used as the API key. This avoids storing
                         secrets in plaintext.
  --api-key-helper       Use a named preset that generates the command for you.
                         Must be paired with --api-key-ref.

Supported helpers and the commands they generate:

  1password    op read <ref>
  vault        vault kv get -field=api_key <ref>
  pass         pass show <ref>
  keychain     security find-generic-password -s <ref> -w

When an api_key_command is set, any existing plaintext api_key is removed from
the context. Flags and environment variables (--api-key, QDRANT_CLOUD_API_KEY)
still take precedence over the command at runtime.

```
qcloud context set <name> [flags]
```

Flags:

```
      --account-id string        Account ID for this context
      --api-key string           API key for this context
      --api-key-command string   Shell command that outputs the API key (e.g. 'op read op://vault/qdrant/api-key')
      --api-key-helper string    Named credential helper (1password, vault, pass, keychain)
      --api-key-ref string       Secret reference for the credential helper
      --endpoint string          API endpoint for this context
```

Examples:

```sh
# Save the current configuration as a named context
qcloud context set production

# Create a context with explicit values
qcloud context set staging --api-key sk-... --account-id acc-123

# Use an external command to resolve the API key
qcloud context set staging --api-key-command 'op read op://vault/qdrant/api-key' --account-id acc-123

# Use a named helper preset
qcloud context set staging --api-key-helper 1password --api-key-ref op://vault/qdrant/api-key --account-id acc-123
```

## qcloud context show

Show the active context configuration

```
qcloud context show
```

Examples:

```sh
# Show the active context
qcloud context show
```

## qcloud context use

Set the active context

```
qcloud context use <name>
```

Examples:

```sh
# Switch to another context
qcloud context use production
```
