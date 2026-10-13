package space

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newAlertListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*serverlessmonitoringv1.ListSpaceAlertsResponse]{
		Use:   "list <space-id>",
		Short: "List alerts of a space",
		Long: `List the alerts of a serverless space, most recently firing first.

Alerts that are tied to the space as a whole rather than to a single collection
are shown with "(space)" in the COLLECTION column. Use --space-only to list only
those alerts, or --collection / --collection-contains to list only the alerts of
matching collections.

By default, all alerts are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.`,
		Example: `# List all alerts of a space
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# List only the alerts that are currently firing
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --state firing

# List the alerts of a single collection
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --collection products

# List alerts that concern the space as a whole
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --space-only

# Show the alert descriptions as JSON
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json`,
		Args: util.ExactArgs(1, "a space ID"),
		Fetch: func(s *state.State, cmd *cobra.Command) (*serverlessmonitoringv1.ListSpaceAlertsResponse, error) {
			var alertState *serverlessmonitoringv1.SpaceAlertState
			if cmd.Flags().Changed("state") {
				v, _ := cmd.Flags().GetString("state")
				st, err := parseAlertState(v)
				if err != nil {
					return nil, err
				}

				alertState = &st
			}

			filter, err := readCollectionFilter(cmd)
			if err != nil {
				return nil, err
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			spaceID := cmd.Flags().Arg(0)
			items, next, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*serverlessmonitoringv1.SpaceAlert, string, error) {
				req := &serverlessmonitoringv1.ListSpaceAlertsRequest{
					AccountId: accountID,
					SpaceId:   spaceID,
					State:     alertState,
					PageSize:  pageSize,
					PageToken: pageToken,
				}
				switch {
				case filter.name != nil:
					req.CollectionFilter = &serverlessmonitoringv1.ListSpaceAlertsRequest_CollectionName{CollectionName: *filter.name}
				case filter.contains != nil:
					req.CollectionFilter = &serverlessmonitoringv1.ListSpaceAlertsRequest_CollectionNameContains{CollectionNameContains: *filter.contains}
				case filter.spaceOnly:
					req.CollectionFilter = &serverlessmonitoringv1.ListSpaceAlertsRequest_SpaceGlobalOnly{SpaceGlobalOnly: true}
				}

				resp, err := client.ServerlessMonitoring().ListSpaceAlerts(ctx, req)
				if err != nil {
					return nil, "", fmt.Errorf("failed to list space alerts: %w", err)
				}

				return resp.GetItems(), resp.GetNextPageToken(), nil
			})
			if err != nil {
				return nil, err
			}

			return &serverlessmonitoringv1.ListSpaceAlertsResponse{Items: items, NextPageToken: next}, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *serverlessmonitoringv1.ListSpaceAlertsResponse) (output.TableRenderer, error) {
			t := output.NewTable[*serverlessmonitoringv1.SpaceAlert](w)
			t.AddField("ID", func(v *serverlessmonitoringv1.SpaceAlert) string {
				return v.GetId()
			})
			t.AddField("SEVERITY", func(v *serverlessmonitoringv1.SpaceAlert) string {
				return output.SpaceAlertSeverity(v.GetSeverity())
			})
			t.AddField("TYPE", func(v *serverlessmonitoringv1.SpaceAlert) string {
				return output.SpaceAlertType(v.GetType())
			})
			t.AddField("STATE", func(v *serverlessmonitoringv1.SpaceAlert) string {
				return output.SpaceAlertState(v.GetState())
			})
			t.AddField("COLLECTION", func(v *serverlessmonitoringv1.SpaceAlert) string {
				return output.OptionalValue(v.CollectionName, "(space)")
			})
			t.AddField("TITLE", func(v *serverlessmonitoringv1.SpaceAlert) string {
				return v.GetTitle()
			})
			t.AddField("LAST FIRING", func(v *serverlessmonitoringv1.SpaceAlert) string {
				if v.GetLastFiringAt() != nil {
					return output.HumanTime(v.GetLastFiringAt().AsTime())
				}

				return "never"
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)

	cmd.Flags().String("state", "", "Only list alerts in this state (firing, resolved)")
	_ = cmd.RegisterFlagCompletionFunc("state", cobra.FixedCompletions([]string{"firing", "resolved"}, cobra.ShellCompDirectiveNoFileComp))
	addCollectionFilterFlags(cmd, true)
	util.AddPaginationFlags(cmd, "alerts")

	return cmd
}

// parseAlertState maps the --state flag value to the proto enum.
func parseAlertState(s string) (serverlessmonitoringv1.SpaceAlertState, error) {
	switch s {
	case "firing":
		return serverlessmonitoringv1.SpaceAlertState_SPACE_ALERT_STATE_FIRING, nil
	case "resolved":
		return serverlessmonitoringv1.SpaceAlertState_SPACE_ALERT_STATE_RESOLVED, nil
	default:
		return 0, fmt.Errorf("invalid --state %q: must be one of firing, resolved", s)
	}
}
