package oac

import (
	"github.com/beclab/Olares/framework/oac/internal/helmrender"
)

// Option mutates a Checker built via New. Options are idempotent and safe to
// apply in any order.
type Option func(*OAC)

// WithValues registers extra helm values that the Checker deep-merges on
// top of the scaffold produced by helmrender.BuildValues for every render
// it performs (Lint, ListImages / ListImagesForMode, CheckResources,
// CheckServiceAccountRules, ...). External keys win on conflicts: scalar
// keys are replaced wholesale, and when both sides are maps the merge
// recurses so siblings the caller did not override are preserved.
//
// Multiple WithValues calls are additive -- each is merged into the
// already-accumulated extra-values map under the same precedence rules.
// Passing nil is a no-op.
//
// The mode argument of ListImagesForMode and the per-mode loop in
// resource-limit checks always set .Values.GPU.Type AFTER WithValues is
// applied, so they keep winning over any GPU.Type the caller injected.
func WithValues(extra map[string]interface{}) Option {
	return func(c *OAC) {
		if len(extra) == 0 {
			return
		}
		if c.extraValues == nil {
			c.extraValues = map[string]interface{}{}
		}
		helmrender.MergeValues(c.extraValues, extra)
	}
}

// WithOwner sets the .Values.bfl.username template value and the owner field
// used when rendering helm charts. When owner is empty the Checker keeps its
// existing value.
func WithOwner(owner string) Option {
	return func(c *OAC) {
		if owner != "" {
			c.owner = owner
		}
	}
}

// WithAdmin sets the .Values.admin template value. An empty admin is ignored.
func WithAdmin(admin string) Option {
	return func(c *OAC) {
		if admin != "" {
			c.admin = admin
		}
	}
}

// WithOwnerAdmin sets both owner and admin to the same value, modelling the
// "installed as admin" scenario where the cluster administrator is also the
// acting user.
func WithOwnerAdmin(value string) Option {
	return func(c *OAC) {
		if value != "" {
			c.owner = value
			c.admin = value
		}
	}
}

// SkipManifestCheck disables OlaresManifest.yaml structural validation.
func SkipManifestCheck() Option {
	return func(c *OAC) { c.skipManifest = true }
}

// SkipResourceCheck disables the container-level resource-limits check.
//
// Note: this option does NOT disable the upload-mount and workload-naming
// checks, which Lint always runs because they guard structural integrity
// (a chart that declares options.upload.dest but mounts it nowhere, or an
// app whose templates produce no Deployment/StatefulSet named after the
// app, is broken regardless of limit accounting).
func SkipResourceCheck() Option {
	return func(c *OAC) { c.skipResource = true }
}

// SkipFolderCheck disables the chart-folder layout check.
func SkipFolderCheck() Option {
	return func(c *OAC) { c.skipFolder = true }
}

// SkipSameVersionCheck disables the Chart.yaml <-> manifest version
// consistency check. By default the check runs; callers that roll their own
// version-alignment step can opt out here.
func SkipSameVersionCheck() Option {
	return func(c *OAC) { c.skipSameVersion = true }
}

// WithSameVersionCheck re-enables the Chart.yaml <-> manifest version
// consistency check. Mostly useful when composing an option set that had
// SkipSameVersionCheck baked in and a particular call-site wants it back on.
func WithSameVersionCheck() Option {
	return func(c *OAC) { c.skipSameVersion = false }
}

// WithServiceAccountRulesCheck enables the RBAC rule inspection which makes
// sure the chart doesn't grant ServiceAccounts forbidden permissions. It is
// disabled by default to match historical Lint behaviour; callers that need
// it can opt in explicitly.
func WithServiceAccountRulesCheck() Option {
	return func(c *OAC) { c.runRBAC = true }
}

// WithAutoOwnerScenarios makes Lint ignore any explicit WithOwner / WithAdmin
// / WithOwnerAdmin values and instead run the chart-rendering portion of the
// pipeline twice:
//
//  1. owner == admin (cluster-admin install)
//  2. owner != admin (regular user install)
//
// Both scenarios must pass. The manifest-level checks (folder layout, ozzo
// validation, custom validators, same-version) are owner-independent and
// still only run once.
//
// This is the programmatic equivalent of the LintBothOwnerScenarios helper —
// use it whenever the caller does not have a concrete owner/admin pair and
// wants the linter to cover both install modes automatically.
func WithAutoOwnerScenarios() Option {
	return func(c *OAC) { c.autoOwner = true }
}

// WithoutAutoOwnerScenarios clears the auto-owner flag, pinning Lint back to
// the explicit owner/admin values. Mostly useful when composing option sets
// that have WithAutoOwnerScenarios baked in and a particular call-site wants
// to opt out.
func WithoutAutoOwnerScenarios() Option {
	return func(c *OAC) { c.autoOwner = false }
}

// CustomValidator is invoked with the chart directory path and the parsed
// Manifest after the built-in structural checks have run.
type CustomValidator func(oacPath string, m Manifest) error

// WithCustomValidator adds a user-defined validator to the Checker.
func WithCustomValidator(fn CustomValidator) Option {
	return func(c *OAC) {
		if fn != nil {
			c.customValidators = append(c.customValidators, fn)
		}
	}
}

// SkipAppDataCheck disables the built-in template-vs-manifest cross-check
// that scans chart templates for .Values.userspace.appdata references and
// requires permission.appData in OlaresManifest.yaml when any are found.
// The check is enabled by default; only opt out when a caller knowingly
// renders appdata via a non-standard path.
func SkipAppDataCheck() Option {
	return func(c *OAC) { c.skipAppData = true }
}

// WithAppDataValidator re-enables the built-in
// .Values.userspace.appdata cross-check after a previous option set
// (re)disabled it. The check is on by default since it is essentially a
// safety net against permission misconfiguration, so calling this on a
// fresh Checker is a no-op. Kept as a named option for backward
// compatibility — old call sites that used it to "register" the
// validator continue to compile and behave as before, modulo the fact
// that the check is no longer wired through customValidators (so it
// runs exactly once even when this option is passed multiple times).
func WithAppDataValidator() Option {
	return func(c *OAC) { c.skipAppData = false }
}
