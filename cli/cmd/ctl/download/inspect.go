package download

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewInspectCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "inspect <url>",
		Short: "probe a URL for provider and available qualities",
		Long: `Probe a URL (GET /api/url/inspect).

Quote the URL. A URL with ?, & or = must be wrapped in single quotes,
otherwise the shell splits it on & and drops the query string:
  olares-cli knowledge download inspect 'https://host/v?a=1&b=2'

Inspect is advisory: the server may return HTTP 200 with data.error set
when the probe fails. That does not block create.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runInspect(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runInspect(ctx context.Context, f *cmdutil.Factory, rawURL, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("url", rawURL)
	var data InspectData
	if err := doGet(ctx, pc.doer, "/api/url/inspect"+encodeQuery(q), &data); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		if err := printJSON(os.Stdout, data); err != nil {
			return err
		}
	default:
		if err := renderInspect(os.Stdout, data); err != nil {
			return err
		}
	}
	// Advisory probe: still exit 0, but always surface the install CTA when
	// the yt-dlp daemon is unreachable so the next step is not buried in
	// skill docs.
	if hint := inspectYTDLPHint(data); hint != "" {
		fmt.Fprintln(os.Stderr, hint)
	}
	return nil
}

// inspectYTDLPHint returns the install CTA when inspect reports the yt-dlp
// provider is unavailable. Empty string means no hint.
func inspectYTDLPHint(d InspectData) string {
	if d.Available != nil && !*d.Available {
		provider := strings.ToLower(strings.TrimSpace(d.Provider))
		if provider == "" || provider == "yt-dlp" || provider == "ytdlp" {
			return ytdlpUnavailableHint()
		}
	}
	if shouldHintYTDLPUnavailable(d.Error) || shouldHintYTDLPUnavailable(d.ErrorCategory) {
		return ytdlpUnavailableHint()
	}
	return ""
}

func renderInspect(w io.Writer, d InspectData) error {
	fmt.Fprintf(w, "Provider:   %s\n", orDash(d.Provider))
	if d.Title != "" {
		fmt.Fprintf(w, "Title:      %s\n", d.Title)
	}
	if len(d.AvailableQualities) > 0 {
		fmt.Fprintf(w, "Qualities:  %s\n", strings.Join(d.AvailableQualities, ", "))
	} else {
		fmt.Fprintf(w, "Qualities:  -\n")
	}
	if d.Available != nil {
		fmt.Fprintf(w, "Available:  %v\n", *d.Available)
	}
	if d.Error != "" {
		fmt.Fprintf(w, "Error:      %s\n", d.Error)
		if d.ErrorCategory != "" {
			fmt.Fprintf(w, "ErrCategory: %s\n", d.ErrorCategory)
		}
	}
	return nil
}
