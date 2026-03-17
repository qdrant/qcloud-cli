package selfupgrade_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/selfupgrade"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

type mockUpdater struct {
	latestRelease  *selfupgrade.ReleaseInfo
	latestFound    bool
	latestErr      error
	versionRelease *selfupgrade.ReleaseInfo
	versionFound   bool
	versionErr     error
	updateErr      error
	updateCalled   bool
}

func (m *mockUpdater) DetectLatest(_ context.Context) (*selfupgrade.ReleaseInfo, bool, error) {
	return m.latestRelease, m.latestFound, m.latestErr
}

func (m *mockUpdater) DetectVersion(_ context.Context, _ string) (*selfupgrade.ReleaseInfo, bool, error) {
	return m.versionRelease, m.versionFound, m.versionErr
}

func (m *mockUpdater) UpdateTo(_ context.Context, _ string, _ string) error {
	m.updateCalled = true
	return m.updateErr
}

func TestSelfUpgrade_NoReleasesFound(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("0.4.0"))
	mock := &mockUpdater{latestFound: false}
	env.State.SetUpdater(mock)

	_, _, err := testutil.Exec(t, env, "self-upgrade")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no releases found")
}

func TestSelfUpgrade_AlreadyUpToDate_SameVersion(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("0.4.0"))
	mock := &mockUpdater{
		latestRelease: &selfupgrade.ReleaseInfo{Version: "0.4.0"},
		latestFound:   true,
	}
	env.State.SetUpdater(mock)

	stdout, _, err := testutil.Exec(t, env, "self-upgrade")
	require.NoError(t, err)
	assert.Contains(t, stdout, "already up to date")
}

func TestSelfUpgrade_CheckOnly(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("0.4.0"))
	mock := &mockUpdater{
		latestRelease: &selfupgrade.ReleaseInfo{Version: "0.5.0"},
		latestFound:   true,
	}
	env.State.SetUpdater(mock)

	stdout, _, err := testutil.Exec(t, env, "self-upgrade", "--check")
	require.NoError(t, err)
	assert.Contains(t, stdout, "New version available: v0.5.0")
	assert.Contains(t, stdout, "current: v0.4.0")
	assert.False(t, mock.updateCalled)
}

func TestSelfUpgrade_DevVersionRequiresForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("0.4.0-dev"))
	mock := &mockUpdater{
		latestRelease: &selfupgrade.ReleaseInfo{Version: "0.5.0"},
		latestFound:   true,
	}
	env.State.SetUpdater(mock)

	stdout, _, err := testutil.Exec(t, env, "self-upgrade")
	require.NoError(t, err)
	assert.Contains(t, stdout, "dev build")
	assert.Contains(t, stdout, "--force")
	assert.False(t, mock.updateCalled)
}

func TestSelfUpgrade_SpecificVersion_NotFound(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("0.4.0"))
	mock := &mockUpdater{versionFound: false}
	env.State.SetUpdater(mock)

	_, _, err := testutil.Exec(t, env, "self-upgrade", "--version", "99.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version 99.0.0 not found")
}

func TestSelfUpgrade_SpecificVersion_CheckOnly(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("0.4.0"))
	mock := &mockUpdater{
		versionRelease: &selfupgrade.ReleaseInfo{Version: "0.3.0"},
		versionFound:   true,
	}
	env.State.SetUpdater(mock)

	stdout, _, err := testutil.Exec(t, env, "self-upgrade", "--check", "--version", "0.3.0")
	require.NoError(t, err)
	assert.Contains(t, stdout, "New version available: v0.3.0")
	assert.False(t, mock.updateCalled)
}
