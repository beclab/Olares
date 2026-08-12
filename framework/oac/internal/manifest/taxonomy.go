package manifest

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/language"
	"sigs.k8s.io/yaml"
)

// The two-tier taxonomy an app declares: `metadata.categories_v2` picks the
// sections it appears under, `metadata.tags` refines that within a section, and
// `spec.locale` says which languages its own text is written in.
//
// These are decoded here rather than through github.com/beclab/api's
// AppMetaData because that type is shared with the cluster runtime, which has
// no use for them: the values are consumed by the market catalog. Decoding the
// same bytes twice is cheaper than a version bump across every consumer of a
// shared schema for three fields none of them read.
//
// What is checked here is shape, not membership. Whether `agents` is a category
// this market actually offers depends on the market's own registry, which oac
// cannot see -- gitbot answers that against adminv2. Checking shape here means
// a developer running `oac lint` locally still learns that a comma-separated
// string is not a list before opening a pull request.

// Taxonomy is the catalog-facing slice of a manifest.
type Taxonomy struct {
	CategoriesV2 []string `json:"categories_v2,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Locale       []string `json:"locale,omitempty"`
}

// taxonomyEnvelope mirrors just enough of the manifest to reach the three
// fields; every other key is ignored.
type taxonomyEnvelope struct {
	Metadata struct {
		CategoriesV2 []string `json:"categories_v2,omitempty"`
		Tags         []string `json:"tags,omitempty"`
	} `json:"metadata"`
	Spec struct {
		Locale []string `json:"locale,omitempty"`
	} `json:"spec"`
}

// A category id and a tag slug are the same shape: they are typed into a
// manifest by hand and then compared for equality across three services, so
// anything that has a second spelling -- case, spaces, underscores -- is a
// mismatch waiting to be reported as "unknown category".
var taxonomySlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ParseTaxonomy reads the taxonomy fields out of an already-rendered manifest.
func ParseTaxonomy(rendered []byte) (Taxonomy, error) {
	var envelope taxonomyEnvelope
	if err := yaml.Unmarshal(rendered, &envelope); err != nil {
		return Taxonomy{}, err
	}
	return Taxonomy{
		CategoriesV2: envelope.Metadata.CategoriesV2,
		Tags:         envelope.Metadata.Tags,
		Locale:       envelope.Spec.Locale,
	}, nil
}

// ValidateTaxonomy checks the shape of the three lists. All three are optional:
// an app that predates the new taxonomy keeps working, and the market falls
// back to its legacy `categories`.
func ValidateTaxonomy(t Taxonomy) error {
	return errors.Join(
		validateSlugList("metadata.categories_v2", t.CategoriesV2),
		validateSlugList("metadata.tags", t.Tags),
		validateLocaleList(t.Locale),
	)
}

func validateSlugList(field string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	var errs []error
	for _, value := range values {
		if value != strings.TrimSpace(value) {
			errs = append(errs, fmt.Errorf("%s value %q must not have surrounding whitespace", field, value))
			continue
		}
		if !taxonomySlug.MatchString(value) {
			errs = append(errs, fmt.Errorf("%s value %q must be lowercase alphanumeric words joined by single hyphens", field, value))
			continue
		}
		if _, dup := seen[value]; dup {
			errs = append(errs, fmt.Errorf("%s value %q is listed more than once", field, value))
			continue
		}
		seen[value] = struct{}{}
	}
	return errors.Join(errs...)
}

func validateLocaleList(values []string) error {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	var errs []error
	for _, value := range values {
		tag, err := language.BCP47.Parse(value)
		if err != nil {
			errs = append(errs, fmt.Errorf("spec.locale value %q must be a valid BCP 47 language tag", value))
			continue
		}
		if tag.String() != value {
			errs = append(errs, fmt.Errorf("spec.locale value %q must use canonical spelling %q", value, tag.String()))
			continue
		}
		if _, dup := seen[value]; dup {
			errs = append(errs, fmt.Errorf("spec.locale value %q is listed more than once", value))
			continue
		}
		seen[value] = struct{}{}
	}
	return errors.Join(errs...)
}
