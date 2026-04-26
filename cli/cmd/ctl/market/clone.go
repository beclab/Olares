package market

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	appserviceapi "github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketClone(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "clone {app-name}",
		Short: "Clone an app as a new instance",
		Long: `Clone an installed application to create a new instance with a different title.
Only apps that support multiple instances can be cloned.

Use --entrance-title NAME=TITLE to override cloned desktop shortcut titles.
For apps with a single visible entrance, the entrance title defaults to --title.

The --title flag is required.

Examples:
  olares-cli market clone firefox --title "Firefox Cloned"
  olares-cli market clone myapp --title "MyApp Cloned" --env API_URL=http://dev.example.com
  olares-cli market clone myapp --title "MyApp Cloned" --entrance-title ui="New UI" --entrance-title api="New API"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClone(opts, args[0])
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addTitleFlag(cmd)
	opts.addEnvFlag(cmd)
	opts.addEntranceTitleFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runClone(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("clone", appName, err)
	}

	source := resolveCatalogSource(opts)
	if strings.TrimSpace(opts.Source) == "" {
		opts.info("Using source: %s", source)
	}

	title := strings.TrimSpace(opts.Title)
	if title == "" {
		return opts.failOp("clone", appName, fmt.Errorf("--title is required for cloning"))
	}
	if len(title) > 30 {
		return opts.failOp("clone", appName, fmt.Errorf("--title cannot exceed 30 characters"))
	}

	ctx := context.Background()
	appInfo, err := fetchAppInfo(ctx, mc, appName, source)
	if err != nil {
		return opts.failOp("clone", appName, err)
	}
	if !appSupportsClone(appInfo) {
		return opts.failOp("clone", appName, fmt.Errorf("app '%s' from source '%s' does not support clone", appName, source))
	}

	entrances, err := buildCloneEntrances(appInfo, title, opts.EntranceTitles)
	if err != nil {
		return opts.failOp("clone", appName, err)
	}

	envs, err := parseEnvFlags(opts.Envs)
	if err != nil {
		return opts.failOp("clone", appName, err)
	}

	opts.info("Cloning '%s' as '%s' from '%s' for user '%s'...", appName, title, source, mc.olaresID)

	resp, err := mc.CloneApp(ctx, appName, source, title, envs, entrances)
	if err != nil {
		if envErr := parseServerEnvError(resp, appName); envErr != nil {
			return opts.failOp("clone", appName, envErr)
		}
		if cloneErr := parseServerCloneError(resp); cloneErr != nil {
			return opts.failOp("clone", appName, cloneErr)
		}
		return opts.failOp("clone", appName, err)
	}

	result := newOperationResult(mc, "clone", appName, source, "", fmt.Sprintf("clone requested with title %q", title), resp)
	// Clone runs through the install lifecycle backend-side, but the row
	// that matters is the *new* instance (TargetApp) — the source app is
	// already running. Fall back to the source appName when the backend
	// hasn't surfaced the cloned uid yet, so the watcher still finds a
	// row to track.
	target := result.TargetApp
	if strings.TrimSpace(target) == "" {
		target = appName
	}
	return runWithWatch(opts, mc, result, newWatchTarget(watchInstall, target, source))
}

type cloneEntranceSpec struct {
	Name  string
	Title string
}

type cloneValidationError struct {
	TitleMessage     string
	MissingEntrances []appserviceapi.EntranceClone
	InvalidEntrances []appserviceapi.EntranceClone
}

func (e *cloneValidationError) Error() string {
	var parts []string

	if msg := strings.TrimSpace(e.TitleMessage); msg != "" {
		parts = append(parts, "app title: "+msg)
	}
	if len(e.MissingEntrances) > 0 {
		parts = append(parts, "missing entrance titles: "+formatBackendCloneEntrances(e.MissingEntrances, false))
	}
	if len(e.InvalidEntrances) > 0 {
		parts = append(parts, "invalid entrance titles: "+formatBackendCloneEntrances(e.InvalidEntrances, true))
	}

	if len(parts) == 0 {
		return "invalid clone titles"
	}
	return strings.Join(parts, "; ")
}

func buildCloneEntrances(appInfo map[string]interface{}, appTitle string, rawTitles []string) ([]AppEntrance, error) {
	cloneEntrances := requiredCloneEntrances(appInfo)

	provided, err := parseCloneEntranceTitles(rawTitles)
	if err != nil {
		return nil, err
	}

	if len(cloneEntrances) == 0 {
		if len(provided) > 0 {
			return nil, fmt.Errorf("--entrance-title is not needed because this app has no visible entrances")
		}
		return nil, nil
	}

	if len(cloneEntrances) == 1 && len(provided) == 0 {
		return []AppEntrance{{
			Name:  cloneEntrances[0].Name,
			Title: appTitle,
		}}, nil
	}

	validNames := make(map[string]struct{}, len(cloneEntrances))
	for _, entrance := range cloneEntrances {
		validNames[entrance.Name] = struct{}{}
	}

	var unknown []string
	for name := range provided {
		if _, ok := validNames[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return nil, fmt.Errorf("unknown --entrance-title target(s): %s; available entrances: %s", strings.Join(unknown, ", "), formatRequiredCloneEntrances(cloneEntrances))
	}

	var missing []string
	entrances := make([]AppEntrance, 0, len(cloneEntrances))
	for _, entrance := range cloneEntrances {
		title, ok := provided[entrance.Name]
		if !ok {
			missing = append(missing, describeCloneEntrance(entrance))
			continue
		}
		entrances = append(entrances, AppEntrance{
			Name:  entrance.Name,
			Title: title,
		})
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing --entrance-title for entrances: %s; repeat --entrance-title NAME=TITLE for each visible entrance", strings.Join(missing, ", "))
	}
	return entrances, nil
}

func parseCloneEntranceTitles(rawTitles []string) (map[string]string, error) {
	titles := make(map[string]string, len(rawTitles))
	for _, raw := range rawTitles {
		parts := strings.SplitN(raw, "=", 2)
		name := strings.TrimSpace(parts[0])
		if len(parts) != 2 || name == "" {
			return nil, fmt.Errorf("invalid --entrance-title value %q: expected NAME=TITLE", raw)
		}

		title := strings.TrimSpace(parts[1])
		if title == "" {
			return nil, fmt.Errorf("invalid --entrance-title for entrance %q: title cannot be empty", name)
		}
		if len(title) > 30 {
			return nil, fmt.Errorf("invalid --entrance-title for entrance %q: title cannot exceed 30 characters", name)
		}
		if _, exists := titles[name]; exists {
			return nil, fmt.Errorf("duplicate --entrance-title for entrance %q", name)
		}
		titles[name] = title
	}
	return titles, nil
}

func requiredCloneEntrances(appInfo map[string]interface{}) []cloneEntranceSpec {
	rawEntrances, _ := getNestedValue(appInfo, "app_info", "app_entry", "entrances").([]interface{})
	entrances := make([]cloneEntranceSpec, 0, len(rawEntrances))
	for _, raw := range rawEntrances {
		entry, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		name := strings.TrimSpace(getStringValue(entry, "name"))
		if name == "" {
			continue
		}

		if invisible, _ := entry["invisible"].(bool); invisible {
			continue
		}

		entrances = append(entrances, cloneEntranceSpec{
			Name:  name,
			Title: strings.TrimSpace(extractLocalizedString(entry["title"])),
		})
	}
	return entrances
}

func describeCloneEntrance(entrance cloneEntranceSpec) string {
	if entrance.Title == "" || entrance.Title == entrance.Name {
		return entrance.Name
	}
	return fmt.Sprintf("%s (current: %s)", entrance.Name, entrance.Title)
}

func formatRequiredCloneEntrances(entrances []cloneEntranceSpec) string {
	names := make([]string, 0, len(entrances))
	for _, entrance := range entrances {
		names = append(names, describeCloneEntrance(entrance))
	}
	return strings.Join(names, ", ")
}

func parseServerCloneError(resp *APIResponse) *cloneValidationError {
	if resp == nil || len(resp.Data) == 0 {
		return nil
	}

	data := parseResponseData(resp)
	checkResult := extractServerCloneCheckResult(data)
	if checkResult == nil {
		return nil
	}

	result := &cloneValidationError{
		MissingEntrances: checkResult.MissingValues,
		InvalidEntrances: checkResult.InvalidValues,
	}
	if !checkResult.TitleValidation.IsValid {
		result.TitleMessage = strings.TrimSpace(checkResult.TitleValidation.Message)
	}

	if result.TitleMessage == "" && len(result.MissingEntrances) == 0 && len(result.InvalidEntrances) == 0 {
		return nil
	}
	return result
}

func extractServerCloneCheckResult(data map[string]interface{}) *appserviceapi.AppEntranceCheckResult {
	if data == nil {
		return nil
	}

	checkPayload := data
	if backendResp, ok := data["backend_response"].(map[string]interface{}); ok {
		backendData, ok := backendResp["data"].(map[string]interface{})
		if !ok {
			return nil
		}
		checkPayload = backendData
	}

	checkType, _ := checkPayload["type"].(string)
	if checkType != appserviceapi.CheckTypeAppEntrance {
		return nil
	}

	if nested, ok := checkPayload["Data"].(map[string]interface{}); ok {
		checkPayload = nested
	}

	payload, err := json.Marshal(checkPayload)
	if err != nil {
		return nil
	}

	var result appserviceapi.AppEntranceCheckResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil
	}
	return &result
}

func formatBackendCloneEntrances(entrances []appserviceapi.EntranceClone, includeMessage bool) string {
	parts := make([]string, 0, len(entrances))
	for _, entrance := range entrances {
		label := strings.TrimSpace(entrance.Name)
		if label == "" {
			label = strings.TrimSpace(entrance.Title)
		}
		if label == "" {
			continue
		}

		message := strings.TrimSpace(entrance.Message)
		if includeMessage && message != "" {
			parts = append(parts, fmt.Sprintf("%s (%s)", label, message))
			continue
		}
		if title := strings.TrimSpace(entrance.Title); title != "" && title != label {
			parts = append(parts, fmt.Sprintf("%s (current: %s)", label, title))
			continue
		}
		parts = append(parts, label)
	}

	sort.Strings(parts)
	return strings.Join(parts, ", ")
}
