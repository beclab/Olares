package download

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewTorrentCommand assembles `olares-cli knowledge download torrent`.
func NewTorrentCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "torrent",
		Short: "torrent inspection and seeding control",
	}
	cmd.AddCommand(newTorrentInspectCommand(f))
	cmd.AddCommand(newTorrentStatsCommand(f))
	cmd.AddCommand(newTorrentPeersCommand(f))
	cmd.AddCommand(newTorrentFilesCommand(f))
	cmd.AddCommand(newTorrentSeedCommand(f))
	return cmd
}

func newTorrentInspectCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		file   string
		output string
	)
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "inspect a local .torrent file (metadata + file list)",
		Long: `Inspect a local .torrent file (POST /api/download/torrent/inspect).

The file is read locally, base64-encoded and uploaded; the server parses
the metainfo and returns the info hash, mode, piece layout and the 1-based
file list used by torrent files --select and create --select-files.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runTorrentInspect(c.Context(), f, file, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&file, "file", "", "path to a local .torrent file (required)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func runTorrentInspect(ctx context.Context, f *cmdutil.Factory, file, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	file = strings.TrimSpace(file)
	if file == "" {
		return fmt.Errorf("--file is required")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read torrent file: %w", err)
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	req := TorrentInspectReq{TorrentFileB64: base64.StdEncoding.EncodeToString(raw)}
	var result TorrentInspectResult
	if err := doMutate(ctx, pc.doer, "POST", "/api/download/torrent/inspect", req, &result); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, result)
	default:
		return renderTorrentInspect(os.Stdout, result)
	}
}

func renderTorrentInspect(w io.Writer, r TorrentInspectResult) error {
	fmt.Fprintf(w, "%-12s %s\n", "Name:", orDash(r.Name))
	fmt.Fprintf(w, "%-12s %s\n", "InfoHash:", orDash(r.InfoHash))
	fmt.Fprintf(w, "%-12s %s\n", "Mode:", orDash(r.Mode))
	fmt.Fprintf(w, "%-12s %d\n", "TotalSize:", r.TotalSize)
	fmt.Fprintf(w, "%-12s %d\n", "NumPieces:", r.NumPieces)
	fmt.Fprintf(w, "%-12s %d\n", "Trackers:", len(r.Trackers))
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "INDEX\tSIZE\tPATH")
	for _, file := range r.Files {
		fmt.Fprintf(tw, "%d\t%d\t%s\n", file.Index, file.Length, orDash(file.Path))
	}
	return tw.Flush()
}

func newTorrentStatsCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "stats <id>",
		Short: "show live BitTorrent stats for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTorrentStats(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runTorrentStats(ctx context.Context, f *cmdutil.Factory, idRaw, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	id, err := parseTaskID(idRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var stats TorrentLiveStats
	if err := doGet(ctx, pc.doer, fmt.Sprintf("/api/download/%d/torrent", id), &stats); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, stats)
	default:
		return renderTorrentStats(os.Stdout, stats)
	}
}

func renderTorrentStats(w io.Writer, s TorrentLiveStats) error {
	fields := [][2]string{
		{"DownloadSpeed", strconv.FormatInt(s.DownloadSpeed, 10)},
		{"UploadSpeed", strconv.FormatInt(s.UploadSpeed, 10)},
		{"UploadedBytes", strconv.FormatInt(s.UploadedBytes, 10)},
		{"ShareRatio", fmt.Sprintf("%.3f", s.ShareRatio)},
		{"Connections", strconv.FormatInt(s.Connections, 10)},
		{"NumSeeders", strconv.FormatInt(s.NumSeeders, 10)},
		{"Pieces", fmt.Sprintf("%d of %d", s.PiecesHave, s.NumPieces)},
		{"VerifiedBytes", strconv.FormatInt(s.VerifiedBytes, 10)},
		{"ETASeconds", strconv.FormatInt(s.ETASeconds, 10)},
		{"IsSeeding", strconv.FormatBool(s.IsSeeding)},
	}
	for _, kv := range fields {
		fmt.Fprintf(w, "%-14s %s\n", kv[0]+":", kv[1])
	}
	return nil
}

func newTorrentPeersCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "peers <id>",
		Short: "list connected BitTorrent peers for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTorrentPeers(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runTorrentPeers(ctx context.Context, f *cmdutil.Factory, idRaw, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	id, err := parseTaskID(idRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var peers TorrentPeers
	if err := doGet(ctx, pc.doer, fmt.Sprintf("/api/download/%d/torrent/peers", id), &peers); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, peers)
	default:
		return renderTorrentPeers(os.Stdout, peers)
	}
}

func renderTorrentPeers(w io.Writer, p TorrentPeers) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "IP:PORT\tPROGRESS\tDOWN\tUP\tSEEDER")
	for _, peer := range p.Peers {
		fmt.Fprintf(tw, "%s\t%.1f%%\t%d\t%d\t%v\n",
			fmt.Sprintf("%s:%d", peer.IP, peer.Port),
			peer.Progress*100,
			peer.DownloadSpeed,
			peer.UploadSpeed,
			peer.Seeder,
		)
	}
	return tw.Flush()
}

func newTorrentFilesCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		selectRaw string
		output    string
	)
	cmd := &cobra.Command{
		Use:   "files <id>",
		Short: "select which files of a multi-file torrent to download",
		Long: `Set the selected files of a multi-file torrent
(PUT /api/download/<id>/torrent/files).

--select takes a comma-separated list of 1-based file indices (as reported
by torrent inspect), e.g. --select 1,3,5. The list is the full selection,
not a delta. Pass --select all to download every file (sends an empty
selection so the server keeps all files).`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTorrentFiles(c.Context(), f, args[0], selectRaw, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&selectRaw, "select", "", "comma-separated 1-based file indices, or \"all\" (required)")
	_ = cmd.MarkFlagRequired("select")
	return cmd
}

func runTorrentFiles(ctx context.Context, f *cmdutil.Factory, idRaw, selectRaw, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	id, err := parseTaskID(idRaw)
	if err != nil {
		return err
	}
	sel, err := parseSelectedIndices(selectRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var res SetTorrentFilesResult
	if err := doMutate(ctx, pc.doer, "PUT", fmt.Sprintf("/api/download/%d/torrent/files", id), SetTorrentFilesReq{Selected: sel}, &res); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, res)
	default:
		fmt.Printf("selected files for task %d: %s\n", id, formatSelected(res.Selected))
		return nil
	}
}

// parseSelectedIndices turns a CSV of 1-based indices into a slice. The literal
// "all" (case-insensitive) yields an empty slice, meaning "all files". Each
// entry must be a positive integer.
func parseSelectedIndices(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("--select is required (comma-separated 1-based indices, or \"all\")")
	}
	if strings.EqualFold(raw, "all") {
		return []int{}, nil
	}
	parts := strings.Split(raw, ",")
	sel := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		n, err := strconv.Atoi(p)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("invalid file index %q (need positive 1-based integers, or \"all\")", p)
		}
		sel = append(sel, n)
	}
	return sel, nil
}

func formatSelected(sel []int) string {
	if len(sel) == 0 {
		return "all"
	}
	sorted := make([]int, len(sel))
	copy(sorted, sel)
	sort.Ints(sorted)
	parts := make([]string, len(sorted))
	for i, n := range sorted {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ",")
}

func newTorrentSeedCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "control seeding for a completed torrent",
		Long: `Start or stop seeding a completed torrent
(POST /api/download/<id>/torrent/seed/stop|resume).

409 semantics:
  seed stop   requires the task to be currently seeding.
  seed resume requires the task to be completed (download finished).`,
	}
	cmd.AddCommand(newTorrentSeedActionCommand(f, "stop"))
	cmd.AddCommand(newTorrentSeedActionCommand(f, "resume"))
	return cmd
}

func newTorrentSeedActionCommand(f *cmdutil.Factory, action string) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   action + " <id>",
		Short: action + " seeding for a completed torrent",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTorrentSeed(c.Context(), f, action, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runTorrentSeed(ctx context.Context, f *cmdutil.Factory, action, idRaw, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	id, err := parseTaskID(idRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var res SeedControlResult
	if err := doMutate(ctx, pc.doer, "POST", fmt.Sprintf("/api/download/%d/torrent/seed/%s", id, action), nil, &res); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, res)
	default:
		fmt.Printf("seed %s task %d: status=%s\n", action, id, orDash(res.Status))
		return nil
	}
}
