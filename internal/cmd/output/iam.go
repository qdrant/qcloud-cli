package output

import (
	"strings"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"
)

// RoleType formats a RoleType enum for display.
func RoleType(v iamv1.RoleType) string {
	return strings.TrimPrefix(v.String(), "ROLE_TYPE_")
}
