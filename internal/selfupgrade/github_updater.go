package selfupgrade

import (
	"context"
	"fmt"

	goselfupdate "github.com/creativeprojects/go-selfupdate"
)

// ReleaseInfo wraps a go-selfupdate Release.
// version is used as a fallback when Release is nil (e.g. in test stubs).
type ReleaseInfo struct {
	Release    *goselfupdate.Release
	version string
}

// NewReleaseInfo creates a ReleaseInfo with only a version string, for use in tests.
func NewReleaseInfo(version string) *ReleaseInfo {
	return &ReleaseInfo{version: version}
}

// NewReleaseInfo creates a ReleaseInfo with only a version string, for use in tests.
func NewReleaseInfoFromSelfUpdate(rel *goselfupdate.Release) *ReleaseInfo {
	return &ReleaseInfo{version: rel.Version(), Release: rel}
}

// Version returns the release version string.
func (r *ReleaseInfo) Version() string {
	if r.Release != nil {
		return r.Release.Version()
	}
	return r.version
}

// Equal reports whether this release's version equals the given version string.
func (r *ReleaseInfo) Equal(version string) bool {
	if r.Release != nil {
		return r.Release.Equal(version)
	}
	return r.version == version
}

var repository = goselfupdate.NewRepositorySlug("qdrant", "qcloud-cli")

// GithubUpdater checks for and applies CLI updates from GitHub releases.
type GithubUpdater struct {
	updater *goselfupdate.Updater
}

// NewGitHubUpdater creates a GithubUpdater.
func NewGitHubUpdater() (*GithubUpdater, error) {
	source, err := goselfupdate.NewGitHubSource(goselfupdate.GitHubConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub source: %w", err)
	}

	updater, err := goselfupdate.NewUpdater(goselfupdate.Config{Source: source})
	if err != nil {
		return nil, fmt.Errorf("failed to create updater: %w", err)
	}

	return &GithubUpdater{updater: updater}, nil
}

// DetectLatest finds the latest release available.
func (g *GithubUpdater) DetectLatest(ctx context.Context) (*ReleaseInfo, bool, error) {
	rel, found, err := g.updater.DetectLatest(ctx, repository)
	if err != nil || !found {
		return nil, found, err
	}
	return NewReleaseInfoFromSelfUpdate(rel), true, nil
}

// UpdateSelf downloads and replaces the running binary with the latest release.
func (g *GithubUpdater) UpdateSelf(ctx context.Context, currentVersion string) (*ReleaseInfo, error) {
	rel, err := g.updater.UpdateSelf(ctx, currentVersion, repository)
	if err != nil {
		return nil, err
	}
	return NewReleaseInfoFromSelfUpdate(rel), nil
}
