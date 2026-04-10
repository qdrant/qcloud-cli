package output

import (
	"strings"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"
)

// UserStatus formats an iamv1.UserStatus enum for display.
func UserStatus(x iamv1.UserStatus) string {
	return strings.TrimPrefix(x.String(), "USER_STATUS_")
}

// RoleType formats a RoleType enum for display.
func RoleType(v iamv1.RoleType) string {
	return strings.TrimPrefix(v.String(), "ROLE_TYPE_")
}
