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
)

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
		return fmt.Errorf("不支持该版本")
	}
}

// ValidateAppConfiguration runs structural and cross-field checks on the manifest.
func ValidateAppConfiguration(c *AppConfiguration) error {
	structErr := validation.ValidateStruct(c,
		validation.Field(&c.ConfigVersion,
			validation.Required.Error("olaresManifest.version is required")),
		validation.Field(&c.APIVersion,
			validation.When(c.APIVersion != "",
				validation.In(validAPIVersions...).Error("不支持该版本"),
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
		validation.Field(&c.Options, validation.By(validateOptions)),
	)
	return errors.Join(structErr, checkSubCharts(c))
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
	)
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
		return validateAppSpec(cfg.ConfigVersion, cfg.APIVersion, s)
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
func validateAppSpec(configVersion, apiVersion string, s AppSpec) error {
	quantityRule := validation.Match(k8sQuantity).Error("must be a valid Kubernetes quantity")
	optionalGPUQuantity := validation.When(s.RequiredGPU != "", quantityRule)
	optionalLimitedGPUQuantity := validation.When(s.LimitedGPU != "", quantityRule)
	optionalLimitedDiskQuantity := validation.When(s.LimitedDisk != "", quantityRule)

	fields := []*validation.FieldRules{
		validation.Field(&s.RequiredGPU, optionalGPUQuantity),
		validation.Field(&s.LimitedGPU, optionalLimitedGPUQuantity),
	}

	api := normalizeAPIVersion(apiVersion)
	var v2ResourcesErr error
	if api == APIVersionV2 && len(s.Resources) > 0 {
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
			validation.Field(&s.Resources,
				validation.Required.Error(
					"spec.resources is required for olaresManifest.version >= 0.12.0; declare at least one entry",
				),
				validation.Each(validation.By(validateResourceModeValue)),
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
		fields = append(fields,
			validation.Field(&s.RequiredMemory,
				validation.Required.Error("spec.requiredMemory is required for olaresManifest.version < 0.12.0"),
				quantityRule),
			validation.Field(&s.RequiredDisk,
				validation.Required.Error("spec.requiredDisk is required for olaresManifest.version < 0.12.0"),
				quantityRule),
			validation.Field(&s.RequiredCPU,
				validation.Required.Error("spec.requiredCpu is required for olaresManifest.version < 0.12.0"),
				quantityRule),
			validation.Field(&s.LimitedMemory,
				validation.Required.Error("spec.limitedMemory is required for olaresManifest.version < 0.12.0"),
				quantityRule),
			validation.Field(&s.LimitedCPU,
				validation.Required.Error("spec.limitedCpu is required for olaresManifest.version < 0.12.0"),
				quantityRule),
			validation.Field(&s.LimitedDisk, optionalLimitedDiskQuantity),
		)
	}

	structErr := validation.ValidateStruct(&s, fields...)
	return errors.Join(v2ResourcesErr, supportedGpuModernErr, structErr, versionGuidance, specResourceCrossFieldRules(configVersion, apiVersion, &s))
}

func validateResourceModeValue(v interface{}) error {
	rm, ok := v.(ResourceMode)
	if !ok {
		return fmt.Errorf("resources: unexpected type %T", v)
	}
	return ValidateResourceMode(rm)
}

func validateOptions(v interface{}) error {
	o, ok := v.(Options)
	if !ok {
		return fmt.Errorf("options: unexpected type %T", v)
	}
	return validation.ValidateStruct(&o,
		validation.Field(&o.Policies, validation.Each(validation.By(validatePolicyValue))),
		validation.Field(&o.ResetCookie),
		validation.Field(&o.Dependencies, validation.Each(validation.By(validateDependencyValue))),
		validation.Field(&o.AppScope),
		validation.Field(&o.WsConfig),
	)
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
	for _, c := range cfg.Spec.SubCharts {
		if c.Shared {
			return nil
		}
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
