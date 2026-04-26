package apps

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps suspend <name>` and `olares-cli settings apps resume <name>`.
//
// These verbs map onto user-service's app.controller.ts:
//
//   GET  /api/app/resume/:name
//   POST /api/app/suspend     body: { "name": <name>, "all": false }
//
// Note the asymmetry: resume is a GET-with-name-in-the-path, while suspend
// is a POST with a JSON body (the SPA's stores/settings/application.ts
// resume() helper matches; the SPA's suspend() uses GET /api/app/suspend/<name>
// which user-service does NOT expose — so we mirror what's actually wired
// in the controller, not the SPA's broken call shape). user-service forwards
// each request to app-service /v1/apps/<name>/{resume,suspend} and returns the
// inner BFL envelope unchanged.
//
// The suspend body's `all` flag is a future-proofing knob that user-service
// hasn't actually plumbed end-to-end (app-service ignores it). We send
// `false` to match the SPA's intent of "suspend just this one app"; an
// `--all` flag on the CLI would be misleading until app-service wires it.
//
// Role: any user with edit rights on the application can suspend/resume
// it. Most cluster setups gate the SPA-side button on `isAdmin`, so a
// soft preflight at the admin level is appropriate. We do NOT use a hard
// gate because the server-side check is the source of truth.

// NewSuspendCommand returns `settings apps suspend <name>`.
func NewSuspendCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend <name>",
		Short: "suspend a running app (POST /api/app/suspend)",
		Long: `Suspend an installed app on the active profile's Olares.

Suspending freezes the app's pods (scale to 0 in app-service terms) without
deleting any persistent state — resume restores them in place. This is the
same action surfaced as "Suspend" on the SPA's per-app Settings page.

The <name> argument is the app's machine-readable name (the same value
shown in the NAME column of "olares-cli settings apps list").
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppSuspend(c.Context(), f, args[0])
		},
	}
	return cmd
}

// NewResumeCommand returns `settings apps resume <name>`.
func NewResumeCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <name>",
		Short: "resume a suspended app (GET /api/app/resume/<name>)",
		Long: `Resume a previously suspended app on the active profile's Olares.

Counterpart to "settings apps suspend": rehydrates the app's pods so its
entrances become reachable again. This is the same action surfaced as
"Resume" on the SPA's per-app Settings page.

The <name> argument is the app's machine-readable name (the same value
shown in the NAME column of "olares-cli settings apps list").
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppResume(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runAppSuspend(ctx context.Context, f *cmdutil.Factory, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("suspend requires a non-empty app name")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"name": name,
		"all":  false,
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/app/suspend", body, nil); err != nil {
		return err
	}
	fmt.Printf("Suspended app %q.\n", name)
	return nil
}

func runAppResume(ctx context.Context, f *cmdutil.Factory, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("resume requires a non-empty app name")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/api/app/resume/" + url.PathEscape(name)
	if err := doMutateEnvelope(ctx, pc.doer, "GET", path, nil, nil); err != nil {
		return err
	}
	fmt.Printf("Resumed app %q.\n", name)
	return nil
}
