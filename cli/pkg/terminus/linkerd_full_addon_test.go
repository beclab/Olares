package terminus

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateLinkerdControlPlaneDeploy_missingFile(t *testing.T) {
	err := validateLinkerdControlPlaneDeploy(t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "linkerd-control-plane.yaml")
}

func TestValidateLinkerdControlPlaneDeploy_present(t *testing.T) {
	root := t.TempDir()
	deploy := osFrameworkDeployPath(root)
	require.NoError(t, os.MkdirAll(deploy, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(deploy, "linkerd-control-plane.yaml"), []byte("apiVersion: v1\n"), 0o644))
	require.NoError(t, validateLinkerdControlPlaneDeploy(root))
}

func TestLinkerdPostInstallEnabled_defaultOn(t *testing.T) {
	t.Setenv("OLARES_LINKERD_POST_INSTALL_ENABLED", "")
	require.True(t, linkerdPostInstallEnabled())
}

func TestLinkerdPostInstallEnabled_optOut(t *testing.T) {
	t.Setenv("OLARES_LINKERD_POST_INSTALL_ENABLED", "0")
	require.False(t, linkerdPostInstallEnabled())
}

func TestGenerateInitialLinkerdPKIMaterial(t *testing.T) {
	mat, err := generateInitialLinkerdPKIMaterial()
	require.NoError(t, err)
	require.NotEmpty(t, mat.CACrt)
	require.NotEmpty(t, mat.CAKey)
	require.NotEmpty(t, mat.IssuerCrt)
	require.NotEmpty(t, mat.IssuerKey)
}
