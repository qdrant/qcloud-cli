package util

import (
	"github.com/spf13/cobra"
)

// AddPaginationFlags registers the --page-size and --page-token flags used by
// list commands. what names the listed resources in the flag help.
func AddPaginationFlags(cmd *cobra.Command, what string) {
	cmd.Flags().Int32("page-size", 0, "Maximum number of "+what+" to return per page (manual pagination mode)")
	cmd.Flags().String("page-token", "", "Page token from a previous response to resume from (manual pagination mode)")
}

// FetchPages runs a paginated list call. Unless --page-size or --page-token is
// set, it follows next-page tokens and returns all items with a nil next token.
// Otherwise it performs a single request with the given flags and returns the
// next page token, if any.
func FetchPages[T any](
	cmd *cobra.Command,
	fetch func(pageSize *int32, pageToken *string) (items []T, nextPageToken string, err error),
) ([]T, *string, error) {
	pageSizeChanged := cmd.Flags().Changed("page-size")
	pageTokenChanged := cmd.Flags().Changed("page-token")

	if !pageSizeChanged && !pageTokenChanged {
		var all []T
		var token *string
		for {
			items, next, err := fetch(nil, token)
			if err != nil {
				return nil, nil, err
			}

			all = append(all, items...)
			if next == "" {
				return all, nil, nil
			}

			token = &next
		}
	}

	var pageSize *int32
	if pageSizeChanged {
		ps, _ := cmd.Flags().GetInt32("page-size")
		pageSize = &ps
	}

	var pageToken *string
	if pageTokenChanged {
		pt, _ := cmd.Flags().GetString("page-token")
		pageToken = &pt
	}

	items, next, err := fetch(pageSize, pageToken)
	if err != nil {
		return nil, nil, err
	}

	if next == "" {
		return items, nil, nil
	}

	return items, &next, nil
}
