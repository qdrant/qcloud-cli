# qcloud iam

Manage IAM resources for the Qdrant Cloud account.

Commands:

- `qcloud iam key create`: Create a cloud management key
- `qcloud iam key delete`: Delete a cloud management key
- `qcloud iam key list`: List cloud management keys
- `qcloud iam permission list`: List all available permissions
- `qcloud iam role assign-permission`: Add permissions to a role
- `qcloud iam role create`: Create a custom role
- `qcloud iam role delete`: Delete a custom role
- `qcloud iam role describe`: Describe a role
- `qcloud iam role list`: List all roles
- `qcloud iam role remove-permission`: Remove permissions from a role
- `qcloud iam role update`: Update a custom role
- `qcloud iam user assign-role`: Assign one or more roles to a user
- `qcloud iam user delete`: Delete a user
- `qcloud iam user describe`: Describe a user and their assigned roles
- `qcloud iam user list`: List users in the account
- `qcloud iam user remove-role`: Remove one or more roles from a user

## qcloud iam key create

Create a new cloud management key for the account.

Management keys grant access to the Qdrant Cloud API. The full key value is returned
only once at creation time — store it securely, as it cannot be retrieved again. If a
key is lost, delete it and create a new one.

```
qcloud iam key create
```

Examples:

```sh
# Create a new management key
qcloud iam key create

# Create and capture the key value in a script
qcloud iam key create --json | jq -r '.key'
```

## qcloud iam key delete

Delete a cloud management key from the account.

Deleting a key immediately revokes its access to the Qdrant Cloud API. Any client
using the deleted key will receive authentication errors. This action cannot be undone.

A confirmation prompt is shown unless --force is passed.

```
qcloud iam key delete <key-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete a management key (with confirmation prompt)
qcloud iam key delete a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Delete without confirmation
qcloud iam key delete a1b2c3d4-e5f6-7890-abcd-ef1234567890 --force
```

## qcloud iam key list

List all cloud management keys for the account.

Management keys grant access to the Qdrant Cloud API and are used to authenticate CLI
and API requests. Each key is identified by its ID and a prefix — the prefix represents
the first bytes of the key value and is safe to display.

```
qcloud iam key list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List all management keys for the account
qcloud iam key list

# Output as JSON
qcloud iam key list --json
```

## qcloud iam permission list

List all permissions known in the system for the account.

Permissions are the individual access rights that can be assigned to roles.
Each permission has a value (e.g. "read:clusters") and a category
(e.g. "Cluster").

```
qcloud iam permission list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List all available permissions
qcloud iam permission list

# Output as JSON
qcloud iam permission list --json
```

## qcloud iam role assign-permission

Add permissions to a custom role.

Fetches the role's current permissions, merges the new ones (deduplicating),
and updates the role. Use "qcloud iam permission list" to see available
permissions.

```
qcloud iam role assign-permission <role-id> [flags]
```

Flags:

```
      --permission strings   Permission to add (repeatable)
```

Examples:

```sh
# Add a single permission
qcloud iam role assign-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe --permission read:clusters

# Add multiple permissions
qcloud iam role assign-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --permission read:clusters --permission read:backups
```

## qcloud iam role create

Create a new custom role for the account.

Custom roles allow fine-grained access control by combining specific permissions.
Use "qcloud iam permission list" to see available permissions.

```
qcloud iam role create [flags]
```

Flags:

```
      --description string   Description of the role
      --name string          Name of the role (4-64 characters)
      --permission strings   Permission to assign (repeatable)
```

Examples:

```sh
# Create a role with specific permissions
qcloud iam role create --name "Cluster Viewer" --permission read:clusters --permission read:cluster-endpoints

# Create a role with a description
qcloud iam role create --name "Backup Manager" --description "Can manage backups" \
  --permission read:clusters --permission read:backups --permission write:backups
```

## qcloud iam role delete

Delete a custom role from the account.

Only custom roles can be deleted. System roles are managed by Qdrant and cannot
be removed.

```
qcloud iam role delete <role-id> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete a role (with confirmation prompt)
qcloud iam role delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud iam role delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force
```

## qcloud iam role describe

Display detailed information about a role, including its name, type,
description, and the full list of assigned permissions.

```
qcloud iam role describe <role-id>
```

Examples:

```sh
# Describe a role
qcloud iam role describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Output as JSON
qcloud iam role describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json
```

## qcloud iam role list

List all roles for the account, including both system and custom roles.

System roles are managed by Qdrant and cannot be modified. Custom roles are
created and managed by the account administrator.

```
qcloud iam role list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List all roles
qcloud iam role list

# Output as JSON
qcloud iam role list --json
```

## qcloud iam role remove-permission

Remove permissions from a custom role.

Fetches the role's current permissions, removes the specified ones, and updates
the role. A role must retain at least one permission.

```
qcloud iam role remove-permission <role-id> [flags]
```

Flags:

```
      --permission strings   Permission to remove (repeatable)
```

Examples:

```sh
# Remove a single permission
qcloud iam role remove-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe --permission read:clusters

# Remove multiple permissions
qcloud iam role remove-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --permission read:clusters --permission read:backups
```

## qcloud iam role update

Update the name or description of a custom role.

Only custom roles can be updated. System roles are managed by Qdrant and cannot
be modified. To change a role's permissions, use the assign-permission and
remove-permission subcommands.

```
qcloud iam role update <role-id> [flags]
```

Flags:

```
      --description string   New description for the role
      --name string          New name for the role
```

Examples:

```sh
# Rename a role
qcloud iam role update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --name "New Name"

# Update the description
qcloud iam role update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --description "Updated description"
```

## qcloud iam user assign-role

Assign one or more roles to a user in the account.

Accepts either a user ID (UUID) or an email address to identify the user.
Each role accepts either a role UUID or a role name, which is
resolved to an ID via the IAM service. Prints the user's resulting roles
after the assignment.

```
qcloud iam user assign-role <user-id-or-email> [flags]
```

Flags:

```
  -r, --role strings   A role ID or name
```

Examples:

```sh
# Assign a role by name
qcloud iam user assign-role user@example.com --role admin

# Assign a role by ID
qcloud iam user assign-role user@example.com --role 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Assign multiple roles at once
qcloud iam user assign-role user@example.com --role admin --role viewer

# Assign multiple roles at once using comma separated values
qcloud iam user assign-role user@example.com --role admin,viewer
```

## qcloud iam user delete

Delete a user from Qdrant Cloud.

Accepts either a user ID (UUID) or an email address to identify the user.
Deleting a user is permanent and cannot be undone. Deletion fails if the user
still owns any accounts; ownership of those accounts must be transferred first.

A confirmation prompt is shown unless --force is passed.

```
qcloud iam user delete <user-id-or-email> [flags]
```

Flags:

```
  -f, --force   Skip confirmation prompt
```

Examples:

```sh
# Delete a user by email (with confirmation prompt)
qcloud iam user delete user@example.com

# Delete a user by ID
qcloud iam user delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud iam user delete user@example.com --force
```

## qcloud iam user describe

Describe a user and their assigned roles.

Accepts either a user ID (UUID) or an email address. Displays the user's
details and the roles currently assigned to them in the account.

```
qcloud iam user describe <user-id-or-email>
```

Examples:

```sh
# Describe a user by ID
qcloud iam user describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Describe a user by email
qcloud iam user describe user@example.com

# Output as JSON
qcloud iam user describe user@example.com --json
```

## qcloud iam user list

List users in the account.

Lists all users who are members of the current account. Requires the read:users
permission.

```
qcloud iam user list [flags]
```

Flags:

```
      --no-headers   Do not print column headers
```

Examples:

```sh
# List all users in the account
qcloud iam user list

# Output as JSON
qcloud iam user list --json
```

## qcloud iam user remove-role

Remove one or more roles from a user in the account.

Accepts either a user ID (UUID) or an email address to identify the user.
Each role accepts either a role UUID or a role name, which is
resolved to an ID via the IAM service. Prints the user's resulting roles
after the removal.

```
qcloud iam user remove-role <user-id-or-email> [flags]
```

Flags:

```
  -r, --role strings   A role ID or name
```

Examples:

```sh
# Remove a role by name
qcloud iam user remove-role user@example.com --role admin

# Remove a role by ID
qcloud iam user remove-role user@example.com --role 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Remove multiple roles at once
qcloud iam user remove-role user@example.com --role admin --role viewer
```
