package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestListVersions_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	remarks := "upgrade recommended"
	env.ClusterServer.ListQdrantReleasesCalls.Returns(&clusterv1.ListQdrantReleasesResponse{
		Items: []*clusterv1.QdrantRelease{
			{Version: "1.14.0", Default: true},
			{Version: "1.13.0", EndOfLife: true},
			{Version: "1.12.0", Unavailable: true},
			{Version: "1.11.0", Remarks: &remarks},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "version", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "VERSION")
	assert.Contains(t, stdout, "DEFAULT")
	assert.Contains(t, stdout, "END OF LIFE")
	assert.Contains(t, stdout, "UNAVAILABLE")
	assert.Contains(t, stdout, "REMARKS")
	assert.Contains(t, stdout, "1.14.0")
	assert.Contains(t, stdout, "1.13.0")
	assert.Contains(t, stdout, "1.12.0")
	assert.Contains(t, stdout, "1.11.0")
	assert.Contains(t, stdout, "upgrade recommended")
}
