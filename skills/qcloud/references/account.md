# qcloud account

Manage Qdrant Cloud accounts and their members.

Use these commands to list, inspect, and update accounts that the current
management key has access to. Account member commands show who belongs to the
current account and whether they are the owner.

Commands:

- `qcloud account describe`: Describe an account
- `qcloud account list`: List accounts
- `qcloud account member describe`: Describe an account member
- `qcloud account member list`: List account members
- `qcloud account update`: Update an account

## qcloud account describe

Describe an account by its ID.

If no account ID is provided, the current account (from --account-id, the
active context, or the QDRANT_CLOUD_ACCOUNT_ID environment variable) is used.

```
qcloud account describe [account-id]
```

Examples:

```sh
# Describe the current account
qcloud account describe

# Describe a specific account
qcloud account describe a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Output as JSON
qcloud account describe --json
```

## qcloud account list

List all accounts associated with the authenticated management key.

Returns every account the current API key has access to. No account ID is
required because the server resolves accounts from the caller's credentials.

```
qcloud account list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List all accessible accounts
qcloud account list

# Output as JSON
qcloud account list --json
```

## qcloud account member describe

Describe a member of the current account by their user ID.

Shows the member's user details and whether they are the account owner.

```
qcloud account member describe <user-id>
```

Examples:

```sh
# Describe a member
qcloud account member describe a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Output as JSON
qcloud account member describe a1b2c3d4-e5f6-7890-abcd-ef1234567890 --json
```

## qcloud account member list

List all members of the current account.

Each member has an associated user record and an ownership flag. Use
"qcloud iam user list" to see users with their status, or this command to
see who is in the account and who owns it.

```
qcloud account member list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List all members
qcloud account member list

# Output as JSON
qcloud account member list --json
```

## qcloud account update

Update an account's name or company information.

If no account ID is provided, the current account (from --account-id, the
active context, or the QDRANT_CLOUD_ACCOUNT_ID environment variable) is used.

Only flags that are explicitly set are applied. Unset flags leave the existing
values unchanged.

```
qcloud account update [account-id] [flags]
```

Flags:

```
      --company-domain string   Company domain
      --company-name string     Company name
      --name string             New account name
```

Examples:

```sh
# Rename the current account
qcloud account update --name "Production Account"

# Update company information on a specific account
qcloud account update a1b2c3d4-e5f6-7890-abcd-ef1234567890 --company-name "Acme Corp" --company-domain acme.com

# Output as JSON
qcloud account update --name "New Name" --json
```
