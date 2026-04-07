package output

import (
	"strings"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"
)

// UserStatus formats an iamv1.UserStatus enum for display.
func UserStatus(x iamv1.UserStatus) string {
	return strings.TrimPrefix(x.String(), "USER_STATUS_")
}
