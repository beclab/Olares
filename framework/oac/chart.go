package oac

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/kube"
	"sigs.k8s.io/yaml"

	"github.com/beclab/Olares/framework/oac/internal/chartfolder"
	"github.com/beclab/Olares/framework/oac/internal/helmrender"
	"github.com/beclab/Olares/framework/oac/internal/manifest"
	"github.com/beclab/Olares/framework/oac/internal/resources"
)

// Lint runs the full lint pipeline against oacPath. The exact set of checks
// executed depends on the Skip* options set on the Checker:
//
//  1. Folder layout (chartfolder.CheckLayout) - skipped by SkipFolderCheck
//  2. Manifest parse + ozzo validation - skipped by SkipManifestCheck
//  3. Built-in chart template cross-checks (permission vs appdata/appCommon/
//     sharedlib; apiVersion=v3 ban on OLARES_USER_* in templates/values.yaml)
//     - permission checks skipped by SkipAppDataCheck (on by default)
//  4. Custom validators registered via WithCustomValidator (none by default)
//  5. Helm dry-run and mandatory workload-integrity checks (upload mount
//     path, `type=app` workload naming, workloadReplicas <-> rendered
//     workload exact match plus values.yaml replicaCount coverage, and
//     overlayGateway entrance workload existence) - ALWAYS run; not
//     governed by any Skip* option
//  6. HostPath + rolling-update incompatibility check - ON by default,
//     turn off with SkipHostPathCheck()
//  7. Rendered-resource namespace check (workloads in app-namespace;
//     other resources in app-namespace or user-system-*) - ON by default,
//     turn off with SkipResourceNamespaceCheck()
//     7b. allowMultipleInstall cluster-scoped fixed-name check (v1/v3 only;
//     skipped for v2) - ALWAYS run when options.allowMultipleInstall is true
//  8. Container-level resource limits check - skipped by SkipResourceCheck
//  9. Chart.yaml <-> manifest same-version check - ON by default, turn off
//     with SkipSameVersionCheck()
//  10. ServiceAccount RBAC inspection - ON by default, turn off with
//     SkipServiceAccountRulesCheck()
//  11. Non-beclab image privileged securityContext check - ON by default,
//     turn off with SkipSecurityContextCheck()
//
// When WithAutoOwnerScenarios() is set, every owner-dependent step runs
// twice — once with owner == admin and once with owner != admin. That
// covers the rendered-chart steps (5/6/7/8/10) AND step 2's manifest
// validation, so manifests that branch on
// `eq .Values.admin .Values.bfl.username` are exercised in both install
// modes. Owner-independent steps (folder layout, appdata cross-check,
// same-version) still run once.
func (c *OAC) Lint(oacPath string) error {
	if !c.skipFolder {
		if err := c.CheckChartFolder(oacPath); err != nil {
			return err
		}
	}

	raw, err := readManifestFile(oacPath)
	if err != nil {
		return err
	}
	m, err := c.parseManifest(raw)
	if err != nil {
		return err
	}

	if !c.skipManifest {
		if err := c.validateManifestBytes(raw, m); err != nil {
			return err
		}
	}

	if err := c.checkChartTemplateRules(oacPath, m); err != nil {
		return err
	}

	for _, v := range c.customValidators {
		if err := v(oacPath, m); err != nil {
			return err
		}
	}
	if m.APIVersion() == "v2" {
		for _, sc := range c.ownerScenarios() {
			if err := c.lintRenderedScenario(oacPath, m, sc); err != nil {
				if sc.label != "" {
					return fmt.Errorf("%s scenario: %w", sc.label, err)
				}
				return err
			}
		}
	} else {
		if err := c.lintRenderedScenario(oacPath, m, ownerScenario{label: "", owner: "owner", admin: "admin"}); err != nil {
			return err
		}
	}

	if !c.skipSameVersion {
		if err := c.CheckSameVersion(oacPath, m); err != nil {
			return err
		}
	}
	return nil
}

// ownerScenario captures a single (owner, admin) pair that the chart must
// render cleanly under. The label is surfaced as a prefix on any returned
// error so a user running both scenarios can tell which one tripped.
type ownerScenario struct {
	label string
	owner string
	admin string
}

// ownerScenarios returns the list of (owner, admin) combinations to run the
// rendered portion of Lint against. It collapses to a single entry using the
// Checker's configured owner/admin unless WithAutoOwnerScenarios() asked for
// both install modes.
func (c *OAC) ownerScenarios() []ownerScenario {
	if c.autoOwner {
		return []ownerScenario{
			{label: "owner==admin", owner: "admin", admin: "admin"},
			{label: "owner!=admin", owner: "owner", admin: "admin"},
		}
	}
	return []ownerScenario{{owner: c.owner, admin: c.admin}}
}

// lintRenderedScenario performs the helm-render-dependent part of Lint for a
// single owner scenario: render, mandatory structural checks, and the
// optional resource/RBAC inspections. Manifest validation and folder layout
// are owner-independent and live on the outer Lint call.
func (c *OAC) lintRenderedScenario(oacPath string, m Manifest, sc ownerScenario) error {
	list, err := c.renderForStructuralChecks(oacPath, m, sc)
	if err != nil {
		return fmt.Errorf("helm render: %w", err)
	}

	// Mandatory workload-integrity checks — not gated by SkipResourceCheck.
	if err := resources.CheckUploadConfig(list, extractUploadDest(m)); err != nil {
		return err
	}
	if err := resources.CheckDeploymentName(list, m.ConfigType(), m.AppName()); err != nil {
		return err
	}

	if err := c.checkManifestWorkloadRefs(oacPath, m, list); err != nil {
		return err
	}

	if err := c.checkAllowMultipleInstallClusterScoped(oacPath, m, sc); err != nil {
		return err
	}

	if !c.skipHostPath {
		if err := resources.CheckHostPath(list); err != nil {
			return err
		}
	}

	if !c.skipResourceNamespace {
		if err := resources.CheckResourceNamespace(list); err != nil {
			return err
		}
	}

	if !c.skipResource {
		if err := c.checkResourceLimits(oacPath, m, sc, list); err != nil {
			return err
		}
	}

	if !c.skipRunRBAC {
		rules, err := resources.LoadForbiddenRules("")
		if err != nil {
			return err
		}
		if err := resources.CheckServiceAccountRules(list, rules); err != nil {
			return err
		}
	}

	if !c.skipRunSecurityContext {
		if err := resources.CheckSecurityContextForNonBeclabImage(list); err != nil {
			return err
		}
	}
	return nil
}

// renderForStructuralChecks produces the kube.ResourceList that drives the
// upload-mount and workload-naming checks. It is the no-mode shortcut for
// renderForMode -- structural checks are GPU-type independent.
func (c *OAC) renderForStructuralChecks(oacPath string, m Manifest, sc ownerScenario) (kube.ResourceList, error) {
	return c.renderForMode(oacPath, m, sc, "")
}

// renderForMode helm-renders oacPath under the given owner scenario,
// optionally injecting .Values.GPU.Type=mode so chart templates that
// branch per GPU family produce the matching workload set.
//
// v1/v3 manifests render the chart at oacPath as a single chart;
// v2 manifests follow the multi-chart install layout -- each entry in
// spec.subCharts is helm-rendered under its own subdirectory and the
// lists are concatenated, mirroring production app-service v2 install.
//
// Passing an empty mode skips the GPU.Type override entirely, which
// matches the historical "no mode" rendering used by structural checks
// and the default ListImages call. Values registered via WithValues are
// always deep-merged in by buildRenderValues; SetGPUType for a non-empty
// mode runs AFTER the merge so the per-mode rule still wins on GPU.Type.
func (c *OAC) renderForMode(oacPath string, m Manifest, sc ownerScenario, mode string) (kube.ResourceList, error) {
	if isV2Manifest(m) {
		cfg, ok := m.Raw().(*manifest.AppConfiguration)
		if !ok {
			return nil, fmt.Errorf("oac: cannot render v2 manifest, Raw() is not *AppConfiguration (got %T)", m.Raw())
		}
		return c.renderAllSubCharts(oacPath, m, sc, cfg.Spec.SubCharts, mode)
	}
	values := c.buildRenderValues(m, sc)
	if mode != "" {
		helmrender.SetGPUType(values, mode)
	}
	return helmrender.Render(oacPath, values, m.AppName())
}

// buildRenderValues returns the helm values map that all of the Checker's
// render paths share. It starts from helmrender.BuildValues for the active
// owner scenario and entrances, then deep-merges any values registered via
// WithValues on top -- external keys winning, scalar keys replaced
// wholesale, map keys recursed into so siblings the caller did not
// override survive (see helmrender.MergeValues).
//
// A fresh map is returned every call so callers can safely mutate the
// result (SetGPUType, ad-hoc overrides) without affecting other renders
// or the OAC's stored extraValues.
func (c *OAC) buildRenderValues(m Manifest, sc ownerScenario) map[string]interface{} {
	values := helmrender.BuildValues(sc.owner, sc.admin, m.Entrances())
	if len(c.extraValues) > 0 {
		helmrender.MergeValues(values, c.extraValues)
	}
	return values
}

// CheckChartFolder validates that oacPath is a structurally-valid chart
// directory (Chart.yaml/values.yaml/templates/OlaresManifest.yaml present,
// folder name well-formed).
func (c *OAC) CheckChartFolder(oacPath string) error {
	return chartfolder.CheckLayout(oacPath)
}

// CheckSameVersion cross-validates the folder name, Chart.yaml metadata, and
// parsed manifest metadata. Provide nil for m to have it loaded on demand.
func (c *OAC) CheckSameVersion(oacPath string, m Manifest) error {
	chartFile, err := chartfolder.LoadChart(oacPath)
	if err != nil {
		return err
	}
	if m == nil {
		m, err = c.LoadManifestFile(oacPath)
		if err != nil {
			return err
		}
	}
	return chartfolder.CheckConsistency(oacPath, chartFile, m)
}

// CheckResources dry-runs the chart and performs the resource-list level
// limit check. The manifest is parsed implicitly.
//
// apiVersion v2 skips this check entirely (returns nil). v1, v3, and empty
// apiVersion (v1 default) share the same logic: one helm render at oacPath
// for the legacy path, and per-mode renders at oacPath for modern manifests.
// A non-empty apiVersion outside v1/v2/v3 yields not supported version.
//
// For legacy manifests (<0.12.0) the chart is rendered once and the
// container-level limits are compared against spec.required*/spec.limited*.
// For modern manifests (>=0.12.0) limits come from spec.resources[]; each
// mode drives its own helm render with .Values.GPU.Type set to rm.Mode.
func (c *OAC) CheckResources(oacPath string) error {
	m, err := c.LoadManifestFile(oacPath)
	if err != nil {
		return err
	}
	if cfg, ok := m.Raw().(*manifest.AppConfiguration); ok {
		if err := manifest.ValidateKnownAPIVersion(cfg.APIVersion); err != nil {
			return err
		}
	}
	if isV2Manifest(m) {
		return nil
	}
	sc := ownerScenario{owner: c.owner, admin: c.admin}
	values := c.buildRenderValues(m, sc)
	defaultList, err := helmrender.Render(oacPath, values, m.AppName())
	if err != nil {
		return err
	}
	return c.checkResourceLimits(oacPath, m, sc, defaultList)
}

// isV2Manifest reports whether the manifest follows the v2 multi-chart
// install layout. The parent OAC root in that layout is not a renderable
// workload chart, so callers should render every spec.subCharts[] entry
// individually instead of calling helmrender.Render(oacPath, ...).
func isV2Manifest(m Manifest) bool {
	cfg, ok := m.Raw().(*manifest.AppConfiguration)
	if !ok {
		return false
	}
	return strings.EqualFold(cfg.APIVersion, manifest.APIVersionV2)
}

// CheckServiceAccountRules inspects Role/ClusterRole bindings in the rendered
// chart and returns an error if any of them grants the ServiceAccount one of
// the built-in forbidden permissions.
func (c *OAC) CheckServiceAccountRules(oacPath string) error {
	m, err := c.LoadManifestFile(oacPath)
	if err != nil {
		return err
	}
	sc := ownerScenario{owner: c.owner, admin: c.admin}
	list, err := c.renderForStructuralChecks(oacPath, m, sc)
	if err != nil {
		return err
	}
	rules, err := resources.LoadForbiddenRules("")
	if err != nil {
		return err
	}
	return resources.CheckServiceAccountRules(list, rules)
}

// ValidateManifestFile parses and validates oacPath/OlaresManifest.yaml. No
// chart rendering is performed. For legacy manifests (<0.12.0) the
// underlying pipeline re-parses the payload under both admin=owner and
// admin!=owner scenarios and aggregates any failures into a single
// ValidationError.
//
// When WithAutoOwnerScenarios() is set, the manifest validation is repeated
// for each (owner, admin) pair (owner==admin / owner!=admin) so manifests
// whose body branches on `eq .Values.admin .Values.bfl.username` are
// exercised in both configurations. Failures from each scenario are
// aggregated into a single *ValidationError.
func (c *OAC) ValidateManifestFile(oacPath string) error {
	raw, err := readManifestFile(oacPath)
	if err != nil {
		return err
	}
	return c.ValidateManifestContent(raw)
}

// ValidateManifestContent is the byte-slice counterpart of ValidateManifestFile.
// It honors WithAutoOwnerScenarios() the same way (manifest validation runs
// once per owner scenario).
func (c *OAC) ValidateManifestContent(content []byte) error {
	m, err := c.parseManifest(content)
	if err != nil {
		return err
	}
	return c.validateManifestBytes(content, m)
}

// checkResourceLimits runs CheckResourceLimits against the right render for
// the manifest's schema version.
//
//   - apiVersion v2: the whole check is **skipped** regardless of
//     olaresManifest.version. v2 is the multi-chart install layout where
//     each spec.subCharts[] entry has its own quota, so summing container
//     limits against the parent manifest's spec.required*/spec.limited*
//     (or any single spec.resources[] row) is meaningless.
//   - Legacy (<0.12.0) + v1/v3/empty: the defaultList that was already
//     rendered by the caller is reused and limits come from the flat
//     spec.required*/spec.limited* fields.
//   - Modern (>=0.12.0) + v1/v3/empty: every entry in spec.resources[]
//     drives dedicated helm renders with .Values.GPU.Type set to rm.Mode,
//     because chart templates may emit different workloads per GPU family.
//     Limits come from the inline ResourceRequirement on each mode row.
//     Each mode renders the chart once at oacPath.
//
// A non-empty apiVersion outside v1/v2/v3 yields not supported version.
//
// Charts that carry no resources[] on a modern manifest skip the check
// entirely — Rule 7 already guaranteed the legacy flat fields are empty,
// so there is nothing to compare container limits against.
func (c *OAC) checkResourceLimits(oacPath string, m Manifest, sc ownerScenario, defaultList kube.ResourceList) error {
	cfg, ok := m.Raw().(*manifest.AppConfiguration)
	if !ok {
		// A future Strategy whose Raw() is not *AppConfiguration would
		// silently bypass the entire limit check if we kept the legacy
		// "fall through to zero limits" behaviour. Surface a hard error
		// instead so the caller cannot miss the drift.
		return fmt.Errorf("oac: cannot check resource limits, manifest Raw() is not *AppConfiguration (got %T)", m.Raw())
	}
	if err := manifest.ValidateKnownAPIVersion(cfg.APIVersion); err != nil {
		return err
	}
	if strings.EqualFold(cfg.APIVersion, manifest.APIVersionV2) {
		return nil
	}
	if !manifest.IsModernResourcesManifest(cfg.ConfigVersion) {
		return resources.CheckResourceLimits(defaultList, resources.ResourceLimits{
			RequiredCPU:    cfg.Spec.RequiredCPU,
			RequiredMemory: cfg.Spec.RequiredMemory,
			LimitedCPU:     cfg.Spec.LimitedCPU,
			LimitedMemory:  cfg.Spec.LimitedMemory,
		})
	}
	if len(cfg.Spec.Accelerator) == 0 {
		return nil
	}
	var errs []error
	for _, rm := range cfg.Spec.Accelerator {
		values := c.buildRenderValues(m, sc)
		helmrender.SetGPUType(values, rm.Mode)
		list, err := helmrender.Render(oacPath, values, m.AppName())
		if err != nil {
			errs = append(errs, fmt.Errorf("resources mode=%s: %w", rm.Mode, err))
			continue
		}
		if err := resources.CheckResourceLimits(list, resourceLimitsFromRequirement(rm.ResourceRequirement)); err != nil {
			errs = append(errs, fmt.Errorf("resources mode=%s: %w", rm.Mode, err))
		}
	}
	return errors.Join(errs...)
}

func resourceLimitsFromRequirement(rr manifest.ResourceRequirement) resources.ResourceLimits {
	return resources.ResourceLimits{
		RequiredCPU:    rr.RequiredCPU,
		RequiredMemory: rr.RequiredMemory,
		LimitedCPU:     rr.LimitedCPU,
		LimitedMemory:  rr.LimitedMemory,
	}
}

// renderAllSubCharts dry-runs every entry in subs and returns the
// concatenation of every per-subchart kube.ResourceList. Each subchart is
// rooted at filepath.Join(oacPath, sub.Name) — matching the on-disk layout
// used by production v2 helm install (framework/app-service/pkg/appinstaller/v2).
// The render values are built fresh per subchart because subcharts have
// their own .Values namespace. When mode is non-empty, .Values.GPU.Type is
// set to it so templates that branch per GPU family render the right
// workload set. Values registered via WithValues are merged into every
// subchart's render via buildRenderValues.
func (c *OAC) renderAllSubCharts(
	oacPath string, m Manifest, sc ownerScenario,
	subs []manifest.Chart, mode string,
) (kube.ResourceList, error) {
	var combined kube.ResourceList
	for _, sub := range subs {
		values := c.buildRenderValues(m, sc)
		if mode != "" {
			helmrender.SetGPUType(values, mode)
		}
		subPath := filepath.Join(oacPath, sub.Name)
		list, err := helmrender.Render(subPath, values, sub.Name)
		if err != nil {
			return nil, fmt.Errorf("helm render subchart=%s: %w", sub.Name, err)
		}
		combined = append(combined, list...)
	}
	return combined, nil
}

// checkChartTemplateRules runs owner-independent chart file scans that do not
// require helm rendering: permission-vs-template cross-checks and, for
// apiVersion=v3, the ban on OLARES_USER_* names in templates/values.yaml.
func (c *OAC) checkChartTemplateRules(oacPath string, m Manifest) error {
	var errs []error
	if !c.skipAppData {
		if err := checkPermissionTemplateUsage(oacPath, m); err != nil {
			errs = append(errs, err)
		}
	}
	if isV3Manifest(m) {
		if err := checkV3OLARESUserInChart(oacPath); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func isV3Manifest(m Manifest) bool {
	return strings.EqualFold(m.APIVersion(), manifest.APIVersionV3)
}

// checkManifestWorkloadRefs runs the manifest checks that depend on the
// rendered workload set and therefore cannot live in the manifest validator:
//
//   - workloadReplicas (when declared) must name exactly the rendered
//     Deployment/StatefulSet set, and values.yaml must carry a matching
//     workloads.<name>.replicaCount for every entry.
//   - every overlayGateway entrance workload must reference a rendered
//     Deployment/StatefulSet.
//
// Charts whose Raw() is not *AppConfiguration are skipped silently -- the
// limit check already surfaces that drift with a hard error.
func (c *OAC) checkManifestWorkloadRefs(oacPath string, m Manifest, list kube.ResourceList) error {
	cfg, ok := m.Raw().(*manifest.AppConfiguration)
	if !ok {
		return nil
	}
	var errs []error
	if cfg.WorkloadReplicas != nil {
		replicas := map[string]int32(*cfg.WorkloadReplicas)
		if err := resources.CheckWorkloadReplicas(list, replicas); err != nil {
			errs = append(errs, err)
		}
		if err := checkWorkloadReplicaValues(oacPath, replicas); err != nil {
			errs = append(errs, err)
		}
	}
	if len(cfg.OverlayGateway.Entrances) > 0 {
		workloads := make([]string, len(cfg.OverlayGateway.Entrances))
		for i, e := range cfg.OverlayGateway.Entrances {
			workloads[i] = e.Workload
		}
		if err := resources.CheckOverlayGatewayWorkloads(list, workloads); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// allowClusterScopedCheckApplies reports whether the rendered
// chart must not declare cluster-scoped resources with release-independent
// metadata.name values. v2 is excluded (multi-chart layout); v1 and v3 run
// when options.allowMultipleInstall is true.
func allowClusterScopedCheckApplies(cfg *manifest.AppConfiguration) bool {
	api := strings.ToLower(strings.TrimSpace(cfg.APIVersion))
	if api == "" || api == manifest.APIVersionV1 {
		return true
	}
	if api == manifest.APIVersionV2 {
		return false
	}
	if api == manifest.APIVersionV3 && cfg.Options.AllowMultipleInstall == true {
		return true
	}
	return false
}

// checkAllowMultipleInstallClusterScoped helm-renders the chart twice with
// synthetic release names and flags any cluster-scoped resource whose name is
// identical across both renders.
func (c *OAC) checkAllowMultipleInstallClusterScoped(oacPath string, m Manifest, sc ownerScenario) error {
	cfg, ok := m.Raw().(*manifest.AppConfiguration)
	if !ok || !allowClusterScopedCheckApplies(cfg) {
		return nil
	}
	values := c.buildRenderValues(m, sc)
	probeA, probeB := resources.ClusterScopedProbeNames()
	listA, err := helmrender.Render(oacPath, values, probeA)
	if err != nil {
		return fmt.Errorf("helm render (allowMultipleInstall cluster-scoped probe %q): %w", probeA, err)
	}
	listB, err := helmrender.Render(oacPath, values, probeB)
	if err != nil {
		return fmt.Errorf("helm render (allowMultipleInstall cluster-scoped probe %q): %w", probeB, err)
	}
	return resources.CheckClusterScopedFixedNames(listA, listB)
}

// checkWorkloadReplicaValues verifies that values.yaml declares a
// workloads.<name>.replicaCount entry for every workload named in
// workloadReplicas. This mirrors the install-time convention that each
// workload's replica count is driven from .Values.workloads.<name>.replicaCount,
// so a missing entry would render the manifest's workloadReplicas value
// unusable.
func checkWorkloadReplicaValues(oacPath string, replicas map[string]int32) error {
	data, err := os.ReadFile(filepath.Join(oacPath, "values.yaml"))
	if err != nil {
		return fmt.Errorf("read values.yaml: %w", err)
	}
	var values struct {
		Workloads map[string]map[string]interface{} `yaml:"workloads"`
	}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("parse values.yaml: %w", err)
	}
	names := make([]string, 0, len(replicas))
	for name := range replicas {
		names = append(names, name)
	}
	sort.Strings(names)
	var errs []error
	for _, name := range names {
		wl, ok := values.Workloads[name]
		if !ok {
			errs = append(errs, fmt.Errorf("values.yaml must define workloads.%s.replicaCount", name))
			continue
		}
		if _, ok := wl["replicaCount"]; !ok {
			errs = append(errs, fmt.Errorf("values.yaml must define workloads.%s.replicaCount", name))
		}
	}
	return errors.Join(errs...)
}

// extractUploadDest pulls the options.upload.dest field out of the active
// manifest, returning "" when no upload stanza is configured.
func extractUploadDest(m Manifest) string {
	if cfg, ok := m.Raw().(*manifest.AppConfiguration); ok {
		return cfg.Options.Upload.Dest
	}
	return ""
}

// -- Top-level convenience functions -----------------------------------------

// ValidateManifestFile is the Checker-less shortcut for one-off callers.
func ValidateManifestFile(oacPath string, opts ...Option) error {
	return New(opts...).ValidateManifestFile(oacPath)
}

// ValidateManifestContent is the byte-slice counterpart of
// ValidateManifestFile.
func ValidateManifestContent(content []byte, opts ...Option) error {
	return New(opts...).ValidateManifestContent(content)
}

// Lint is the Checker-less shortcut for (*Checker).Lint.
func Lint(oacPath string, opts ...Option) error {
	return New(opts...).Lint(oacPath)
}

// LintBothOwnerScenarios runs Lint twice: once with owner == admin (cluster
// admin install) and once with owner != admin (regular user install). Both
// scenarios must pass.
//
// This is kept as a named shortcut for Lint with WithAutoOwnerScenarios
// appended to the caller's options.
func LintBothOwnerScenarios(oacPath string, extraOpts ...Option) error {
	// Build a fresh slice rather than appending to extraOpts directly: the
	// variadic argument may share its backing array with the caller, and
	// appending to it could overwrite memory the caller still references.
	opts := make([]Option, 0, len(extraOpts)+1)
	opts = append(opts, extraOpts...)
	opts = append(opts, WithAutoOwnerScenarios())
	return Lint(oacPath, opts...)
}
