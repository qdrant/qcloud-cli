package context

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// apiKeyHelpers maps --api-key-helper names to command templates.
// The placeholder %s is replaced with the --api-key-ref value.
var apiKeyHelpers = map[string]string{
	"1password": "op read %s",
	"vault":     "vault kv get -field=api_key %s",
	"pass":      "pass show %s",
	"keychain":  "security find-generic-password -s %s -w",
}

func newSetCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Example: `# Save the current configuration as a named context
qcloud context set production

# Create a context with explicit values
qcloud context set staging --api-key sk-... --account-id acc-123

# Use an external command to resolve the API key
qcloud context set staging --api-key-command 'op read op://vault/qdrant/api-key' --account-id acc-123

# Use a named helper preset
qcloud context set staging --api-key-helper 1password --api-key-ref op://vault/qdrant/api-key --account-id acc-123`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "set <name>",
				Short: "Create or update a context",
				Long: `Create or update a named context in the configuration file.

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
still take precedence over the command at runtime.`,
				Args: util.ExactArgs(1, "a context name"),
			}
			cmd.Flags().String("endpoint", "", "API endpoint for this context")
			cmd.Flags().String("api-key", "", "API key for this context")
			cmd.Flags().String("api-key-command", "", "Shell command that outputs the API key (e.g. 'op read op://vault/qdrant/api-key')")
			cmd.Flags().String("api-key-helper", "", "Named credential helper (1password, vault, pass, keychain)")
			_ = cmd.RegisterFlagCompletionFunc("api-key-helper", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				names := make([]string, 0, len(apiKeyHelpers))
				for k := range apiKeyHelpers {
					names = append(names, k)
				}
				sort.Strings(names)
				return names, cobra.ShellCompDirectiveNoFileComp
			})
			cmd.Flags().String("api-key-ref", "", "Secret reference for the credential helper")
			cmd.Flags().String("account-id", "", "Account ID for this context")
			cmd.MarkFlagsRequiredTogether("api-key-helper", "api-key-ref")
			cmd.MarkFlagsMutuallyExclusive("api-key", "api-key-command")
			cmd.MarkFlagsMutuallyExclusive("api-key", "api-key-helper")
			cmd.MarkFlagsMutuallyExclusive("api-key-command", "api-key-helper")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			name := args[0]

			ctx, _ := s.Config.GetContext(name)
			ctx.Name = name

			if ep, ok := flagChangedWithValue(cmd, "endpoint"); ok {
				ctx.Endpoint = ep
			} else {
				ctx.Endpoint = s.Config.Endpoint()
			}

			apiKeyCommand, err := resolveAPIKeyFlags(cmd)
			if err != nil {
				return err
			}

			if apiKeyCommand != "" {
				ctx.APIKeyCommand = apiKeyCommand
				ctx.APIKey = ""
			} else if key, ok := flagChangedWithValue(cmd, "api-key"); ok {
				ctx.APIKey = key
				ctx.APIKeyCommand = ""
			} else if ctx.APIKey == "" && ctx.APIKeyCommand == "" {
				ctx.APIKey = s.Config.APIKey()
			}

			if id, ok := flagChangedWithValue(cmd, "account-id"); ok {
				ctx.AccountID = id
			} else {
				ctx.AccountID = s.Config.AccountID()
			}

			if ctx.Endpoint == "" {
				return errors.New("cannot set a context with an empty endpoint")
			}

			if ctx.APIKey == "" && ctx.APIKeyCommand == "" {
				return errors.New("cannot set a context with an empty API key (use --api-key, --api-key-command, or --api-key-helper)")
			}

			if ctx.AccountID == "" {
				return errors.New("cannot set a context with an empty account id")
			}

			s.Config.UpsertContext(ctx)
			activated := false
			if s.Config.CurrentContext() == "" {
				s.Config.SetCurrentContext(name)
				activated = true
			}
			if err := s.Config.WriteToFile(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Context %q saved.\n", name)
			if activated {
				fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", name)
			}
			return nil
		},
	}.CobraCommand(s)
}

// resolveAPIKeyFlags returns the api_key_command string from either
// --api-key-command or --api-key-helper + --api-key-ref.
// Returns empty string if neither is set.
func resolveAPIKeyFlags(cmd *cobra.Command) (string, error) {
	if cmdStr, ok := flagChangedWithValue(cmd, "api-key-command"); ok {
		return cmdStr, nil
	}

	helper, helperSet := flagChangedWithValue(cmd, "api-key-helper")
	ref, _ := flagChangedWithValue(cmd, "api-key-ref")

	if !helperSet {
		return "", nil
	}

	tmpl, ok := apiKeyHelpers[helper]
	if !ok {
		names := make([]string, 0, len(apiKeyHelpers))
		for k := range apiKeyHelpers {
			names = append(names, k)
		}
		sort.Strings(names)
		return "", fmt.Errorf("unknown api-key-helper %q (supported: %s)", helper, strings.Join(names, ", "))
	}

	return fmt.Sprintf(tmpl, quoteShellArg(ref)), nil
}

func flagChangedWithValue(cmd *cobra.Command, name string) (string, bool) {
	if !cmd.Flags().Changed(name) {
		return "", false
	}
	v, _ := cmd.Flags().GetString(name)
	return v, true
}

var shellArgRegex = regexp.MustCompile(`[^\w@%+=:,./-]`)

// Quote returns a shell-escaped version of the string s. The returned value
// is a string that can safely be used as one token in a shell command line.
func quoteShellArg(s string) string {
	if len(s) == 0 {
		return "''"
	}

	if shellArgRegex.MatchString(s) {
		return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
	}

	return s
}

