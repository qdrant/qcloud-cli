package main

import (
	"log"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/qdrant/qcloud-cli/internal/cli"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func main() {
	outDir := "./docs/reference"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	s := state.New("dev")
	root := cli.NewRootCommand(s)
	root.DisableAutoGenTag = true

	if err := doc.GenMarkdownTree(root, outDir); err != nil {
		log.Fatalf("generate docs: %v", err)
	}
}
