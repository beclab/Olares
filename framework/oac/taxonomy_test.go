package oac_test

import (
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/oac"
)

const taxonomyManifest = `olaresManifest.version: 0.12.0
olaresManifest.type: app
apiVersion: v1
metadata:
  name: demo
  title: Demo
  version: 1.0.0
  categories: 
    - Productivity
  categories_v2:
    - agents
    - workspace
  tags:
    - coding
    - self-hosted
spec:
  versionName: v1
  locale:
    - en-US
    - zh-CN
`

// gitbot and the market tools read the taxonomy off a manifest they already
// loaded, so it has to be reachable through the exported Manifest.
func TestLoadedManifestCarriesItsTaxonomy(t *testing.T) {
	m, err := oac.New().LoadManifestContent([]byte(taxonomyManifest))
	if err != nil {
		t.Fatalf("LoadManifestContent: %v", err)
	}

	got := m.Taxonomy()
	want := oac.Taxonomy{
		CategoriesV2: []string{"agents", "workspace"},
		Tags:         []string{"coding", "self-hosted"},
		Locale:       []string{"en-US", "zh-CN"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Taxonomy() = %+v, want %+v", got, want)
	}
}

func TestParseTaxonomyReadsRenderedBytes(t *testing.T) {
	got, err := oac.ParseTaxonomy([]byte(taxonomyManifest))
	if err != nil {
		t.Fatalf("ParseTaxonomy: %v", err)
	}
	want := oac.Taxonomy{
		CategoriesV2: []string{"agents", "workspace"},
		Tags:         []string{"coding", "self-hosted"},
		Locale:       []string{"en-US", "zh-CN"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseTaxonomy = %+v, want %+v", got, want)
	}
}

// spec.locale is a field of the shared AppConfiguration, and the taxonomy view
// of it must be that same value rather than a second decode of the same bytes:
// two decoders of one field can disagree, and then lint and the market read
// different languages off the same manifest.
func TestTaxonomyLocaleIsTheParsedSpecLocale(t *testing.T) {
	m, err := oac.New().LoadManifestContent([]byte(taxonomyManifest))
	if err != nil {
		t.Fatalf("LoadManifestContent: %v", err)
	}
	cfg, ok := oac.AsAppConfiguration(m)
	if !ok {
		t.Fatal("expected an *AppConfiguration")
	}
	if !reflect.DeepEqual(m.Taxonomy().Locale, cfg.Spec.Locale) {
		t.Fatalf("Taxonomy().Locale = %v, spec.locale = %v", m.Taxonomy().Locale, cfg.Spec.Locale)
	}
}

// An app that predates the taxonomy declares none of it, and reading it back
// must not invent empty lists a caller would have to distinguish from real ones.
func TestTaxonomyIsAbsentWhenUndeclared(t *testing.T) {
	const bare = `olaresManifest.version: 0.12.0
olaresManifest.type: app
apiVersion: v1
metadata:
  name: demo
  title: Demo
  version: 1.0.0
spec:
  versionName: v1
`
	m, err := oac.New().LoadManifestContent([]byte(bare))
	if err != nil {
		t.Fatalf("LoadManifestContent: %v", err)
	}
	if got := m.Taxonomy(); len(got.CategoriesV2) != 0 || len(got.Tags) != 0 || len(got.Locale) != 0 {
		t.Fatalf("Taxonomy() = %+v, want empty", got)
	}
}
