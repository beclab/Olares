// Package oachecker lints an OlaresApp (`oac`) chart directory. It
// validates the OlaresManifest.yaml against its declared apiVersion,
// dry-runs the helm chart to inspect the resulting workloads, and produces
// the deduped image list that downstream tooling needs.
//
// A typical caller uses one of the top-level convenience functions:
//
//	oac.LintChart("./myapp", oac.WithOwnerAdmin("alice"))
//	images, err := oac.ListImagesFromOAC("./myapp")
//
// For more control, construct a Checker with New and call its methods
// directly.
package oac

// OAC is a reusable lint context. All fields are private; build one via
// New(opts...) and the With*/Skip* option helpers.
type OAC struct {
	owner string
	admin string

	// autoOwner, when true, makes Lint ignore the owner/admin fields set
	// above and instead run the chart-rendering checks twice — once with
	// owner == admin (cluster-admin install) and once with owner != admin
	// (regular user install). Both scenarios must pass. Turn it on with
	// WithAutoOwnerScenarios().
	autoOwner bool

	skipManifest    bool
	skipResource    bool
	skipFolder      bool
	skipSameVersion bool
	skipAppData     bool
	runRBAC         bool

	customValidators []CustomValidator

	// extraValues, when non-empty, is deep-merged on top of the values
	// produced by helmrender.BuildValues for every helm dry-run the
	// Checker performs (Lint, ListImages, CheckResources, etc.). External
	// keys win on conflicts; see WithValues.
	extraValues map[string]interface{}
}

// New builds a Checker with the given options applied on top of the default
// configuration:
//   - no owner/admin override (templates fall back to the "default" placeholder)
//   - owner/admin scenarios are NOT auto-expanded — the explicit owner/admin
//     values are used as-is (see WithAutoOwnerScenarios)
//   - all built-in checks enabled EXCEPT ServiceAccount rule inspection
//     (RBAC is off by default; the Chart.yaml <-> manifest same-version
//     check runs by default — turn it off with SkipSameVersionCheck;
//     the .Values.userspace.appdata template-vs-manifest cross-check runs
//     by default — turn it off with SkipAppDataCheck)
func New(opts ...Option) *OAC {
	c := &OAC{}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// Owner returns the configured .Values.bfl.username value, or "" when unset.
func (c *OAC) Owner() string { return c.owner }

// Admin returns the configured .Values.admin value, or "" when unset.
func (c *OAC) Admin() string { return c.admin }
