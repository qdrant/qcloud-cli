package space_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceKeyList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Returns(&spaceauthv1.ListSpaceApiKeysResponse{
		Items: []*spaceauthv1.SpaceApiKey{
			{
				Id:        "key-global",
				Name:      "admin-key",
				Postfix:   "abcd",
				CreatedAt: timestamppb.Now(),
				ExpiresAt: timestamppb.New(time2027()),
				State:     &spaceauthv1.SpaceApiKeyState{Phase: spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_READY},
				AccessRules: []*spaceauthv1.AccessRule{{
					Scope: &spaceauthv1.AccessRule_GlobalAccess{GlobalAccess: &spaceauthv1.GlobalAccessRule{
						AccessType: spaceauthv1.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_MANAGE,
					}},
				}},
			},
			{
				Id:      "key-coll",
				Name:    "app-key",
				Postfix: "wxyz",
				State:   &spaceauthv1.SpaceApiKeyState{Phase: spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_PROCESSING},
				AccessRules: []*spaceauthv1.AccessRule{
					{Scope: &spaceauthv1.AccessRule_CollectionAccess{CollectionAccess: &spaceauthv1.CollectionAccessRule{
						CollectionName: "products",
						AccessType:     spaceauthv1.CollectionAccessRuleAccessType_COLLECTION_ACCESS_RULE_ACCESS_TYPE_READ_WRITE,
					}}},
					{Scope: &spaceauthv1.AccessRule_CollectionAccess{CollectionAccess: &spaceauthv1.CollectionAccessRule{
						CollectionName: "reviews",
						AccessType:     spaceauthv1.CollectionAccessRuleAccessType_COLLECTION_ACCESS_RULE_ACCESS_TYPE_READ_ONLY,
					}}},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "key", "list", "space-abc")
	require.NoError(t, err)
	for _, h := range []string{"ID", "NAME", "PHASE", "ACCESS", "POSTFIX", "EXPIRES", "CREATED"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "key-global")
	assert.Contains(t, stdout, "admin-key")
	assert.Contains(t, stdout, "READY")
	assert.Contains(t, stdout, "MANAGE")
	assert.Contains(t, stdout, "abcd")
	assert.Contains(t, stdout, "2027-06-15")
	assert.Contains(t, stdout, "PROCESSING")
	assert.Contains(t, stdout, "products:READ_WRITE, reviews:READ_ONLY")

	req, ok := env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
}

func TestSpaceKeyList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Returns(&spaceauthv1.ListSpaceApiKeysResponse{
		Items: []*spaceauthv1.SpaceApiKey{{Id: "key-json", Name: "json-key", SpaceId: "space-abc"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "key", "list", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			SpaceID string `json:"spaceId"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "key-json", result.Items[0].ID)
	assert.Equal(t, "json-key", result.Items[0].Name)
	assert.Equal(t, "space-abc", result.Items[0].SpaceID)
}

func TestSpaceKeyList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "list", "space-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list API keys")
}

func TestSpaceKeyList_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "list")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceApiKeyServer.ListSpaceApiKeysCalls.Count())
}
