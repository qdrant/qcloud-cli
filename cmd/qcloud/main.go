package main

import (
	"context"
	"fmt"
	"os"

	"github.com/qdrant/qcloud-cli/internal/cli"
	"github.com/qdrant/qcloud-cli/internal/state"
)

var version = "dev"

func main() {
	ctx := context.Background()
	s := state.New(version)
	cmd := cli.NewRootCommand(s)
	cmd.SetContext(ctx)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
