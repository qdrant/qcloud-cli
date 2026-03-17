package selfupgrade

import (
	"context"
	"fmt"

	goselfupdate "github.com/creativeprojects/go-selfupdate"
)

// ReleaseInfo contains version information about a release.
type ReleaseInfo struct {
	Version string
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

	return &ReleaseInfo{Version: rel.Version()}, true, nil
}

// DetectVersion finds a specific release by version string.
func (g *GithubUpdater) DetectVersion(ctx context.Context, version string) (*ReleaseInfo, bool, error) {
	rel, found, err := g.updater.DetectVersion(ctx, repository, version)
	if err != nil || !found {
		return nil, found, err
	}

	return &ReleaseInfo{Version: rel.Version()}, true, nil
}

// UpdateTo downloads and replaces the binary at execPath with the given version.
func (g *GithubUpdater) UpdateTo(ctx context.Context, version string, execPath string) error {
	rel, found, err := g.updater.DetectVersion(ctx, repository, version)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("version %s not found", version)
	}

	return g.updater.UpdateTo(ctx, rel, execPath)
}
