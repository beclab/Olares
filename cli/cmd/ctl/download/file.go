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
		Short: "check / remove downloaded files and pre-check URL destinations",
	}
	cmd.AddCommand(newFileExistsCommand(f))
	cmd.AddCommand(newFileCheckCommand(f))
	cmd.AddCommand(newFileRemoveCommand(f))
	return cmd
}

func newFileExistsCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app    string
		path   string
		name   string
		output string
	)
	cmd := &cobra.Command{
		Use:   "exists <url>",
		Short: "pre-check whether a URL download would collide at the destination",
		Long: `Pre-check a URL download destination (GET /api/url/file-exists).

Quote the URL. A URL with ?, & or = must be wrapped in single quotes,
otherwise the shell splits it on & and drops the query string:
  olares-cli knowledge download file exists 'https://host/v?a=1&b=2'

The server resolves the target file name from the URL (or --name) under
--path for the given --app and reports whether it already exists.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFileExists(c.Context(), f, args[0], app, path, name, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&path, "path", "", "destination path (e.g. drive/Home/Downloads/)")
	cmd.Flags().StringVar(&name, "name", "", "expected file_name override")
	return cmd
}

func runFileExists(ctx context.Context, f *cmdutil.Factory, rawURL, app, path, name, outputRaw string) error {
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
	if a := strings.TrimSpace(app); a != "" {
		q.Set("app", a)
	}
	if p := strings.TrimSpace(path); p != "" {
		q.Set("path", p)
	}
	if n := strings.TrimSpace(name); n != "" {
		q.Set("file_name", n)
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

func newFileCheckCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		path   string
		output string
	)
	cmd := &cobra.Command{
		Use:   "check",
		Short: "check whether a downloaded file exists on the PVC",
		Long: `Check whether a file-manager resource exists
(GET /api/download/file_check).

--path is a file-manager resource path such as drive/Home/xxx.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runFileCheck(c.Context(), f, path, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&path, "path", "", "file-manager resource path, e.g. drive/Home/xxx (required)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runFileCheck(ctx context.Context, f *cmdutil.Factory, path, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("--path is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q := url.Values{}
	// user only satisfies the IDL binding (required query field); the real
	// identity is the gateway-injected X-Bfl-User, so the profile OlaresID
	// is just a placeholder here.
	q.Set("user", pc.profile.OlaresID)
	q.Set("path", path)
	var res FileCheckResult
	// /none is a mandatory placeholder suffix in the route; do not drop it.
	if err := doGet(ctx, pc.doer, "/api/download/file_check/none"+encodeQuery(q), &res); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, res)
	default:
		fmt.Printf("Exist:  %v\n", res.Exist)
		return nil
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
		Long: `Remove a file-manager resource (DELETE /api/download/file_remove).

--path is a file-manager resource path such as drive/Home/xxx. A file
that does not exist is still treated as success by the server.`,
		Args: cobra.NoArgs,
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
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("--path is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q := url.Values{}
	// user only satisfies the IDL binding (required query field); the real
	// identity is the gateway-injected X-Bfl-User, so the profile OlaresID
	// is just a placeholder here.
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
