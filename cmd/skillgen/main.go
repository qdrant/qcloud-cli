package main

import (
	"log"
	"os"

	"github.com/qdrant/qcloud-cli/internal/agentskill"
	"github.com/qdrant/qcloud-cli/internal/cli"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func main() {
	outDir := "./skills"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	s := state.New("dev")
	root := cli.NewRootCommand(s)

	files, err := agentskill.Generate(root, s.Version)
	if err != nil {
		log.Fatalf("generate skill: %v", err)
	}

	if _, err := agentskill.Write(outDir, files); err != nil {
		log.Fatalf("write skill: %v", err)
	}
}
