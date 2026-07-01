package network

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	marketcmd "github.com/beclab/Olares/cli/cmd/ctl/market"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings network overlay ...`
//
// 1:1 mirror of the SPA's Settings -> Network -> Overlay Gateway page
// (apps/packages/app/src/pages/settings/Network/OverlayGatewayPage.vue,
// stores/settings/overlayGateway.ts, api/settings/overlayGateway.ts). The
// Overlay Gateway gives supported apps a dedicated LAN IP (via a virtual
// bridge/macvlan gateway) for screen mirroring, DLNA, and device discovery.
//
// Backed by user-service's overlay-gateway.controller.ts, which forwards to the
// olaresd daemon:
//
//	GET  /api/system/overlay-gateway-status/{user}   -> OverlayGatewayStatus
//	POST /api/command/enable-overlay-gateway         (owner-only)
//	POST /api/command/disable-overlay-gateway        (owner-only)
//	POST /api/command/enable-app-overlay-gateway     {app_id, user}
//	POST /api/command/disable-app-overlay-gateway    {app_id, user}
//
// All responses are the uniform BFL {code, message, data} envelope, so the
// same doGetEnvelope / doMutateEnvelope helpers this package already uses for
// reverse-proxy / frp / hosts-file apply here. Unlike hosts-file / frp / ssl
// writes (JWS-gated), these overlay endpoints only require Authorization (plus
// RequireOwner on the gateway master switch), so the CLI can implement them in
// full — exactly like `reverse-proxy set`.
func NewOverlayCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overlay",
		Short: "Overlay gateway (Settings -> Network)",
		Long: `Inspect and control the Overlay Gateway, which gives supported apps a
dedicated LAN IP for screen mirroring, DLNA, and device discovery.

Subcommands:
  status              show gateway state + per-app overlay + LAN endpoints
  enable              turn the overlay gateway on   (owner-only)
  disable             turn the overlay gateway off  (owner-only)
  app enable  <app>   turn an app's overlay on
  app disable <app>   turn an app's overlay off

Toggling an app's overlay restarts the app when it is running (so the
change takes effect), mirroring the Settings page; stopped apps keep the
persisted setting and pick it up on their next start.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newOverlayStatusCommand(f))
	cmd.AddCommand(newOverlayEnableCommand(f))
	cmd.AddCommand(newOverlayDisableCommand(f))
	cmd.AddCommand(newOverlayAppCommand(f))
	return cmd
}

// overlayStatus mirrors olaresd's OverlayGatewayStatus
// (daemon handler_overlay_gateway_status.go) and the SPA's
// OverlayGatewayStatus (api/settings/overlayGateway.ts).
type overlayStatus struct {
	Status        string                `json:"status"` // on|off|activating|deactivating
	Disable       bool                  `json:"disable"`
	DisableReason string                `json:"disable_reason"`
	SupportedApps []overlaySupportedApp `json:"supported_apps"`
	ErrorMessage  string                `json:"error_message"`
}

type overlaySupportedApp struct {
	AppName          string            `json:"app_name"`
	AppID            string            `json:"app_id"`
	Enabled          bool              `json:"enabled"`
	SharedApp        bool              `json:"shared_app"`
	UnderlayNetworks []underlayNetwork `json:"underlay_networks"`
}

type underlayNetwork struct {
	IP    string         `json:"ip"`
	Ports []underlayPort `json:"ports"`
}

type underlayPort struct {
	Title       string `json:"title"`
	Port        int    `json:"port"`
	Workload    string `json:"workload"`
	Description string `json:"description,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
}

// appOverlayCommand is the {app_id, user} body of the per-app enable/disable
// commands (api/settings/overlayGateway.ts AppOverlayGatewayCommand).
type appOverlayCommand struct {
	AppID string `json:"app_id"`
	User  string `json:"user"`
}

const overlayStatusPath = "/api/system/overlay-gateway-status/"

// currentUser resolves the active profile's Olares user name — the value the
// daemon's `itsMe` check requires in the status path's {user} segment, and the
// `user` field the per-app commands send for non-shared apps. It reuses the
// whoami round-trip (BFL /api/backend/v1/user-info) but does NOT persist to the
// role cache (cfg=nil): we only need the name, in-memory.
func currentUser(ctx context.Context, pc *preparedClient) (string, error) {
	res, err := whoami.FetchAndCache(ctx, pc.doer, nil, pc.profile.OlaresID, time.Now)
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(res.Info.Name)
	if name == "" {
		return "", fmt.Errorf("could not resolve current user name from %s", whoami.Endpoint)
	}
	return name, nil
}

func fetchOverlayStatus(ctx context.Context, pc *preparedClient, user string) (overlayStatus, error) {
	var st overlayStatus
	err := doGetEnvelope(ctx, pc.doer, overlayStatusPath+url.PathEscape(user), &st)
	return st, err
}

// -------- status --------

func newOverlayStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show overlay gateway status and per-app overlay",
		Long: `Show the overlay gateway status as a key/value summary plus, when the
gateway is on, a table of overlay-capable apps and their LAN endpoints:

  Status:       on / off / activating / deactivating
  Available:    yes / no
  Reason:       why it's unavailable (WSL / macOS / no ethernet), if any
  Error:        last gateway operation error, if any

  NAME  OVERLAY  SHARED  APP-ID  LAN ENDPOINTS

Use --output json for the raw daemon status (includes per-port workload,
protocol, and description).
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			return preflight.Wrap(ctx, f, runOverlayStatus(ctx, f, output), "show overlay gateway status")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runOverlayStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	user, err := currentUser(ctx, pc)
	if err != nil {
		return err
	}
	st, err := fetchOverlayStatus(ctx, pc, user)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, st)
	default:
		return renderOverlayStatus(os.Stdout, st)
	}
}

func renderOverlayStatus(w io.Writer, st overlayStatus) error {
	rows := [][2]string{
		{"Status", nonEmpty(st.Status)},
		{"Available", boolStr(!st.Disable)},
	}
	if st.Disable && strings.TrimSpace(st.DisableReason) != "" {
		rows = append(rows, [2]string{"Reason", st.DisableReason})
	}
	if strings.TrimSpace(st.ErrorMessage) != "" {
		rows = append(rows, [2]string{"Error", st.ErrorMessage})
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-11s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}

	// The daemon only populates supported_apps when the gateway is on
	// (mirrored by the SPA store's `apps` getter, which returns [] otherwise).
	if st.Status != "on" {
		return nil
	}
	if _, err := fmt.Fprintln(w, "\nApps:"); err != nil {
		return err
	}
	return renderOverlayApps(w, st.SupportedApps)
}

func renderOverlayApps(w io.Writer, apps []overlaySupportedApp) error {
	if len(apps) == 0 {
		_, err := fmt.Fprintln(w, "no overlay-capable apps")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tOVERLAY\tSHARED\tAPP-ID\tLAN ENDPOINTS"); err != nil {
		return err
	}
	for _, a := range apps {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(a.AppName),
			overlayOnOff(a.Enabled),
			boolStr(a.SharedApp),
			nonEmpty(a.AppID),
			nonEmpty(overlayEndpoints(a)),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func overlayOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// overlayEndpoints flattens underlay_networks into a comma-separated list of
// ip:port, matching OverlayAppItem.vue's networkPortRows. Only enabled apps
// carry LAN endpoints.
func overlayEndpoints(a overlaySupportedApp) string {
	if !a.Enabled {
		return ""
	}
	var eps []string
	for _, net := range a.UnderlayNetworks {
		if strings.TrimSpace(net.IP) == "" {
			continue
		}
		for _, p := range net.Ports {
			eps = append(eps, fmt.Sprintf("%s:%d", net.IP, p.Port))
		}
	}
	return strings.Join(eps, ", ")
}

// overlayHasIP reports whether an app has at least one underlay network with an
// assigned IP. Mirrors the store's hasUnderlayIp: enabling is only "done" once
// an IP shows up (the app finished restarting).
func overlayHasIP(a *overlaySupportedApp) bool {
	for _, net := range a.UnderlayNetworks {
		if strings.TrimSpace(net.IP) != "" {
			return true
		}
	}
	return false
}

// -------- --watch (poll until the change converges) --------

// overlayGatewayPollInterval mirrors the SPA's OVERLAY_GATEWAY_POLL_MS (8s):
// how often OverlayGatewayPage re-fetches status while a gateway/app change is
// syncing.
const overlayGatewayPollInterval = 8 * time.Second

type overlayWatchFlags struct {
	watch    bool
	interval time.Duration
	timeout  time.Duration
}

func addOverlayWatchFlags(cmd *cobra.Command, wf *overlayWatchFlags) {
	cmd.Flags().BoolVarP(&wf.watch, "watch", "w", false,
		"poll status until the change converges (gateway settles on/off; app reports enabled + an assigned LAN IP)")
	cmd.Flags().DurationVar(&wf.interval, "watch-interval", overlayGatewayPollInterval,
		"polling interval when --watch is set (default matches the SPA's 8s cadence); no-op without --watch")
	cmd.Flags().DurationVar(&wf.timeout, "watch-timeout", 5*time.Minute,
		"max wall-clock to wait when --watch is set; no-op without --watch")
}

// pollOverlay re-fetches overlay status until done(status) is true, the timeout
// elapses, or ctx is canceled (e.g. Ctrl-C via main.go's signal context). The
// first check happens immediately so an already-converged state returns without
// a wasted sleep.
func pollOverlay(ctx context.Context, pc *preparedClient, user string, wf overlayWatchFlags, done func(overlayStatus) bool) (overlayStatus, error) {
	interval := wf.interval
	if interval <= 0 {
		interval = overlayGatewayPollInterval
	}
	timeout := wf.timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	deadline := time.Now().Add(timeout)

	for {
		st, err := fetchOverlayStatus(ctx, pc, user)
		if err != nil {
			return st, err
		}
		if done(st) {
			return st, nil
		}
		if time.Now().After(deadline) {
			return st, fmt.Errorf("timed out after %s waiting for the overlay change to converge (current status: %s)", timeout, nonEmpty(st.Status))
		}
		select {
		case <-ctx.Done():
			return st, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// -------- enable / disable (gateway master switch, owner-only) --------

func newOverlayEnableCommand(f *cmdutil.Factory) *cobra.Command {
	var wf overlayWatchFlags
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "turn the overlay gateway on (owner-only)",
		Long: `Turn the overlay gateway on. Owner-only: non-owner callers hit a 403.

The gateway comes up asynchronously; it may report 'activating' for a
while. Re-run 'status' (or the SPA polls every 8s) to see it settle at
'on', or pass --watch to block until it does.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleOwner, "enable overlay gateway"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runOverlayGatewayToggle(ctx, f, true, wf), "enable overlay gateway")
		},
	}
	addOverlayWatchFlags(cmd, &wf)
	return cmd
}

func newOverlayDisableCommand(f *cmdutil.Factory) *cobra.Command {
	var wf overlayWatchFlags
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "turn the overlay gateway off (owner-only)",
		Long: `Turn the overlay gateway off. Owner-only: non-owner callers hit a 403.

The gateway goes down asynchronously; it may report 'deactivating' for a
while. Re-run 'status' to see it settle at 'off', or pass --watch to block
until it does.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleOwner, "disable overlay gateway"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runOverlayGatewayToggle(ctx, f, false, wf), "disable overlay gateway")
		},
	}
	addOverlayWatchFlags(cmd, &wf)
	return cmd
}

func runOverlayGatewayToggle(ctx context.Context, f *cmdutil.Factory, enable bool, wf overlayWatchFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/api/command/disable-overlay-gateway"
	if enable {
		path = "/api/command/enable-overlay-gateway"
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", path, nil, nil); err != nil {
		return err
	}

	target := "off"
	transient := "deactivating"
	if enable {
		target = "on"
		transient = "activating"
	}

	if !wf.watch {
		fmt.Fprintf(os.Stdout, "overlay gateway %s requested (%s); run 'settings network overlay status' to check\n",
			map[bool]string{true: "enable", false: "disable"}[enable], transient)
		return nil
	}

	// --watch: poll status until the gateway settles at the target
	// (activating/deactivating are the transient states we wait through),
	// mirroring OverlayGatewayPage's 8s polling.
	user, err := currentUser(ctx, pc)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "overlay gateway %s requested (%s); waiting until it settles at %q (timeout: %s)...\n",
		map[bool]string{true: "enable", false: "disable"}[enable], transient, target, wf.timeout)
	if _, err := pollOverlay(ctx, pc, user, wf, func(s overlayStatus) bool {
		return s.Status == target
	}); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "overlay gateway is now %s\n", target)
	return nil
}

// -------- app enable / disable --------

func newOverlayAppCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "per-app overlay (enable / disable)",
		Long: `Turn a single app's overlay on or off. The app must be overlay-capable
(declared in its manifest and visible under 'status' while the gateway is
on).

When the app is running it is restarted so the change takes effect
(mirrors the Settings page); stopped apps keep the persisted setting and
pick it up on their next start.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newOverlayAppToggleCommand(f, true))
	cmd.AddCommand(newOverlayAppToggleCommand(f, false))
	return cmd
}

func newOverlayAppToggleCommand(f *cmdutil.Factory, enable bool) *cobra.Command {
	use := "disable {app-name}"
	short := "turn an app's overlay off"
	verb := "disable app overlay"
	if enable {
		use = "enable {app-name}"
		short = "turn an app's overlay on"
		verb = "enable app overlay"
	}
	var wf overlayWatchFlags
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			return preflight.Wrap(ctx, f, runOverlayAppToggle(ctx, f, args[0], enable, wf), verb)
		},
	}
	addOverlayWatchFlags(cmd, &wf)
	return cmd
}

func runOverlayAppToggle(ctx context.Context, f *cmdutil.Factory, appName string, enable bool, wf overlayWatchFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	user, err := currentUser(ctx, pc)
	if err != nil {
		return err
	}

	// Look up the supported app so we send the right app_id and the correct
	// `user` (empty for shared apps), exactly as the store's toggleAppOverlay
	// does. The daemon only reports supported_apps while the gateway is on.
	st, err := fetchOverlayStatus(ctx, pc, user)
	if err != nil {
		return err
	}
	if st.Status != "on" {
		return fmt.Errorf("overlay gateway is not on (status: %s); enable it first with 'settings network overlay enable'", nonEmpty(st.Status))
	}
	var found *overlaySupportedApp
	for i := range st.SupportedApps {
		if st.SupportedApps[i].AppName == appName || st.SupportedApps[i].AppID == appName {
			found = &st.SupportedApps[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("%q is not an overlay-capable app (run 'settings network overlay status' to list them)", appName)
	}

	appID := found.AppID
	if appID == "" {
		appID = appName
	}
	body := appOverlayCommand{AppID: appID, User: user}
	if found.SharedApp {
		body.User = ""
	}

	path := "/api/command/disable-app-overlay-gateway"
	if enable {
		path = "/api/command/enable-app-overlay-gateway"
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", path, body, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "overlay %s for app %q\n", overlayOnOff(enable), found.AppName)

	// Mirror OverlayGatewayPage: restart running apps so the change takes
	// effect. Best-effort — the overlay setting is already persisted, so a
	// restart hiccup is a warning, not a failure of this command.
	restarted, rerr := marketcmd.RestartApp(ctx, f, found.AppName)
	switch {
	case rerr != nil:
		fmt.Fprintf(os.Stderr, "warning: overlay %s but restarting %q failed: %v\n  the setting is persisted; restart the app manually with 'olares-cli market restart %s'\n",
			overlayOnOff(enable), found.AppName, rerr, found.AppName)
	case restarted:
		fmt.Fprintf(os.Stdout, "restarting %q to apply the change\n", found.AppName)
	default:
		fmt.Fprintf(os.Stdout, "%q is not running; the change is persisted and applies on next start\n", found.AppName)
	}

	if !wf.watch {
		return nil
	}

	// --watch: poll status until the app converges, mirroring the store's
	// evaluateStatusSyncTargets — enabling is "done" once the app reports
	// enabled AND an underlay IP is assigned (i.e. the restart finished);
	// disabling is done as soon as it is no longer enabled (or drops off the
	// supported list).
	appName = found.AppName
	fmt.Fprintf(os.Stdout, "waiting for %q overlay to converge (timeout: %s)...\n", appName, wf.timeout)
	if _, err := pollOverlay(ctx, pc, user, wf, func(s overlayStatus) bool {
		var sup *overlaySupportedApp
		for i := range s.SupportedApps {
			if s.SupportedApps[i].AppName == appName || (appID != "" && s.SupportedApps[i].AppID == appID) {
				sup = &s.SupportedApps[i]
				break
			}
		}
		if enable {
			return sup != nil && sup.Enabled && overlayHasIP(sup)
		}
		return sup == nil || !sup.Enabled
	}); err != nil {
		return err
	}
	if enable {
		fmt.Fprintf(os.Stdout, "%q overlay is active (LAN IP assigned)\n", appName)
	} else {
		fmt.Fprintf(os.Stdout, "%q overlay is now off\n", appName)
	}
	return nil
}
