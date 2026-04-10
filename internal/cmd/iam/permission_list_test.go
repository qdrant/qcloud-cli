package iam_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestPermissionList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.IAMServer.ListPermissionsCalls.Returns(&iamv1.ListPermissionsResponse{
		Permissions: []*iamv1.Permission{
			{Value: "read:clusters", Category: new("Cluster")},
			{Value: "write:roles", Category: new("IAM")},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "permission", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "PERMISSION")
	assert.Contains(t, stdout, "CATEGORY")
	assert.Contains(t, stdout, "read:clusters")
	assert.Contains(t, stdout, "Cluster")
	assert.Contains(t, stdout, "write:roles")
	assert.Contains(t, stdout, "IAM")

	req, ok := env.IAMServer.ListPermissionsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestPermissionList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListPermissionsCalls.Returns(&iamv1.ListPermissionsResponse{
		Permissions: []*iamv1.Permission{
			{Value: "read:clusters", Category: new("Cluster")},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "permission", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Permissions []struct {
			Value    string `json:"value"`
			Category string `json:"category"`
		} `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Permissions, 1)
	assert.Equal(t, "read:clusters", result.Permissions[0].Value)
	assert.Equal(t, "Cluster", result.Permissions[0].Category)
}

func TestPermissionList_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListPermissionsCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "iam", "permission", "list")
	require.Error(t, err)
}
