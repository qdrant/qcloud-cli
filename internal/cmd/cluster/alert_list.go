package cluster

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// alertDescriptionWidth is the width at which alert descriptions, which can be
// up to 300 characters long, wrap onto the next line of the table.
const alertDescriptionWidth = 60

func newAlertListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*monitoringv1.ListClusterAlertsResponse]{
		Use:   "list <cluster-id>",
		Short: "List alerts of a cluster",
		Long: `List the alerts of a cluster, most recently firing first.

By default, only the alerts that are currently firing are listed. Use
--show-resolved to also list the alerts that have been resolved.

Long descriptions wrap onto multiple lines; use
--json to get each alert on a single record.`,
		Example: `# List the alerts of a cluster that are currently firing
qcloud cluster alert list 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# List both firing and resolved alerts
qcloud cluster alert list 7b2ea926-724b-4de2-b73a-8675c42a6ebe --show-resolved

# Output the alerts as JSON
qcloud cluster alert list 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json`,
		Args: util.ExactArgs(1, "a cluster ID"),
		Fetch: func(s *state.State, cmd *cobra.Command) (*monitoringv1.ListClusterAlertsResponse, error) {
			// Leaving the state unset makes the API return alerts in all states.
			alertState := monitoringv1.ClusterAlertState_CLUSTER_ALERT_STATE_FIRING.Enum()
			if showResolved, _ := cmd.Flags().GetBool("show-resolved"); showResolved {
				alertState = nil
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

			resp, err := client.Monitoring().ListClusterAlerts(ctx, &monitoringv1.ListClusterAlertsRequest{
				AccountId: accountID,
				ClusterId: cmd.Flags().Arg(0),
				State:     alertState,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list cluster alerts: %w", err)
			}

			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *monitoringv1.ListClusterAlertsResponse) (output.TableRenderer, error) {
			t := output.NewTable[*monitoringv1.ClusterAlert](w)
			t.AddField("SEVERITY", func(v *monitoringv1.ClusterAlert) string {
				return output.ClusterAlertSeverity(v.GetSeverity())
			})
			t.AddField("TYPE", func(v *monitoringv1.ClusterAlert) string {
				return output.ClusterAlertType(v.GetType())
			})
			t.AddField("STATE", func(v *monitoringv1.ClusterAlert) string {
				return output.ClusterAlertState(v.GetState())
			})
			t.AddField("TITLE", func(v *monitoringv1.ClusterAlert) string {
				return v.GetTitle()
			})
			t.AddField("LAST FIRING", func(v *monitoringv1.ClusterAlert) string {
				if v.GetLastFiringAt() != nil {
					return output.HumanTime(v.GetLastFiringAt().AsTime())
				}

				return "never"
			})
			t.AddWrappedField("DESCRIPTION", alertDescriptionWidth, func(v *monitoringv1.ClusterAlert) string {
				return v.GetDescription()
			})
			t.SetItems(resp.GetAlerts())
			return t, nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)

	cmd.Flags().Bool("show-resolved", false, "Also list alerts that have been resolved")

	return cmd
}
