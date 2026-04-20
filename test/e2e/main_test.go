package e2e_test

import (
	"log"
	"os"
	"testing"

	"github.com/qdrant/qcloud-cli/test/e2e/framework"
)

// TestMain pre-resolves the qcloud binary once before any test runs. Without
// this, every test that calls framework.NewEnv would otherwise trigger a
// fresh download attempt on its first invocation.
//
// When QCLOUD_E2E is unset we skip setup entirely so `go test ./...` stays
// free of network calls.
func TestMain(m *testing.M) {
	if os.Getenv("QCLOUD_E2E") != "" {
		if _, err := framework.Binary(); err != nil {
			log.Fatalf("e2e: resolving qcloud binary: %v", err)
		}
	}
	os.Exit(m.Run())
}
