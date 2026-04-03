package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmdAppSync() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:   "sync {app-name}",
		Short: "Sync an app from a remote source into the local source",
		Long: `Download an app chart from a remote market source and re-upload it to the local source.

Examples:
  olares-cli app sync myapp
  olares-cli app sync myapp --from-source market.olares`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(opts, args[0])
		},
	}
	opts.addConnectionFlags(cmd)
	opts.addSourceFlag(cmd, "local source id to sync the chart into (auto-detected when omitted)")
	opts.addOutputFlags(cmd)
	opts.addVersionFlag(cmd)
	opts.addFromSourceFlag(cmd)
	return cmd
}

func runSync(opts *AppOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("sync", appName, err)
	}

	if s := strings.TrimSpace(opts.Source); s != "" {
		if err := validateLocalSource(s); err != nil {
			return opts.failOp("sync", appName, err)
		}
	}

	toSource := resolveLocalSource(opts)
	if strings.TrimSpace(opts.Source) == "" {
		opts.info("Using target source: %s", toSource)
	}

	fromSource := resolveFromSource(opts)
	if strings.TrimSpace(opts.FromSource) == "" {
		opts.info("Using remote source: %s", fromSource)
	}

	if fromSource == toSource {
		return opts.failOp("sync", appName, fmt.Errorf("source '%s' is already local target source, choose a different --from-source", fromSource))
	}

	version := strings.TrimSpace(opts.Version)
	if version != "" {
		if err := validateVersion(version); err != nil {
			return opts.failOp("sync", appName, err)
		}
	} else {
		v, err := resolveVersionInSource(mc, appName, fromSource)
		if err != nil {
			return opts.failOp("sync", appName, fmt.Errorf("cannot determine version in source '%s': %w", fromSource, err))
		}
		version = v
		opts.info("Using version: %s", version)
	}

	opts.info("Syncing '%s' v%s from '%s' to '%s'...", appName, version, fromSource, toSource)

	kubeClient, err := newKubeClient(strings.TrimSpace(opts.KubeConfig))
	if err != nil {
		return opts.failOp("sync", appName, fmt.Errorf("failed to create kubernetes client: %w", err))
	}
	chartRepoHost, err := discoverChartRepoEndpoint(kubeClient)
	if err != nil {
		return opts.failOp("sync", appName, fmt.Errorf("cannot discover chart-repo-service: %w", err))
	}

	filename := fmt.Sprintf("%s-%s.tgz", appName, version)
	opts.info("Downloading chart '%s'...", filename)

	chartData, err := downloadChartFromRepo(chartRepoHost, filename, mc.user, fromSource)
	if err != nil {
		return opts.failOp("sync", appName, fmt.Errorf("failed to download chart: %w", err))
	}

	opts.info("Uploading chart '%s' (%d bytes) to '%s'...", filename, len(chartData), toSource)

	ctx := context.Background()
	_, err = mc.UploadChartFromReader(ctx, filename, bytes.NewReader(chartData), toSource)
	if err != nil {
		return opts.failOp("sync", appName, fmt.Errorf("failed to upload chart: %w", err))
	}

	result := OperationResult{
		App:       appName,
		Operation: "sync",
		Status:    "success",
		Message:   fmt.Sprintf("version %s synced from '%s' to '%s'", version, fromSource, toSource),
		Source:    toSource,
		Version:   version,
	}
	if !opts.Quiet {
		opts.printResult(result)
	}
	return nil
}

func downloadChartFromRepo(chartRepoHost, filename, user, source string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/charts/%s", chartRepoHost, filename)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Market-User", user)
	req.Header.Set("X-Market-Source", source)

	client := &http.Client{Timeout: defaultRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.TrimSpace(string(body))
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("chart-repo-service returned HTTP %d: %s", resp.StatusCode, bodyStr)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read chart data: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from chart-repo-service")
	}
	return data, nil
}
