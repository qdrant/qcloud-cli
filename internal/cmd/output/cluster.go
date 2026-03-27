package output

import (
	"strings"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
)

// ClusterPhase returns a concise label for a ClusterPhase.
func ClusterPhase(p clusterv1.ClusterPhase) string {
	return strings.TrimPrefix(p.String(), "CLUSTER_PHASE_")
}

// ClusterNodeState returns a concise label for a ClusterNodeState.
func ClusterNodeState(s clusterv1.ClusterNodeState) string {
	return strings.TrimPrefix(s.String(), "CLUSTER_NODE_STATE_")
}

// TolerationOperator returns a concise label for a TolerationOperator.
func TolerationOperator(op clusterv1.TolerationOperator) string {
	return strings.TrimPrefix(op.String(), "TOLERATION_OPERATOR_")
}

// TolerationEffect returns a concise label for a TolerationEffect.
func TolerationEffect(eff clusterv1.TolerationEffect) string {
	return strings.TrimPrefix(eff.String(), "TOLERATION_EFFECT_")
}
