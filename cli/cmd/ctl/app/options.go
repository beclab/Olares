package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errReported = errors.New("(already reported)")

type AppOptions struct {
	User       string
	Host       string
	KubeConfig string
	Source     string
	FromSource string

	Output    string
	Quiet     bool
	NoHeaders bool

	Version        string
	AllSources     bool
	Cascade        bool
	Category       string
	Envs           []string
	EntranceTitles []string
	DeleteData     bool
	Title          string
}

func (o *AppOptions) isJSON() bool {
	return strings.EqualFold(strings.TrimSpace(o.Output), "json")
}

// info prints an informational message to stderr.
// Suppressed in JSON and quiet modes.
func (o *AppOptions) info(format string, args ...interface{}) {
	if o.Quiet || o.isJSON() {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (o *AppOptions) addConnectionFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.User, "user", "u", "", "target user (auto-detected if only one user exists)")
	cmd.Flags().StringVar(&o.Host, "host", "", "market service address (auto-discovered from k8s)")
	cmd.Flags().StringVar(&o.KubeConfig, "kubeconfig", "", "path to kubeconfig file")
}

func (o *AppOptions) addSourceFlag(cmd *cobra.Command, desc string) {
	if desc == "" {
		desc = "market source id (auto-detected when omitted)"
	}
	cmd.Flags().StringVarP(&o.Source, "source", "s", "", desc)
}

func (o *AppOptions) addCommonFlags(cmd *cobra.Command) {
	o.addConnectionFlags(cmd)
	o.addSourceFlag(cmd, "")
}

func (o *AppOptions) addOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Output, "output", "o", "table", "output format: table, json")
	cmd.Flags().BoolVarP(&o.Quiet, "quiet", "q", false, "suppress output; exit code indicates success/failure")
	cmd.Flags().BoolVar(&o.NoHeaders, "no-headers", false, "omit table headers (useful for scripting)")
}

func (o *AppOptions) addVersionFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Version, "version", "", "app version (default: latest available)")
}

func (o *AppOptions) addAllSourcesFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&o.AllSources, "all-sources", "a", false, "include apps from all sources")
}

func (o *AppOptions) addCascadeFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.Cascade, "cascade", false, "apply to all sub-charts (for v2 multi-chart apps)")
}

func (o *AppOptions) addEnvFlag(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&o.Envs, "env", nil, "set env var in KEY=VALUE format (repeatable)")
}

func (o *AppOptions) addEntranceTitleFlag(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&o.EntranceTitles, "entrance-title", nil, "set cloned entrance title in NAME=TITLE format (repeatable)")
}

func (o *AppOptions) addDeleteDataFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.DeleteData, "delete-data", false, "delete persistent data when uninstalling")
}

func (o *AppOptions) addTitleFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Title, "title", "", "display title for the cloned app instance")
}

func (o *AppOptions) addFromSourceFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.FromSource, "from-source", "", "remote source to sync from")
}

func (o *AppOptions) prepare() (*MarketClient, error) {
	user := strings.TrimSpace(o.User)
	host := strings.TrimSpace(o.Host)

	var kubeClient client.Client
	if user == "" || host == "" {
		c, err := newKubeClient(strings.TrimSpace(o.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
		}
		kubeClient = c
	}

	if user == "" {
		u, err := resolveUser(kubeClient)
		if err != nil {
			return nil, fmt.Errorf("cannot auto-detect user: %w", err)
		}
		user = u
	}

	if host == "" {
		endpoint, err := discoverMarketEndpoint(kubeClient)
		if err != nil {
			return nil, fmt.Errorf("cannot discover market service: %w", err)
		}
		host = endpoint
	}

	return NewMarketClient(host, user, strings.TrimSpace(o.Source)), nil
}

func (o *AppOptions) failOp(op, app string, err error) error {
	if o.Quiet {
		return errReported
	}
	result := OperationResult{
		App:       app,
		Operation: op,
		Status:    "failed",
		Message:   err.Error(),
	}
	o.printResult(result)
	return errReported
}

func (o *AppOptions) printJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func (o *AppOptions) printResult(result OperationResult) {
	if o.Quiet {
		return
	}
	if o.isJSON() {
		if err := o.printJSON(result); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode JSON output: %v\n", err)
		}
		return
	}

	writer := os.Stdout
	appLabel := result.App
	if result.TargetApp != "" && result.TargetApp != result.App {
		appLabel = fmt.Sprintf("%s -> %s", result.App, result.TargetApp)
	}

	message := strings.TrimSpace(result.Message)
	if message == "" {
		switch result.Status {
		case "accepted":
			message = "request accepted"
		case "success":
			message = "completed successfully"
		case "failed":
			message = "request failed"
		default:
			message = "completed"
		}
	}

	if result.Status == "failed" {
		writer = os.Stderr
		fmt.Fprintf(writer, "%s '%s' failed: %s\n", result.Operation, appLabel, message)
	} else {
		fmt.Fprintf(writer, "%s '%s': %s\n", result.Operation, appLabel, message)
	}

	if result.Source != "" {
		fmt.Fprintf(writer, "  source: %s\n", result.Source)
	}
	if result.Version != "" {
		fmt.Fprintf(writer, "  version: %s\n", result.Version)
	}
	if result.State != "" {
		fmt.Fprintf(writer, "  state: %s\n", result.State)
	}
	if result.Progress != "" && result.Progress != "0.00" {
		fmt.Fprintf(writer, "  progress: %s\n", result.Progress)
	}
}
