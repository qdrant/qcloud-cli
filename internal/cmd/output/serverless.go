package output

import (
	"strings"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"
	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"
	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"
	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"
)

// SpacePhase returns a concise label for a serverless SpaceStatePhase.
func SpacePhase(p spacev1.SpaceStatePhase) string {
	return strings.TrimPrefix(p.String(), "SPACE_STATE_PHASE_")
}

// SpaceApiKeyPhase returns a concise label for a serverless SpaceApiKeyStatePhase.
func SpaceApiKeyPhase(p spaceauthv1.SpaceApiKeyStatePhase) string {
	return strings.TrimPrefix(p.String(), "SPACE_API_KEY_STATE_PHASE_")
}

// SpaceGlobalAccessType returns a concise label for a serverless GlobalAccessRuleAccessType.
func SpaceGlobalAccessType(t spaceauthv1.GlobalAccessRuleAccessType) string {
	return strings.TrimPrefix(t.String(), "GLOBAL_ACCESS_RULE_ACCESS_TYPE_")
}

// SpaceCollectionAccessType returns a concise label for a serverless CollectionAccessRuleAccessType.
func SpaceCollectionAccessType(t spaceauthv1.CollectionAccessRuleAccessType) string {
	return strings.TrimPrefix(t.String(), "COLLECTION_ACCESS_RULE_ACCESS_TYPE_")
}

// SpaceBackupStatus returns a concise label for a serverless BackupStatus.
func SpaceBackupStatus(s spacebackupv1.BackupStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_STATUS_")
}

// SpaceBackupScheduleStatus returns a concise label for a serverless BackupScheduleStatus.
func SpaceBackupScheduleStatus(s spacebackupv1.BackupScheduleStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_SCHEDULE_STATUS_")
}

// SpaceBackupRestoreStatus returns a concise label for a serverless BackupRestoreStatus.
func SpaceBackupRestoreStatus(s spacebackupv1.BackupRestoreStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_RESTORE_STATUS_")
}

// SpaceAlertType returns a concise label for a serverless SpaceAlertType.
func SpaceAlertType(t serverlessmonitoringv1.SpaceAlertType) string {
	return strings.TrimPrefix(t.String(), "SPACE_ALERT_TYPE_")
}

// SpaceAlertSeverity returns a concise label for a serverless SpaceAlertSeverity.
func SpaceAlertSeverity(s serverlessmonitoringv1.SpaceAlertSeverity) string {
	return strings.TrimPrefix(s.String(), "SPACE_ALERT_SEVERITY_")
}

// SpaceAlertState returns a concise label for a serverless SpaceAlertState.
func SpaceAlertState(s serverlessmonitoringv1.SpaceAlertState) string {
	return strings.TrimPrefix(s.String(), "SPACE_ALERT_STATE_")
}
