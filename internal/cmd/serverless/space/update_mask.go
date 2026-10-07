package space

import (
	"slices"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// flagPath maps a command flag to the field-mask path of the resource field it
// changes. Several flags may map to the same path.
type flagPath struct {
	flag string
	path string
}

// spaceUpdatePaths lists the "space update" flags and the Space fields they change.
var spaceUpdatePaths = []flagPath{
	{"name", "name"},
	{"label", "labels"},
	{"cost-allocation-label", "cost_allocation_label"},
	{"allowed-ip", "configuration.allowed_ip_source_ranges"},
	{"allowed-origin", "configuration.allowed_origins"},
	{"max-collection-size", "configuration.collection_settings.max_size"},
	{"searcher-idle-timeout", "configuration.searcher_settings.idle_timeout"},
	{"searcher-max-workers", "configuration.searcher_settings.max_workers"},
}

// backupScheduleUpdatePaths lists the "backup schedule update" flags and the
// BackupSchedule fields they change.
var backupScheduleUpdatePaths = []flagPath{
	{"name", "name"},
	{"schedule", "schedule"},
	{"retention-days", "retention_period"},
	{"pause", "paused_at"},
	{"resume", "paused_at"},
}

// updateMask builds a field mask from the flags that were set on cmd, so that
// the server only touches the fields the user asked to change. It returns nil
// when no mapped flag was set.
func updateMask(cmd *cobra.Command, mapping []flagPath) *fieldmaskpb.FieldMask {
	var paths []string
	for _, m := range mapping {
		if cmd.Flags().Changed(m.flag) && !slices.Contains(paths, m.path) {
			paths = append(paths, m.path)
		}
	}

	if len(paths) == 0 {
		return nil
	}

	return &fieldmaskpb.FieldMask{Paths: paths}
}
