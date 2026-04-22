package selfupgrade_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/cmdexec"
	"github.com/qdrant/qcloud-cli/internal/selfupgrade"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

type mockUpdater struct {
	latestRelease *selfupgrade.ReleaseInfo
	latestFound   bool
	latestErr     error
	updateSelfRel *selfupgrade.ReleaseInfo
	updateSelfErr error
	updateCalled  bool
}

func (m *mockUpdater) DetectLatest(_ context.Context) (*selfupgrade.ReleaseInfo, bool, error) {
	return m.latestRelease, m.latestFound, m.latestErr
}

func (m *mockUpdater) UpdateSelf(_ context.Context, _ string) (*selfupgrade.ReleaseInfo, error) {
	m.updateCalled = true
	return m.updateSelfRel, m.updateSelfErr
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
		latestRelease: selfupgrade.NewReleaseInfo("0.4.0"),
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
		latestRelease: selfupgrade.NewReleaseInfo("0.5.0"),
		latestFound:   true,
	}
	env.State.SetUpdater(mock)

	stdout, _, err := testutil.Exec(t, env, "self-upgrade", "--check")
	require.NoError(t, err)
	assert.Contains(t, stdout, "New version available: 0.5.0")
	assert.Contains(t, stdout, "current: 0.4.0")
	assert.False(t, mock.updateCalled)
}

func TestSelfUpgrade_DevVersionRequiresForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("0.4.0-dev"))
	mock := &mockUpdater{
		latestRelease: selfupgrade.NewReleaseInfo("0.5.0"),
		latestFound:   true,
	}
	env.State.SetUpdater(mock)

	stdout, _, err := testutil.Exec(t, env, "self-upgrade")
	require.NoError(t, err)
	assert.Contains(t, stdout, "dev build")
	assert.Contains(t, stdout, "--force")
	assert.False(t, mock.updateCalled)
}

func TestSelfUpgrade_HomebrewCheckPassesRunnerThrough(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("0.4.0"))

	// Simulate "brew not found" — the homebrew check should return false,
	// allowing the upgrade flow to proceed normally.
	runner := cmdexec.NewMockRunner().
		Respond("brew", []string{"--prefix"}, nil, fmt.Errorf("brew not found"))

	env.State.CmdRunner = runner

	mock := &mockUpdater{
		latestRelease: selfupgrade.NewReleaseInfo("0.4.0"),
		latestFound:   true,
	}
	env.State.SetUpdater(mock)

	stdout, _, err := testutil.Exec(t, env, "self-upgrade")
	require.NoError(t, err)
	assert.Contains(t, stdout, "already up to date")
	require.Equal(t, 1, runner.CallCount())
	assert.Equal(t, cmdexec.Invocation{Name: "brew", Args: []string{"--prefix"}}, runner.Call(0))
}
