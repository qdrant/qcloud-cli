package output

import (
	"strings"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"
)

// HybridEnvironmentPhase returns a concise label for a HybridCloudEnvironmentStatusPhase.
func HybridEnvironmentPhase(p hybridv1.HybridCloudEnvironmentStatusPhase) string {
	return strings.TrimPrefix(p.String(), "HYBRID_CLOUD_ENVIRONMENT_STATUS_PHASE_")
}

// ClusterCreationStatus returns a concise label for a QdrantClusterCreationStatus.
func ClusterCreationStatus(s hybridv1.QdrantClusterCreationStatus) string {
	return strings.TrimPrefix(s.String(), "QDRANT_CLUSTER_CREATION_STATUS_")
}

// HybridComponentPhase returns a concise label for a HybridCloudEnvironmentComponentStatusPhase.
func HybridComponentPhase(p hybridv1.HybridCloudEnvironmentComponentStatusPhase) string {
	return strings.TrimPrefix(p.String(), "HYBRID_CLOUD_ENVIRONMENT_COMPONENT_STATUS_PHASE_")
}
