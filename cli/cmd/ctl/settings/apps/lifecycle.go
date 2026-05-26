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
//   POST /api/app/suspend     body: { "name": <name>, "all": <bool> }
//
// Note the asymmetry: resume is a GET-with-name-in-the-path, while suspend
// is a POST with a JSON body. user-service forwards each request to
// app-service /v1/apps/<name>/{resume,suspend} and returns the inner
// envelope unchanged.
//
// The suspend body's `all` flag is plumbed end-to-end: app-service's
// handler reads `StopRequest{All bool}` from the request body and
// writes ApplicationManager annotations (AppStopAllKey) accordingly.
// Semantically `all` toggles which half of a V2 cluster-shared app gets
// stopped — the shared server side (S-side / shared charts) or only the
// caller's user-side slice — not "how many users are affected":
//
//	`true`  → also stop the V2 cluster-shared server charts. Other
//	          users' client pods get stopped as a downstream consequence
//	          (the suspending_app reconciler walks the ApplicationManager
//	          list and stops every other owner's client because they
//	          can't function without the shared server); that propagation
//	          is a side effect of the server going down, not the primary
//	          intent of the flag.
//	`false` → only stop the caller's slice: a V1 app's user instance,
//	          or a V2 client. The shared server (and other users'
//	          clients) keep running.
//
// The CLI exposes the flag as `--all` and auto-picks a default by app
// scope when unset (CS app -> true, user-scoped -> false; one extra
// /api/myapps round-trip to resolve).
//
// Resume is read-only on the wire (GET, no body). app-service's resume
// handler decides `AppResumeAllKey` from the caller's isAdmin bit
// (admin -> all=true, non-admin -> all=false). There's no body to
// override that, so the CLI does NOT expose an `--all` flag on resume
// — it would be a misleading no-op.
//
// Role: any user with edit rights on the application can suspend/resume
// it. Most cluster setups gate the SPA-side button on `isAdmin`, so a
// soft preflight at the admin level is appropriate. We do NOT use a hard
// gate because the server-side check is the source of truth.

// NewSuspendCommand returns `settings apps suspend <name> [--all]`.
func NewSuspendCommand(f *cmdutil.Factory) *cobra.Command {
	var allFlag bool
	cmd := &cobra.Command{
		Use:   "suspend <name>",
		Short: "suspend a running app (POST /api/app/suspend)",
		Long: `Suspend an installed app on the active profile's Olares.

Suspending freezes the app's pods (scale to 0 in app-service terms) without
deleting any persistent state — resume restores them in place. This is the
same action surfaced as "Suspend" on the SPA's per-app Settings page.

The <name> argument is the app's machine-readable name (the same value
shown in the NAME column of "olares-cli settings apps list").

The --all flag controls which half of a V2 cluster-shared app gets stopped.
It maps to the body field "all" (forwarded by user-service to app-service's
StopRequest.All and stored on the ApplicationManager as AppStopAllKey).
Picture a V2 cluster-scoped app as two halves: the shared server side
("S-side", the cluster-shared charts that every user talks to) and each
user's own client side (the per-user pods that connect to that shared
server). --all picks which half is suspended:

  --all=true   also suspend the V2 cluster-shared server side (the
               "S-side" / shared charts). Because every other user's
               client of this app cannot function without the shared
               server, suspending the server side stops their client
               pods as a downstream side effect — that propagation is
               NOT the primary intent of the flag, just its consequence.
               Only meaningful for V2 cluster-scoped apps; for a V1 or
               user-scoped app there is no shared server to stop, so the
               server-side stop is a no-op and the call degrades to the
               same behavior as --all=false.

  --all=false  suspend only the caller's own slice: a V1 app's user
               instance or the V2 client (user-side) pods. The shared
               server (and therefore every other user's client) keeps
               running.

  (unset)      auto-pick: cluster-scoped apps default to --all=true,
               user-scoped apps default to --all=false. The auto path
               costs one extra "apps list" call to read isClusterScoped
               for the named app; pass the flag explicitly to skip it.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			var allOpt *bool
			if c.Flag("all").Changed {
				v := allFlag
				allOpt = &v
			}
			return runAppSuspend(c.Context(), f, args[0], allOpt)
		},
	}
	cmd.Flags().BoolVar(&allFlag, "all", false, "also suspend the V2 cluster-shared server side (S-side); other users' clients then stop as a downstream side effect. Auto-picks true for cluster-scoped apps and false for user-scoped when unset")
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

Note: there's no --all flag on resume. The wire is GET-with-no-body, and
app-service's resume handler decides AppResumeAllKey from the caller's
isAdmin role (admin -> resume across all users, non-admin -> only this
user's instance). Adding an --all flag here would be a silent no-op, so
we omit it. If you need finer control, use suspend with --all=true /
--all=false to gate the corresponding suspend half.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppResume(c.Context(), f, args[0])
		},
	}
	return cmd
}

// runAppSuspend posts /api/app/suspend with body {name, all}. allOpt
// nil means "auto" — resolve via the app's isClusterScoped flag (one
// extra GET /api/myapps); non-nil means "the user passed --all
// explicitly, use that value verbatim".
func runAppSuspend(ctx context.Context, f *cmdutil.Factory, name string, allOpt *bool) error {
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
	all, source, err := resolveSuspendAll(ctx, pc.doer, name, allOpt)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"name": name,
		"all":  all,
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/app/suspend", body, nil); err != nil {
		return err
	}
	fmt.Printf("Suspended app %q (all=%t, %s).\n", name, all, source)
	return nil
}

// resolveSuspendAll returns the final value of `all` plus a short
// human-readable explanation of where it came from. Split out so tests
// can drive the auto path without going through cobra.
func resolveSuspendAll(ctx context.Context, d Doer, name string, allOpt *bool) (bool, string, error) {
	if allOpt != nil {
		return *allOpt, fmt.Sprintf("--all=%t set explicitly", *allOpt), nil
	}
	scoped, err := lookupClusterScoped(ctx, d, name)
	if err != nil {
		// Surface the cause but make it actionable: the caller can
		// just retry with an explicit --all=true|false to skip the
		// auto-resolve path.
		return false, "", fmt.Errorf("resolve --all default for %q: %w (pass --all=true or --all=false to skip auto-detection)", name, err)
	}
	return scoped, fmt.Sprintf("auto from isClusterScoped=%t", scoped), nil
}

// lookupClusterScoped fetches /api/myapps and reports whether the
// named app is cluster-scoped. We share /api/myapps with `apps list`
// rather than introducing a new per-app endpoint — there isn't one,
// and the list is already cheap to round-trip (matches what apps/get.go
// does for the same reason).
func lookupClusterScoped(ctx context.Context, d Doer, name string) (bool, error) {
	var rows []appInfo
	if err := doGetEnvelope(ctx, d, "/api/myapps", &rows); err != nil {
		return false, err
	}
	for _, r := range rows {
		if r.Name == name {
			return r.IsClusterScoped, nil
		}
	}
	return false, fmt.Errorf("app %q not found in /api/myapps", name)
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
