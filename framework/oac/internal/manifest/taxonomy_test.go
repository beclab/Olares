package manifest

import (
	"strings"
	"testing"
)

func TestParseTaxonomyReadsBothLevelsAndLocale(t *testing.T) {
	raw := []byte(`
metadata:
  name: wise
  categories_v2:
    - agents
    - workspace
  tags:
    - coding
    - self-hosted
spec:
  locale:
    - en-US
    - zh-CN
`)

	got, err := ParseTaxonomy(raw)
	if err != nil {
		t.Fatalf("ParseTaxonomy: %v", err)
	}
	if len(got.CategoriesV2) != 2 || got.CategoriesV2[0] != "agents" {
		t.Fatalf("categories_v2 = %v", got.CategoriesV2)
	}
	if len(got.Tags) != 2 || got.Tags[1] != "self-hosted" {
		t.Fatalf("tags = %v", got.Tags)
	}
	if len(got.Locale) != 2 || got.Locale[1] != "zh-CN" {
		t.Fatalf("locale = %v", got.Locale)
	}
	if err := ValidateTaxonomy(got); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
}

// An app written before the new taxonomy declares none of it and must still
// pass: the market falls back to its legacy categories.
func TestValidateTaxonomyAcceptsAbsence(t *testing.T) {
	if err := ValidateTaxonomy(Taxonomy{}); err != nil {
		t.Fatalf("empty taxonomy rejected: %v", err)
	}
}

func TestValidateTaxonomyRejectsShape(t *testing.T) {
	cases := []struct {
		name     string
		taxonomy Taxonomy
		want     string
	}{
		{
			name:     "category with capitals",
			taxonomy: Taxonomy{CategoriesV2: []string{"Agents"}},
			want:     "lowercase",
		},
		{
			name:     "category with a space",
			taxonomy: Taxonomy{CategoriesV2: []string{"social network"}},
			want:     "lowercase",
		},
		{
			name:     "category with an underscore",
			taxonomy: Taxonomy{CategoriesV2: []string{"social_network"}},
			want:     "lowercase",
		},
		{
			name:     "duplicate category",
			taxonomy: Taxonomy{CategoriesV2: []string{"agents", "agents"}},
			want:     "more than once",
		},
		{
			name:     "padded tag",
			taxonomy: Taxonomy{Tags: []string{" coding"}},
			want:     "whitespace",
		},
		{
			name:     "duplicate tag",
			taxonomy: Taxonomy{Tags: []string{"coding", "coding"}},
			want:     "more than once",
		},
		{
			name:     "locale is not a language code",
			taxonomy: Taxonomy{Locale: []string{"english"}},
			want:     "BCP 47",
		},
		{
			name:     "locale region is lowercase",
			taxonomy: Taxonomy{Locale: []string{"zh-cn"}},
			want:     "canonical",
		},
		{
			name:     "locale includes a suppressed script",
			taxonomy: Taxonomy{Locale: []string{"en-Latn"}},
			want:     "canonical",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTaxonomy(tc.taxonomy)
			if err == nil {
				t.Fatalf("%+v was accepted", tc.taxonomy)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestValidateTaxonomyDoesNotImposeBusinessCountLimits(t *testing.T) {
	taxonomy := Taxonomy{
		CategoriesV2: []string{"a", "b", "c", "d", "e", "f"},
		Tags:         []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"},
	}
	if err := ValidateTaxonomy(taxonomy); err != nil {
		t.Fatalf("valid taxonomy rejected: %v", err)
	}
}

func TestValidateTaxonomyAcceptsCanonicalBCP47Locales(t *testing.T) {
	locales := []string{
		"en",
		"es-419",
		"de-DE-1996",
		"sl-rozaj",
		"en-US-u-ca-gregory",
		"x-private",
	}
	if err := ValidateTaxonomy(Taxonomy{Locale: locales}); err != nil {
		t.Fatalf("valid BCP 47 locales rejected: %v", err)
	}
}
