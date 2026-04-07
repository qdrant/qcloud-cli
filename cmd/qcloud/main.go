package main

import (
	"context"
	"fmt"
	"os"

	"github.com/qdrant/qcloud-cli/internal/cli"
	"github.com/qdrant/qcloud-cli/internal/state"
)

var (
	version           = "0.19.0" // x-releaser-pleaser-version
	versionPrerelease = "dev"
)

func versionString() string {
	if versionPrerelease != "" {
		return version + "-" + versionPrerelease
	}
	return version
}

func main() {
	ctx := context.Background()
	s := state.New(versionString())
	cmd := cli.NewRootCommand(s)
	cmd.SetContext(ctx)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
