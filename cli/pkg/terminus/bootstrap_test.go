package terminus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The join script follows the same convention as every other versioned artifact
// on the CDN: version in the file name, at the bucket root, no vendor path.
func TestJoinClusterScriptURL(t *testing.T) {
	assert.Equal(t,
		"https://cdn.olares.cn/joincluster-v1.12.7-20260728.sh",
		JoinClusterScriptURL("https://cdn.olares.cn", "1.12.7-20260728"))
	assert.Equal(t,
		"https://cdn.olares.cn/joincluster-v1.12.7.sh",
		JoinClusterScriptURL("https://cdn.olares.cn/", "1.12.7"))
	// an unset CDN must still yield a usable URL rather than a relative path
	assert.Equal(t,
		"https://cdn.olares.com/joincluster-v1.12.7.sh",
		JoinClusterScriptURL("", "1.12.7"))
}
