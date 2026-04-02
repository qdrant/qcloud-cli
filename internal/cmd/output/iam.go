package output

import (
	"strings"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"
	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"
)

// UserStatus formats an iamv1.UserStatus enum for display.
func UserStatus(x iamv1.UserStatus) string {
	return strings.TrimPrefix(x.String(), "USER_STATUS_")
}

// AccountInviteStatus formats an accountv1.AccountInviteStatus enum for display.
func AccountInviteStatus(x accountv1.AccountInviteStatus) string {
	return strings.TrimPrefix(x.String(), "ACCOUNT_INVITE_STATUS_")
}
