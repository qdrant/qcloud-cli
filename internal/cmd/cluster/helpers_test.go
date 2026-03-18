package cluster

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGpuFlagToMillicores(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "integer 1", input: "1", want: "1000m"},
		{name: "integer 2", input: "2", want: "2000m"},
		{name: "millicore passthrough", input: "1000m", want: "1000m"},
		{name: "empty", input: "", want: ""},
		{name: "invalid", input: "bad", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("gpu", "", "")
			_ = cmd.Flags().Set("gpu", tt.input)
			got, err := gpuFlagToMillicores(cmd)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
