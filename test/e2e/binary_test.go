package e2e_test

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

const downloadURL = "https://github.com/qdrant/qcloud-cli/releases/latest/download"

// setupBinary returns the path to a qcloud binary. If QCLOUD_E2E_BINARY is
// set it uses that path directly; otherwise it downloads the latest release
// from GitHub.
func setupBinary(t *testing.T) string {
	t.Helper()

	if p := os.Getenv("QCLOUD_E2E_BINARY"); p != "" {
		t.Logf("using binary from QCLOUD_E2E_BINARY: %s", p)
		return p
	}

	archiveName := fmt.Sprintf("qcloud-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	url := downloadURL + "/" + archiveName

	t.Logf("downloading %s", url)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusOK, resp.StatusCode, "GET %s returned %s", url, resp.Status)

	dir := t.TempDir()
	binaryPath := extractQcloud(t, resp.Body, dir)

	require.NoError(t, os.Chmod(binaryPath, 0o755))
	t.Logf("binary at %s", binaryPath)

	return binaryPath
}

// extractQcloud reads a gzip-compressed tar archive from r and extracts the
// "qcloud" binary into dir. It returns the path to the extracted binary.
func extractQcloud(t *testing.T, r io.Reader, dir string) string {
	t.Helper()

	gr, err := gzip.NewReader(r)
	require.NoError(t, err)
	defer func() { require.NoError(t, gr.Close()) }()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)

		if filepath.Base(hdr.Name) != "qcloud" {
			continue
		}

		dst := filepath.Join(dir, "qcloud")
		f, err := os.Create(dst)
		require.NoError(t, err)

		_, err = io.Copy(f, tr)
		require.NoError(t, f.Close())
		require.NoError(t, err)

		return dst
	}

	t.Fatal("qcloud binary not found in archive")
	return ""
}
