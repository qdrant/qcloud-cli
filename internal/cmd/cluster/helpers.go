package cluster

import (
	"strings"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
)

func phaseString(phase clusterv1.ClusterPhase) string {
	return strings.TrimPrefix(phase.String(), "CLUSTER_PHASE_")
}
