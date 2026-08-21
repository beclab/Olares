package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router key …` — the credential other software uses.
//
// GET    /console/api/api-keys
// POST   /console/api/api-keys
// PATCH  /console/api/api-keys/:id
// DELETE /console/api/api-keys/:id
//
// A key is what turns Router into something an application can call. The
// console plane this CLI talks to is authenticated by the Olares session and
// needs no key; Router's /v1 data plane accepts only `Authorization: Bearer
// sk-…` or a platform-injected caller app. So the reason to issue a key is
// always to hand it to something else — a script, a container, a service
// outside Olares.
//
// The plaintext exists once, in the answer to the create call. Router keeps
// only its hash, so a key that was not recorded then cannot be recovered and
// has to be replaced.

type apiKeyView struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	UserID        string     `json:"user_id"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	IsExpired     bool       `json:"is_expired"`
	AllowedModels []string   `json:"allowed_models,omitempty"`
}

// createdKey is the create answer: the view plus the one plaintext.
type createdKey struct {
	apiKeyView
	Key string `json:"key"`
}

func NewKeyCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "API keys for software that calls Router",
		Long: `Issue and manage the sk- keys other software uses to call models.

Configuring Router needs no key: its console plane trusts your Olares session.
A key is what you hand to something that has no Olares session of its own — a
script, a container, a service elsewhere. Applications installed on this Olares
are recognised by the platform and do not need one either.

"router call" needs no key either: the platform vouches for you the same way it
does for the console plane. A key is worth issuing for it when the call has to
carry a budget or a model allowlist of its own, or comes from somewhere the
platform cannot vouch for.

The plaintext is shown once, when the key is issued, because Router stores only
its hash. Nothing can recover it afterwards; a lost key is replaced, not looked
up.

Subcommands:
  list             the keys you can see
  issue <name>     create one and print the plaintext once
  update <key>     rename, enable, disable, re-expire, or change its models
  revoke <key>     stop it working, keeping the row and its usage history
  current          which credential "router call" would present right now

A key can be restricted to named models, which is how a key given to one
application stops being a key to everything. Without that restriction it reaches
every model in the workspace.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newKeyListCommand(f))
	cmd.AddCommand(newKeyIssueCommand(f))
	cmd.AddCommand(newKeyUpdateCommand(f))
	cmd.AddCommand(newKeyRevokeCommand(f))
	cmd.AddCommand(newKeyCurrentCommand(f))
	cmd.AddCommand(newDeprecatedKeyLocalCommand(f))
	return cmd
}

func newKeyCurrentCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		forget bool
	)
	cmd := &cobra.Command{
		Use:   "current",
		Short: "which credential \"router call\" would present right now",
		Long: `Say what the next "router call" will authenticate with.

Usually nothing: Router takes the caller's identity from the platform, the same
way the console plane does, so a call carries no key at all and is attributed to
you. A key is presented only when one is named — --api-key, or the
OLARES_ROUTER_API_KEY environment variable.

An older olares-cli saved a key of its own here on first use. Calls no longer
use it, but it is still a working, unrestricted key in Router: "router key list"
shows it and "router key revoke" is what ends it.

--forget deletes this machine's copy and nothing else. Router keeps only the
hash, so once the copy is gone the plaintext is unrecoverable — a key that
should stop working has to be revoked, and a revoked key cannot be un-revoked.
This command never revokes anything on its own.

Examples:
  olares-cli router key current
  olares-cli router key current --forget
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runKeyLocal(c.Context(), f, forget, output)
		},
	}
	cmd.Flags().BoolVar(&forget, "forget", false, "delete this machine's saved copy; the key itself stays valid")
	addOutputFlag(cmd, &output)
	return cmd
}

func runKeyLocal(ctx context.Context, f *cmdutil.Factory, forget bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	if forget {
		if err := forgetDataPlaneKey(rp.OlaresID); err != nil {
			return fmt.Errorf("drop the saved key: %w", err)
		}
		fmt.Printf("forgot the saved key for %s. Calls were not using it anyway, and it still works: "+
			"`olares-cli router key list` shows it and `router key revoke` is what ends it. "+
			"This copy was the only plaintext — Router stores a hash — so it cannot be read back.\n",
			rp.OlaresID)
		return nil
	}

	stored, gerr := cachedDataPlaneKey(rp.OlaresID)
	state := map[string]any{
		"profile": rp.OlaresID,
		"saved":   stored != "",
		// What the next call actually presents. Reusing the calling path's own
		// vocabulary keeps this from becoming a second opinion about it.
		"credential": string(resolveDataPlaneAuth("").Mode),
	}
	if stored != "" {
		state["key_prefix"] = keyPrefixOf(stored)
	}
	if gerr != nil {
		state["keychain_error"] = gerr.Error()
	}
	if env := strings.TrimSpace(os.Getenv(dataPlaneKeyEnv)); env != "" {
		state["env_override"] = dataPlaneKeyEnv
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, state)
	}
	return renderKeyLocal(os.Stdout, state)
}

func renderKeyLocal(w io.Writer, state map[string]any) error {
	env, _ := state["env_override"].(string)
	keyed := state["credential"] == string(authKey)
	t := newTable(w)
	t.row("PROFILE", fmt.Sprintf("%v", state["profile"]))
	if keyed {
		t.row("CALLS USE", "the key in "+env)
	} else {
		t.row("CALLS USE", "your platform identity, no key")
	}
	saved, _ := state["saved"].(bool)
	t.row("SAVED KEY", boolStr(saved))
	if p, ok := state["key_prefix"].(string); ok {
		t.row("PREFIX", p)
	}
	if e, ok := state["keychain_error"].(string); ok {
		t.row("KEYCHAIN", "unreadable: "+e)
	}
	if err := t.flush(); err != nil {
		return err
	}
	switch {
	case keyed:
		_, err := fmt.Fprintf(w, "\n%s is set, so calls present that key. Unset it to call as yourself again.\n", env)
		return err
	case saved:
		_, err := fmt.Fprintln(w, "\nThe saved key is left over from an older olares-cli and is no longer used, "+
			"but it still works in Router: `router key list` shows it, `router key revoke` ends it, "+
			"and --forget only discards this copy.")
		return err
	default:
		_, err := fmt.Fprintln(w, "\nNothing to manage here. --api-key or OLARES_ROUTER_API_KEY is how a call "+
			"presents a key instead.")
		return err
	}
}

// keyPrefixOf shows enough of a key to match it against a row in `key list`
// without reproducing the secret. Router's own prefix is the first characters,
// so the same slice lines the two up.
func keyPrefixOf(key string) string {
	const shown = 11
	if len(key) <= shown {
		return key
	}
	return key[:shown] + "…"
}

func newKeyListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "the keys you can see",
		Long: `List API keys, newest first.

An admin sees every key a person holds. Anyone else sees their own. Neither sees
the keys Router issues to applications for itself — those belong to the app rows
and are not listed anywhere in this tree. What an application has actually spent
is still visible: "olares-cli router usage summary --by caller_app".

MODELS says what a key may reach: "all" means every model in the workspace, and
anything else is the allowlist it was given.

Examples:
  olares-cli router key list
  olares-cli router key list -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runKeyList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runKeyList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	keys, err := listKeys(ctx, pc)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": keys})
	}
	return renderKeyList(ctx, pc, os.Stdout, keys)
}

func listKeys(ctx context.Context, pc *preparedClient) ([]apiKeyView, error) {
	return collection[apiKeyView](ctx, pc, epAPIKeys)
}

func renderKeyList(ctx context.Context, pc *preparedClient, w io.Writer, keys []apiKeyView) error {
	if len(keys) == 0 {
		_, err := fmt.Fprintln(w, "no keys. `olares-cli router key issue <name>` creates one; "+
			"you only need one for software that has no Olares session of its own.")
		return err
	}
	users := userLabels(ctx, pc)
	t := newTable(w, "NAME", "PREFIX", "OWNER", "STATE", "EXPIRES", "LAST USED", "MODELS")
	for i := range keys {
		k := &keys[i]
		owner := k.UserID
		if label, ok := users[owner]; ok && label != "" {
			owner = label
		}
		models := "all"
		if len(k.AllowedModels) > 0 {
			models = strings.Join(k.AllowedModels, ", ")
		}
		t.row(
			nonEmpty(k.Name), nonEmpty(k.KeyPrefix), nonEmpty(owner), keyState(k),
			keyExpiry(k), keyLastUsed(k), clip(models, 48))
	}
	if err := t.flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "\nPREFIX is the only part of a key Router keeps, and is how a key in a log or a config is identified.")
	return err
}

// keyState folds status and expiry into the one answer a caller wants, since an
// active key past its expiry does not work and reads as active in the raw row.
//
// Disabled and revoked are one state, not two: Router stamps revoked_at
// whichever route asked, and both are undone by enabling the key again.
func keyState(k *apiKeyView) string {
	switch {
	case k.Status == "disabled":
		return "disabled"
	case k.IsExpired:
		return "expired"
	default:
		return nonEmpty(k.Status)
	}
}

func keyExpiry(k *apiKeyView) string {
	if k.ExpiresAt == nil {
		return "never"
	}
	return k.ExpiresAt.Local().Format("2006-01-02 15:04")
}

func keyLastUsed(k *apiKeyView) string {
	if k.LastUsedAt == nil {
		return "never"
	}
	return k.LastUsedAt.Local().Format("2006-01-02 15:04")
}

func newKeyIssueCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		forUser   string
		ttl       string
		expiresAt string
		models    []string
	)
	cmd := &cobra.Command{
		Use:   "issue <name>",
		Short: "create a key and print the plaintext once",
		Long: `Issue an API key.

The plaintext is printed once and never again — Router keeps only its hash. Put
it where it is going before you close the terminal.

The name is for you: it is what identifies the key in this list and in an audit
entry, so name it after what will hold it rather than after yourself.

--model restricts the key to what it names. Two kinds of name may be given, and
they grant different things. "<provider>/<model>" grants one backend and nothing
else. A route name — an alias, a group, or a default category — grants the name:
whatever serves it today, and whatever an admin attaches to it tomorrow. Grant a
route when the key should follow a decision somebody else is making, and a
qualified name when it should not.

Names are checked against what is configured, because Router only checks the
shape: a typo otherwise produces a key that authenticates and then cannot reach
anything. Without --model the key reaches every model in the workspace.

--ttl and --expires-at set an expiry; without either the key never expires.
--for-user issues on someone else's behalf and is admin-only.

Examples:
  olares-cli router key issue "wise indexer"
  olares-cli router key issue ci --ttl 30d --model Olares/qwen3-8b
  olares-cli router key issue app --model default-chat --model default-embedding
  olares-cli router key issue bot --expires-at 2027-01-01T00:00:00Z -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runKeyIssue(c.Context(), f, args[0], forUser, ttl, expiresAt, models, output)
		},
	}
	cmd.Flags().StringVar(&forUser, "for-user", "", "issue for this Olares user instead of yourself (admin only)")
	cmd.Flags().StringVar(&ttl, "ttl", "", "expire this long from now, e.g. 30d, 12h, 90m")
	cmd.Flags().StringVar(&expiresAt, "expires-at", "", "expire at this RFC3339 instant, e.g. 2027-01-01T00:00:00Z")
	cmd.Flags().StringArrayVar(&models, "model", nil,
		"restrict to this model, as <provider>/<model>, or to a route name; repeatable")
	addOutputFlag(cmd, &output)
	return cmd
}

type createKeyRequest struct {
	Name          string   `json:"name"`
	ForUserID     string   `json:"for_user_id,omitempty"`
	ExpiresAt     *string  `json:"expires_at,omitempty"`
	TTLSeconds    *int64   `json:"ttl_seconds,omitempty"`
	AllowedModels []string `json:"allowed_models,omitempty"`
}

func runKeyIssue(ctx context.Context, f *cmdutil.Factory, name, forUser, ttl, expiresAt string, models []string, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("a key name is required; it is how you will recognise this key later")
	}
	if strings.TrimSpace(ttl) != "" && strings.TrimSpace(expiresAt) != "" {
		return fmt.Errorf("pass either --ttl or --expires-at, not both")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	req := createKeyRequest{Name: name}
	if s := strings.TrimSpace(expiresAt); s != "" {
		when, perr := parseInstant(s)
		if perr != nil {
			return perr
		}
		iso := when.UTC().Format(time.RFC3339)
		req.ExpiresAt = &iso
	}
	if s := strings.TrimSpace(ttl); s != "" {
		d, perr := parseTTL(s)
		if perr != nil {
			return perr
		}
		secs := int64(d.Seconds())
		req.TTLSeconds = &secs
	}
	if len(models) > 0 {
		allowed, verr := verifyAllowedModels(ctx, pc, models)
		if verr != nil {
			return verr
		}
		req.AllowedModels = allowed
	}
	if s := strings.TrimSpace(forUser); s != "" {
		id, uerr := resolveUserID(ctx, pc, s)
		if uerr != nil {
			return uerr
		}
		req.ForUserID = id
	}

	var created createdKey
	if err := pc.router.doJSON(ctx, "POST", epAPIKeys, req, &created); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, created)
	}
	return renderCreatedKey(os.Stdout, &created)
}

func renderCreatedKey(w io.Writer, k *createdKey) error {
	if _, err := fmt.Fprintf(w, "%s\n\n", k.Key); err != nil {
		return err
	}
	t := newTable(w)
	t.row("NAME", nonEmpty(k.Name))
	t.row("PREFIX", nonEmpty(k.KeyPrefix))
	t.row("EXPIRES", keyExpiry(&k.apiKeyView))
	models := "all models in this workspace"
	if len(k.AllowedModels) > 0 {
		models = strings.Join(k.AllowedModels, ", ")
	}
	t.row("REACHES", models)
	if err := t.flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "\nThe key above is shown once. Router stores only its hash, so it cannot be read back — "+
		"issue a new one if it is lost.")
	return err
}

func newKeyUpdateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		name        string
		enable      bool
		disable     bool
		ttl         string
		expiresAt   string
		noExpiry    bool
		models      []string
		clearModels bool
	)
	cmd := &cobra.Command{
		Use:   "update <key>",
		Short: "rename, enable, disable, re-expire, or change a key's models",
		Long: `Change a key without replacing it.

The key keeps working through all of this — its plaintext does not change — so
this is the way to correct a mistake, extend a deadline, or narrow what something
reaches, without redeploying whatever holds it.

--enable brings back a key that was disabled or revoked; those are one state in
Router, and nothing about either is permanent. --no-expiry removes an expiry;
--clear-models removes an allowlist, and a key with no allowlist reaches
everything.

--model replaces the allowlist rather than adding to it, and takes either a
"<provider>/<model>" or a route name, as "key issue" does.

Name the key by its prefix, its name, or its id.

Examples:
  olares-cli router key update sk-abc123 --disable
  olares-cli router key update "wise indexer" --ttl 90d
  olares-cli router key update ci --model "openai/gpt-4o-mini" --model "openai/gpt-4o"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if enable && disable {
				return fmt.Errorf("pass either --enable or --disable, not both")
			}
			if noExpiry && (strings.TrimSpace(ttl) != "" || strings.TrimSpace(expiresAt) != "") {
				return fmt.Errorf("pass either --no-expiry or one of --ttl / --expires-at, not both")
			}
			if strings.TrimSpace(ttl) != "" && strings.TrimSpace(expiresAt) != "" {
				return fmt.Errorf("pass either --ttl or --expires-at, not both")
			}
			if clearModels && len(models) > 0 {
				return fmt.Errorf("pass either --clear-models or --model, not both")
			}
			return runKeyUpdate(c.Context(), f, args[0], keyPatchInput{
				Name:        name,
				Enable:      enable,
				Disable:     disable,
				TTL:         ttl,
				ExpiresAt:   expiresAt,
				NoExpiry:    noExpiry,
				Models:      models,
				ClearModels: clearModels,
			}, output)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "rename the key")
	cmd.Flags().BoolVar(&enable, "enable", false, "make the key usable again")
	cmd.Flags().BoolVar(&disable, "disable", false, "stop the key working, reversibly")
	cmd.Flags().StringVar(&ttl, "ttl", "", "expire this long from now, e.g. 30d")
	cmd.Flags().StringVar(&expiresAt, "expires-at", "", "expire at this RFC3339 instant")
	cmd.Flags().BoolVar(&noExpiry, "no-expiry", false, "remove the expiry")
	cmd.Flags().StringArrayVar(&models, "model", nil,
		"replace the allowlist with these, as <provider>/<model> or a route name; repeatable")
	cmd.Flags().BoolVar(&clearModels, "clear-models", false, "remove the allowlist, letting the key reach every model")
	addOutputFlag(cmd, &output)
	return cmd
}

type keyPatchInput struct {
	Name        string
	Enable      bool
	Disable     bool
	TTL         string
	ExpiresAt   string
	NoExpiry    bool
	Models      []string
	ClearModels bool
}

type patchKeyRequest struct {
	Name          *string   `json:"name,omitempty"`
	Status        *string   `json:"status,omitempty"`
	ExpiresAt     *string   `json:"expires_at,omitempty"`
	ClearExpiry   bool      `json:"clear_expiry,omitempty"`
	AllowedModels *[]string `json:"allowed_models,omitempty"`
}

func runKeyUpdate(ctx context.Context, f *cmdutil.Factory, ref string, in keyPatchInput, outputRaw string) error {
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
	found, err := resolveKey(ctx, pc, ref)
	if err != nil {
		return err
	}

	var req patchKeyRequest
	if s := strings.TrimSpace(in.Name); s != "" {
		req.Name = &s
	}
	if in.Enable {
		s := "active"
		req.Status = &s
	}
	if in.Disable {
		s := "disabled"
		req.Status = &s
	}
	if in.NoExpiry {
		req.ClearExpiry = true
	}
	if s := strings.TrimSpace(in.ExpiresAt); s != "" {
		when, perr := parseInstant(s)
		if perr != nil {
			return perr
		}
		iso := when.UTC().Format(time.RFC3339)
		req.ExpiresAt = &iso
	}
	if s := strings.TrimSpace(in.TTL); s != "" {
		d, perr := parseTTL(s)
		if perr != nil {
			return perr
		}
		iso := time.Now().Add(d).UTC().Format(time.RFC3339)
		req.ExpiresAt = &iso
	}
	if in.ClearModels {
		empty := []string{}
		req.AllowedModels = &empty
	}
	if len(in.Models) > 0 {
		allowed, verr := verifyAllowedModels(ctx, pc, in.Models)
		if verr != nil {
			return verr
		}
		req.AllowedModels = &allowed
	}
	if req.Name == nil && req.Status == nil && req.ExpiresAt == nil && !req.ClearExpiry && req.AllowedModels == nil {
		return fmt.Errorf("nothing to change; pass --name, --enable, --disable, --ttl, --expires-at, --no-expiry, --model or --clear-models")
	}

	var updated apiKeyView
	path := epAPIKey(found.ID)
	if err := pc.router.doJSON(ctx, "PATCH", path, req, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	models := "all models in this workspace"
	if len(updated.AllowedModels) > 0 {
		models = strings.Join(updated.AllowedModels, ", ")
	}
	_, err = fmt.Fprintf(os.Stdout, "%s (%s) is %s, expires %s, reaches %s\n",
		nonEmpty(updated.Name), nonEmpty(updated.KeyPrefix), keyState(&updated), keyExpiry(&updated), models)
	return err
}

func newKeyRevokeCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "revoke <key>",
		Short: "stop a key working, keeping its history",
		Long: `Revoke an API key.

Whatever holds this key starts being refused immediately, with no way to tell it
why, so revoke a key that is in use only when that is the point.

Nothing is destroyed. The row survives, and so does everything it spent —
"olares-cli router usage" still attributes past calls to it — and the key can be
brought back with "key update <key> --enable". Router keeps no harder deletion
than this: revoking is the same reversible disable that "key update --disable"
performs, spelled the way people ask for it and with a prompt attached.

Confirmation is required. --yes skips the prompt and is mandatory when stdin is
not a terminal.

Examples:
  olares-cli router key revoke sk-abc123 --yes
  olares-cli router key revoke "wise indexer"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runKeyRevoke(c.Context(), f, args[0], assumeYes, output)
		},
	}
	addConfirmFlag(cmd, &assumeYes)
	addOutputFlag(cmd, &output)
	return cmd
}

func runKeyRevoke(ctx context.Context, f *cmdutil.Factory, ref string, assumeYes bool, outputRaw string) error {
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
	found, err := resolveKey(ctx, pc, ref)
	if err != nil {
		return err
	}
	if !assumeYes {
		used := "it has never been used"
		if found.LastUsedAt != nil {
			used = "it was last used " + found.LastUsedAt.Local().Format("2006-01-02 15:04")
		}
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Revoke key %q (%s)? Whatever holds it starts being refused; %s.",
				found.Name, found.KeyPrefix, used),
			false); err != nil {
			return err
		}
	}
	var revoked apiKeyView
	path := epAPIKey(found.ID)
	if err := pc.router.doJSON(ctx, "DELETE", path, nil, &revoked); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, revoked)
	}
	_, err = fmt.Fprintf(os.Stdout, "%s (%s) is revoked; its usage history is kept, and "+
		"`olares-cli router key update %s --enable` brings it back\n",
		nonEmpty(revoked.Name), nonEmpty(revoked.KeyPrefix), nonEmpty(revoked.KeyPrefix))
	return err
}

// resolveKey accepts an id, a key prefix, or a name. The prefix is what appears
// in a log or a config file, and the name is what a person remembers; only the
// id is in neither place, which is why it cannot be the only handle.
//
// A plaintext key is accepted too, by matching its prefix, so pasting the thing
// you are holding works. It is not sent anywhere.
func resolveKey(ctx context.Context, pc *preparedClient, ref string) (*apiKeyView, error) {
	ref, err := requireRef(ref, "a key name, prefix or id")
	if err != nil {
		return nil, err
	}
	keys, err := listKeys(ctx, pc)
	if err != nil {
		return nil, err
	}
	var byName []*apiKeyView
	known := make([]string, 0, len(keys))
	for i := range keys {
		k := &keys[i]
		if k.ID == ref || (k.KeyPrefix != "" && (k.KeyPrefix == ref || strings.HasPrefix(ref, k.KeyPrefix))) {
			return k, nil
		}
		if strings.EqualFold(k.Name, ref) {
			byName = append(byName, k)
		}
		known = append(known, nonEmpty(k.Name))
	}
	switch len(byName) {
	case 1:
		return byName[0], nil
	case 0:
		return nil, missing{
			noun:  "key",
			ref:   ref,
			known: known,
			have:  "the ones you can see are",
			none:  "you have no keys",
			note:  "`olares-cli router key issue <name>` creates one",
		}.err()
	default:
		prefixes := make([]string, 0, len(byName))
		for _, k := range byName {
			prefixes = append(prefixes, k.KeyPrefix)
		}
		return nil, fmt.Errorf("%d keys are named %q; name one by its prefix instead: %s",
			len(byName), ref, strings.Join(prefixes, ", "))
	}
}

// verifyAllowedModels checks each entry against what is actually configured.
// Router validates only the shape, so an unchecked typo yields a key that
// authenticates and then reaches nothing — a failure that shows up later, in
// something else's logs.
//
// An entry is either a qualified `<provider>/<model>` or a route name, which is
// how the data plane reads one, and the two are told apart by the slash: a route
// name never contains one and a qualified name always does. Both are accepted
// because they are different grants rather than two spellings of one. Neither is
// rewritten into the other — expanding a route into its current members would
// re-close it every time somebody moved it, which is the whole reason a key
// would name a route.
func verifyAllowedModels(ctx context.Context, pc *preparedClient, refs []string) ([]string, error) {
	rows, err := listAllModels(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("check the models named by --model: %w", err)
	}
	known := make(map[string]string, len(rows))
	qualified := make([]string, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		q := r.ProviderName + "/" + r.Model.Name
		if _, dup := known[strings.ToLower(q)]; !dup {
			qualified = append(qualified, q)
		}
		known[strings.ToLower(q)] = q
	}

	// Routes are read only when one is named, so restricting a key to models
	// costs the same round trips it always did.
	var routeNames map[string]string

	out := make([]string, 0, len(refs))
	seen := map[string]struct{}{}
	for _, raw := range refs {
		m := strings.TrimSpace(raw)
		var canonical string
		if strings.Contains(m, "/") {
			var ok bool
			canonical, ok = known[strings.ToLower(m)]
			if !ok {
				return nil, fmt.Errorf("no model %q is configured; `olares-cli router model list` shows what is, "+
					"and the name to use here is the provider and the model joined by a slash, e.g. %q",
					m, exampleQualified(qualified))
			}
		} else {
			if routeNames == nil {
				routes, rerr := listRoutes(ctx, pc, "")
				if rerr != nil {
					return nil, fmt.Errorf("check the route named by --model %q: %w", raw, rerr)
				}
				routeNames = make(map[string]string, len(routes))
				for i := range routes {
					routeNames[strings.ToLower(routes[i].Name)] = routes[i].Name
				}
			}
			var ok bool
			canonical, ok = routeNames[strings.ToLower(m)]
			if !ok {
				names := make([]string, 0, len(routeNames))
				for _, n := range routeNames {
					names = append(names, n)
				}
				sort.Strings(names)
				return nil, missing{
					noun:  "model or route",
					ref:   m,
					known: names,
					have:  "the routes are",
					none:  "no route exists",
					note: fmt.Sprintf("A model is named as <provider>/<model>, e.g. %q; a name without a "+
						"slash has to be a route.", exampleQualified(qualified)),
				}.err()
			}
		}
		if _, dup := seen[canonical]; dup {
			continue
		}
		seen[canonical] = struct{}{}
		out = append(out, canonical)
	}
	return out, nil
}

func exampleQualified(qualified []string) string {
	if len(qualified) > 0 {
		return qualified[0]
	}
	return "openai/gpt-4o-mini"
}

// parseTTL accepts Go's duration syntax plus a day suffix, because a key's
// lifetime is usually stated in days and "720h" is nobody's first thought.
func parseTTL(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if rest, ok := strings.CutSuffix(strings.ToLower(s), "d"); ok {
		var days float64
		if _, err := fmt.Sscanf(rest, "%g", &days); err == nil && days > 0 {
			return time.Duration(days * 24 * float64(time.Hour)), nil
		}
		return 0, fmt.Errorf("--ttl %q is not a number of days", s)
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("--ttl %q is not a duration; use 30d, 12h or 90m", s)
	}
	if d <= 0 {
		return 0, fmt.Errorf("--ttl %q is not in the future", s)
	}
	return d, nil
}

// parseInstant takes RFC3339, and also a bare date, which is what people type
// for a deadline. A bare date means the start of that day, in local time.
func parseInstant(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("%q is not a time; use 2027-01-01 or 2027-01-01T00:00:00Z", s)
}
