package files

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/share"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// shareInternalOptions captures the flags exclusive to
// `files share internal`. Members is parsed from a single comma-
// joined --users flag rather than a repeated flag because the web
// app's UI batches "select N users at once" → one POST, and a
// single comma-joined flag mirrors that one-shot semantic.
type shareInternalOptions struct {
	usersRaw string
	perm     string
}

// newShareInternalCommand: `share internal <path> [--users
// alice:edit,bob:view] [--permission admin]`.
//
// Creates an Internal share at the target path with the given owner
// permission, then optionally calls AddInternalMembers to grant the
// listed users access. The two calls are split server-side, so we
// surface the share id even if the member-add half fails — that
// way the user can retry without a second create.
func newShareInternalCommand(f *cmdutil.Factory) *cobra.Command {
	o := &shareInternalOptions{}
	cmd := &cobra.Command{
		Use:   "internal <remote-path>",
		Short: "create an Internal (cross-user) share for a folder or file",
		Long: `Create an Internal share — visible to other Olares users on the same
node — for a folder or file under the per-user files-backend.

Wire shape (two calls, both required when --users is given):

    POST /api/share/share_path/<fileType>/<extend><subPath>/
        body: {name, share_type:"internal", permission:<owner-perm>, password:""}
    POST /api/share/share_member/
        body: {path_id, share_members: [{share_member, permission}]}

The first call creates the share record (and yields the share id).
The second call grants the listed users access. If --users is not
passed, only the share record is created — useful when you want to
hand the id off to a workflow that adds members separately, or
when the share is private to its owner for now.

--users uses the format "name:perm" with multiple users joined by
",", e.g. "alice:edit,bob:view,charlie:admin". perm is one of
view / upload / edit / admin (or 0..4). Default per-user perm is
view.

--permission sets the OWNER's permission on the share record; the
sensible default (matching the web app) is admin (full control).

Examples:

    olares-cli files share internal drive/Home/Backups/
    olares-cli files share internal drive/Home/Reports/Q1.pdf \
        --users alice:edit,bob:view
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShareInternal(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.usersRaw, "users", "",
		"comma-joined member list: name[:perm],name[:perm]... (perm: view/upload/edit/admin or 0..4)")
	cmd.Flags().StringVar(&o.perm, "permission", "admin",
		"owner permission on the share record: view/upload/edit/admin (default admin, matches web app)")
	return cmd
}

func runShareInternal(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	pathArg string,
	o *shareInternalOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tgt, err := frontendPathToShareTarget(pathArg)
	if err != nil {
		return err
	}
	ownerPerm, err := share.ParsePermission(o.perm)
	if err != nil {
		return fmt.Errorf("--permission: %w", err)
	}
	if ownerPerm == share.PermNone {
		// Defense in depth: ParsePermission returns PermNone for
		// empty input; for owner permission "none" makes no sense
		// (it would lock the creator out of their own share).
		return errors.New("--permission must be view/upload/edit/admin (got 'none')")
	}

	members, err := parseShareMembers(o.usersRaw)
	if err != nil {
		return fmt.Errorf("--users: %w", err)
	}

	client, rp, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}

	res, err := client.Create(ctx, tgt, share.CreateOptions{
		Name:       shareNameFromPath(tgt),
		ShareType:  share.TypeInternal,
		Permission: ownerPerm,
		Password:   "",
	})
	if err != nil {
		return reformatShareHTTPErr(err, rp.OlaresID, "create internal share for "+pathArg)
	}
	fmt.Fprintf(out, "created internal share:\n")
	fmt.Fprintf(out, "  id        : %s\n", res.ID)
	fmt.Fprintf(out, "  path      : %s/%s%s\n", tgt.FileType, tgt.Extend, tgt.SubPath)
	fmt.Fprintf(out, "  owner     : %s (%s)\n", res.Owner, res.Permission)

	if len(members) == 0 {
		fmt.Fprintln(out, "  members   : (none — share is private until --users is added or `share internal` is re-run)")
		return nil
	}
	if err := client.AddInternalMembers(ctx, res.ID, members); err != nil {
		// Important: the share record DID get created. Surface the
		// id so the user can recover by calling AddInternalMembers
		// directly (or re-running with --users on a tighter set).
		return fmt.Errorf("share %s created, but adding members failed: %w (re-run `share internal <path> --users ...` once the issue is resolved)",
			res.ID,
			reformatShareHTTPErr(err, rp.OlaresID, "add internal share members"))
	}
	fmt.Fprintf(out, "  members   :\n")
	for _, m := range members {
		fmt.Fprintf(out, "    - %s  (%s)\n", m.ShareMember, m.Permission)
	}
	return nil
}

// sharePublicOptions captures the flags exclusive to `share public`.
// Either expireDays OR expireTime must be set (the web app forces
// the choice; we mirror it). Password defaults to a random 8-byte
// base32-ish value when omitted, mirroring the LarePass app's
// generatePassword(6) default — except we go a couple of bytes
// longer because a typed password on the CLI is friendlier with a
// little more entropy.
type sharePublicOptions struct {
	password        string
	expireDays      int
	expireTime      string
	uploadOnly      bool
	uploadSizeLimit string // raw "100M" / "1G" / etc.
}

// newSharePublicCommand: `share public <path> [--password] [--expire-days N | --expire-time RFC3339] [--upload-only] [--upload-size-limit 100M]`.
func newSharePublicCommand(f *cmdutil.Factory) *cobra.Command {
	o := &sharePublicOptions{}
	cmd := &cobra.Command{
		Use:     "public <remote-path>",
		Aliases: []string{"link"},
		Short:   "create a Public-link share for a folder or file",
		Long: `Create a Public-link share. The link is opaque and can be sent to
anyone who has the password; recipients open it through the
LarePass app at the share host's /sharable-link/<id>/ path.

Wire shape:

    POST /api/share/share_path/<fileType>/<extend><subPath>/
        body: {name, share_type:"external", permission:<edit|upload>,
               password, expire_in?|expire_time?, upload_size_limit?}

Required: --password OR auto-generated; AND one of --expire-days /
--expire-time. Public links without an expiration are not supported
by the backend.

Permission defaults to "edit" (recipients can read AND upload). Pass
--upload-only to lock recipients into "upload-only" mode (they can
drop files in but can't browse the shared folder).

--upload-size-limit accepts a human-readable size: 100M, 1G, 500K,
512 (raw bytes). Passing 0 or omitting the flag means no per-upload
cap.

Examples:

    # 7-day expiration, auto-generated password.
    olares-cli files share public drive/Home/Photos/ --expire-days 7

    # Explicit password, 30 days, 100 MB upload cap.
    olares-cli files share public drive/Home/Photos/ \
        --password "s3cret-pw-1" --expire-days 30 \
        --upload-size-limit 100M

    # Upload-only portal, expires at a specific UTC time.
    olares-cli files share public drive/Home/Inbox/ --upload-only \
        --password drop --expire-time 2026-12-31T23:59:00Z
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSharePublic(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.password, "password", "",
		"share access password; if omitted, an 8-character random password is generated and printed")
	cmd.Flags().IntVar(&o.expireDays, "expire-days", 0,
		"expiration window in days from now (mutually exclusive with --expire-time)")
	cmd.Flags().StringVar(&o.expireTime, "expire-time", "",
		"explicit expiration time as RFC3339 (e.g. 2026-12-31T23:59:00Z); mutually exclusive with --expire-days")
	cmd.Flags().BoolVar(&o.uploadOnly, "upload-only", false,
		"recipients can ONLY upload (no listing / download); useful for inbox-style links")
	cmd.Flags().StringVar(&o.uploadSizeLimit, "upload-size-limit", "",
		"per-upload size cap: 100M, 1G, 500K, or raw bytes (default unlimited)")
	return cmd
}

func runSharePublic(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	pathArg string,
	o *sharePublicOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tgt, err := frontendPathToShareTarget(pathArg)
	if err != nil {
		return err
	}
	if o.expireDays > 0 && o.expireTime != "" {
		return errors.New("--expire-days and --expire-time are mutually exclusive; pass exactly one")
	}
	if o.expireDays <= 0 && o.expireTime == "" {
		return errors.New("Public shares require an expiration; pass --expire-days N or --expire-time RFC3339")
	}

	password := o.password
	autoGenerated := false
	if password == "" {
		password, err = generatePassword(8)
		if err != nil {
			return fmt.Errorf("generate random password: %w", err)
		}
		autoGenerated = true
	}
	// Web app rule (passwordLimitRule in Public/public.ts): minimum
	// 6 chars. Mirror it client-side so the user gets a clean error
	// instead of a generic 4xx from the server.
	if len(password) < 6 {
		return errors.New("--password must be at least 6 characters")
	}

	perm := share.PermEdit
	if o.uploadOnly {
		perm = share.PermUpload
	}

	uploadLimit := int64(0)
	if o.uploadSizeLimit != "" {
		uploadLimit, err = parseSizeWithSuffix(o.uploadSizeLimit)
		if err != nil {
			return fmt.Errorf("--upload-size-limit: %w", err)
		}
	}

	opts := share.CreateOptions{
		Name:            shareNameFromPath(tgt),
		ShareType:       share.TypePublic,
		Permission:      perm,
		Password:        password,
		UploadSizeLimit: uploadLimit,
	}
	if o.expireDays > 0 {
		// Web app multiplies by 24 * 3600 * 1000 (ms). Match that
		// shape on the wire so the server's interval math behaves
		// identically (same TTL semantics regardless of which client
		// posted).
		opts.ExpireIn = int64(o.expireDays) * 24 * 3600 * 1000
	} else {
		// Validate parseability before sending so a typo doesn't
		// reach the server as an opaque 400.
		if _, err := time.Parse(time.RFC3339, o.expireTime); err != nil {
			return fmt.Errorf("--expire-time: must be RFC3339 (e.g. 2026-12-31T23:59:00Z); got %q: %w",
				o.expireTime, err)
		}
		opts.ExpireTime = o.expireTime
	}

	client, rp, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	res, err := client.Create(ctx, tgt, opts)
	if err != nil {
		return reformatShareHTTPErr(err, rp.OlaresID, "create public share for "+pathArg)
	}

	fmt.Fprintf(out, "created public share:\n")
	fmt.Fprintf(out, "  id              : %s\n", res.ID)
	fmt.Fprintf(out, "  path            : %s/%s%s\n", tgt.FileType, tgt.Extend, tgt.SubPath)
	fmt.Fprintf(out, "  permission      : %s\n", res.Permission)
	fmt.Fprintf(out, "  password        : %s", password)
	if autoGenerated {
		fmt.Fprintf(out, "  (auto-generated; copy now, the server doesn't echo it back on subsequent reads)")
	}
	fmt.Fprintln(out)
	if opts.ExpireIn > 0 {
		fmt.Fprintf(out, "  expires in (ms) : %d  (~%d days)\n", opts.ExpireIn, o.expireDays)
	}
	if opts.ExpireTime != "" {
		fmt.Fprintf(out, "  expires at      : %s\n", opts.ExpireTime)
	}
	if uploadLimit > 0 {
		fmt.Fprintf(out, "  upload limit    : %s\n", formatBytes(uploadLimit))
	}
	// We don't have shareBaseUrl on the CLI side (the web app derives
	// it from window.location.hostname; on the CLI we'd have to
	// reverse-engineer the share host). Tell the user how to compose
	// the link rather than trying to guess wrong.
	fmt.Fprintf(out, "  link template   : <share-host>/sharable-link/%s/\n", res.ID)
	fmt.Fprintln(out, "  (the share host is the LarePass-app share subdomain; the LarePass app constructs it from the user's hostname)")
	return nil
}

// shareSMBOptions captures the flags exclusive to `share smb`.
// publicSMB and the per-user list are mutually exclusive: the web
// app's SMB modal lets the user pick "specific users" OR "anyone on
// the local network", and we mirror that constraint at flag-parse
// time.
type shareSMBOptions struct {
	publicSMB bool
	readOnly  bool
	usersRaw  string // smb-account-id[:perm], joined by ","
}

// newShareSMBCommand: `share smb <path> [--public] [--read-only]
// [--users smb-id:edit,...]`.
func newShareSMBCommand(f *cmdutil.Factory) *cobra.Command {
	o := &shareSMBOptions{}
	cmd := &cobra.Command{
		Use:   "smb <remote-path>",
		Short: "create an SMB (Samba) network share for a folder",
		Long: `Create an SMB (Samba) share. The result includes a UNC-style smb_link
(e.g. \\smb-host\share-name) plus the smb_user / smb_password the
client should use to mount it.

Wire shape:

    POST /api/share/share_path/<fileType>/<extend><subPath>/
        body: {name, share_type:"smb", permission, password:"",
               expire_in:0, expire_time:"",
               users?: [{id, permission}], public_smb: <bool>}

Recipient model (mutually exclusive):

    --public         Anyone on the local network can mount the share
                     using the returned smb_user / smb_password. No
                     per-user list is sent.

    --users id:perm  Specific SMB accounts only. IDs are SMB-account
                     IDs (NOT Olares user names) — list available IDs
                     via ` + "`olares-cli files share smb-users list`" + ` and
                     create new ones via ` + "`smb-users create`" + `.
                     perm is view / edit (or 1 / 3); upload-only and
                     admin do not apply to SMB shares.

Permission shape:

    --read-only      Force every member's effective permission to
                     view (read-only). Default (without this flag) is
                     edit (read+write). Web app's "readOnly" toggle.

SMB shares don't take a password (the password lives on the
returned record per share, not in the request body) and don't take
an expiration — they're mounted-resource-style and live until you
` + "`share rm`" + ` them.

Examples:

    olares-cli files share smb drive/Home/Movies/ --public

    olares-cli files share smb drive/Home/Backups/ \
        --users smb-uid-1:edit,smb-uid-2:view

    olares-cli files share smb drive/Home/Reports/ \
        --users smb-uid-1 --read-only
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShareSMB(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().BoolVar(&o.publicSMB, "public", false,
		"public-SMB mode: anyone on the local network can mount the share (mutually exclusive with --users)")
	cmd.Flags().BoolVar(&o.readOnly, "read-only", false,
		"force every member's permission to view (read-only); default is edit (read+write)")
	cmd.Flags().StringVar(&o.usersRaw, "users", "",
		"comma-joined SMB-account list: id[:perm],id[:perm]... (perm: view/edit, default edit)")
	return cmd
}

func runShareSMB(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	pathArg string,
	o *shareSMBOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tgt, err := frontendPathToShareTarget(pathArg)
	if err != nil {
		return err
	}
	if o.publicSMB && o.usersRaw != "" {
		return errors.New("--public and --users are mutually exclusive (public-SMB mode does not take a member list)")
	}
	if !o.publicSMB && o.usersRaw == "" {
		return errors.New("`share smb` requires either --public OR --users (the share has to be visible to someone)")
	}

	users, err := parseSMBUsers(o.usersRaw, o.readOnly)
	if err != nil {
		return fmt.Errorf("--users: %w", err)
	}

	// Per the web app's createSMBShare logic
	// (components/files/share/SMB/smb.ts L114-135):
	//
	//   - public_smb=true → permission=Edit, no users in body
	//   - public_smb=false + read-only → permission=View
	//   - public_smb=false + read-write → permission=Edit
	//
	// We mirror that mapping exactly so the server sees the same
	// shape regardless of whether the request came from the web
	// app or the CLI.
	perm := share.PermEdit
	if !o.publicSMB && o.readOnly {
		perm = share.PermView
	}

	publicSMB := o.publicSMB
	opts := share.CreateOptions{
		Name:       shareNameFromPath(tgt),
		ShareType:  share.TypeSMB,
		Permission: perm,
		Password:   "",
		ExpireIn:   0,
		ExpireTime: "",
		PublicSMB:  &publicSMB,
	}
	if !o.publicSMB {
		opts.Users = users
	}

	client, rp, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	res, err := client.Create(ctx, tgt, opts)
	if err != nil {
		return reformatShareHTTPErr(err, rp.OlaresID, "create SMB share for "+pathArg)
	}

	fmt.Fprintf(out, "created SMB share:\n")
	fmt.Fprintf(out, "  id            : %s\n", res.ID)
	fmt.Fprintf(out, "  path          : %s/%s%s\n", tgt.FileType, tgt.Extend, tgt.SubPath)
	fmt.Fprintf(out, "  permission    : %s\n", res.Permission)
	fmt.Fprintf(out, "  public access : %t\n", o.publicSMB)
	if res.SMBLink != "" {
		fmt.Fprintf(out, "  smb link      : %s\n", res.SMBLink)
	}
	if res.SMBUser != "" {
		fmt.Fprintf(out, "  smb user      : %s\n", res.SMBUser)
	}
	if res.SMBPassword != "" {
		fmt.Fprintf(out, "  smb password  : %s\n", res.SMBPassword)
	}
	if !o.publicSMB && len(users) > 0 {
		fmt.Fprintf(out, "  members       :\n")
		for _, u := range users {
			fmt.Fprintf(out, "    - %s  (%s)\n", u.ID, u.Permission)
		}
	}
	return nil
}

// parseShareMembers parses --users for `share internal`. Format:
//
//	"name1:perm1,name2:perm2,name3"
//
// Empty input → nil (no members; the cobra layer treats this as
// "create the share record only, no AddInternalMembers call").
//
// Each entry is split on the LAST ':' (not the first) so a name
// containing a ':' is unlikely to be misinterpreted — names with
// ':' aren't really a thing in Olares user names but defense in
// depth keeps the parser robust to a future schema change.
//
// Permission defaults to view if omitted ("alice" → alice:view).
// This matches the web app's selectedUsers default in
// components/files/share/Internal/internal.ts L130.
func parseShareMembers(raw string) ([]share.Member, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]share.Member, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		name, permStr := splitUserPerm(p)
		if name == "" {
			return nil, fmt.Errorf("entry %q has no user name (expected name[:perm])", p)
		}
		perm := share.PermView
		if permStr != "" {
			parsed, err := share.ParsePermission(permStr)
			if err != nil {
				return nil, fmt.Errorf("entry %q: %w", p, err)
			}
			if parsed != share.PermNone {
				perm = parsed
			}
		}
		out = append(out, share.Member{ShareMember: name, Permission: perm})
	}
	return out, nil
}

// parseSMBUsers parses --users for `share smb`. Same format as
// parseShareMembers but the entry is an SMB-account ID (not a user
// name) and the result is []share.SMBUser.
//
// readOnlyOverride forces every entry's permission to View; this
// matches the web app's --read-only flag, which doesn't take per-
// user perms when set (it's an all-or-nothing toggle).
func parseSMBUsers(raw string, readOnlyOverride bool) ([]share.SMBUser, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]share.SMBUser, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, permStr := splitUserPerm(p)
		if id == "" {
			return nil, fmt.Errorf("entry %q has no SMB-account id (expected id[:perm])", p)
		}
		perm := share.PermEdit
		if permStr != "" {
			parsed, err := share.ParsePermission(permStr)
			if err != nil {
				return nil, fmt.Errorf("entry %q: %w", p, err)
			}
			// SMB shares accept only View / Edit on the wire.
			// Reject upload / admin so the user gets a clean
			// error instead of an opaque server rejection.
			switch parsed {
			case share.PermView, share.PermEdit:
				perm = parsed
			case share.PermNone:
				// Fall through to the default perm.
			default:
				return nil, fmt.Errorf("entry %q: SMB shares accept only view or edit (got %s)", p, parsed)
			}
		}
		if readOnlyOverride {
			perm = share.PermView
		}
		out = append(out, share.SMBUser{ID: id, Permission: perm})
	}
	return out, nil
}

// splitUserPerm splits "name:perm" on the LAST ':'. "name" → ("name", "").
// Used by both parseShareMembers and parseSMBUsers.
func splitUserPerm(s string) (string, string) {
	if i := strings.LastIndex(s, ":"); i >= 0 {
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
	}
	return s, ""
}

// parseSizeWithSuffix turns "100M" / "1G" / "512" into a byte count.
// Suffixes are case-insensitive single letters (K/M/G/T) following
// the binary base (1024 multiplier — same as the web app's
// fileLimitSize() which uses 1024-based KiB/MiB/GiB/TiB despite the
// "M"/"G" labels). A bare integer with no suffix is treated as
// raw bytes.
//
// Negative values are rejected; zero is allowed (used by some
// callers as "no limit").
func parseSizeWithSuffix(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty size string")
	}
	mult := int64(1)
	last := s[len(s)-1]
	switch last {
	case 'K', 'k':
		mult = 1024
		s = s[:len(s)-1]
	case 'M', 'm':
		mult = 1024 * 1024
		s = s[:len(s)-1]
	case 'G', 'g':
		mult = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	case 'T', 't':
		mult = 1024 * 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("size must be non-negative (got %d)", n)
	}
	return n * mult, nil
}

// generatePassword returns a URL-safe random string of approximately
// `byteLen * 4 / 3` characters. Used as the auto-generated default
// for `share public --password`. We use crypto/rand because a
// guessable share password is the entire point of attack on a
// public link.
//
// base64.RawURLEncoding is chosen over base32 (or hex) because it's
// (a) URL-safe — important if the password ends up in a hand-copied
// share URL — and (b) shorter for the same entropy.
func generatePassword(byteLen int) (string, error) {
	if byteLen <= 0 {
		return "", errors.New("password byte length must be positive")
	}
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
