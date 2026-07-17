package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewThirdLevelDomainCommand backs `olares-cli doctor thirdleveldomain`.
func NewThirdLevelDomainCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		kubeconfig  string
		output      string
		quiet       bool
		noHeaders   bool
		forceDedupe bool
	)
	cmd := &cobra.Command{
		Use:     "thirdleveldomain",
		Aliases: []string{"tld"},
		Short:   "check per-user third_level_domain conflicts",
		Long: `Audit Application customDomain.third_level_domain values for
conflicts within each user zone (via kubeconfig).

Reports:
  - duplicate  same prefix used by 2+ app/entrance pairs in one zone
  - reserved   prefix is auth, desktop, or wizard

Shared apps use EffectiveSettings(user); per-user apps contribute only
to their Owner's zone. Does not check default appid prefixes or
third_party_domain.

With --force-dedupe (writes Application CRs):
  - duplicate: keep the lexicographically first (app, entrance) in each
    user zone; clear third_level_domain on the rest
  - reserved: clear every third_level_domain that is auth, desktop, or wizard`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runThirdLevelDomain(c.Context(), kubeconfig, output, quiet, noHeaders, forceDedupe)
		},
	}
	cmd.SilenceUsage = true
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file (optional; falls back to KUBECONFIG / default)")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "output format: table, json")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output; exit code indicates success/failure")
	cmd.Flags().BoolVar(&noHeaders, "no-headers", false, "omit table headers")
	cmd.Flags().BoolVar(&forceDedupe, "force-dedupe", false, "clear duplicate losers (keep one) and reserved auth/desktop/wizard third_level_domain values")
	_ = f
	return cmd
}

func runThirdLevelDomain(ctx context.Context, kubeconfig, output string, quiet, noHeaders, forceDedupe bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c, err := newAppClientFromKubeConfig(kubeconfig)
	if err != nil {
		return err
	}

	apps, users, err := listAppsAndUsers(ctx, c)
	if err != nil {
		return err
	}

	if forceDedupe {
		ops := PlanForceDedupeClears(apps, users)
		if len(ops) > 0 {
			if err := applyForceDedupeClears(ctx, c, ops); err != nil {
				return err
			}
			if !quiet && !strings.EqualFold(strings.TrimSpace(output), "json") {
				renderClearOps(noHeaders, ops)
			}
			apps, users, err = listAppsAndUsers(ctx, c)
			if err != nil {
				return err
			}
		}
	}

	issues := FindThirdLevelDomainIssues(apps, users)
	if err := renderThirdLevelDomainIssues(output, quiet, noHeaders, issues); err != nil {
		return err
	}
	if len(issues) > 0 {
		return fmt.Errorf("found %d third_level_domain issue(s)", len(issues))
	}
	return nil
}

func listAppsAndUsers(ctx context.Context, c client.Client) ([]appv1alpha1.Application, []string, error) {
	var appList appv1alpha1.ApplicationList
	if err := c.List(ctx, &appList); err != nil {
		return nil, nil, fmt.Errorf("list applications: %w", err)
	}

	var iamUsers []string
	var userList iamv1alpha2.UserList
	if err := c.List(ctx, &userList); err == nil {
		iamUsers = make([]string, 0, len(userList.Items))
		for i := range userList.Items {
			iamUsers = append(iamUsers, userList.Items[i].Name)
		}
	}
	return appList.Items, CollectZoneUsers(appList.Items, iamUsers), nil
}

func applyForceDedupeClears(ctx context.Context, c client.Client, ops []ClearThirdLevelOp) error {
	byObject := make(map[string][]ClearThirdLevelOp)
	for _, op := range ops {
		byObject[op.ObjectName] = append(byObject[op.ObjectName], op)
	}
	for objectName, objectOps := range byObject {
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			var app appv1alpha1.Application
			if err := c.Get(ctx, types.NamespacedName{Name: objectName}, &app); err != nil {
				return err
			}
			original := app.DeepCopy()
			if err := ApplyClearOpsToApp(&app, objectOps); err != nil {
				return err
			}
			return c.Patch(ctx, &app, client.MergeFrom(original))
		}); err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("application %q not found while clearing third_level_domain: %w", objectName, err)
			}
			return fmt.Errorf("clear third_level_domain on application %q: %w", objectName, err)
		}
	}
	return nil
}

func renderClearOps(noHeaders bool, ops []ClearThirdLevelOp) {
	fmt.Fprintf(os.Stderr, "force-dedupe: cleared %d third_level_domain value(s)\n", len(ops))
	w := tabwriter.NewWriter(os.Stderr, 0, 0, 2, ' ', 0)
	if !noHeaders {
		fmt.Fprintln(w, "USER\tCLEARED_APP\tENTRANCE\tTHIRD_LEVEL\tREASON\tKEPT")
	}
	for _, op := range ops {
		kept := "-"
		if op.Reason == "duplicate" {
			kept = op.KeepApp + "/" + op.KeepEntrance
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			op.User, op.App, op.Entrance, op.ThirdLevel, op.Reason, kept)
	}
	_ = w.Flush()
}

func renderThirdLevelDomainIssues(output string, quiet, noHeaders bool, issues []DomainIssue) error {
	if quiet {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(output), "json") {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(issues)
	}
	if len(issues) == 0 {
		fmt.Fprintln(os.Stdout, "ok: no third_level_domain conflicts")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if !noHeaders {
		fmt.Fprintln(w, "USER\tAPP\tENTRANCE\tTHIRD_LEVEL\tISSUE")
	}
	for _, iss := range issues {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			iss.User, iss.App, iss.Entrance, iss.ThirdLevel, iss.Issue)
	}
	return w.Flush()
}

func newAppClientFromKubeConfig(kubeconfig string) (client.Client, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add application scheme: %w", err)
	}
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add user scheme: %w", err)
	}
	c, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return c, nil
}
