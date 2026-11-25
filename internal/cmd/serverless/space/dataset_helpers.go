package space

import (
	"context"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"time"

	spacedatasetv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/dataset/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

// sampleDatasetVectorLabel renders a vector name; the empty name is the
// default (unnamed) vector.
func sampleDatasetVectorLabel(name string) string {
	if name == "" {
		return "default"
	}

	return name
}

// sampleDatasetVectorNames renders the vector names of a sample dataset
// collection, dense vectors first, e.g. "default, bm25 (sparse)".
func sampleDatasetVectorNames(cfg *spacedatasetv1.CollectionConfig) string {
	parts := make([]string, 0, len(cfg.GetDenseVectors())+len(cfg.GetSparseVectors()))
	for _, name := range sortedKeys(cfg.GetDenseVectors()) {
		parts = append(parts, sampleDatasetVectorLabel(name))
	}

	for _, name := range sortedKeys(cfg.GetSparseVectors()) {
		parts = append(parts, sampleDatasetVectorLabel(name)+" (sparse)")
	}

	return strings.Join(parts, ", ")
}

func sortedKeys[V any](m map[string]V) []string {
	return slices.Sorted(maps.Keys(m))
}

// printSampleDatasetLoad prints the details of a sample dataset load.
func printSampleDatasetLoad(w io.Writer, l *spacedatasetv1.SampleDatasetLoad) {
	fmt.Fprintf(w, "ID:          %s\n", l.GetId())
	fmt.Fprintf(w, "Space:       %s\n", l.GetSpaceId())
	fmt.Fprintf(w, "Dataset:     %s\n", l.GetSampleDatasetId())
	fmt.Fprintf(w, "Version:     %s\n", l.GetSampleDatasetVersion())
	fmt.Fprintf(w, "Collection:  %s\n", l.GetCollectionName())
	fmt.Fprintf(w, "Status:      %s\n", output.SampleDatasetLoadStatus(l.GetStatus()))
	if l.GetReason() != "" {
		fmt.Fprintf(w, "Reason:      %s\n", l.GetReason())
	}

	if st := l.GetStats(); st != nil {
		if st.Progress != nil {
			fmt.Fprintf(w, "Progress:    %s\n", st.GetProgress())
		}

		if st.Duration != nil {
			fmt.Fprintf(w, "Duration:    %s\n", formatOptionalDuration(st.GetDuration()))
		}
	}

	if l.GetCreatedAt() != nil {
		t := l.GetCreatedAt().AsTime()
		fmt.Fprintf(w, "Created:     %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
	}
}

// sampleDatasetLoadFailureStatuses are the final statuses of a load that did
// not succeed.
var sampleDatasetLoadFailureStatuses = map[spacedatasetv1.SampleDatasetLoadStatus]bool{
	spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_FAILED:         true,
	spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_FAILED_TO_SYNC: true,
	spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_NOT_FOUND:      true,
	spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_SKIPPED:        true,
}

// waitForSampleDatasetLoad polls the load until it succeeds, times out, or ends
// in a failure status. Progress lines are written to out. On failure the last
// state of the load is returned together with the error.
func waitForSampleDatasetLoad(
	ctx context.Context,
	svc spacedatasetv1.SampleDatasetServiceClient,
	out io.Writer,
	accountID, spaceID, loadID string,
	timeout, pollInterval time.Duration,
) (*spacedatasetv1.SampleDatasetLoad, error) {
	start := time.Now()
	var last *spacedatasetv1.SampleDatasetLoad
	done, err := util.PollUntilDone(ctx, timeout, pollInterval, "sample dataset load to finish",
		func(ctx context.Context) (*spacedatasetv1.SampleDatasetLoad, bool, error) {
			resp, err := svc.GetSampleDatasetLoad(ctx, &spacedatasetv1.GetSampleDatasetLoadRequest{
				AccountId:           accountID,
				SpaceId:             spaceID,
				SampleDatasetLoadId: loadID,
			})
			if err != nil {
				return nil, false, fmt.Errorf("failed to get sample dataset load status: %w", err)
			}

			load := resp.GetSampleDatasetLoad()
			last = load
			status := load.GetStatus()
			statusStr := output.SampleDatasetLoadStatus(status)
			if p := load.GetStats().GetProgress(); p != "" {
				fmt.Fprintf(out, "status=%s progress=%s (%s)\n", statusStr, p, time.Since(start).Round(time.Second))
			} else {
				fmt.Fprintf(out, "status=%s (%s)\n", statusStr, time.Since(start).Round(time.Second))
			}

			if status == spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_SUCCEEDED {
				return load, true, nil
			}

			if sampleDatasetLoadFailureStatuses[status] {
				return nil, false, fmt.Errorf("status=%s, reason=%s", statusStr, load.GetReason())
			}

			return nil, false, nil
		})
	if err != nil {
		return last, err
	}

	return done, nil
}
