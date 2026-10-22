package auth

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

// maxContextNameLength keeps generated context names readable.
const maxContextNameLength = 60

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

// slug turns an account name into a context name part: lowercase, with dashes.
func slug(name string) string {
	s := strings.Trim(nonSlugChars.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(s) > maxContextNameLength {
		s = strings.TrimRight(s[:maxContextNameLength], "-")
	}

	return s
}

// contextPrefix names the environment of an endpoint: none for production (grpc.cloud.qdrant.io),
// "dev-" for grpc.development-cloud.qdrant.io, "staging-" for grpc.staging-cloud.qdrant.io, and the
// host otherwise.
func contextPrefix(endpoint string) string {
	host := endpoint
	if h, _, err := net.SplitHostPort(endpoint); err == nil {
		host = h
	}

	cluster, ok := strings.CutPrefix(host, "grpc.")
	if ok {
		if cluster, ok = strings.CutSuffix(cluster, ".qdrant.io"); ok {
			switch cluster = strings.TrimSuffix(cluster, "-cloud"); cluster {
			case "cloud", "":
				return ""
			case "development":
				return "dev-"
			default:
				return slug(cluster) + "-"
			}
		}
	}

	return slug(host) + "-"
}

// syncOAuthContexts keeps one OAuth context per granted account for endpoint: it creates a context for
// a new account, keeps the name of an existing one, and removes the OAuth contexts of accounts that are
// no longer granted. Contexts with an API key, and OAuth contexts of other endpoints, are never touched.
// When no context is current, the first account's context becomes the current one.
// It returns the context names in account order, and the name it made current, if any.
func syncOAuthContexts(s *state.State, endpoint string, accounts []*accountv1.Account) ([]string, string, error) {
	existingByAccount := map[string]string{}
	taken := map[string]bool{}
	for _, ctx := range s.Config.Contexts() {
		taken[ctx.Name] = true
		if ctx.IsOAuth() && ctx.Endpoint == endpoint {
			existingByAccount[ctx.AccountID] = ctx.Name
		}
	}

	granted := map[string]bool{}
	names := make([]string, 0, len(accounts))
	for _, account := range accounts {
		granted[account.GetId()] = true

		name, ok := existingByAccount[account.GetId()]
		if !ok {
			name = uniqueContextName(contextPrefix(endpoint), account, taken)
			taken[name] = true
		}

		s.Config.UpsertContext(config.ContextEntry{
			Name:      name,
			Endpoint:  endpoint,
			AccountID: account.GetId(),
			Auth:      config.AuthOAuth,
		})
		names = append(names, name)
	}

	for accountID, name := range existingByAccount {
		if !granted[accountID] {
			s.Config.DeleteContext(name)
		}
	}

	selected := ""
	if len(names) > 0 && s.Config.CurrentContext() == "" {
		s.Config.SetCurrentContext(names[0])
		selected = names[0]
	}

	if err := s.Config.WriteToFile(); err != nil {
		return nil, "", err
	}

	return names, selected, nil
}

// uniqueContextName is prefix + the account's name, with a suffix when that name is taken.
func uniqueContextName(prefix string, account *accountv1.Account, taken map[string]bool) string {
	base := slug(account.GetName())
	if base == "" {
		base = "account-" + slug(account.GetId())
	}

	name := prefix + base
	for i := 2; taken[name]; i++ {
		name = fmt.Sprintf("%s%s-%d", prefix, base, i)
	}

	return name
}

// removeOAuthContexts deletes the OAuth contexts of endpoint and returns their names.
func removeOAuthContexts(s *state.State, endpoint string) ([]string, error) {
	var removed []string
	for _, ctx := range s.Config.Contexts() {
		if ctx.IsOAuth() && ctx.Endpoint == endpoint {
			s.Config.DeleteContext(ctx.Name)
			removed = append(removed, ctx.Name)
		}
	}

	if len(removed) == 0 {
		return nil, nil
	}

	return removed, s.Config.WriteToFile()
}
