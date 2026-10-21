package output

import (
	"strings"

	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"
)

// ClusterAlertType returns a concise label for a ClusterAlertType.
func ClusterAlertType(t monitoringv1.ClusterAlertType) string {
	return strings.TrimPrefix(t.String(), "CLUSTER_ALERT_TYPE_")
}

// ClusterAlertSeverity returns a concise label for a ClusterAlertSeverity.
func ClusterAlertSeverity(s monitoringv1.ClusterAlertSeverity) string {
	return strings.TrimPrefix(s.String(), "CLUSTER_ALERT_SEVERITY_")
}

// ClusterAlertState returns a concise label for a ClusterAlertState.
func ClusterAlertState(s monitoringv1.ClusterAlertState) string {
	return strings.TrimPrefix(s.String(), "CLUSTER_ALERT_STATE_")
}
