package hybrid

import (
	"fmt"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"
)

const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
)

func parseLogLevel(s string) (hybridv1.HybridCloudEnvironmentConfigurationLogLevel, error) {
	switch s {
	case logLevelDebug:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_DEBUG, nil
	case logLevelInfo:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_INFO, nil
	case logLevelWarn:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_WARN, nil
	case logLevelError:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_ERROR, nil
	default:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_UNSPECIFIED,
			fmt.Errorf("invalid log level %q: must be one of %s, %s, %s, %s", s, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError)
	}
}

func logLevelString(l hybridv1.HybridCloudEnvironmentConfigurationLogLevel) string {
	switch l {
	case hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_DEBUG:
		return logLevelDebug
	case hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_INFO:
		return logLevelInfo
	case hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_WARN:
		return logLevelWarn
	case hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_ERROR:
		return logLevelError
	default:
		return ""
	}
}
