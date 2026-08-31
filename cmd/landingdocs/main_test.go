package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files in testdata/golden")

// TestRun_GeneratesGoldenPages runs the converter end-to-end against a
// small fixture command tree (testdata/src) and checks every generated
// page byte-for-byte against its golden file (testdata/golden).
func TestRun_GeneratesGoldenPages(t *testing.T) {
	srcDir := filepath.Join("testdata", "src")
	goldenDir := filepath.Join("testdata", "golden")
	destDir := t.TempDir()

	if err := run(srcDir, destDir); err != nil {
		t.Fatalf("run() returned error: %v", err)
	}

	wantFiles := []string{"_index.md", "qcloud_cluster.md", "qcloud_cluster_create.md"}

	got, err := filepath.Glob(filepath.Join(destDir, "*.md"))
	if err != nil {
		t.Fatalf("glob generated files: %v", err)
	}

	if len(got) != len(wantFiles) {
		t.Fatalf("generated %d files, want %d: %v", len(got), len(wantFiles), got)
	}

	for _, name := range wantFiles {
		gotPath := filepath.Join(destDir, name)
		goldenPath := filepath.Join(goldenDir, name)

		gotContent, err := os.ReadFile(gotPath)
		if err != nil {
			t.Fatalf("read generated file %s: %v", gotPath, err)
		}

		if *update {
			if err := os.WriteFile(goldenPath, gotContent, 0o644); err != nil {
				t.Fatalf("update golden file %s: %v", goldenPath, err)
			}

			continue
		}

		wantContent, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("read golden file %s: %v", goldenPath, err)
		}

		if string(gotContent) != string(wantContent) {
			t.Errorf("generated %s does not match golden file %s\n--- got ---\n%s\n--- want ---\n%s",
				name, goldenPath, gotContent, wantContent)
		}
	}
}

// TestRun_EmptySourceDir checks that run() fails loudly rather than
// silently producing an empty reference section, which would otherwise
// go unnoticed until the landing page shipped with a missing command
// tree.
func TestRun_EmptySourceDir(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	err := run(srcDir, destDir)
	if err == nil {
		t.Fatal("run() with an empty source dir returned nil error, want an error")
	}
}
