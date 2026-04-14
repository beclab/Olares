package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmdAppGet() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:     "get {app-name}",
		Aliases: []string{"info", "show"},
		Short:   "Get detailed information about an app",
		Long: `Get detailed information about an app from the market.

Table output shows a curated summary. JSON output includes the full API response.

Examples:
  olares-cli app get firefox
  olares-cli app get firefox -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(opts, args[0])
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	return cmd
}

func runGet(opts *AppOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("get", appName, err)
	}

	source := resolveCatalogSource(opts)
	if strings.TrimSpace(opts.Source) == "" {
		opts.info("Using source: %s", source)
	}

	ctx := context.Background()
	appInfo, err := fetchAppInfo(ctx, mc, appName, source)
	if err != nil {
		return opts.failOp("get", appName, err)
	}

	if opts.Quiet {
		return nil
	}

	if opts.isJSON() {
		return opts.printJSON(appInfo)
	}

	printAppDetail(appInfo, source)
	return nil
}

func printAppDetail(raw interface{}, source string) {
	m, ok := raw.(map[string]interface{})
	if !ok {
		fmt.Fprintf(os.Stdout, "%v\n", raw)
		return
	}

	name := getNestedString(m, "app_info", "app_entry", "name")
	if name == "" {
		name = getNestedString(m, "app_simple_info", "app_name")
	}

	title := resolveI18nField(m, "app_info", "app_entry", "i18n", "title")
	if title == "" {
		title = extractLocalizedString(getNestedValue(m, "app_simple_info", "app_title"))
	}

	version := getNestedString(m, "version")
	cloneable := appSupportsClone(m)

	description := resolveI18nField(m, "app_info", "app_entry", "i18n", "description")
	developer := getNestedString(m, "app_info", "app_entry", "developer")
	cfgType := getNestedString(m, "app_info", "app_entry", "cfgType")

	var categories []string
	if cats, ok := getNestedValue(m, "app_info", "app_entry", "categories").([]interface{}); ok {
		for _, c := range cats {
			if s, ok := c.(string); ok {
				categories = append(categories, s)
			}
		}
	}

	var entrances []string
	if ents, ok := getNestedValue(m, "app_info", "app_entry", "entrances").([]interface{}); ok {
		for _, e := range ents {
			if em, ok := e.(map[string]interface{}); ok {
				eName, _ := em["name"].(string)
				eTitle, _ := em["title"].(string)
				eHost, _ := em["host"].(string)
				ePort, _ := em["port"].(float64)
				label := eName
				if eTitle != "" && eTitle != eName {
					label = fmt.Sprintf("%s (%s)", eTitle, eName)
				}
				entrances = append(entrances, fmt.Sprintf("%s -> %s:%d", label, eHost, int(ePort)))
			}
		}
	}
	envSpecs := decodeAppEnvSpecs(getNestedValue(m, "raw_data", "envs"))

	fmt.Printf("Name:         %s\n", name)
	if title != "" {
		fmt.Printf("Title:        %s\n", title)
	}
	if version != "" {
		fmt.Printf("Version:      %s\n", version)
	}
	fmt.Printf("Source:        %s\n", source)
	fmt.Printf("Cloneable:    %t\n", cloneable)
	if cfgType != "" {
		fmt.Printf("Type:         %s\n", cfgType)
	}
	if developer != "" {
		fmt.Printf("Developer:    %s\n", developer)
	}
	if len(categories) > 0 {
		fmt.Printf("Categories:   %s\n", strings.Join(categories, ", "))
	}
	if len(entrances) > 0 {
		fmt.Println("Entrances:")
		for _, e := range entrances {
			fmt.Printf("  - %s\n", e)
		}
	}
	if details := formatAppEnvDetails(envSpecs); details != "" {
		fmt.Println(details)
	}
	if description != "" {
		if len(description) > 200 {
			description = description[:200] + "..."
		}
		fmt.Printf("Description:  %s\n", strings.TrimSpace(description))
	}
}

func resolveI18nField(m map[string]interface{}, path ...string) string {
	if len(path) < 2 {
		return ""
	}
	fieldName := path[len(path)-1]
	i18nPath := append(path[:len(path)-1], "i18n")

	i18n := getNestedValue(m, i18nPath...)
	i18nMap, ok := i18n.(map[string]interface{})
	if !ok {
		return ""
	}

	for _, locale := range []string{"en-US", "en", "zh-CN"} {
		localeData, ok := i18nMap[locale].(map[string]interface{})
		if !ok {
			continue
		}
		meta, ok := localeData["metadata"].(map[string]interface{})
		if !ok {
			continue
		}
		if s, ok := meta[fieldName].(string); ok && s != "" {
			return s
		}
	}
	return ""
}
