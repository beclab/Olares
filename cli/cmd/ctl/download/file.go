package download

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewFileCommand assembles `olares-cli knowledge download file`.
func NewFileCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "pre-check URL destinations and remove downloaded files",
	}
	cmd.AddCommand(newFileExistsCommand(f))
	cmd.AddCommand(newFileRemoveCommand(f))
	return cmd
}

func newFileExistsCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app    string
		path   string
		name   string
		hfDest string
		output string
	)
	cmd := &cobra.Command{
		Use:   "exists <url>",
		Short: "pre-check whether a URL download would collide at the destination",
		Long: `Check whether downloading a URL would overwrite an existing file.

Quote the URL. A URL with ?, & or = must be wrapped in single quotes,
otherwise your shell may split it.

The target name is taken from the URL unless --name is provided. Use --path
and --app to check the same destination you intend to use with create.

For HuggingFace URLs, --hf-dest selects cache vs local destination
semantics. Omit it to use the default.`,
		Example: `  olares-cli knowledge download file exists 'https://host/v?a=1&b=2'
  olares-cli knowledge download file exists https://huggingface.co/org/repo --hf-dest cache`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFileExists(c.Context(), f, args[0], app, path, name, hfDest, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&path, "path", "", "destination path (e.g. drive/Home/Downloads/)")
	cmd.Flags().StringVar(&name, "name", "", "expected file_name override")
	cmd.Flags().StringVar(&hfDest, "hf-dest", "", "HuggingFace destination: cache or local (optional)")
	return cmd
}

func runFileExists(ctx context.Context, f *cmdutil.Factory, rawURL, app, path, name, hfDest, outputRaw string) error {
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
	app, err = validateApp(app)
	if err != nil {
		return err
	}
	hfDest, err = validateHFDest(hfDest)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("url", rawURL)
	q.Set("app", app)
	if p := strings.TrimSpace(path); p != "" {
		q.Set("path", p)
	}
	if n := strings.TrimSpace(name); n != "" {
		q.Set("file_name", n)
	}
	if hfDest != "" {
		q.Set("hf_dest", hfDest)
	}
	var data FileExistsData
	if err := doGet(ctx, pc.doer, "/api/url/file-exists"+encodeQuery(q), &data); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, data)
	default:
		fmt.Printf("Exists:  %v\n", data.Exists)
		if strings.TrimSpace(data.ConflictPath) != "" {
			fmt.Printf("Conflict: %s\n", data.ConflictPath)
		}
		return nil
	}
}

// validateHFDest accepts empty (omit query), "cache", or "local".
func validateHFDest(raw string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "", "cache", "local":
		return v, nil
	default:
		return "", fmt.Errorf("unsupported --hf-dest %q (allowed: cache, local)", raw)
	}
}

func newFileRemoveCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		path   string
		output string
	)
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "remove a downloaded file from the PVC",
		Long: `Remove a downloaded file from Olares storage.

--path is a file-manager resource path such as drive/Home/xxx. A file
that does not exist is still treated as successfully removed.`,
		Example: `  olares-cli knowledge download file remove --path drive/Home/Downloads/video.mp4`,
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runFileRemove(c.Context(), f, path, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&path, "path", "", "file-manager resource path, e.g. drive/Home/xxx (required)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runFileRemove(ctx context.Context, f *cmdutil.Factory, path, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	path, err = normalizeDownloadPath(path, false)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q := url.Values{}
	// A manager still serving file_remove from the generated IDL binder
	// rejects the request without "user"; newer builds ignore it and take
	// identity from X-Bfl-User.
	q.Set("user", pc.profile.OlaresID)
	q.Set("path", path)
	// /none is a mandatory placeholder suffix in the route; do not drop it.
	if err := doMutate(ctx, pc.doer, "DELETE", "/api/download/file_remove/none"+encodeQuery(q), nil, nil); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, RemoveActionResult{Removed: true, Path: path})
	default:
		fmt.Printf("removed %s\n", path)
		return nil
	}
}
