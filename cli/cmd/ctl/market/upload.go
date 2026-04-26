package market

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketUpload(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "upload {chart-file-or-dir}",
		Short: "Upload app chart package(s) to the market",
		Long: `Upload Helm-style chart package(s) (.tgz or .tar.gz) to the market.
If the path is a directory, all chart files in the directory are uploaded.

Examples:
  olares-cli market upload myapp-1.0.0.tgz
  olares-cli market upload ./charts/`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpload(opts, args[0])
		},
	}
	opts.addSourceFlag(cmd, "local source id to upload charts into (auto-detected when omitted)")
	opts.addOutputFlags(cmd)
	return cmd
}

type uploadItemResult struct {
	File    string `json:"file"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func isChartFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".tgz") || strings.HasSuffix(lower, ".tar.gz")
}

func runUpload(opts *MarketOptions, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return opts.failOp("upload", path, fmt.Errorf("cannot access '%s': %w", path, err))
	}

	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("upload", path, err)
	}

	if s := strings.TrimSpace(opts.Source); s != "" {
		if err := validateLocalSource(s); err != nil {
			return opts.failOp("upload", path, err)
		}
	}

	source := resolveLocalSource(opts)
	if strings.TrimSpace(opts.Source) == "" {
		opts.info("Using source: %s", source)
	}

	if info.IsDir() {
		return uploadDir(opts, mc, path, source)
	}

	if !isChartFile(info.Name()) {
		return opts.failOp("upload", path, fmt.Errorf("unsupported file format: expected .tgz or .tar.gz"))
	}
	return uploadFile(opts, mc, path, source)
}

func uploadDir(opts *MarketOptions, mc *MarketClient, dir, source string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return opts.failOp("upload", dir, fmt.Errorf("failed to read directory: %w", err))
	}

	var charts []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if isChartFile(e.Name()) {
			charts = append(charts, filepath.Join(dir, e.Name()))
		}
	}

	if len(charts) == 0 {
		return opts.failOp("upload", dir, fmt.Errorf("no chart files (.tgz / .tar.gz) found in '%s'", dir))
	}

	opts.info("Found %d chart(s) in '%s'", len(charts), dir)

	var failed int
	results := make([]uploadItemResult, 0, len(charts))
	for i, f := range charts {
		opts.info("[%d/%d] Uploading %s ...", i+1, len(charts), filepath.Base(f))
		if err := doUploadFile(opts, mc, f, source); err != nil {
			results = append(results, uploadItemResult{
				File:    filepath.Base(f),
				Status:  "failed",
				Message: err.Error(),
			})
			opts.info("  ERROR: %v", err)
			failed++
			continue
		}
		results = append(results, uploadItemResult{
			File:   filepath.Base(f),
			Status: "success",
		})
	}

	if opts.Quiet {
		if failed > 0 {
			return errReported
		}
		return nil
	}

	if opts.isJSON() {
		return opts.printJSON(results)
	}

	if failed > 0 {
		fmt.Fprintf(os.Stderr, "upload '%s': %d of %d uploads failed\n", dir, failed, len(charts))
		return errReported
	}
	fmt.Fprintf(os.Stdout, "upload '%s': all %d chart(s) uploaded\n", dir, len(charts))
	return nil
}

func uploadFile(opts *MarketOptions, mc *MarketClient, filePath, source string) error {
	if err := doUploadFile(opts, mc, filePath, source); err != nil {
		return opts.failOp("upload", filepath.Base(filePath), err)
	}
	result := OperationResult{
		App:       filepath.Base(filePath),
		Operation: "upload",
		Status:    "success",
		Message:   "chart uploaded",
		Source:    source,
	}
	if !opts.Quiet {
		opts.printResult(result)
	}
	return nil
}

func doUploadFile(opts *MarketOptions, mc *MarketClient, filePath, source string) error {
	absPath, _ := filepath.Abs(filePath)
	opts.info("Uploading '%s' to source '%s'...", filepath.Base(absPath), source)
	ctx := context.Background()
	_, err := mc.UploadChart(ctx, absPath, source)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	return nil
}
