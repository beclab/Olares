package oac

import (
	"github.com/beclab/Olares/framework/oac/internal/manifest"
)

// Taxonomy is the catalog-facing slice of a manifest: the sections an app
// appears under (`metadata.categories_v2`), the tags that refine that within a
// section (`metadata.tags`), and the languages its own text is written in
// (`spec.locale`).
//
// oac checks the shape of these values; whether a given category or tag is one
// the market actually offers lives in the market's own registry, so gitbot
// answers that question and reads the declared values from here.
type Taxonomy = manifest.Taxonomy

// ParseTaxonomy reads the taxonomy out of an already-rendered manifest, for
// callers holding manifest bytes rather than a parsed Manifest. A Manifest
// loaded through this package carries the same values on its Taxonomy method.
func ParseTaxonomy(rendered []byte) (Taxonomy, error) {
	return manifest.ParseTaxonomy(rendered)
}
