package manifest

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var validChartName = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

var (
	errInvalidSubChartName = fmt.Errorf(
		"invalid subchart name, must match regex %s and the length must not be longer than 53",
		validChartName.String())
	errMissingSubChartName = errors.New("no name provided")
)

var isSemver = validation.NewStringRuleWithError(func(s string) bool {
	if s == "" {
		return false
	}
	_, err := semver.NewVersion(s)
	return err == nil
}, validation.NewError("validation_is_semver", "must be a valid semantic version (e.g. 1.2.3)"))

var isHTTPURL = validation.NewStringRuleWithError(func(s string) bool {
	if s == "" {
		return true
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}, validation.NewError("validation_is_http_url", "must be a valid http(s) URL"))

var k8sQuantity = regexp.MustCompile(`^(?:\d+(?:\.\d+)?(?:[eE][-+]?(\d+|i))?(?:[kKMGTP]?i?|[mMGTPE])?|[kKMGTP]i|[mMGTPE])$`)

var (
	entranceNameRegex    = regexp.MustCompile(`^([a-z0-9A-Z-]*)$`)
	entranceHostRegex    = regexp.MustCompile(`^([a-z]([-a-z0-9]*[a-z0-9]))$`)
	validDurationRegex   = regexp.MustCompile(`^((?:[-+]?\d+(?:\.\d+)?([smhdwy]|us|ns|ms))+)$`)
	validAPIVersions     = []any{APIVersionV1, APIVersionV2, APIVersionV3}
	validDependencyTypes = []any{"system", "application", "middleware"}
	validAuthLevels      = []any{"", "public", "private", "internal"}
	validOpenMethods     = []any{"", "default", "iframe", "window"}
	// validOverlayProtocols enumerates the protocols an overlayGateway
	// entrance may declare. An empty value is allowed and means the
	// entrance is reachable over both tcp and udp.
	validOverlayProtocols = []any{"", "tcp", "udp"}
)

// validSupportArchSet enumerates the architectures spec.supportArch may
// declare. The check intentionally rejects empty strings and case variants
// (e.g. "AMD64") since the downstream installer compares against exactly
// these two lowercase values.
var validSupportArchSet = map[string]struct{}{
	"amd64": {},
	"arm64": {},
}

// ValidateKnownAPIVersion returns nil if api is empty (treated as v1) or one
// of v1, v2, v3 (case-insensitive). Otherwise it returns an error with the
// message used by manifest validation and resource checks.
func ValidateKnownAPIVersion(api string) error {
	if strings.TrimSpace(api) == "" {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(api)) {
	case APIVersionV1, APIVersionV2, APIVersionV3:
		return nil
	default:
		return fmt.Errorf("not supported version")
	}
}

// ValidateAppConfiguration runs structural and cross-field checks on the manifest.
func ValidateAppConfiguration(c *AppConfiguration) error {
	structErr := validation.ValidateStruct(c,
		validation.Field(&c.ConfigVersion,
			validation.Required.Error("olaresManifest.version is required")),
		validation.Field(&c.APIVersion,
			validation.When(c.APIVersion != "",
				validation.In(validAPIVersions...).Error("not supported version"),
			),
		),
		validation.Field(&c.Metadata, validation.By(validateAppMetaData)),
		validation.Field(&c.Entrances,
			validation.Required.Error("entrances is required"),
			validation.Length(1, 10).Error("entrances must have between 1 and 10 items"),
			validation.Each(validation.By(validateEntranceValue)),
			validation.By(uniqueEntranceNames),
		),
		validation.Field(&c.Spec,
			validation.Required.Error("spec is required"),
			validation.By(validateAppSpecFor(c)),
		),
		validation.Field(&c.Options, validation.By(validateOptionsFor(c))),
		validation.Field(&c.OverlayGateway, validation.By(validateOverlayGateway)),
	)
	return errors.Join(
		structErr,
		checkSubCharts(c),
		validatePermission(c.ConfigVersion, c.Permission),
		validateRootProvider(c.ConfigVersion, c.Provider),
		validateWorkloadReplicas(c.ConfigVersion, c.APIVersion, c.WorkloadReplicas),
		validateOlaresDependency(c),
		validateSharedAppRequirements(c),
		validateV3Configuration(c),
	)
}

func validateAppMetaData(v interface{}) error {
	m, ok := v.(AppMetaData)
	if !ok {
		return fmt.Errorf("metadata: unexpected type %T", v)
	}
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name,
			validation.Required.Error("metadata.name is required"),
			validation.Length(1, 30).Error("metadata.name must be 1-30 characters"),
		),
		validation.Field(&m.Icon,
			validation.Required.Error("metadata.icon is required"),
			isHTTPURL.Error("metadata.icon must be a valid http(s) URL"),
		),
		validation.Field(&m.Description,
			validation.Required.Error("metadata.description is required"),
		),
		validation.Field(&m.Title,
			validation.Required.Error("metadata.title is required"),
			validation.Length(1, 30).Error("metadata.title must be 1-30 characters"),
		),
		validation.Field(&m.Version,
			validation.Required.Error("metadata.version is required"),
			isSemver.Error("metadata.version must be a valid semantic version (e.g. 1.2.3)"),
		),
		validation.Field(&m.AppID,
			validation.By(validateMetadataAppID),
		),
	)
}

// validateMetadataAppID rejects a metadata.appid value that collides with a
// reserved built-in system app id. Empty appid is permitted -- the loader
// normalizes it to md5(metadata.name)[:8] at LoadAppConfiguration time, and
// downstream consumers that require a non-empty appid surface their own
// errors (e.g. "market upload" rejects a missing field).
func validateMetadataAppID(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("metadata.appid: unexpected type %T", v)
	}
	if s == "" {
		return nil
	}
	if IsReservedSystemAppID(s) {
		return fmt.Errorf(
			"metadata.appid %q collides with a reserved system app id; choose a different value (the loader normalizes appid to md5(metadata.name)[:8] anyway, so leaving the field empty is also fine)",
			s,
		)
	}
	return nil
}

func validateEntranceValue(v interface{}) error {
	e, ok := v.(appv1.Entrance)
	if !ok {
		return fmt.Errorf("entrance: unexpected type %T", v)
	}
	return validation.ValidateStruct(&e,
		validation.Field(&e.Name,
			validation.Match(entranceNameRegex).Error("entrance.name must match ^[a-z0-9A-Z-]*$"),
			validation.Length(0, 63).Error("entrance.name must be <= 63 characters"),
		),
		validation.Field(&e.Host,
			validation.Match(entranceHostRegex).Error("entrance.host must match ^[a-z]([-a-z0-9]*[a-z0-9])$"),
			validation.Length(0, 63).Error("entrance.host must be <= 63 characters"),
		),
		validation.Field(&e.Port,
			validation.Min(int32(1)).Error("entrance.port must be > 0"),
		),
		validation.Field(&e.Icon,
			isHTTPURL.Error("entrance.icon must be a valid http(s) URL"),
		),
		validation.Field(&e.Title,
			validation.Required.Error("entrance.title is required"),
			validation.Length(1, 30).Error("entrance.title must be 1-30 characters"),
		),
		validation.Field(&e.AuthLevel,
			validation.In(validAuthLevels...).Error(`entrance.authLevel must be one of "", "public", "private"`),
		),
		validation.Field(&e.OpenMethod,
			validation.In(validOpenMethods...).Error(`entrance.openMethod must be one of "", "default", "iframe", "window"`),
		),
	)
}

// validateAppSpecFor binds olaresManifest.version so spec validation can
// branch: below 0.12.0 it checks legacy flat quantities; from 0.12.0 onward it
// checks spec.resources[] (per-element rules) plus Rule 7 and mode→arch via
// specResourceCrossFieldRules.
func validateAppSpecFor(cfg *AppConfiguration) validation.RuleFunc {
	return func(v interface{}) error {
		s, ok := v.(AppSpec)
		if !ok {
			return fmt.Errorf("spec: unexpected type %T", v)
		}
		return validateAppSpec(cfg.ConfigVersion, cfg.APIVersion, cfg.Options.TemplateOnly, s)
	}
}

// normalizeAPIVersion returns a canonical lowercase apiVersion; empty means v1.
func normalizeAPIVersion(api string) string {
	s := strings.ToLower(strings.TrimSpace(api))
	if s == "" {
		return APIVersionV1
	}
	return s
}

// validateAppSpec branches on olaresManifest.version and apiVersion:
//
//   - < 0.12.0 (legacy): the flat spec.requiredCpu/limitedCpu/requiredMemory/
//     limitedMemory/requiredDisk fields are required (and quantity-checked);
//     spec.limitedDisk is optional (quantity-checked when set).
//     spec.resources[] is not consulted on this branch.
//   - >= 0.12.0 (modern), apiVersion not v2: spec.resources is required and
//     every declared entry must form a complete envelope via per-element
//     ValidateResourceMode plus the empty-section check in
//     specResourceCrossFieldRules. The legacy flat
//     spec.required*/spec.limited* fields must be empty (Rule 7, enforced
//     via specResourceCrossFieldRules) and therefore aren't quantity-
//     checked at the spec level on this branch.
//   - apiVersion=v2: spec.resources is never allowed (even when
//     olaresManifest.version >= 0.12.0); use the legacy flat fields instead.
//   - olaresManifest.version >= 0.12.0: spec.supportedGpu must be empty
//     (GPU shapes belong in spec.resources[] via the appropriate mode).
//
// As an empty-spec optimisation, validateAppSpec emits a single
// consolidated guidance message instead of letting the per-field
// "is required" rules cascade:
//
//   - At >= 0.12.0 the canonical declaration is spec.resources; its own
//     Required rule already produces a single message pointing the user
//     at the right place.
//   - At < 0.12.0 when none of the five required legacy fields is set,
//     a single guidance error replaces the five "is required" cascade
//     so the user gets one obvious next step instead of repeated noise.
//     If they did set some but not all of the five, the per-field
//     required rules still fire as before -- pinpointed feedback is more
//     useful than a vague "fill the envelope" message in that case.
//
// spec.requiredGpu / spec.limitedGpu are never required; when set they are
// only validated as Kubernetes quantities. They live at the spec level for
// backwards compatibility but should be expressed inside a
// spec.resources[] entry on modern manifests.
func validateAppSpec(configVersion, apiVersion string, templateOnly bool, s AppSpec) error {
	optionalGPUQuantity := validation.When(s.RequiredGPU != "", flatResourceQuantityRule("spec.requiredGpu", templateOnly))
	optionalLimitedGPUQuantity := validation.When(s.LimitedGPU != "", flatResourceQuantityRule("spec.limitedGpu", templateOnly))
	optionalLimitedDiskQuantity := validation.When(s.LimitedDisk != "", flatResourceQuantityRule("spec.limitedDisk", templateOnly))

	fields := []*validation.FieldRules{
		validation.Field(&s.RequiredGPU, optionalGPUQuantity),
		validation.Field(&s.LimitedGPU, optionalLimitedGPUQuantity),
		validation.Field(&s.SupportArch,
			validation.Each(validation.By(validateSupportArchEntry)),
			validation.By(uniqueSupportArches),
		),
		validation.Field(&s.SubCharts),
	}

	api := normalizeAPIVersion(apiVersion)
	var v2ResourcesErr error
	if api == APIVersionV2 && len(s.Accelerator) > 0 {
		v2ResourcesErr = fmt.Errorf(
			"spec.resources is not supported for apiVersion=v2 (including when olaresManifest.version >= 0.12.0); use spec.requiredCpu, spec.limitedCpu, spec.requiredMemory, spec.limitedMemory, and spec.requiredDisk instead",
		)
	}

	var supportedGpuModernErr error
	if resourcesCheckApplies(configVersion) && len(s.SupportedGpu) > 0 {
		supportedGpuModernErr = fmt.Errorf(
			"spec.supportedGpu must be empty for olaresManifest.version >= 0.12.0; declare GPU resources in spec.resources[] with the appropriate mode (e.g. nvidia, amd-gpu)",
		)
	}

	modern := resourcesCheckApplies(configVersion) && api != APIVersionV2
	var versionGuidance error

	switch {
	case modern:
		fields = append(fields,
			validation.Field(&s.Accelerator,
				validation.By(validateResourceModeValueFor(templateOnly, &s)),
			),
		)
	case isLegacyEnvelopeMissing(&s):
		versionGuidance = fmt.Errorf(
			"spec.requiredCpu / spec.limitedCpu / spec.requiredMemory / spec.limitedMemory / spec.requiredDisk are required for olaresManifest.version < 0.12.0; populate the legacy resource envelope",
		)
		fields = append(fields,
			validation.Field(&s.LimitedDisk, optionalLimitedDiskQuantity),
		)
	default:
		fields = append(fields, legacyEnvelopeFieldRules(&s, templateOnly)...)
	}

	structErr := validation.ValidateStruct(&s, fields...)
	return errors.Join(v2ResourcesErr, supportedGpuModernErr, structErr, versionGuidance, specResourceCrossFieldRules(configVersion, apiVersion, &s))
}

// legacyEnvelopeFieldRules builds the FieldRules for the legacy flat
// resource envelope: requiredCpu / limitedCpu / requiredMemory /
// limitedMemory / requiredDisk are all required (with quantity validation)
// while limitedDisk stays optional but still quantity-checked when set.
//
// Factored out of validateAppSpec so the legacy branch can install it with
// a single append, and so any future caller (or test) that wants the same
// shape doesn't have to redeclare every Field rule.
func legacyEnvelopeFieldRules(s *AppSpec, templateOnly bool) []*validation.FieldRules {
	return []*validation.FieldRules{
		validation.Field(&s.RequiredMemory,
			validation.Required.Error("spec.requiredMemory is required for olaresManifest.version < 0.12.0"),
			flatResourceQuantityRule("spec.requiredMemory", templateOnly)),
		validation.Field(&s.RequiredDisk,
			validation.Required.Error("spec.requiredDisk is required for olaresManifest.version < 0.12.0"),
			flatResourceQuantityRule("spec.requiredDisk", templateOnly)),
		validation.Field(&s.RequiredCPU,
			validation.Required.Error("spec.requiredCpu is required for olaresManifest.version < 0.12.0"),
			flatResourceQuantityRule("spec.requiredCpu", templateOnly)),
		validation.Field(&s.LimitedMemory,
			validation.Required.Error("spec.limitedMemory is required for olaresManifest.version < 0.12.0"),
			flatResourceQuantityRule("spec.limitedMemory", templateOnly)),
		validation.Field(&s.LimitedCPU,
			validation.Required.Error("spec.limitedCpu is required for olaresManifest.version < 0.12.0"),
			flatResourceQuantityRule("spec.limitedCpu", templateOnly)),
		validation.Field(&s.LimitedDisk,
			validation.When(s.LimitedDisk != "", flatResourceQuantityRule("spec.limitedDisk", templateOnly))),
	}
}

// flatResourceQuantityRule validates a legacy flat spec.* quantity field.
// When templateOnly is true, non-disk fields may use AutoResourceValue ("-1").
func flatResourceQuantityRule(field string, templateOnly bool) validation.Rule {
	return validation.By(func(value interface{}) error {
		s, _ := value.(string)
		return validateResourceQuantity(s, field, templateOnly, false)
	})
}

// validateResourceModeValueFor binds spec so the modern (>= 0.12.0,
// apiVersion != v2) branch can dispatch validation onto whichever shape
// the manifest actually uses, from a single validation.By rule wired onto
// spec.Resources:
//
//  1. spec.resources[] is non-empty: iterate every entry and apply
//     per-element validation (ValidateResourceMode -- mode enum, quantity
//     validity, section completeness, gpu-section gate, limit >= required).
//  2. spec.resources[] is empty AND at least one flat field is set: the
//     manifest opted into the legacy envelope on a modern version, so
//     format-validate every populated spec.required* / spec.limited*
//     flat field via validateFlatResourceQuantities. Required-ness of
//     individual flat fields is deliberately not enforced here -- users
//     who fill out only part of the envelope still get format feedback
//     plus the Rule 5 limit >= required pairing applied to whatever they
//     wrote down.
//  3. spec.resources[] is empty AND no flat field is set: emit a single
//     consolidated guidance message listing both shapes; the manifest
//     must declare at least one.
//
// Rule 7 mutex (spec.resources[] cannot coexist with the eight flat
// fields) is still enforced by ensureLegacyAndResourcesAreMutuallyExclusive
// from specResourceCrossFieldRules, which runs for every version, so this
// closure does not need to repeat it.
func validateResourceModeValueFor(templateOnly bool, spec *AppSpec) validation.RuleFunc {
	return func(v interface{}) error {
		var errs []error
		for i, rm := range spec.Accelerator {
			if err := ValidateResourceMode(rm, templateOnly); err != nil {
				errs = append(errs, fmt.Errorf("spec.resources[%d]: %w", i, err))
			}
		}
		if len(spec.Accelerator) == 0 {
			if err := validateFlatResourceQuantities(spec, templateOnly); err != nil {
				errs = append(errs, err)
			}
			if !hasAnyFlatResourceQuantity(spec) {
				errs = append(errs, fmt.Errorf(
					"either spec.resources[] or the legacy envelope (spec.requiredCpu / spec.limitedCpu / spec.requiredMemory / spec.limitedMemory / spec.requiredDisk) is required for olaresManifest.version >= 0.12.0",
				))
			}
		}

		return errors.Join(errs...)
	}
}

// hasAnyFlatResourceQuantity reports whether any of the eight legacy
// spec.required* / spec.limited* flat fields carries a value. Used by
// the modern-branch closure to decide whether an empty spec.resources[]
// means "the manifest opted into the flat envelope" (some flat field is
// set) or "the manifest forgot to declare a resource envelope at all"
// (nothing is set).
func hasAnyFlatResourceQuantity(spec *AppSpec) bool {
	return spec.RequiredCPU != "" ||
		spec.LimitedCPU != "" ||
		spec.RequiredMemory != "" ||
		spec.LimitedMemory != "" ||
		spec.RequiredDisk != "" ||
		spec.LimitedDisk != ""
}

// validateFlatResourceQuantities reports each populated
// spec.required* / spec.limited* flat field whose value does not parse
// as a Kubernetes quantity. Empty fields are skipped: this helper only
// answers "is the value I wrote down legal?", not "did I write down
// enough fields?". Required-ness is a separate concern owned by the
// caller (legacyEnvelopeFieldRules for the legacy branch; the modern
// branch currently accepts any subset since the user can also opt into
// spec.resources[] instead).
//
// The eight fields covered match the legacy envelope plus the two GPU
// quantities, so a manifest that mixes any of them will get format
// feedback before Rule 7 weighs in on whether they were allowed to be
// set at all.
func validateFlatResourceQuantities(spec *AppSpec, templateOnly bool) error {
	pairs := []struct {
		name  string
		value string
	}{
		{"spec.requiredCpu", spec.RequiredCPU},
		{"spec.limitedCpu", spec.LimitedCPU},
		{"spec.requiredMemory", spec.RequiredMemory},
		{"spec.limitedMemory", spec.LimitedMemory},
		{"spec.requiredDisk", spec.RequiredDisk},
		{"spec.limitedDisk", spec.LimitedDisk},
		{"spec.requiredGpu", spec.RequiredGPU},
		{"spec.limitedGpu", spec.LimitedGPU},
	}
	var errs []error
	for _, p := range pairs {
		if err := validateResourceQuantity(p.value, p.name, templateOnly, false); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// validatePermission gates the manifest-level permission flags whose
// acceptance changes across olaresManifest.version boundaries:
//
//   - permission.externalData (grants access to the External directory) was
//     introduced with olaresManifest.version 0.12.0; declaring it on an
//     older manifest is rejected.
//   - permission.provider (cross-app provider access) was retired starting
//     with olaresManifest.version 0.12.0; the field is still accepted
//     structurally for backwards compatibility on legacy manifests, but a
//     modern manifest must leave it empty so the platform can finish
//     migrating callers away from it.
//
// permission.appCommon (access to the Common directory) is accepted at
// every version and needs no extra validation here.
func validatePermission(configVersion string, p Permission) error {
	var errs []error
	if p.ExternalData && !resourcesCheckApplies(configVersion) {
		errs = append(errs, fmt.Errorf(
			"permission.externalData is only supported for olaresManifest.version >= %s",
			minResourcesManifestVersion,
		))
	}
	if len(p.Provider) > 0 && resourcesCheckApplies(configVersion) {
		errs = append(errs, fmt.Errorf(
			"permission.provider must be empty for olaresManifest.version >= %s; cross-app provider access is no longer granted via permission.provider",
			minResourcesManifestVersion,
		))
	}
	return errors.Join(errs...)
}

// olaresSystemDepName is the canonical name of the Olares system entry
// inside options.dependencies. validateOlaresDependency uses this exact
// (name, type) tuple — name="olares", type="system" — to locate the entry
// whose version constraint is gated by the rules below.
const olaresSystemDepName = "olares"

// olaresDepConstraintRule describes the version-constraint range that the
// options.dependencies[name=olares].version must stay inside on a modern
// (olaresManifest.version >= 0.12.0) manifest. The check works by sampling
// representative versions just outside the allowed range and asserting
// none of them satisfy the manifest's constraint — anything that does
// would leak the supported Olares range outside the documented window.
type olaresDepConstraintRule struct {
	// requirement is the human-readable constraint expression embedded in
	// error messages, e.g. ">=1.12.3-0,<1.12.6".
	requirement string
	// forbidden lists versions that are JUST outside the required range
	// (one or two below the floor, one at/above the ceiling, plus a wide-
	// margin sample). A constraint that allows any of these would be too
	// permissive for the active rule.
	forbidden []string
}

// olaresDepRulePreV3 is the constraint window for legacy-schema modern
// manifests (apiVersion in {empty, v1, v2} AND no 1.12.6-only feature
// is declared). The floor 1.12.3-0 is the minimum Olares that ships
// olaresManifest.version 0.12.0; the ceiling 1.12.6 locks these manifests
// out of systems that ship the v3 runtime.
var olaresDepRulePreV3 = olaresDepConstraintRule{
	requirement: ">=1.12.3-0,<1.12.6",
	forbidden:   []string{"1.12.2", "1.12.6", "2.0.0"},
}

// olaresDepRulePostV3 is the constraint window selected when the manifest
// either is apiVersion=v3 or declares any of the 1.12.6-only feature
// fields enumerated in detectOlares1126OnlyFields. 1.12.6-0 is the first
// Olares release that ships those features.
var olaresDepRulePostV3 = olaresDepConstraintRule{
	requirement: ">=1.12.6-0",
	forbidden:   []string{"1.12.5", "1.12.0", "0.1.0"},
}

// detectOlares1126OnlyFields returns the set of manifest field labels
// that — when declared on c — pin the Olares system requirement to
// >=1.12.6-0 regardless of apiVersion (rule 4). apiVersion=v3 itself is
// intentionally not in this list: it is handled as a separate trigger by
// the caller so the error message can distinguish "v3 manifest demands
// 1.12.6+" from "v1/v2 manifest opts into a 1.12.6-only feature".
//
// Labels are the user-facing names the manifest author writes, so they
// can be threaded straight into error messages.
func detectOlares1126OnlyFields(c *AppConfiguration) []string {
	var fields []string
	if len(c.Spec.Accelerator) > 0 {
		fields = append(fields, "spec.accelerator")
	}
	if c.WorkloadReplicas != nil && len(*c.WorkloadReplicas) > 0 {
		fields = append(fields, "workloadReplicas")
	}
	if c.OverlayGateway.Enable || len(c.OverlayGateway.Entrances) > 0 {
		fields = append(fields, "overlayGateway")
	}
	if c.Options.LLMGatewaySupported {
		fields = append(fields, "options.LLMGatewaySupported")
	}
	if c.Options.TemplateOnly {
		fields = append(fields, "options.templateOnly")
	}
	if c.Options.Shared {
		fields = append(fields, "options.shared")
	}
	if c.Permission.AppCommon {
		fields = append(fields, "permission.appCommon")
	}
	return fields
}

// pickOlaresDepRule selects the constraint window that applies to c.
// apiVersion=v3 or any 1.12.6-only feature field forces the post-v3
// (>=1.12.6-0) window; everything else falls back to the legacy
// pre-v3 (>=1.12.3-0,<1.12.6) window. The returned trigger list is
// non-empty iff a feature field promoted a v1/v2 manifest into the
// post-v3 window; callers use it to explain the promotion in error
// messages without surprising users whose manifests are already v3.
func pickOlaresDepRule(c *AppConfiguration) (rule olaresDepConstraintRule, featureTriggers []string) {
	api := normalizeAPIVersion(c.APIVersion)
	if api == APIVersionV3 {
		return olaresDepRulePostV3, nil
	}
	triggers := detectOlares1126OnlyFields(c)
	if len(triggers) > 0 {
		return olaresDepRulePostV3, triggers
	}
	return olaresDepRulePreV3, nil
}

// validateOlaresDependency enforces the rules on the Olares system
// dependency (the options.dependencies entry with name="olares" and
// type="system"). The gate fires only on modern manifests
// (olaresManifest.version >= 0.12.0); legacy manifests retain their
// existing freedom to declare any version constraint they like.
//
// Four rules combine into a per-manifest required range, captured by
// pickOlaresDepRule:
//
//  1. apiVersion in {empty, v1, v2}: legacy schemas default to <1.12.6
//     (must not install on systems that ship the v3 runtime).
//  2. apiVersion=v3: the constraint must restrict Olares to >=1.12.6-0
//     (the release that introduced the v3 runtime).
//  3. Across every apiVersion the constraint's lower bound must be at
//     least 1.12.3-0, the floor under which modern manifests are not
//     supported at all.
//  4. If the manifest declares any 1.12.6-only feature field
//     (spec.accelerator, workloadReplicas, overlayGateway,
//     options.LLMGatewaySupported, options.templateOnly, options.shared,
//     permission.appCommon), the constraint must restrict to >=1.12.6-0
//     even on a v1/v2 manifest — those fields are not honoured by older
//     Olares releases.
//
// A missing Olares system dependency is itself an error on modern
// manifests — the platform always pins the host Olares version, so
// omitting it makes the manifest non-portable.
func validateOlaresDependency(c *AppConfiguration) error {
	if !resourcesCheckApplies(c.ConfigVersion) {
		return nil
	}
	var olaresDep *Dependency
	for i := range c.Options.Dependencies {
		d := &c.Options.Dependencies[i]
		if d.Name == olaresSystemDepName && d.Type == "system" {
			olaresDep = d
			break
		}
	}
	if olaresDep == nil {
		return fmt.Errorf(
			"options.dependencies must declare an entry with name=%q and type=\"system\" for olaresManifest.version >= %s",
			olaresSystemDepName, minResourcesManifestVersion,
		)
	}
	constraint, err := semver.NewConstraint(olaresDep.Version)
	if err != nil {
		return fmt.Errorf(
			"options.dependencies[name=%s].version %q is not a valid semver constraint: %w",
			olaresSystemDepName, olaresDep.Version, err,
		)
	}
	api := normalizeAPIVersion(c.APIVersion)
	if _, knownAPI := map[string]struct{}{
		APIVersionV1: {}, APIVersionV2: {}, APIVersionV3: {},
	}[api]; !knownAPI {
		// apiVersion outside the supported set; the struct-level validator
		// already surfaces "not supported version" elsewhere.
		return nil
	}
	rule, triggers := pickOlaresDepRule(c)
	promotion := ""
	if len(triggers) > 0 {
		promotion = fmt.Sprintf(
			" because the manifest declares %s, which requires Olares 1.12.6+",
			strings.Join(triggers, " / "),
		)
	}
	var errs []error
	for _, v := range rule.forbidden {
		sv, parseErr := semver.NewVersion(v)
		if parseErr != nil {
			continue
		}
		if constraint.Check(sv) {
			errs = append(errs, fmt.Errorf(
				"options.dependencies[name=%s].version %q must restrict the Olares system version to %q for apiVersion=%s%s; the constraint currently allows %s, which is outside the supported range",
				olaresSystemDepName, olaresDep.Version, rule.requirement, api, promotion, sv.String(),
			))
		}
	}
	return errors.Join(errs...)
}

// validateSharedAppRequirements enforces the cross-field rules that an
// options.shared=true manifest must satisfy. A shared install services
// every user on the cluster from a single deployment, which is only
// compatible with a narrow shape:
//
//   - spec.onlyAdmin must be true — shared apps span every user, so the
//     installer must be restricted to admins; allowing a regular user to
//     install one would let them install on behalf of everyone else.
//   - spec.subCharts must be empty — subCharts is the v2-only multi-chart
//     delivery mechanism. Shared apps require apiVersion=v3 (enforced in
//     validateOptions), and v3 ships every workload in a single chart;
//     declaring subCharts here is meaningless and almost certainly a
//     manifest authored against the wrong schema.
//   - options.appScope.clusterScoped must be false AND
//     options.appScope.appRef must be empty — both fields express a
//     scoping intent (cluster-wide singleton, or shared access with
//     specific other apps) that overlaps with "shared". Declaring them
//     together makes the install topology ambiguous. The two are
//     checked independently so the error message pinpoints the offender
//     instead of forcing the user to clear the whole appScope block to
//     discover which field tripped the rule.
//
// The apiVersion=v3 requirement on shared apps is enforced separately in
// validateOptions; these gates run alongside it so a manifest that
// violates multiple constraints surfaces every offender in a single Lint
// run.
func validateSharedAppRequirements(c *AppConfiguration) error {
	if !c.Options.Shared {
		return nil
	}
	var errs []error
	if !c.Spec.OnlyAdmin {
		errs = append(errs, fmt.Errorf(
			"spec.onlyAdmin must be true when options.shared=true; shared apps service every user on the cluster and may only be installed by an admin",
		))
	}
	if len(c.Spec.SubCharts) > 0 {
		errs = append(errs, fmt.Errorf(
			"spec.subCharts must be empty when options.shared=true; shared apps require apiVersion=v3 and deliver every workload in a single chart",
		))
	}
	if c.Options.AppScope.ClusterScoped {
		errs = append(errs, fmt.Errorf(
			"options.appScope.clusterScoped must be false when options.shared=true; the shared scope already covers every user and is incompatible with cluster scoping",
		))
	}
	if len(c.Options.AppScope.AppRef) > 0 {
		errs = append(errs, fmt.Errorf(
			"options.appScope.appRef must be empty when options.shared=true; shared apps do not declare cross-app scoping",
		))
	}
	return errors.Join(errs...)
}

// validateRootProvider enforces that the top-level AppConfiguration.Provider
// section (the per-app published interfaces, declared at the document root
// rather than under permission) is empty on olaresManifest.version >= 0.12.0.
// The section was retired alongside permission.provider once the platform
// stopped granting cross-app access through manifest-declared provider lists;
// legacy manifests still accept arbitrary entries for backwards
// compatibility.
func validateRootProvider(configVersion string, providers []Provider) error {
	if !resourcesCheckApplies(configVersion) {
		return nil
	}
	if len(providers) == 0 {
		return nil
	}
	return fmt.Errorf(
		"provider must be empty for olaresManifest.version >= %s; the top-level provider section is no longer accepted on modern manifests",
		minResourcesManifestVersion,
	)
}

// validateWorkloadReplicas enforces that workloadReplicas is declared (non-nil
// and non-empty) on modern manifests (olaresManifest.version >= 0.12.0). The
// rule mirrors the install-time convention that every Deployment/StatefulSet's
// replica count is sourced from .Values.workloads.<name>.replicaCount, so a
// modern app that omits the field would have no way to express its replica
// envelope.
//
// apiVersion=v2 is exempt: v2 manifests render workloads inside subCharts, so
// the parent-level workloadReplicas does not apply. Below the 0.12.0 gate the
// field stays optional. Whenever the field is declared, the existing
// chart-render lint phase still verifies the per-entry name/values.yaml
// correspondence — this check only adds the "must be declared" gate.
func validateWorkloadReplicas(configVersion, apiVersion string, wr *WorkloadReplicas) error {
	if !resourcesCheckApplies(configVersion) {
		return nil
	}
	if normalizeAPIVersion(apiVersion) == APIVersionV2 {
		return nil
	}
	if wr == nil || len(*wr) == 0 {
		return fmt.Errorf(
			"workloadReplicas is required for olaresManifest.version >= %s; declare a workloadReplicas.<workload>: <count> entry for every Deployment/StatefulSet",
			minResourcesManifestVersion,
		)
	}
	return nil
}

// validateOverlayGateway runs structural checks on every overlayGateway
// entrance. The workload reference is required here but its existence against
// the rendered Deployment/StatefulSet set is verified separately during the
// chart-render lint phase (the manifest validator has no rendered workloads
// to compare against). Entrances are validated whenever they are declared,
// regardless of overlayGateway.enable, so a disabled-but-declared gateway is
// still well-formed.
func validateOverlayGateway(v interface{}) error {
	og, ok := v.(OverlayGateway)
	if !ok {
		return fmt.Errorf("overlayGateway: unexpected type %T", v)
	}
	var errs []error
	for i, e := range og.Entrances {
		if err := validation.ValidateStruct(&e,
			validation.Field(&e.Title,
				validation.Required.Error("overlayGateway.entrances.title is required"),
				validation.Length(1, 30).Error("overlayGateway.entrances.title must be 1-30 characters"),
			),
			validation.Field(&e.Port,
				validation.Required.Error("overlayGateway.entrances.port is required"),
				validation.Min(int32(1)).Error("overlayGateway.entrances.port must be > 0"),
			),
			validation.Field(&e.Workload,
				validation.Required.Error("overlayGateway.entrances.workload is required"),
			),
			validation.Field(&e.Protocol,
				validation.In(validOverlayProtocols...).Error(`overlayGateway.entrances.protocol must be one of "", "tcp", "udp"`),
			),
		); err != nil {
			errs = append(errs, fmt.Errorf("overlayGateway.entrances[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

// validateOptionsFor binds the manifest's apiVersion so options-level
// cross-field checks (templateOnly => allowMultipleInstall, shared => v3)
// can reach it. ozzo's validation.By only sees the Options value, so the
// outer AppConfiguration must be captured here in a closure.
func validateOptionsFor(cfg *AppConfiguration) validation.RuleFunc {
	return func(v interface{}) error {
		o, ok := v.(Options)
		if !ok {
			return fmt.Errorf("options: unexpected type %T", v)
		}
		return validateOptions(cfg.APIVersion, o)
	}
}

// validateOptions runs the options-level validation. It is parameterised on
// apiVersion because two cross-field rules depend on it:
//
//   - options.templateOnly=true requires options.allowMultipleInstall=true.
//     Template-only apps are installed as multiple clones from a single
//     template chart; without allowMultipleInstall the platform would only
//     ever install a single instance, defeating the purpose of the flag.
//   - options.shared=true is only meaningful on apiVersion=v3. v3 is the
//     only schema where a single install services multiple users, so a
//     shared install on v1/v2 would be silently ignored at install time
//     and is rejected here to avoid the foot-gun.
//
// Both cross-field errors are aggregated alongside the existing per-field
// struct validation so a manifest that violates more than one rule sees
// every offender in a single Lint run.
func validateOptions(apiVersion string, o Options) error {
	structErr := validation.ValidateStruct(&o,
		validation.Field(&o.Policies, validation.Each(validation.By(validatePolicyValue))),
		validation.Field(&o.ResetCookie),
		validation.Field(&o.Dependencies, validation.Each(validation.By(validateDependencyValue))),
		validation.Field(&o.AppScope),
		validation.Field(&o.WsConfig),
	)
	var crossErrs []error
	if o.TemplateOnly && !o.AllowMultipleInstall {
		crossErrs = append(crossErrs, fmt.Errorf(
			"options.allowMultipleInstall must be true when options.templateOnly is true; template-only apps are installed as multiple clones",
		))
	}
	if o.Shared && normalizeAPIVersion(apiVersion) != APIVersionV3 {
		crossErrs = append(crossErrs, fmt.Errorf(
			"options.shared=true is only supported for apiVersion=v3 (got %q)",
			apiVersion,
		))
	}
	return errors.Join(structErr, errors.Join(crossErrs...))
}

func validatePolicyValue(v interface{}) error {
	p, ok := v.(Policy)
	if !ok {
		return fmt.Errorf("policy: unexpected type %T", v)
	}
	return validation.ValidateStruct(&p,
		validation.Field(&p.URIRegex, validation.Required.Error("policy.uriRegex is required")),
		validation.Field(&p.Level, validation.Required.Error("policy.level is required")),
		validation.Field(&p.Duration,
			validation.When(p.Duration != "",
				validation.Match(validDurationRegex).Error("policy.validDuration is malformed"),
			),
		),
	)
}

func validateDependencyValue(v interface{}) error {
	d, ok := v.(Dependency)
	if !ok {
		return fmt.Errorf("dependency: unexpected type %T", v)
	}
	return validation.ValidateStruct(&d,
		validation.Field(&d.Name, validation.Required.Error("dependency.name is required")),
		validation.Field(&d.Version, validation.Required.Error("dependency.version is required")),
		validation.Field(&d.Type,
			validation.Required.Error("dependency.type is required"),
			validation.In(validDependencyTypes...).Error("dependency.type must be system, application or middleware"),
		),
	)
}

func checkSubCharts(cfg *AppConfiguration) error {
	if cfg.APIVersion != APIVersionV2 {
		return nil
	}
	if len(cfg.Spec.SubCharts) == 0 {
		return fmt.Errorf("spec.subCharts is required for apiVersion=v2")
	}
	hasSharedChart := false
	for _, c := range cfg.Spec.SubCharts {
		err := isSafeSubChartName(c.Name)
		if err != nil {
			return err
		}
		if c.Shared {
			hasSharedChart = true
		}
	}
	if hasSharedChart {
		return nil
	}
	return fmt.Errorf("spec.subCharts must contain at least one entry with shared=true")
}

func uniqueEntranceNames(value interface{}) error {
	entrances, ok := value.([]appv1.Entrance)
	if !ok {
		return fmt.Errorf("entrances: unexpected type %T", value)
	}
	seen := make(map[string]struct{}, len(entrances))
	for i, e := range entrances {
		if _, dup := seen[e.Name]; dup {
			return fmt.Errorf("entrances[%d].name: duplicate entrance name %q", i, e.Name)
		}
		seen[e.Name] = struct{}{}
	}
	return nil
}

// validateSupportArchEntry enforces that each spec.supportArch element is
// one of the two architectures the downstream installer accepts. Empty
// strings, case variants ("AMD64"), and other CPU families are rejected
// here so the manifest fails fast instead of silently misbehaving at
// install time.
func validateSupportArchEntry(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("spec.supportArch: unexpected type %T", value)
	}
	if _, allowed := validSupportArchSet[s]; allowed {
		return nil
	}
	return fmt.Errorf(`spec.supportArch entries must be "amd64" or "arm64" (got %q)`, s)
}

// uniqueSupportArches reports duplicate entries in spec.supportArch. The
// per-element "amd64"/"arm64" enum is enforced separately by
// validation.Each; this rule guards against the same architecture being
// listed twice, which is silently ignored downstream but almost always a
// manifest authoring mistake.
func uniqueSupportArches(value interface{}) error {
	arches, ok := value.([]string)
	if !ok {
		return fmt.Errorf("spec.supportArch: unexpected type %T", value)
	}
	seen := make(map[string]struct{}, len(arches))
	for i, a := range arches {
		if _, dup := seen[a]; dup {
			return fmt.Errorf("spec.supportArch[%d]: duplicate value %q", i, a)
		}
		seen[a] = struct{}{}
	}
	return nil
}

func isSafeSubChartName(name string) error {
	if name == "" {
		return errMissingSubChartName
	}
	if len(name) > 53 || !validChartName.MatchString(name) {
		return errInvalidSubChartName
	}
	return nil
}
