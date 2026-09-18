package chart

import (
	"fmt"
	"os"

	pkgchart "github.com/beclab/Olares/cli/pkg/chart"
	oac "github.com/beclab/Olares/framework/oac"
	"github.com/spf13/cobra"
)

type lintOpts struct {
	Owner     string
	Admin     string
	AutoOwner bool

	SkipFolder      bool
	SkipManifest    bool
	SkipResource    bool
	SkipHostPath    bool
	SkipNamespace   bool
	SkipAppData     bool
	SkipSameVersion bool

	WithRBAC            bool
	WithSecurityContext bool
}

func NewCmdChartLint() *cobra.Command {
	o := &lintOpts{AutoOwner: true}
	cmd := &cobra.Command{
		Use:   "lint <path>",
		Short: "Validate a directory or .tgz/.tar.gz as a valid Olares chart",
		Long: `Validate that a chart directory or .tgz / .tar.gz package is a valid
Olares chart. The check runs the same pipeline the app store uses to
ingest a chart:

  - chart folder layout (Chart.yaml / values.yaml / templates / OlaresManifest.yaml)
  - OlaresManifest.yaml structural + cross-field validation
  - helm dry-run + workload-integrity, hostPath, namespace checks
  - container-level resource-limit checks
  - Chart.yaml <-> OlaresManifest.yaml version consistency

By default the chart is rendered under both owner==admin and owner!=admin
install scenarios; use --owner / --admin (with --auto-owner=false) to pin
a specific scenario instead.

Examples:
  olares-cli chart lint ./my-app
  olares-cli chart lint ./my-app-1.0.0.tgz
  olares-cli chart lint ./my-app --skip-resource --with-rbac
  olares-cli chart lint ./my-app --auto-owner=false --owner alice --admin root`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(args[0], o)
		},
	}
	fs := cmd.Flags()
	fs.StringVar(&o.Owner, "owner", "", "set .Values.bfl.username for the helm dry-run (overridden by --auto-owner)")
	fs.StringVar(&o.Admin, "admin", "", "set .Values.admin for the helm dry-run (overridden by --auto-owner)")
	fs.BoolVar(&o.AutoOwner, "auto-owner", true, "lint under both owner==admin and owner!=admin scenarios; ignores --owner/--admin when on")

	fs.BoolVar(&o.SkipFolder, "skip-folder", false, "skip folder-layout check")
	fs.BoolVar(&o.SkipManifest, "skip-manifest", false, "skip OlaresManifest.yaml structural validation")
	fs.BoolVar(&o.SkipResource, "skip-resource", false, "skip container-level resource-limit checks")
	fs.BoolVar(&o.SkipHostPath, "skip-host-path", false, "skip hostPath + rolling-update incompatibility check")
	fs.BoolVar(&o.SkipNamespace, "skip-namespace", false, "skip rendered-resource namespace check")
	fs.BoolVar(&o.SkipAppData, "skip-app-data", false, "skip .Values.userspace.appdata cross-check")
	fs.BoolVar(&o.SkipSameVersion, "skip-same-version", false, "skip Chart.yaml <-> OlaresManifest.yaml version match")

	fs.BoolVar(&o.WithRBAC, "with-rbac", false, "enable ServiceAccount RBAC forbidden-rules check (off by default)")
	fs.BoolVar(&o.WithSecurityContext, "with-security-context", false, "explicitly enable non-beclab privileged securityContext check (already on by default)")

	return cmd
}

func runLint(input string, o *lintOpts) error {
	chartDir, cleanup, err := pkgchart.ResolveDir(input)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := oac.Lint(chartDir, buildOACOptions(o)...); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s: OK\n", input)
	return nil
}

func buildOACOptions(o *lintOpts) []oac.Option {
	opts := []oac.Option{}
	if o.AutoOwner {
		opts = append(opts, oac.WithAutoOwnerScenarios())
	} else {
		if o.Owner != "" {
			opts = append(opts, oac.WithOwner(o.Owner))
		}
		if o.Admin != "" {
			opts = append(opts, oac.WithAdmin(o.Admin))
		}
	}
	if o.SkipFolder {
		opts = append(opts, oac.SkipFolderCheck())
	}
	if o.SkipManifest {
		opts = append(opts, oac.SkipManifestCheck())
	}
	if o.SkipResource {
		opts = append(opts, oac.SkipResourceCheck())
	}
	if o.SkipHostPath {
		opts = append(opts, oac.SkipHostPathCheck())
	}
	if o.SkipNamespace {
		opts = append(opts, oac.SkipResourceNamespaceCheck())
	}
	if o.SkipAppData {
		opts = append(opts, oac.SkipAppDataCheck())
	}
	if o.SkipSameVersion {
		opts = append(opts, oac.SkipSameVersionCheck())
	}
	if o.WithRBAC {
		opts = append(opts, oac.WithServiceAccountRulesCheck())
	}
	if o.WithSecurityContext {
		opts = append(opts, oac.WithSecurityContextCheck())
	}
	return opts
}
