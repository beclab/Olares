package terminus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateInitialLinkerdPKIMaterial(t *testing.T) {
	mat, err := generateInitialLinkerdPKIMaterial()
	require.NoError(t, err)
	require.NotEmpty(t, mat.CACrt)
	require.NotEmpty(t, mat.CAKey)
	require.NotEmpty(t, mat.IssuerCrt)
	require.NotEmpty(t, mat.IssuerKey)
}
