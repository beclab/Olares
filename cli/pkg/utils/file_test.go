package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateFileVersion(t *testing.T) {
	dir := t.TempDir()

	prepared := filepath.Join(dir, ".prepared")
	assert.NoError(t, os.WriteFile(prepared, []byte("1.12.7-20260728\n"), 0644))
	assert.Equal(t, "1.12.7-20260728", StateFileVersion(prepared))

	// .installed holds "<version> <kubetype>", so only the first field is the
	// version; reading the whole line would never match anything.
	installed := filepath.Join(dir, ".installed")
	assert.NoError(t, os.WriteFile(installed, []byte("1.12.7-20260728 k3s\n"), 0644))
	assert.Equal(t, "1.12.7-20260728", StateFileVersion(installed))

	empty := filepath.Join(dir, ".empty")
	assert.NoError(t, os.WriteFile(empty, []byte("  \n"), 0644))
	assert.Empty(t, StateFileVersion(empty))

	assert.Empty(t, StateFileVersion(filepath.Join(dir, "nonexistent")))
}
