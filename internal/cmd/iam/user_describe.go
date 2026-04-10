package iam

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUserDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*iamv1.User]{
		Use:               "describe <user-id-or-email>",
		Short:             "Describe a user and their assigned roles",
		Args:              util.ExactArgs(1, "a user ID or email"),
		ValidArgsFunction: userCompletion(s),
		Long: `Describe a user and their assigned roles.

Accepts either a user ID (UUID) or an email address. Displays the user's
details and the roles currently assigned to them in the account.`,
		Example: `# Describe a user by ID
qcloud iam user describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Describe a user by email
qcloud iam user describe user@example.com

# Output as JSON
qcloud iam user describe user@example.com --json`,
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*iamv1.User, error) {
			client, err := s.Client(cmd.Context())
			if err != nil {
				return nil, err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}
			return resolveUser(cmd, client, accountID, args[0])
		},
		PrintText: func(cmd *cobra.Command, w io.Writer, user *iamv1.User) error {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return err
			}
			rolesResp, err := client.IAM().ListUserRoles(ctx, &iamv1.ListUserRolesRequest{
				AccountId: accountID,
				UserId:    user.GetId(),
			})
			if err != nil {
				return fmt.Errorf("failed to list user roles: %w", err)
			}
			roles := rolesResp.GetRoles()
			return printUserWithRoles(w, user, roles, effectivePermissions(roles))
		},
	}.CobraCommand(s)
}

func printUserWithRoles(w io.Writer, user *iamv1.User, roles []*iamv1.Role, permissions []rolePermission) error {
	fmt.Fprintf(w, "ID:      %s\n", user.GetId())
	fmt.Fprintf(w, "Email:   %s\n", user.GetEmail())
	fmt.Fprintf(w, "Status:  %s\n", output.UserStatus(user.GetStatus()))
	if user.GetCreatedAt() != nil {
		t := user.GetCreatedAt().AsTime()
		fmt.Fprintf(w, "Created: %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
	}
	if len(roles) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Roles:")
		printRoles(w, roles)
	}
	if len(permissions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Effective Permissions:")
		printPermissions(w, permissions)
	}
	return nil
}

func printRoles(w io.Writer, roles []*iamv1.Role) {
	t := output.NewTable[*iamv1.Role](w)
	t.AddField("ID", func(v *iamv1.Role) string { return v.GetId() })
	t.AddField("NAME", func(v *iamv1.Role) string { return v.GetName() })
	t.Write(roles)
}

type rolePermission struct {
	permission *iamv1.Permission
	roleNames  []string
}

func printPermissions(w io.Writer, rps []rolePermission) {
	t := output.NewTable[rolePermission](w)
	t.AddField("PERMISSION", func(v rolePermission) string { return v.permission.GetValue() })
	t.AddField("CATEGORY", func(v rolePermission) string { return v.permission.GetCategory() })
	t.AddField("FROM ROLES", func(v rolePermission) string { return strings.Join(v.roleNames, ", ") })
	t.Write(rps)
}

// effectivePermissions collects unique permissions across all roles, with the
// sorted list of role names that grant each permission. Results are sorted by
// permission value.
func effectivePermissions(roles []*iamv1.Role) []rolePermission {
	type entry struct {
		permission *iamv1.Permission
		roleNames  []string
	}
	seen := make(map[string]*entry)
	order := []string{}
	for _, role := range roles {
		for _, p := range role.GetPermissions() {
			v := p.GetValue()
			if e, ok := seen[v]; ok {
				e.roleNames = append(e.roleNames, role.GetName())
			} else {
				seen[v] = &entry{permission: p, roleNames: []string{role.GetName()}}
				order = append(order, v)
			}
		}
	}
	sort.Strings(order)
	out := make([]rolePermission, 0, len(order))
	for _, v := range order {
		e := seen[v]
		sort.Strings(e.roleNames)
		out = append(out, rolePermission{permission: e.permission, roleNames: e.roleNames})
	}
	return out
}
