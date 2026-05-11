package vpn

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings vpn devices rename | delete | tags set ...`
//
// Backed by user-service's HeadScaleController, which forwards each call
// to the per-Olares headscale sidecar verbatim:
//
//   POST   /headscale/machine/<id>/rename/<name>      (no body)
//   DELETE /headscale/machine/<id>                    (no body)
//   POST   /headscale/machine/<id>/tags               { tags: ["tag:<t>"...] }
//
// Wire shape: Headscale returns raw JSON (no BFL envelope). user-service
// proxies it through ProviderClient.execute. We don't bother decoding the
// response on writes — the SPA itself ignores it (`stores/settings/
// headscale.ts:85+`) and just refreshes the device list afterward, which
// the user can do via `settings vpn devices list`.
//
// Role: device write verbs (rename / delete / tags) require admin floor.
// HeadScaleDeviceCard.vue itself has no isAdmin guard, but the device-
// management surface lives under the admin-gated VPN page and is owner/
// admin territory in practice. We soft-preflight to admin so a normal
// caller fails fast with a friendly hint; Headscale stays authoritative.

// NewDevicesCommand is defined in devices.go; we extend it here by
// registering the write subcommands during package init via a pattern
// where devices.go calls registerDeviceWriteCommands(cmd, f). To keep
// the existing devices.go untouched on the read side, we inline the
// write registrations into NewDevicesCommand below — the original
// definition lives in devices.go and must call this helper.

// addDevicesWriteCommands wires the device-write verbs onto an
// existing `devices` parent. devices.go's NewDevicesCommand must call
// this so the read+write surface stays under one parent.
func addDevicesWriteCommands(parent *cobra.Command, f *cmdutil.Factory) {
	parent.AddCommand(newDevicesRenameCommand(f))
	parent.AddCommand(newDevicesDeleteCommand(f))
	parent.AddCommand(newDevicesTagsCommand(f))
}

// `vpn devices rename <id> <new-name>`
func newDevicesRenameCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <device-id> <new-name>",
		Short: "rename a Headscale device",
		Long: `Rename a Headscale device on this Olares user's mesh.

The new name is sent verbatim to Headscale, which uses it as both the
display name (givenName) and the canonical hostname for tailnet DNS.
Headscale enforces uniqueness across the mesh; if the name is taken or
contains characters Headscale rejects, the upstream error message is
forwarded as-is.

Example:
  olares-cli settings vpn devices rename 7 alices-laptop
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "rename Headscale device"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runDevicesRename(ctx, f, args[0], args[1]), "rename Headscale device")
		},
	}
	return cmd
}

func runDevicesRename(ctx context.Context, f *cmdutil.Factory, deviceID, newName string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	deviceID = strings.TrimSpace(deviceID)
	newName = strings.TrimSpace(newName)
	if deviceID == "" {
		return fmt.Errorf("device-id is required")
	}
	if newName == "" {
		return fmt.Errorf("new-name is required")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := doRenameViaDoer(ctx, pc.doer, deviceID, newName); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "renamed device %s to %q\n", deviceID, newName)
	return nil
}

// doRenameViaDoer is the wire-only core of `vpn devices rename`,
// extracted so unit tests can assert path/method/body without spinning
// up a Factory. Both path segments are user-supplied; we PathEscape
// them so a name like "alice's laptop" or a numeric id with reserved
// characters can't break the URL.
func doRenameViaDoer(ctx context.Context, d Doer, deviceID, newName string) error {
	path := "/headscale/machine/" + url.PathEscape(deviceID) + "/rename/" + url.PathEscape(newName)
	return d.DoJSON(ctx, "POST", path, nil, nil)
}

// `vpn devices delete <id>` — destructive, gated on confirmation.
func newDevicesDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "delete <device-id>",
		Short: "remove a Headscale device from this Olares user's mesh",
		Long: `Remove a Headscale device. The device immediately loses access to the
mesh; any TermiPass session bound to that device is invalidated and
must re-enroll.

This verb prompts for confirmation by default. Pass --yes to skip the
prompt for automation. Non-TTY stdin without --yes is a hard error so
unattended scripts don't silently destroy state.

Example:
  olares-cli settings vpn devices delete 7
  olares-cli settings vpn devices delete 7 --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "delete Headscale device"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runDevicesDelete(ctx, f, args[0], assumeYes), "delete Headscale device")
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required for non-TTY automation)")
	return cmd
}

func runDevicesDelete(ctx context.Context, f *cmdutil.Factory, deviceID string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return fmt.Errorf("device-id is required")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Delete Headscale device %q? Bound TermiPass sessions will be invalidated.", deviceID),
		assumeYes); err != nil {
		return err
	}
	if err := doDeleteViaDoer(ctx, pc.doer, deviceID); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "deleted device %s\n", deviceID)
	return nil
}

// doDeleteViaDoer is the wire-only core of `vpn devices delete`.
func doDeleteViaDoer(ctx context.Context, d Doer, deviceID string) error {
	path := "/headscale/machine/" + url.PathEscape(deviceID)
	return d.DoJSON(ctx, "DELETE", path, nil, nil)
}

// `vpn devices tags ...` parent — only `set` for now (replace-the-list
// semantics, matching the upstream POST /tags signature). add/rm wrappers
// can come later if anyone asks; in practice replacing the whole list is
// the natural shape because Headscale's "ForcedTags" is a set, not a log.
func newDevicesTagsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "manage forcedTags on a Headscale device",
		Long: `Manage the forcedTags Headscale stores against a device.

Subcommands:
  set   replace the device's tag list
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDevicesTagsSetCommand(f))
	return cmd
}

// `vpn devices tags set <id> [--tag x ...]`
func newDevicesTagsSetCommand(f *cmdutil.Factory) *cobra.Command {
	var tags []string
	cmd := &cobra.Command{
		Use:   "set <device-id>",
		Short: "replace the device's forcedTags list",
		Long: `Replace the forcedTags list on a Headscale device.

Use --tag <name> for each tag (repeatable). Tags are sent as the SPA
sends them: each tag is automatically prefixed with "tag:" before
hitting Headscale, so callers should pass the bare name (e.g. "ops",
not "tag:ops"). Tags already prefixed with "tag:" are passed through
unchanged so this stays idempotent against the SPA's stored values.

Passing zero --tag flags clears all tags on the device (mirrors what
the SPA's "Remove all tags" path does).

Examples:
  olares-cli settings vpn devices tags set 7 --tag ops --tag laptop
  olares-cli settings vpn devices tags set 7                # clear all
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "set Headscale device tags"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runDevicesTagsSet(ctx, f, args[0], tags), "set Headscale device tags")
		},
	}
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "tag to set on the device (repeat --tag for multiple); pass zero --tag flags to clear")
	return cmd
}

func runDevicesTagsSet(ctx context.Context, f *cmdutil.Factory, deviceID string, tags []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return fmt.Errorf("device-id is required")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	wire := normalizeHeadscaleTags(tags)
	if err := doTagsSetViaDoer(ctx, pc.doer, deviceID, wire); err != nil {
		return err
	}
	if len(wire) == 0 {
		fmt.Fprintf(os.Stdout, "cleared tags on device %s\n", deviceID)
	} else {
		fmt.Fprintf(os.Stdout, "set tags on device %s: %s\n", deviceID, strings.Join(wire, ","))
	}
	return nil
}

// doTagsSetViaDoer is the wire-only core of `vpn devices tags set`.
// Callers should normalizeHeadscaleTags(...) first so the wire payload
// uses the upstream "tag:<name>" form; this helper does NOT re-normalize
// to keep the test surface honest about what's actually sent.
func doTagsSetViaDoer(ctx context.Context, d Doer, deviceID string, prefixedTags []string) error {
	body := map[string][]string{"tags": prefixedTags}
	path := "/headscale/machine/" + url.PathEscape(deviceID) + "/tags"
	return d.DoJSON(ctx, "POST", path, body, nil)
}

// normalizeHeadscaleTags applies the SPA's tag-formatting convention:
//
//   - trim whitespace
//   - drop empties
//   - de-dupe (preserves first-seen order)
//   - prefix each survivor with "tag:" unless the caller already did
//
// Callers may pass either the bare name ("ops") or the prefixed form
// ("tag:ops"); both end up as "tag:ops" on the wire, which is what
// Headscale stores.
func normalizeHeadscaleTags(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if !strings.HasPrefix(t, "tag:") {
			t = "tag:" + t
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
