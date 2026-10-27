package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/language"
	"sigs.k8s.io/yaml"
)

// The two-tier taxonomy an app declares: `metadata.categories_v2` picks the
// sections it appears under, `metadata.tags` refines that within a section, and
// `spec.locale` says which languages its own text is written in.
//
// The first two are decoded here rather than through github.com/beclab/api's
// AppMetaData because that type is shared with the cluster runtime, which has
// no use for them: the values are consumed by the market catalog. Decoding the
// same bytes twice is cheaper than a version bump across every consumer of a
// shared schema for two fields none of them read. `spec.locale` is not in that
// position -- AppSpec already carries it -- so it is read off the parsed
// AppConfiguration instead of decoded again.
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

const (
	categoriesV2Field = "categories_v2"
	tagsField         = "tags"
)

// catalogMetadata mirrors the two catalog fields inside `metadata`; every other
// key is decoded through AppMetaData.
type catalogMetadata struct {
	CategoriesV2 []string `json:"categories_v2,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// metadataEnvelope keeps `metadata` as raw bytes so it can be read twice: once
// for the catalog fields, once as a plain mapping whose keys are checked
// against the schema.
type metadataEnvelope struct {
	Metadata json.RawMessage `json:"metadata"`
}

// A category id and a tag slug are the same shape: they are typed into a
// manifest by hand and then compared for equality across three services, so
// anything that has a second spelling -- case, spaces, underscores -- is a
// mismatch waiting to be reported as "unknown category".
var taxonomySlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ParseTaxonomy reads the taxonomy fields out of an already-rendered manifest.
func ParseTaxonomy(rendered []byte) (Taxonomy, error) {
	var cfg AppConfiguration
	if err := yaml.Unmarshal(rendered, &cfg); err != nil {
		return Taxonomy{}, err
	}
	taxonomy, _, err := parseTaxonomy(rendered, &cfg)
	return taxonomy, err
}

// parseTaxonomy also returns the keys the manifest spelled under `metadata`,
// which is what lets a misspelt catalog field be reported rather than silently
// dropped by the YAML decoder.
func parseTaxonomy(rendered []byte, cfg *AppConfiguration) (Taxonomy, []string, error) {
	var envelope metadataEnvelope
	if err := yaml.Unmarshal(rendered, &envelope); err != nil {
		return Taxonomy{}, nil, err
	}
	if len(envelope.Metadata) == 0 {
		return Taxonomy{Locale: cfg.Spec.Locale}, nil, nil
	}

	var catalog catalogMetadata
	if err := json.Unmarshal(envelope.Metadata, &catalog); err != nil {
		return Taxonomy{}, nil, fmt.Errorf("read metadata: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(envelope.Metadata, &fields); err != nil {
		return Taxonomy{}, nil, fmt.Errorf("metadata must be a mapping: %w", err)
	}

	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)

	return Taxonomy{
		CategoriesV2: catalog.CategoriesV2,
		Tags:         catalog.Tags,
		Locale:       cfg.Spec.Locale,
	}, names, nil
}

// ValidateTaxonomy checks the shape of the three lists. All three are optional:
// an app that predates the new taxonomy keeps working, and the market falls
// back to its legacy `categories`.
func ValidateTaxonomy(t Taxonomy) error {
	return errors.Join(
		validateSlugList("metadata."+categoriesV2Field, t.CategoriesV2),
		validateSlugList("metadata."+tagsField, t.Tags),
		validateLocaleList(t.Locale),
	)
}

// validateMetadataFields rejects keys `metadata` has no schema for. A YAML
// decoder drops them without a word, so `categoriesV2` produces an app with no
// categories and no explanation; the near-miss spelling is named in the error
// because that is the mistake this check is here to catch.
func validateMetadataFields(fields []string) error {
	var errs []error
	for _, field := range fields {
		if _, known := knownMetadataFields[field]; known {
			continue
		}
		if suggestion, ok := nearestMetadataField(field); ok {
			errs = append(errs, fmt.Errorf("metadata field %q is not part of the manifest schema; did you mean %q?", field, suggestion))
			continue
		}
		errs = append(errs, fmt.Errorf("metadata field %q is not part of the manifest schema", field))
	}
	return errors.Join(errs...)
}

// knownMetadataFields is every key `metadata` may spell: the json names of the
// shared AppMetaData plus the two catalog fields that type does not carry.
// Reading the names off the struct means a field added upstream is accepted
// here without a second list to keep in step.
var knownMetadataFields = collectMetadataFields()

func collectMetadataFields() map[string]struct{} {
	fields := map[string]struct{}{
		categoriesV2Field: {},
		tagsField:         {},
	}
	meta := reflect.TypeOf(AppMetaData{})
	for i := 0; i < meta.NumField(); i++ {
		field := meta.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			name = field.Name
		}
		fields[name] = struct{}{}
	}
	return fields
}

// nearestMetadataField finds the schema field an unknown key differs from only
// in case and separators -- `categoriesV2` for `categories_v2`.
func nearestMetadataField(field string) (string, bool) {
	target := foldMetadataField(field)
	for known := range knownMetadataFields {
		if foldMetadataField(known) == target {
			return known, true
		}
	}
	return "", false
}

func foldMetadataField(field string) string {
	var folded strings.Builder
	for _, r := range strings.ToLower(field) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			folded.WriteRune(r)
		}
	}
	return folded.String()
}

// namesUndefinedLanguage reports whether a canonical tag has `und` as its
// language subtag. It parses and canonicalizes cleanly, but a catalog cannot
// offer it to anybody: an app declaring it has declared nothing. A wholly
// private-use tag such as `x-private` carries no language subtag at all and is
// left alone -- it is unknown to the standard, not undefined by the app.
func namesUndefinedLanguage(canonical string) bool {
	return canonical == "und" || strings.HasPrefix(canonical, "und-")
}

func validateSlugList(field string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	// Duplicates are looked for before shape, so a value that is both
	// misspelt and repeated is reported as two distinct mistakes rather than
	// as the same shape complaint twice.
	seen := make(map[string]struct{}, len(values))
	var errs []error
	for _, value := range values {
		if _, dup := seen[value]; dup {
			errs = append(errs, fmt.Errorf("%s value %q is listed more than once", field, value))
			continue
		}
		seen[value] = struct{}{}

		if value == "" {
			errs = append(errs, fmt.Errorf("%s must not be empty", field))
			continue
		}
		if value != strings.TrimSpace(value) {
			errs = append(errs, fmt.Errorf("%s value %q must not have surrounding whitespace", field, value))
			continue
		}
		if !taxonomySlug.MatchString(value) {
			errs = append(errs, fmt.Errorf("%s value %q must be lowercase alphanumeric words joined by single hyphens", field, value))
		}
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
		if _, dup := seen[value]; dup {
			errs = append(errs, fmt.Errorf("spec.locale value %q is listed more than once", value))
			continue
		}
		seen[value] = struct{}{}

		if value == "" {
			errs = append(errs, errors.New("spec.locale must not be empty"))
			continue
		}
		tag, err := language.BCP47.Parse(value)
		if err != nil {
			errs = append(errs, fmt.Errorf("spec.locale value %q must be a valid BCP 47 language tag", value))
			continue
		}
		if tag.String() != value {
			errs = append(errs, fmt.Errorf("spec.locale value %q must use canonical spelling %q", value, tag.String()))
			continue
		}
		if namesUndefinedLanguage(value) {
			errs = append(errs, fmt.Errorf("spec.locale value %q names the undefined language; list the languages the app is actually written in", value))
		}
	}
	return errors.Join(errs...)
}
