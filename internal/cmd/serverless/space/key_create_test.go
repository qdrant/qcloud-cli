package space_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func time2027() time.Time {
	return time.Date(2027, 6, 15, 0, 0, 0, 0, time.UTC)
}

func createdKeyResponse(id string) *spaceauthv1.CreateSpaceApiKeyResponse {
	return &spaceauthv1.CreateSpaceApiKeyResponse{SpaceApiKey: &spaceauthv1.SpaceApiKey{
		Id:   id,
		Name: "my-key",
		Key:  "secret-key-value",
		State: &spaceauthv1.SpaceApiKeyState{
			Phase: spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_PROCESSING,
		},
	}}
}

func TestSpaceKeyCreate_Basic(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-new"), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc", "--name", "my-key")
	require.NoError(t, err)
	assert.Contains(t, stdout, "API key key-new (my-key) created.")
	assert.Contains(t, stdout, "not be shown again")
	assert.Contains(t, stdout, "secret-key-value")

	req, ok := env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Last()
	require.True(t, ok)
	key := req.GetSpaceApiKey()
	assert.Equal(t, "test-account-id", key.GetAccountId())
	assert.Equal(t, "space-abc", key.GetSpaceId())
	assert.Equal(t, "my-key", key.GetName())
	assert.Empty(t, key.GetAccessRules())
	assert.Nil(t, key.GetExpiresAt())
}

func TestSpaceKeyCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-json"), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc", "--name", "my-key", "--json")
	require.NoError(t, err)

	var result struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "key-json", result.ID)
	assert.Equal(t, "secret-key-value", result.Key)
}

func TestSpaceKeyCreate_GlobalAccessTypes(t *testing.T) {
	tests := []struct {
		flag string
		want spaceauthv1.GlobalAccessRuleAccessType
	}{
		{"manage", spaceauthv1.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_MANAGE},
		{"read-only", spaceauthv1.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_READ_ONLY},
		{"metrics-read-only", spaceauthv1.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_METRICS_READ_ONLY},
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-new"), nil)

			_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc",
				"--name", "my-key", "--access-type", tt.flag)
			require.NoError(t, err)

			req, ok := env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Last()
			require.True(t, ok)
			rules := req.GetSpaceApiKey().GetAccessRules()
			require.Len(t, rules, 1)
			require.NotNil(t, rules[0].GetGlobalAccess())
			assert.Equal(t, tt.want, rules[0].GetGlobalAccess().GetAccessType())
		})
	}
}

func TestSpaceKeyCreate_CollectionAccess(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-new"), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc", "--name", "my-key",
		"--collection", "products=read-write", "--collection", "reviews=read-only")
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Last()
	require.True(t, ok)
	rules := req.GetSpaceApiKey().GetAccessRules()
	require.Len(t, rules, 2)
	assert.Equal(t, "products", rules[0].GetCollectionAccess().GetCollectionName())
	assert.Equal(t, spaceauthv1.CollectionAccessRuleAccessType_COLLECTION_ACCESS_RULE_ACCESS_TYPE_READ_WRITE,
		rules[0].GetCollectionAccess().GetAccessType())
	assert.Equal(t, "reviews", rules[1].GetCollectionAccess().GetCollectionName())
	assert.Equal(t, spaceauthv1.CollectionAccessRuleAccessType_COLLECTION_ACCESS_RULE_ACCESS_TYPE_READ_ONLY,
		rules[1].GetCollectionAccess().GetAccessType())
}

func TestSpaceKeyCreate_WithExpires(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-new"), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc",
		"--name", "my-key", "--expires", "2027-06-15")
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.GetSpaceApiKey().GetExpiresAt())
	assert.Equal(t, time2027(), req.GetSpaceApiKey().GetExpiresAt().AsTime().UTC())
}

func TestSpaceKeyCreate_InputErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"missing name", []string{"space-abc"}, "name"},
		{"missing space id", []string{"--name", "my-key"}, "space ID"},
		{"invalid access type", []string{"space-abc", "--name", "k", "--access-type", "superuser"}, "superuser"},
		{"invalid expires", []string{"space-abc", "--name", "k", "--expires", "not-a-date"}, "YYYY-MM-DD"},
		{"collection without access", []string{"space-abc", "--name", "k", "--collection", "products"}, "name=read-only|read-write"},
		{"collection without name", []string{"space-abc", "--name", "k", "--collection", "=read-only"}, "name=read-only|read-write"},
		{"invalid collection access", []string{"space-abc", "--name", "k", "--collection", "products=manage"}, "read-only or read-write"},
		{"access type with collection", []string{
			"space-abc", "--name", "k", "--access-type", "manage", "--collection", "products=read-only",
		}, "none of the others can be"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "key", "create"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Equal(t, 0, env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Count())
		})
	}
}

func TestSpaceKeyCreate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc", "--name", "my-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create API key")
}

func listedKey(phase spaceauthv1.SpaceApiKeyStatePhase) *spaceauthv1.ListSpaceApiKeysResponse {
	return &spaceauthv1.ListSpaceApiKeysResponse{Items: []*spaceauthv1.SpaceApiKey{
		{Id: "other-key", State: &spaceauthv1.SpaceApiKeyState{Phase: spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_READY}},
		{Id: "key-new", Name: "my-key", State: &spaceauthv1.SpaceApiKeyState{Phase: phase, Reason: "boom"}},
	}}
}

func TestSpaceKeyCreate_WaitSuccess(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-new"), nil)
	env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.
		OnCall(0, func(_ context.Context, _ *spaceauthv1.ListSpaceApiKeysRequest) (*spaceauthv1.ListSpaceApiKeysResponse, error) {
			return &spaceauthv1.ListSpaceApiKeysResponse{}, nil
		}).
		OnCall(1, func(_ context.Context, _ *spaceauthv1.ListSpaceApiKeysRequest) (*spaceauthv1.ListSpaceApiKeysResponse, error) {
			return listedKey(spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_PROCESSING), nil
		}).
		Always(func(_ context.Context, _ *spaceauthv1.ListSpaceApiKeysRequest) (*spaceauthv1.ListSpaceApiKeysResponse, error) {
			return listedKey(spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_READY), nil
		})

	stdout, stderr, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc",
		"--name", "my-key", "--wait", "--wait-poll-interval", "10ms")
	require.NoError(t, err)
	assert.Contains(t, stderr, "phase=NOT_FOUND")
	assert.Contains(t, stderr, "phase=PROCESSING")
	assert.Contains(t, stderr, "phase=READY")
	assert.Contains(t, stdout, "API key key-new (my-key) created.")
	assert.Contains(t, stdout, "secret-key-value")

	req, ok := env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
}

func TestSpaceKeyCreate_WaitFailurePhase(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-new"), nil)
	env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Returns(
		listedKey(spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_DISABLED), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc",
		"--name", "my-key", "--wait", "--wait-poll-interval", "10ms")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "phase=DISABLED")
	assert.Contains(t, err.Error(), "reason=boom")
}

func TestSpaceKeyCreate_WaitTimeout(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-new"), nil)
	env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Returns(
		listedKey(spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_PROCESSING), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc",
		"--name", "my-key", "--wait", "--wait-timeout", "50ms", "--wait-poll-interval", "10ms")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for API key to become ready")
}

func TestSpaceKeyCreate_WaitListError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.CreateSpaceApiKeyCalls.Returns(createdKeyResponse("key-new"), nil)
	env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "create", "space-abc",
		"--name", "my-key", "--wait", "--wait-poll-interval", "10ms")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get API key status")
}
