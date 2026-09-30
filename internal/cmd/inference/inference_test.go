package inference_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func models() []*bookingv1.InferenceModel {
	return []*bookingv1.InferenceModel{
		{
			Id:                  "3b0f4a4e-6b4a-4f2a-9f5a-1f2f3a4b5c6d",
			Name:                "mixedbread-ai/mxbai-embed-large-v1",
			Title:               "mxbai-embed-large-v1",
			Description:         "General purpose English embedding model",
			VectorType:          bookingv1.VectorType_VECTOR_TYPE_DENSE,
			Modality:            bookingv1.ModelModality_MODEL_MODALITY_TEXT,
			Vendor:              "MixedBread",
			UnitIntPrice:        5000,
			IsExternal:          false,
			Dimensionality:      new(uint32(1024)),
			MaxTokensPerRequest: new(uint32(512)),
		},
		{
			Id:           "9a1c2d3e-4f5a-6b7c-8d9e-0f1a2b3c4d5e",
			Name:         "cohere/embed-v4.0",
			Title:        "Cohere Embed v4.0",
			Description:  "Multimodal embedding model from Cohere",
			VectorType:   bookingv1.VectorType_VECTOR_TYPE_MULTI,
			Modality:     bookingv1.ModelModality_MODEL_MODALITY_IMAGE,
			Vendor:       "Cohere",
			UnitIntPrice: 0,
			IsExternal:   true,
		},
	}
}

func TestListInferenceModels_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListGlobalInferenceModelsCalls.Returns(&bookingv1.ListGlobalInferenceModelsResponse{
		Items: models(),
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"inference", "models", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
	)
	require.NoError(t, err)

	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "VECTOR-TYPE")
	assert.Contains(t, stdout, "MODALITY")
	assert.Contains(t, stdout, "DIMENSIONS")
	assert.Contains(t, stdout, "MAX-TOKENS")
	assert.Contains(t, stdout, "EXTERNAL")
	assert.Contains(t, stdout, "PRICE/1M-TOKENS")

	assert.Contains(t, stdout, "mixedbread-ai/mxbai-embed-large-v1")
	assert.Contains(t, stdout, "DENSE")
	assert.Contains(t, stdout, "TEXT")
	assert.Contains(t, stdout, "1024")
	assert.Contains(t, stdout, "512")
	assert.Contains(t, stdout, "no")
	assert.Contains(t, stdout, "0.0500 USD")

	assert.Contains(t, stdout, "cohere/embed-v4.0")
	assert.Contains(t, stdout, "MULTI")
	assert.Contains(t, stdout, "IMAGE")
	// Unset optional dimensionality / token limit render as n/a, free models as "free".
	assert.Contains(t, stdout, "n/a")
	assert.Contains(t, stdout, "free")
	assert.Contains(t, stdout, "yes")
}

func TestListInferenceModels_SortedByName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	// Returned in non-alphabetical order; the command sorts them by name.
	env.BookingServer.ListGlobalInferenceModelsCalls.Returns(&bookingv1.ListGlobalInferenceModelsResponse{
		Items: []*bookingv1.InferenceModel{
			{Name: "qdrant/bm25"},
			{Name: "cohere/embed-v4.0"},
			{Name: "mixedbread-ai/mxbai-embed-large-v1"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"inference", "models", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
		"--no-headers",
	)
	require.NoError(t, err)

	names := make([]string, 0, 3)
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		names = append(names, strings.Fields(line)[0])
	}

	assert.Equal(t, []string{
		"cohere/embed-v4.0",
		"mixedbread-ai/mxbai-embed-large-v1",
		"qdrant/bm25",
	}, names)
}

func TestListInferenceModels_SendsRequestFields(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListGlobalInferenceModelsCalls.Returns(&bookingv1.ListGlobalInferenceModelsResponse{}, nil)

	_, _, err := testutil.Exec(t, env,
		"inference", "models", "list",
		"--cloud-provider", "gcp",
		"--cloud-region", "us-east4",
	)
	require.NoError(t, err)

	req, ok := env.BookingServer.ListGlobalInferenceModelsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "gcp", req.GetCloudProviderId())
	assert.Equal(t, "us-east4", req.GetCloudProviderRegionId())
}

func TestListInferenceModels_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListGlobalInferenceModelsCalls.Returns(&bookingv1.ListGlobalInferenceModelsResponse{
		Items: models(),
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"inference", "models", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
		"--json",
	)
	require.NoError(t, err)

	var resp struct {
		Items []struct {
			Name           string `json:"name"`
			Vendor         string `json:"vendor"`
			VectorType     string `json:"vectorType"`
			Dimensionality uint32 `json:"dimensionality"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &resp))
	require.Len(t, resp.Items, 2)
	// JSON output is sorted by name too, so cohere comes before mixedbread-ai.
	assert.Equal(t, "cohere/embed-v4.0", resp.Items[0].Name)
	assert.Equal(t, "Cohere", resp.Items[0].Vendor)
	assert.Equal(t, "VECTOR_TYPE_MULTI", resp.Items[0].VectorType)
	assert.Equal(t, "mixedbread-ai/mxbai-embed-large-v1", resp.Items[1].Name)
	assert.Equal(t, "VECTOR_TYPE_DENSE", resp.Items[1].VectorType)
	assert.Equal(t, uint32(1024), resp.Items[1].Dimensionality)
}

func TestListInferenceModels_NoHeaders(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListGlobalInferenceModelsCalls.Returns(&bookingv1.ListGlobalInferenceModelsResponse{
		Items: models(),
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"inference", "models", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
		"--no-headers",
	)
	require.NoError(t, err)
	assert.NotContains(t, stdout, "VECTOR-TYPE")
	assert.Contains(t, stdout, "cohere/embed-v4.0")
}

func TestListInferenceModels_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListGlobalInferenceModelsCalls.Returns(nil, fmt.Errorf("boom"))

	_, _, err := testutil.Exec(t, env,
		"inference", "models", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list inference models")
}

func TestListInferenceModels_RequiresCloudProviderAndRegion(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "inference", "models", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cloud-provider")

	_, _, err = testutil.Exec(t, env, "inference", "models", "list", "--cloud-provider", "aws")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cloud-region")
}
