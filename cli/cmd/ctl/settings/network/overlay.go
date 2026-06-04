package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/edge"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/olares"
	"github.com/beclab/Olares/cli/pkg/olaresclient"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// overlayFloor is the first Olares line that ships the overlay gateway. This is
// the canonical "brand new feature" example for the version framework: there is
// no overlay surface at all below 1.12.6, so the whole `overlay` subtree is
// hidden (advisory) on older backends and every olaresclient.OverlayOps method
// returns *ErrUnsupportedVersion (authoritative) there.
var overlayFloor = semver.MustParse("1.12.6")

func supportsOverlay(v *semver.Version) bool {
	if v == nil {
		return true
	}
	return !utils.CoreVersion(v).LessThan(overlayFloor)
}

// overlayPollInterval mirrors the SPA's OVERLAY_GATEWAY_POLL_MS.
const overlayPollInterval = 8 * time.Second

// overlayWatchTimeout caps a --watch poll. A gateway toggle restarts the edge
// network (which briefly returns HTTP 530), so the window is generous.
const overlayWatchTimeout = 3 * time.Minute

// --- wire shapes (overlay-gateway.types.ts) ---

type overlayUnderlayPort struct {
	Title       string `json:"title"`
	Port        int    `json:"port"`
	Workload    string `json:"workload"`
	Description string `json:"description,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
}

type overlayUnderlayNetwork struct {
	IP    string                `json:"ip"`
	Ports []overlayUnderlayPort `json:"ports,omitempty"`
}

type overlaySupportedApp struct {
	AppName          string                   `json:"app_name"`
	AppID            string                   `json:"app_id"`
	Enabled          bool                     `json:"enabled"`
	SharedApp        bool                     `json:"shared_app"`
	UnderlayNetworks []overlayUnderlayNetwork `json:"underlay_networks,omitempty"`
}

type overlayStatus struct {
	Status        string                `json:"status"` // on | off | activating | deactivating
	Disable       bool                  `json:"disable"`
	DisableReason string                `json:"disable_reason"`
	SupportedApps []overlaySupportedApp `json:"supported_apps,omitempty"`
	ErrorMessage  string                `json:"error_message,omitempty"`
}

// NewOverlayCommand builds the `settings network overlay` subtree. Backed by
// the overlay-gateway endpoints (user-service overlay-gateway.controller.ts ->
// olaresd). Enable/disable are owner-only; status is read-only.
func NewOverlayCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overlay",
		Short: "Overlay gateway: assign LAN IPs to supported apps (Olares 1.12.6+)",
		Long: `Inspect and control the overlay gateway, which assigns reachable LAN (underlay)
IP addresses to supported apps.

Subcommands:
  status              show gateway status, availability, and per-app overlay state
  enable [--watch]    turn the gateway on (owner only)
  disable [--watch]   turn the gateway off (owner only)

Toggling the gateway restarts the edge network, during which the API is briefly
unavailable (HTTP 530); --watch polls until the gateway reaches the target
state, treating that window as "restarting".

Requires Olares 1.12.6 or newer.`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newOverlayStatusCommand(f))
	cmd.AddCommand(newOverlayEnableCommand(f))
	cmd.AddCommand(newOverlayDisableCommand(f))
	return cmd
}

func newOverlayStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show overlay gateway status (Olares 1.12.6+)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "read overlay gateway status"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runOverlayStatus(ctx, f, output), "read overlay gateway status")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newOverlayEnableCommand(f *cmdutil.Factory) *cobra.Command {
	var watch bool
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "turn the overlay gateway on (owner only, Olares 1.12.6+)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleOwner, "enable overlay gateway"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runOverlayToggle(ctx, f, true, watch), "enable overlay gateway")
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, "poll until the gateway is on (tolerates the restart window)")
	return cmd
}

func newOverlayDisableCommand(f *cmdutil.Factory) *cobra.Command {
	var watch bool
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "turn the overlay gateway off (owner only, Olares 1.12.6+)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleOwner, "disable overlay gateway"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runOverlayToggle(ctx, f, false, watch), "disable overlay gateway")
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, "poll until the gateway is off (tolerates the restart window)")
	return cmd
}

// overlayUser derives the username the status path expects (the Olares local
// name, matching the SPA's meStore.user.name) from the resolved profile.
func overlayUser(olaresID string) (string, error) {
	id, err := olares.ParseID(olaresID)
	if err != nil {
		return "", fmt.Errorf("parse olaresId %q: %w", olaresID, err)
	}
	return id.Local(), nil
}

func runOverlayStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	doer, rp, err := edge.New(ctx, f)
	if err != nil {
		return err
	}
	user, err := overlayUser(rp.OlaresID)
	if err != nil {
		return err
	}

	return f.WithOlaresClient(ctx, func(c olaresclient.OlaresClient) error {
		st, err := fetchOverlayStatus(ctx, c, doer, user)
		if err != nil {
			return err
		}
		if format == FormatJSON {
			return printJSON(os.Stdout, st)
		}
		return renderOverlayStatus(os.Stdout, st)
	})
}

// fetchOverlayStatus calls OverlayGatewayStatus and decodes the payload.
func fetchOverlayStatus(ctx context.Context, c olaresclient.OlaresClient, doer olaresclient.Doer, user string) (*overlayStatus, error) {
	raw, err := c.OverlayGatewayStatus(ctx, doer, user)
	if err != nil {
		return nil, err
	}
	var st overlayStatus
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &st); err != nil {
			return nil, fmt.Errorf("decode overlay status: %w", err)
		}
	}
	return &st, nil
}

func runOverlayToggle(ctx context.Context, f *cmdutil.Factory, enable, watch bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	doer, rp, err := edge.New(ctx, f)
	if err != nil {
		return err
	}
	user, err := overlayUser(rp.OlaresID)
	if err != nil {
		return err
	}
	target := "off"
	verb := "disable"
	if enable {
		target = "on"
		verb = "enable"
	}

	return f.WithOlaresClient(ctx, func(c olaresclient.OlaresClient) error {
		var toggleErr error
		if enable {
			_, toggleErr = c.EnableOverlayGateway(ctx, doer)
		} else {
			_, toggleErr = c.DisableOverlayGateway(ctx, doer)
		}
		// A 530 means the toggle was accepted and the edge network is
		// restarting (mirrors the SPA's toggleGateway swallow). Any other
		// error is a genuine failure.
		if toggleErr != nil && !isOverlayTransient(toggleErr) {
			return toggleErr
		}
		fmt.Fprintf(os.Stdout, "overlay gateway %s requested\n", verb)

		if !watch {
			fmt.Fprintf(os.Stdout, "the edge network is restarting; run `olares-cli settings network overlay status` to check\n")
			return nil
		}
		return watchOverlayUntil(ctx, c, doer, user, target)
	})
}

// watchOverlayUntil polls the gateway status until it reaches target ('on' /
// 'off'), tolerating the HTTP 530 restart window. Returns when the target is
// reached, or an error on timeout / a non-transient failure.
func watchOverlayUntil(ctx context.Context, c olaresclient.OlaresClient, doer olaresclient.Doer, user, target string) error {
	deadline := time.Now().Add(overlayWatchTimeout)
	restartingNoted := false
	for {
		st, err := fetchOverlayStatus(ctx, c, doer, user)
		switch {
		case err == nil:
			if st.Status == target {
				fmt.Fprintf(os.Stdout, "overlay gateway is %s\n", target)
				if target == "on" {
					_ = renderOverlayStatus(os.Stdout, st)
				}
				return nil
			}
			if st.ErrorMessage != "" {
				return fmt.Errorf("overlay gateway error: %s", st.ErrorMessage)
			}
			fmt.Fprintf(os.Stdout, "  ... gateway is %s, waiting for %s\n", nonEmpty(st.Status), target)
		case isOverlayTransient(err):
			if !restartingNoted {
				fmt.Fprintf(os.Stdout, "  ... edge network restarting (HTTP 530), waiting\n")
				restartingNoted = true
			}
		default:
			return err
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for overlay gateway to become %s", overlayWatchTimeout, target)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(overlayPollInterval):
		}
	}
}

// isOverlayTransient reports whether err is the HTTP 530 the edge ingress
// returns while the network restarts (mirrors the SPA's isOverlayTransientError,
// which keys on response.status === 530). whoami's DoJSON renders non-2xx as a
// message containing "HTTP 530", so we key on that.
func isOverlayTransient(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "HTTP 530")
}

func renderOverlayStatus(w io.Writer, st *overlayStatus) error {
	fmt.Fprintf(w, "GATEWAY:   %s\n", nonEmpty(st.Status))
	fmt.Fprintf(w, "AVAILABLE: %s\n", boolStr(!st.Disable))
	if st.Disable && st.DisableReason != "" {
		fmt.Fprintf(w, "REASON:    %s\n", st.DisableReason)
	}
	if st.ErrorMessage != "" {
		fmt.Fprintf(w, "ERROR:     %s\n", st.ErrorMessage)
	}
	if st.Status != "on" || len(st.SupportedApps) == 0 {
		return nil
	}
	fmt.Fprintln(w, "\nAPPS:")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tOVERLAY\tSHARED\tLAN-IP"); err != nil {
		return err
	}
	for _, a := range st.SupportedApps {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			nonEmpty(a.AppName), boolStr(a.Enabled), boolStr(a.SharedApp), overlayIPs(a.UnderlayNetworks),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func overlayIPs(nets []overlayUnderlayNetwork) string {
	ips := make([]string, 0, len(nets))
	for _, n := range nets {
		if strings.TrimSpace(n.IP) != "" {
			ips = append(ips, n.IP)
		}
	}
	if len(ips) == 0 {
		return "-"
	}
	return strings.Join(ips, ",")
}
