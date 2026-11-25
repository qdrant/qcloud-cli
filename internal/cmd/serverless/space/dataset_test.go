package space_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	spacedatasetv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/dataset/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func sampleDatasets() *spacedatasetv1.ListSampleDatasetsResponse {
	return &spacedatasetv1.ListSampleDatasetsResponse{
		Items: []*spacedatasetv1.SampleDataset{
			{
				Id:               "midjourney-styles",
				Name:             "Midjourney Styles",
				Description:      "Image styles with text and image embeddings",
				PointCount:       5000,
				SizeBytes:        75 * 1024 * 1024,
				IndexedSizeBytes: 12 * 1024 * 1024,
				Version:          "v2",
				CollectionConfig: &spacedatasetv1.CollectionConfig{
					DenseVectors: map[string]*spacedatasetv1.DenseVectorConfig{
						"":      {Size: 384, Distance: spacedatasetv1.Distance_DISTANCE_COSINE, Model: "sentence-transformers/all-MiniLM-L6-v2"},
						"image": {Size: 512, Distance: spacedatasetv1.Distance_DISTANCE_DOT, Multivector: true, Model: "Qdrant/clip-ViT-B-32-vision"},
					},
					SparseVectors: map[string]*spacedatasetv1.SparseVectorConfig{
						"bm25": {UseIdf: true, Model: "Qdrant/bm25"},
					},
				},
			},
			{Id: "arxiv-titles", Name: "arXiv Titles", PointCount: 100, SizeBytes: 1024, Version: "v1"},
		},
	}
}

func TestSpaceDatasetList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.ListSampleDatasetsCalls.Returns(sampleDatasets(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "list", "space-abc")
	require.NoError(t, err)
	for _, h := range []string{"ID", "NAME", "POINTS", "SIZE", "INDEXED SIZE", "VECTORS", "VERSION"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "midjourney-styles")
	assert.Contains(t, stdout, "Midjourney Styles")
	assert.Contains(t, stdout, "5000")
	assert.Contains(t, stdout, "75 MiB")
	assert.Contains(t, stdout, "12 MiB")
	assert.Contains(t, stdout, "default, image, bm25 (sparse)")
	assert.Contains(t, stdout, "arxiv-titles")

	req, ok := env.ServerlessDatasetServer.ListSampleDatasetsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
}

func TestSpaceDatasetList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.ListSampleDatasetsCalls.Returns(sampleDatasets(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "list", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID         string `json:"id"`
			PointCount string `json:"pointCount"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 2)
	assert.Equal(t, "midjourney-styles", result.Items[0].ID)
	assert.Equal(t, "5000", result.Items[0].PointCount)
}

func TestSpaceDatasetList_Errors(t *testing.T) {
	t.Run("missing space ID", func(t *testing.T) {
		env := testutil.NewTestEnv(t)
		_, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "list")
		require.Error(t, err)
		assert.Equal(t, 0, env.ServerlessDatasetServer.ListSampleDatasetsCalls.Count())
	})

	t.Run("backend error", func(t *testing.T) {
		env := testutil.NewTestEnv(t)
		env.ServerlessDatasetServer.ListSampleDatasetsCalls.Returns(nil, fmt.Errorf("boom"))
		_, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "list", "space-abc")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list sample datasets")
	})
}

func TestSpaceDatasetDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.ListSampleDatasetsCalls.Returns(sampleDatasets(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "describe", "space-abc", "midjourney-styles")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID:            midjourney-styles")
	assert.Contains(t, stdout, "Version:       v2")
	assert.Contains(t, stdout, "Points:        5000")
	assert.Contains(t, stdout, "Size:          75 MiB")
	assert.Contains(t, stdout, "Indexed Size:  12 MiB")
	assert.Contains(t, stdout, "Description:   Image styles with text and image embeddings")
	assert.Contains(t, stdout, "Dense Vectors:")
	assert.Contains(t, stdout, "COSINE")
	assert.Contains(t, stdout, "sentence-transformers/all-MiniLM-L6-v2")
	assert.Contains(t, stdout, "DOT")
	assert.Contains(t, stdout, "Qdrant/clip-ViT-B-32-vision")
	assert.Contains(t, stdout, "Sparse Vectors:")
	assert.Contains(t, stdout, "Qdrant/bm25")
	assert.Contains(t, stdout, "default")
	assert.NotContains(t, stdout, "arxiv-titles")
}

func TestSpaceDatasetDescribe_WithoutVectors(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.ListSampleDatasetsCalls.Returns(sampleDatasets(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "describe", "space-abc", "arxiv-titles")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID:            arxiv-titles")
	assert.NotContains(t, stdout, "Description:")
	assert.NotContains(t, stdout, "Dense Vectors:")
	assert.NotContains(t, stdout, "Sparse Vectors:")
}

func TestSpaceDatasetDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.ListSampleDatasetsCalls.Returns(sampleDatasets(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "describe", "space-abc", "midjourney-styles", "--json")
	require.NoError(t, err)

	var result struct {
		ID               string `json:"id"`
		CollectionConfig struct {
			DenseVectors map[string]struct {
				Size  string `json:"size"`
				Model string `json:"model"`
			} `json:"denseVectors"`
		} `json:"collectionConfig"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "midjourney-styles", result.ID)
	assert.Equal(t, "512", result.CollectionConfig.DenseVectors["image"].Size)
}

func TestSpaceDatasetDescribe_NotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.ListSampleDatasetsCalls.Returns(sampleDatasets(), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "describe", "space-abc", "nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `sample dataset "nope" is not available in space space-abc`)
}

func TestSpaceDatasetDescribe_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)
	_, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "describe", "space-abc")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessDatasetServer.ListSampleDatasetsCalls.Count())
}

func sampleDatasetLoad(status spacedatasetv1.SampleDatasetLoadStatus) *spacedatasetv1.SampleDatasetLoad {
	return &spacedatasetv1.SampleDatasetLoad{
		Id:                   "load-123",
		CreatedAt:            timestamppb.Now(),
		AccountId:            "test-account-id",
		SpaceId:              "space-abc",
		SampleDatasetId:      "midjourney-styles",
		SampleDatasetVersion: "v2",
		CollectionName:       "styles",
		Status:               status,
	}
}

func TestSpaceDatasetLoadStart_DefaultCollection(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.LoadSampleDatasetCalls.Returns(&spacedatasetv1.LoadSampleDatasetResponse{
		SampleDatasetLoad: sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "start", "space-abc", "midjourney-styles")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Sample dataset load load-123 started")
	assert.Contains(t, stdout, "qcloud serverless space dataset load describe space-abc load-123")

	req, ok := env.ServerlessDatasetServer.LoadSampleDatasetCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Equal(t, "midjourney-styles", req.GetSampleDatasetId())
	assert.Equal(t, "midjourney-styles", req.GetCollectionName(), "the collection defaults to the dataset ID")
	assert.Equal(t, 0, env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Count())
}

func TestSpaceDatasetLoadStart_CustomCollectionJSON(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.LoadSampleDatasetCalls.Returns(&spacedatasetv1.LoadSampleDatasetResponse{
		SampleDatasetLoad: sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "start", "space-abc", "midjourney-styles",
		"--collection", "styles", "--json")
	require.NoError(t, err)

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "load-123", result.ID)
	assert.Equal(t, "SAMPLE_DATASET_LOAD_STATUS_RUNNING", result.Status)

	req, ok := env.ServerlessDatasetServer.LoadSampleDatasetCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "styles", req.GetCollectionName())
}

func TestSpaceDatasetLoadStart_InputErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing dataset ID", []string{"space-abc"}},
		{"empty collection", []string{"space-abc", "midjourney-styles", "--collection", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			args := append([]string{"serverless", "space", "dataset", "load", "start"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.ServerlessDatasetServer.LoadSampleDatasetCalls.Count())
		})
	}
}

func TestSpaceDatasetLoadStart_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.LoadSampleDatasetCalls.Returns(nil, fmt.Errorf("collection already exists"))

	_, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "start", "space-abc", "midjourney-styles")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load sample dataset")
}

func TestSpaceDatasetLoadStart_WaitSucceeds(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.LoadSampleDatasetCalls.Returns(&spacedatasetv1.LoadSampleDatasetResponse{
		SampleDatasetLoad: sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING),
	}, nil)

	running := sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING)
	running.Stats = &spacedatasetv1.SampleDatasetLoadStats{Progress: new("35.2/75.4 MiB (46%)")}
	succeeded := sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_SUCCEEDED)
	env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.
		OnCall(0, func(_ context.Context, _ *spacedatasetv1.GetSampleDatasetLoadRequest) (*spacedatasetv1.GetSampleDatasetLoadResponse, error) {
			return &spacedatasetv1.GetSampleDatasetLoadResponse{SampleDatasetLoad: running}, nil
		}).
		OnCall(1, func(_ context.Context, _ *spacedatasetv1.GetSampleDatasetLoadRequest) (*spacedatasetv1.GetSampleDatasetLoadResponse, error) {
			return &spacedatasetv1.GetSampleDatasetLoadResponse{SampleDatasetLoad: succeeded}, nil
		})

	stdout, stderr, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "start", "space-abc", "midjourney-styles",
		"--collection", "styles", "--wait", "--wait-poll-interval", "10ms")
	require.NoError(t, err)
	assert.Contains(t, stderr, "waiting for it to finish")
	assert.Contains(t, stderr, "status=RUNNING progress=35.2/75.4 MiB (46%)")
	assert.Contains(t, stderr, "status=SUCCEEDED")
	assert.Contains(t, stdout, "Sample dataset midjourney-styles loaded into collection styles (load load-123).")

	req, ok := env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Equal(t, "load-123", req.GetSampleDatasetLoadId())
	assert.Equal(t, 2, env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Count())
}

func TestSpaceDatasetLoadStart_WaitFails(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.LoadSampleDatasetCalls.Returns(&spacedatasetv1.LoadSampleDatasetResponse{
		SampleDatasetLoad: sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING),
	}, nil)

	failed := sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_FAILED)
	failed.Reason = "collection limit"
	env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Returns(&spacedatasetv1.GetSampleDatasetLoadResponse{SampleDatasetLoad: failed}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "start", "space-abc", "midjourney-styles",
		"--wait", "--wait-poll-interval", "10ms")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sample dataset load load-123 did not succeed")
	assert.Contains(t, err.Error(), "status=FAILED, reason=collection limit")
	// The load is printed with its last status before the error.
	assert.Contains(t, stdout, "ID:          load-123")
	assert.Contains(t, stdout, "Status:      FAILED")
	assert.Contains(t, stdout, "Reason:      collection limit")
}

func TestSpaceDatasetLoadStart_WaitTimeoutJSON(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.LoadSampleDatasetCalls.Returns(&spacedatasetv1.LoadSampleDatasetResponse{
		SampleDatasetLoad: sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING),
	}, nil)
	env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Returns(&spacedatasetv1.GetSampleDatasetLoadResponse{
		SampleDatasetLoad: sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "start", "space-abc", "midjourney-styles",
		"--wait", "--wait-timeout", "50ms", "--wait-poll-interval", "10ms", "--json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for sample dataset load to finish")

	var result struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "load-123", result.ID)
}

func TestSpaceDatasetLoadList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	running := sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING)
	running.Stats = &spacedatasetv1.SampleDatasetLoadStats{Progress: new("35.2/75.4 MiB (46%)")}
	done := sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_SUCCEEDED)
	done.Id = "load-456"
	done.Stats = &spacedatasetv1.SampleDatasetLoadStats{Duration: durationpb.New(95_400_000_000)}
	env.ServerlessDatasetServer.ListSampleDatasetLoadsCalls.Returns(&spacedatasetv1.ListSampleDatasetLoadsResponse{
		Items: []*spacedatasetv1.SampleDatasetLoad{running, done},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "list", "space-abc")
	require.NoError(t, err)
	for _, h := range []string{"ID", "DATASET", "VERSION", "COLLECTION", "STATUS", "PROGRESS", "DURATION", "CREATED"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "load-123")
	assert.Contains(t, stdout, "RUNNING")
	assert.Contains(t, stdout, "35.2/75.4 MiB (46%)")
	assert.Contains(t, stdout, "load-456")
	assert.Contains(t, stdout, "SUCCEEDED")
	assert.Contains(t, stdout, "1m35s")

	req, ok := env.ServerlessDatasetServer.ListSampleDatasetLoadsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Nil(t, req.PageSize)
}

func TestSpaceDatasetLoadList_FetchesAllPages(t *testing.T) {
	env := testutil.NewTestEnv(t)
	first := sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_SUCCEEDED)
	second := sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_FAILED)
	second.Id = "load-789"
	env.ServerlessDatasetServer.ListSampleDatasetLoadsCalls.
		OnCall(0, func(_ context.Context, _ *spacedatasetv1.ListSampleDatasetLoadsRequest) (*spacedatasetv1.ListSampleDatasetLoadsResponse, error) {
			return &spacedatasetv1.ListSampleDatasetLoadsResponse{Items: []*spacedatasetv1.SampleDatasetLoad{first}, NextPageToken: new("page-2")}, nil
		}).
		OnCall(1, func(_ context.Context, _ *spacedatasetv1.ListSampleDatasetLoadsRequest) (*spacedatasetv1.ListSampleDatasetLoadsResponse, error) {
			return &spacedatasetv1.ListSampleDatasetLoadsResponse{Items: []*spacedatasetv1.SampleDatasetLoad{second}}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "list", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
		NextPageToken *string `json:"nextPageToken"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 2)
	assert.Equal(t, "load-123", result.Items[0].ID)
	assert.Equal(t, "load-789", result.Items[1].ID)
	assert.Nil(t, result.NextPageToken)

	reqs := env.ServerlessDatasetServer.ListSampleDatasetLoadsCalls.All()
	require.Len(t, reqs, 2)
	assert.Equal(t, "page-2", reqs[1].GetPageToken())
}

func TestSpaceDatasetLoadList_ManualPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.ListSampleDatasetLoadsCalls.Returns(&spacedatasetv1.ListSampleDatasetLoadsResponse{
		Items:         []*spacedatasetv1.SampleDatasetLoad{sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING)},
		NextPageToken: new("page-3"),
		TotalSize:     new(int32(5)),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "list", "space-abc",
		"--page-size", "1", "--page-token", "page-2", "--json")
	require.NoError(t, err)

	var result struct {
		NextPageToken string `json:"nextPageToken"`
		TotalSize     int32  `json:"totalSize"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "page-3", result.NextPageToken)
	assert.Equal(t, int32(5), result.TotalSize)

	req, ok := env.ServerlessDatasetServer.ListSampleDatasetLoadsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, int32(1), req.GetPageSize())
	assert.Equal(t, "page-2", req.GetPageToken())
	assert.Equal(t, 1, env.ServerlessDatasetServer.ListSampleDatasetLoadsCalls.Count())
}

func TestSpaceDatasetLoadList_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.ListSampleDatasetLoadsCalls.Returns(nil, fmt.Errorf("boom"))

	_, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "list", "space-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list sample dataset loads")
}

func TestSpaceDatasetLoadDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	failed := sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_FAILED)
	failed.Reason = "collection limit"
	failed.Stats = &spacedatasetv1.SampleDatasetLoadStats{Duration: durationpb.New(12_000_000_000), Progress: new("10/75 MiB (13%)")}
	env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Returns(&spacedatasetv1.GetSampleDatasetLoadResponse{SampleDatasetLoad: failed}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "describe", "space-abc", "load-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID:          load-123")
	assert.Contains(t, stdout, "Space:       space-abc")
	assert.Contains(t, stdout, "Dataset:     midjourney-styles")
	assert.Contains(t, stdout, "Version:     v2")
	assert.Contains(t, stdout, "Collection:  styles")
	assert.Contains(t, stdout, "Status:      FAILED")
	assert.Contains(t, stdout, "Reason:      collection limit")
	assert.Contains(t, stdout, "Progress:    10/75 MiB (13%)")
	assert.Contains(t, stdout, "Duration:    12s")
	assert.Contains(t, stdout, "Created:")

	req, ok := env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Equal(t, "load-123", req.GetSampleDatasetLoadId())
}

func TestSpaceDatasetLoadDescribe_RunningWithoutReason(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Returns(&spacedatasetv1.GetSampleDatasetLoadResponse{
		SampleDatasetLoad: sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_RUNNING),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "describe", "space-abc", "load-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Status:      RUNNING")
	assert.NotContains(t, stdout, "Reason:")
	assert.NotContains(t, stdout, "Duration:")
}

func TestSpaceDatasetLoadDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Returns(&spacedatasetv1.GetSampleDatasetLoadResponse{
		SampleDatasetLoad: sampleDatasetLoad(spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_SUCCEEDED),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "describe", "space-abc", "load-123", "--json")
	require.NoError(t, err)

	var result struct {
		ID             string `json:"id"`
		CollectionName string `json:"collectionName"`
		Status         string `json:"status"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "load-123", result.ID)
	assert.Equal(t, "styles", result.CollectionName)
	assert.Equal(t, "SAMPLE_DATASET_LOAD_STATUS_SUCCEEDED", result.Status)
}

func TestSpaceDatasetLoadDescribe_Errors(t *testing.T) {
	t.Run("missing load ID", func(t *testing.T) {
		env := testutil.NewTestEnv(t)
		_, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "describe", "space-abc")
		require.Error(t, err)
		assert.Equal(t, 0, env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Count())
	})

	t.Run("backend error", func(t *testing.T) {
		env := testutil.NewTestEnv(t)
		env.ServerlessDatasetServer.GetSampleDatasetLoadCalls.Returns(nil, fmt.Errorf("not found"))
		_, _, err := testutil.Exec(t, env, "serverless", "space", "dataset", "load", "describe", "space-abc", "load-123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get sample dataset load")
	})
}
