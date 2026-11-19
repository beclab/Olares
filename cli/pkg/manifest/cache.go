package manifest

import (
	"errors"

	"github.com/beclab/Olares/cli/pkg/core/cache"
)

const CacheKey = "installation.manifest"

// FromCache returns the manifest loaded by the pipeline. Actions receive this
// shared cache through Init; reading it never opens the manifest file again.
func FromCache(pipelineCache *cache.Cache) (InstallationManifest, error) {
	if pipelineCache != nil {
		value, _ := pipelineCache.Get(CacheKey)
		if items, ok := value.(InstallationManifest); ok && items != nil {
			return items, nil
		}
	}
	return nil, errors.New("installation manifest is not initialized in pipeline cache")
}
