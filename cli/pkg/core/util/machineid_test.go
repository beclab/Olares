package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirstNonEmptyFileContent(t *testing.T) {
	tempDir := t.TempDir()
	empty := filepath.Join(tempDir, "empty")
	identifier := filepath.Join(tempDir, "identifier")
	require.NoError(t, os.WriteFile(empty, nil, 0o600))
	require.NoError(t, os.WriteFile(identifier, []byte("  ABCD-1234\x00\n"), 0o600))

	require.Equal(t, "ABCD-1234", firstNonEmptyFileContent([]string{
		filepath.Join(tempDir, "missing"),
		empty,
		identifier,
	}))
}

func TestFirstNonEmptyFileContentReturnsEmpty(t *testing.T) {
	require.Empty(t, firstNonEmptyFileContent([]string{
		filepath.Join(t.TempDir(), "missing"),
	}))
}
