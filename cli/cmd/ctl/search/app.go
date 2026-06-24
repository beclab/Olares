package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

var uninstalledAppStates = map[string]struct{}{
	"pendingCanceled":      {},
	"downloadingCanceled":  {},
	"downloadFailed":       {},
	"installingCanceled":   {},
	"installFailed":        {},
	"uninstalled":          {},
}

type appOptions struct {
	pagingOptions
}

type appEntrance struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Icon      string `json:"icon"`
	Invisible bool   `json:"invisible"`
	State     string `json:"state"`
}

type appRaw struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Name      string        `json:"name"`
	State     string        `json:"state"`
	URL       string        `json:"url"`
	Icon      string        `json:"icon"`
	Entrances []appEntrance `json:"entrances"`
}

type appItem struct {
	Title      string `json:"title"`
	Name       string `json:"name"`
	AppID      string `json:"appid"`
	EntranceID string `json:"entrance_id"`
	State      string `json:"state"`
	URL        string `json:"url,omitempty"`
	Icon       string `json:"icon,omitempty"`
}

func newAppCommand(f *cmdutil.Factory) *cobra.Command {
	o := &appOptions{}
	cmd := &cobra.Command{
		Use:   "app <keyword>",
		Short: "Search installed applications by title",
		Long: `Search installed applications, mirroring the Desktop global search
"Application" category.

Applications are fetched from /server/myApps and filtered locally by
entrance title (case-insensitive substring match).

Examples:
  olares-cli search app wise
  olares-cli search app drive --limit 10
  olares-cli search app notes -o json
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			keyword, err := parseKeyword(args)
			if err != nil {
				return err
			}
			return runAppSearch(c.Context(), f, keyword, o)
		},
	}
	cmd.SilenceUsage = true
	registerPagingFlags(cmd, &o.pagingOptions)
	return cmd
}

func runAppSearch(ctx context.Context, f *cmdutil.Factory, keyword string, o *appOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := o.validate()
	if err != nil {
		return err
	}

	doer, err := newDoer(ctx, f)
	if err != nil {
		return err
	}

	apps, err := fetchMyApps(ctx, doer)
	if err != nil {
		return err
	}

	matched := filterAppsByKeyword(apps, keyword)
	items := paginateAppItems(matched, o.offset, o.limit)

	switch format {
	case FormatJSON:
		return printAppResultsJSON(os.Stdout, items)
	default:
		return renderAppResults(os.Stdout, items)
	}
}

func fetchMyApps(ctx context.Context, doer *whoami.HTTPClient) ([]appRaw, error) {
	var apps []appRaw
	if err := doEnvelope(ctx, doer, "POST", "/server/myApps", map[string]interface{}{}, &apps); err == nil {
		return apps, nil
	}

	// Fallback: some deployments return a bare JSON array without the BFL envelope.
	if err := doer.DoJSON(ctx, "POST", "/server/myApps", map[string]interface{}{}, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func filterAppsByKeyword(apps []appRaw, keyword string) []appItem {
	needle := strings.ToLower(keyword)
	var items []appItem
	for _, app := range apps {
		if _, skip := uninstalledAppStates[app.State]; skip {
			continue
		}
		if len(app.Entrances) == 0 {
			continue
		}
		for _, entrance := range app.Entrances {
			if entrance.Invisible {
				continue
			}
			title := entrance.Title
			if title == "" {
				title = app.Title
			}
			if !strings.Contains(strings.ToLower(title), needle) {
				continue
			}
			state := app.State
			if app.State == "running" {
				state = entrance.State
			}
			icon := entrance.Icon
			if icon == "" {
				icon = app.Icon
			}
			url := entrance.URL
			if url == "" {
				url = app.URL
			}
			items = append(items, appItem{
				Title:      title,
				Name:       entrance.Name,
				AppID:      app.ID,
				EntranceID: entrance.ID,
				State:      state,
				URL:        url,
				Icon:       icon,
			})
		}
	}
	return items
}

func paginateAppItems(items []appItem, offset, limit int) []appItem {
	if offset >= len(items) {
		return nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func printAppResultsJSON(w io.Writer, items []appItem) error {
	if items == nil {
		items = []appItem{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

func renderAppResults(w io.Writer, items []appItem) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no results")
		return err
	}
	for i, it := range items {
		title := it.Title
		if title == "" {
			title = "(untitled)"
		}
		if _, err := fmt.Fprintf(w, "%d. %s\n", i+1, title); err != nil {
			return err
		}
		if it.Name != "" || it.AppID != "" {
			if _, err := fmt.Fprintf(w, "   app=%s entrance=%s\n", it.AppID, it.Name); err != nil {
				return err
			}
		}
		if it.State != "" {
			if _, err := fmt.Fprintf(w, "   state: %s\n", it.State); err != nil {
				return err
			}
		}
		if it.URL != "" {
			if _, err := fmt.Fprintf(w, "   %s\n", it.URL); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(w, "\n%d result(s)\n", len(items))
	return err
}
