package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/internal/files/edit"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// editOptions holds per-invocation flags for `files edit`. Kept
// small on purpose — `edit` is a focused verb (open editor, save
// back) and adding too many knobs blurs that contract.
type editOptions struct {
	// editor overrides the editor program. Without this we follow
	// the standard cascade $VISUAL → $EDITOR → "vi" (Unix) /
	// "notepad" (Windows). Same precedence `git commit` and
	// `crontab -e` use.
	editor string
	// contentType lets the caller pick a non-default Content-Type
	// for the PUT body. Default is text/plain (matches the web
	// app's saveFile / updateFile / put helpers). YAML / JSON /
	// markdown can carry a more specific type for any
	// content-aware caching layer between us and the storage
	// driver.
	contentType string
	// create allows editing a file that doesn't exist on the
	// server yet — start with an empty buffer instead of
	// erroring out. Without this flag, a 404 from /api/raw is a
	// hard error so a typo in the path doesn't silently land
	// content somewhere unexpected.
	create bool
	// keepTemp retains the temp file on no-change / error so the
	// user can recover whatever they typed. Without it we always
	// clean up — the no-change path is silent and the error
	// path's message tells you the temp file was removed.
	keepTemp bool
	// maxSize is the upper bound (in bytes) on both the
	// pre-edit remote size and the post-edit local size.
	// `edit` is meant for text-editing — config files, notes,
	// short logs — and a hard cap protects users from
	// accidentally streaming a multi-GB binary through their
	// editor (vim's "swap file" warning helps but doesn't stop
	// it) and from PUT-ing a runaway buffer back to the server.
	// Set to 0 to disable both checks; anything > 0 caps both
	// directions at that exact byte count. Default is
	// DefaultMaxSize (1 MiB) — comfortably above any reasonable
	// hand-edited config but tight enough that "I `cat`-ed an
	// image into vim" stops at the door.
	maxSize int64
	// allowBinary disables BOTH binary-detection checks
	// (extension deny-list AND post-fetch content sniff).
	// `edit` is for text formats — pictures, PDFs, archives,
	// executables produce a corrupted blob the moment $EDITOR
	// touches them, so the default policy refuses up-front.
	// Power users editing odd-but-real cases (UTF-16 with
	// embedded NULs, an .iso table-of-contents they really do
	// understand byte-by-byte) can pass --allow-binary to opt
	// out. The cap from --max-size still applies independently.
	allowBinary bool
}

// DefaultMaxSize is the default ceiling enforced by `files edit`
// on both the remote pre-edit size and the post-edit local size.
// 1 MiB is large enough to cover any realistic text-edit workflow
// (kubeconfigs, app.yaml, .env, multi-thousand-line markdown
// notes — the largest config we've seen in the wild is ~256 KiB)
// while bracketing out the "oops I tried to edit a binary" foot-
// gun before the editor even spawns. Override via --max-size.
const DefaultMaxSize int64 = 1 << 20

// binarySniffLen is the byte window we read from the head of the
// fetched buffer to decide whether the content looks binary. 8 KiB
// matches git's binary-detection window (`buffer_is_binary` in
// git/diff.c) and is large enough to catch the binary streams that
// follow a PDF / Office / archive header while still trivial for
// any text file regardless of encoding.
const binarySniffLen = 8 * 1024

// binaryExtensions is the deny-list of file suffixes we refuse to
// open in $EDITOR without --allow-binary. The set was distilled
// from the LarePass GUI's preview classifier (image / pdf / video
// / audio / blob) plus the obvious archive / executable / db
// formats that share the "edits in vi turn it into garbage" foot-
// gun. Extensions are stored lower-case; the lookup folds case at
// call time. Keep entries to single dot-suffixes here and put any
// compound suffixes (.tar.gz, .tar.xz, …) in compoundBinarySuffix
// below — filepath.Ext only returns the trailing component, so a
// flat-map check would miss the `.tar` part.
var binaryExtensions = map[string]struct{}{
	// images
	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".bmp": {},
	".webp": {}, ".tiff": {}, ".tif": {}, ".ico": {}, ".heic": {},
	".heif": {}, ".raw": {}, ".psd": {}, ".ai": {}, ".eps": {},
	// portable / office documents (binary container formats)
	".pdf": {}, ".doc": {}, ".docx": {}, ".xls": {}, ".xlsx": {},
	".ppt": {}, ".pptx": {}, ".odt": {}, ".ods": {}, ".odp": {},
	".pages": {}, ".numbers": {}, ".key": {}, ".rtf": {},
	".epub": {}, ".mobi": {}, ".azw": {}, ".azw3": {},
	// video
	".mp4": {}, ".m4v": {}, ".mov": {}, ".avi": {}, ".mkv": {},
	".webm": {}, ".flv": {}, ".wmv": {}, ".mpeg": {}, ".mpg": {},
	".3gp": {},
	// audio
	".mp3": {}, ".wav": {}, ".flac": {}, ".aac": {}, ".ogg": {},
	".m4a": {}, ".wma": {}, ".opus": {}, ".aiff": {}, ".aif": {},
	// archives & disk images
	".zip": {}, ".tar": {}, ".gz": {}, ".bz2": {}, ".xz": {},
	".7z": {}, ".rar": {}, ".tgz": {}, ".tbz2": {}, ".txz": {},
	".dmg": {}, ".iso": {}, ".img": {}, ".pkg": {}, ".deb": {},
	".rpm": {}, ".apk": {}, ".ipa": {}, ".cab": {}, ".msi": {},
	// executables / shared objects / bytecode
	".exe": {}, ".dll": {}, ".so": {}, ".dylib": {}, ".bin": {},
	".o": {}, ".a": {}, ".lib": {}, ".obj": {}, ".class": {},
	".jar": {}, ".war": {}, ".ear": {}, ".pyc": {}, ".pyo": {},
	".pyd": {}, ".wasm": {},
	// databases & on-disk indexes
	".db": {}, ".sqlite": {}, ".sqlite3": {}, ".mdb": {},
	".accdb": {}, ".pst": {}, ".ost": {},
	// fonts
	".ttf": {}, ".otf": {}, ".woff": {}, ".woff2": {}, ".eot": {},
}

// compoundBinarySuffix holds the multi-component archive suffixes
// filepath.Ext can't recognise on its own — same effect as
// binaryExtensions, just matched by HasSuffix on the lower-cased
// name.
var compoundBinarySuffix = []string{
	".tar.gz", ".tar.bz2", ".tar.xz", ".tar.zst", ".tar.lz",
}

// hasBinaryExtension reports whether `name` ends with a suffix the
// deny-list flags as binary. Lookup is case-insensitive (Windows
// users routinely type FOO.PDF) and respects compound suffixes
// like ".tar.gz" via the dedicated list. Pure-text formats that
// happen to live in extension-rich namespaces (.svg / .html /
// .xml / .ts (TypeScript) / .csv / .yaml) are intentionally NOT
// in the deny-list — the post-fetch content sniff will let them
// through if and only if the bytes really are textual.
func hasBinaryExtension(name string) bool {
	lower := strings.ToLower(name)
	for _, suf := range compoundBinarySuffix {
		if strings.HasSuffix(lower, suf) {
			return true
		}
	}
	ext := filepath.Ext(lower)
	if ext == "" {
		return false
	}
	_, ok := binaryExtensions[ext]
	return ok
}

// looksBinary reports whether the first up-to-binarySniffLen bytes
// of `buf` contain a NUL (0x00) byte. This is the same heuristic
// git, diff(1), and grep(1) use: real text never carries a NUL,
// and every binary container format we care about (PNG / JPEG /
// PDF / ELF / Mach-O / ZIP / Office .docx) hits one within its
// first kilobyte. Empty buffers (--create with no remote file)
// return false — there's nothing binary about an empty file.
func looksBinary(buf []byte) bool {
	n := len(buf)
	if n > binarySniffLen {
		n = binarySniffLen
	}
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

// NewEditCommand: `olares-cli files edit <remote-path>`
//
// Edit a file in place on the per-user files-backend. Wire flow:
//
//  1. GET /api/raw/<encPath>          — pull current bytes (or 404
//                                       → empty buffer when --create)
//  2. write to a temp file under $TMPDIR/olares-files-edit-*/
//  3. spawn $EDITOR on the temp file (foreground, inherits the
//     parent's stdin/stdout/stderr so vi / nano / hx all work)
//  4. PUT /api/resources/<encPath>    — only if the SHA-256 of the
//                                       post-edit bytes differs
//                                       from the pre-edit bytes
//
// This mirrors the LarePass web app's "Edit" affordance for text
// files — same endpoint pair (/api/raw to read, /api/resources to
// PUT), same Content-Type default ("text/plain"). The CLI difference
// is the editor handoff: the web app pops a Monaco-style code
// editor; we hand the bytes to whatever the user already trusts on
// their machine.
//
// Supported namespaces: drive / sync / cache / external + cloud
// drives awss3 / google / dropbox. Cloud drives are supported
// even though the LarePass GUI's `onSaveFile` plumbing has known
// wiring bugs there (see internal/files/edit's package docstring
// for the gory detail) — the underlying wire endpoint
// `PUT /api/resources/<fileType><subPath>` is uniform and the CLI
// hits it directly. tencent is intentionally excluded (its
// upload protocol diverges from the standard resources handler);
// share / internal are also refused — they're cross-user / read-
// only views in the LarePass UX with no documented write surface.
//
// Size cap: by default `edit` refuses to download or upload a
// file larger than 1 MiB (DefaultMaxSize) so the verb stays
// scoped to its real use case — text editing — and a typo like
// `files edit drive/Home/Photos/big.jpg` doesn't stream a 5 MB
// JPEG through the user's editor. Override via --max-size; pass
// `--max-size 0` to disable the check entirely.
//
// CLI semantics:
//
//   - <remote-path>: full 3-segment frontend path, identical to
//     `files cat` / `files download` (e.g.
//     `drive/Home/Documents/notes.md`). MUST point at a file —
//     a trailing '/' or a directory target is rejected before the
//     temp file is even created.
//   - $EDITOR is the cascade $VISUAL → $EDITOR → fallback. The
//     fallback is "vi" on POSIX and "notepad" on Windows. Pass
//     --editor to override.
//   - The temp file's BASENAME matches the remote basename (so
//     editor-side syntax detection picks up the right mode for
//     `.md` / `.json` / `.yaml`). It lives in a fresh
//     $TMPDIR/olares-files-edit-NNNNN/ directory which is rm'd on
//     exit (or retained when --keep-temp is set).
//   - No-change detection: we hash the pre- and post-edit bytes
//     with SHA-256 and skip the PUT when they match. This makes
//     `:q` / `:q!` workflows cheap and avoids touching the
//     server's modified-time when the user just looked at the file.
//
// Why no `--no-editor` / stdin mode here: piping bytes into a
// remote file is `olares-cli files upload <local> <remote>` (which
// also handles directories, resume, parallelism). `edit` is the
// verb for "open it in my editor"; making it a second upload path
// would dilute that contract.
func NewEditCommand(f *cmdutil.Factory) *cobra.Command {
	o := &editOptions{}
	cmd := &cobra.Command{
		Use:   "edit <remote-path>",
		Short: "edit a remote file in place via $EDITOR",
		Long: `Edit a single file on the per-user files-backend by opening it in $EDITOR.

The CLI fetches the file's current contents into a fresh temp file,
spawns ` + "`$EDITOR`" + ` on it, and PUTs the new contents back to the server
when you save. If you exit the editor without changes, no upload
happens — the no-change check is a SHA-256 comparison so it's
robust against editors that always rewrite the file (e.g. vim's
default backup behavior).

Editor cascade (matches ` + "`git commit`" + ` / ` + "`crontab -e`" + `):

    --editor flag  →  $VISUAL  →  $EDITOR  →  vi (POSIX) / notepad (Windows)

Wire shape:

    GET  /api/raw/<encPath>            → pull current bytes
    PUT  /api/resources/<encPath>      Content-Type: text/plain
                                       <body: full new contents>

Supported namespaces:

    drive/Home/<sub>/<file>
    drive/Data/<sub>/<file>
    sync/<repo_id>/<sub>/<file>
    cache/<node>/<sub>/<file>
    external/<node>/<volume>/<sub>/<file>
    awss3/<account>/<bucket>/<sub>/<file>
    google/<account>/<sub>/<file>
    dropbox/<account>/<sub>/<file>

Cloud drives (awss3 / google / dropbox) are supported on the wire
even though the LarePass GUI's "Save" flow has a known wiring
bug for them (it routes through drive's saveFile and misses the
cloud bucket). The CLI bypasses the GUI plumbing and PUTs
directly to ` + "`/api/resources/<fileType><subPath>`" + `, which is the
uniform write endpoint across every namespace the backend's
resources handler covers.

tencent is intentionally NOT supported — its upload-side protocol
diverges from the standard resources handler and we don't have a
wire-shape signoff that small-PUT edits are honored end-to-end.
share / internal are refused as cross-user / read-only views.

Size cap (default 1 MiB):

    By default ` + "`edit`" + ` refuses files larger than 1 MiB on either
    side of the editor. This bracket-checks the verb's intent
    (text editing — configs, notes, short logs) so a typo like
    ` + "`files edit drive/Home/Photos/big.jpg`" + ` doesn't accidentally
    pour a binary into your editor. Override with ` + "`--max-size <bytes>`" + `
    or pass ` + "`--max-size 0`" + ` to disable the check entirely.

Text-only policy:

    ` + "`edit`" + ` refuses non-text files in two layers, in the same
    spirit (and with the same heuristics) git, diff(1), and grep(1)
    use to detect binaries:

      1. Extension deny-list  — refuses common binary suffixes
                                (.jpg, .png, .pdf, .mp4, .zip,
                                .exe, .so, …) BEFORE Stat / fetch.
      2. NUL-byte content sniff — after the GET, if the first 8 KiB
                                of the buffer contains a NUL byte
                                we refuse to spawn the editor.

    Pure-text formats with binary-looking neighbors (.svg, .html,
    .xml, .csv, .yaml, .ts) pass the extension layer and the
    content sniff lets them through if the bytes really are
    textual. Pass ` + "`--allow-binary`" + ` to disable BOTH layers when
    you have a real reason (UTF-16 with embedded NULs, hand-
    auditing a small ELF, …).

Flags:

    --editor string         override the editor program (default $EDITOR cascade)
    --content-type string   PUT Content-Type header (default "text/plain")
    --create                start with an empty buffer if the file does not exist
    --keep-temp             retain the temp file on no-change / error for recovery
    --max-size int          max bytes for both the remote (pre-edit) and local
                            (post-edit) sizes; 0 disables the check (default 1 MiB)
    --allow-binary          disable the binary-content guard (extension deny-list
                            AND post-fetch NUL-byte sniff) — only use this when
                            you specifically want to edit non-text bytes

Examples:

    olares-cli files edit drive/Home/Documents/notes.md
    olares-cli files edit drive/Home/.config/app.yaml --editor nano
    olares-cli files edit sync/<repo_id>/Notes/draft.md
    olares-cli files edit drive/Home/new.txt --create
    olares-cli files edit awss3/<account>/<bucket>/config.json
    olares-cli files edit dropbox/<account>/Notes/draft.md
    olares-cli files edit google/<account>/Documents/.env
    olares-cli files edit drive/Home/Logs/today.log --max-size 5242880  # 5 MiB

Notes:

  - The temp file's basename matches the remote basename so
    editor-side syntax highlighting picks the right mode for
    .md / .json / .yaml / .ts / etc.
  - Saving an empty file is allowed (the wire endpoint accepts a
    zero-byte PUT). To bail out without saving anything, leave the
    file untouched (or delete its contents and exit — that's a
    real "make this an empty file" save).
  - Concurrent edits aren't coordinated: if someone else updates
    the same file between our GET and PUT, the PUT wins. There's
    no ETag / If-Match support on the wire to do better client-
    side, so this is consistent with the LarePass GUI's behavior.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(cmd.Context(), f, cmd.OutOrStdout(), os.Stdin, os.Stdout, os.Stderr, args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.editor, "editor", "",
		"editor program to spawn (default: $VISUAL / $EDITOR / vi)")
	cmd.Flags().StringVar(&o.contentType, "content-type", edit.DefaultContentType,
		"Content-Type header for the PUT body (default text/plain)")
	cmd.Flags().BoolVar(&o.create, "create", false,
		"start with an empty buffer when the remote file does not exist (404)")
	cmd.Flags().BoolVar(&o.keepTemp, "keep-temp", false,
		"retain the temp file on no-change / error for recovery")
	cmd.Flags().Int64Var(&o.maxSize, "max-size", DefaultMaxSize,
		"max bytes for both the remote (pre-edit) and local (post-edit) sizes; 0 disables the check")
	cmd.Flags().BoolVar(&o.allowBinary, "allow-binary", false,
		"disable the binary-content guard (extension deny-list and NUL-byte content sniff)")
	return cmd
}

// runEdit is the cobra-side glue for the edit verb. We split out
// the editor I/O readers / writers (rather than always using
// os.Stdin / os.Stdout / os.Stderr) so a test harness can swap in
// fakes — the editor child process itself still inherits the file
// handles, which keeps interactive editors like vim usable, but
// the temp-file writer / hasher logic is tested without a TTY.
func runEdit(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	editorStdin io.Reader,
	editorStdout io.Writer,
	editorStderr io.Writer,
	rawPath string,
	o *editOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Up-front TTY guard so a CI pipeline gets a clear error
	// instead of having the editor child process either hang
	// forever waiting for input or write garbage to a non-TTY.
	// Mirrors `rm`'s non-TTY refusal pattern (cli/cmd/ctl/files/rm.go).
	if !term.IsTerminal(int(syscall.Stdin)) {
		return errors.New(
			"refusing to spawn an editor without a TTY (no interactive stdin); " +
				"`files edit` is interactive — use `files download` + `files upload` for scripted edits")
	}

	tgt, err := frontendPathToEditTarget(rawPath)
	if err != nil {
		return err
	}

	op, err := edit.Plan(tgt)
	if err != nil {
		return err
	}

	// Pre-Stat extension guard (layer 1 of the text-only policy).
	// Stop here for the obvious "I tried to edit a JPEG / PDF /
	// .zip" case BEFORE we round-trip to the server. The
	// post-fetch NUL-byte sniff is the second, content-aware
	// layer for files whose extension lies (or is missing).
	if !o.allowBinary && hasBinaryExtension(op.DisplayPath) {
		return fmt.Errorf(
			"refusing to edit %s: extension looks like a non-text format "+
				"(images, PDFs, video, audio, archives, executables — `edit` "+
				"is for text). Use `files download` to copy it locally, or pass "+
				"--allow-binary if you really meant to open it in $EDITOR",
			op.DisplayPath)
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}

	// We use the download package to Stat the remote (parent-
	// listing strategy is the only one that doesn't trigger the
	// Content=true payload bug — see download/stat.go's docstring
	// for why a direct GET on /api/resources/<file> can return
	// HTTP 500 for many real files). Both clients share BaseURL
	// + HTTPClient, so this is a cheap projection rather than a
	// re-auth dance.
	statClient := &download.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}
	editClient := &edit.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}

	// Stat-then-fetch flow:
	//   - file exists  → fetch current bytes (download.Stat
	//                    rejects directories with a friendly
	//                    message before we touch the temp dir).
	//   - file missing → if --create, start with an empty buffer;
	//                    otherwise hard error with a "did you mean
	//                    --create?" hint.
	//
	// statAndFetch ALSO short-circuits the size cap: when
	// o.maxSize > 0 and Stat reports Size > maxSize we error out
	// BEFORE pulling bytes, so a typo like
	// `files edit drive/Home/Photos/big.jpg` doesn't waste a
	// multi-MB download just to refuse at the client.
	currentBytes, isDir, err := statAndFetch(ctx, statClient, editClient, op.DisplayPath, o.create, o.maxSize)
	if err != nil {
		return reformatEditHTTPErr(err, rp.OlaresID, "fetch", op.DisplayPath)
	}
	if isDir {
		return fmt.Errorf("%s is a directory: edit only works on files (use `files ls %s` to list it)",
			op.DisplayPath, op.DisplayPath)
	}

	// Post-fetch content sniff (layer 2 of the text-only policy).
	// Catches files whose extension lies — `myfile` (no extension)
	// that's actually a JPEG, `.log` files that are really ELF
	// core dumps, `.dat` blobs, etc. Empty buffers (--create with
	// a 404) trivially pass. Skipped under --allow-binary.
	if !o.allowBinary && looksBinary(currentBytes) {
		return fmt.Errorf(
			"refusing to edit %s: content looks binary (NUL byte in the first %d bytes); "+
				"`edit` is for text formats. Use `files download` to copy it locally, "+
				"or pass --allow-binary if you really meant to open it in $EDITOR",
			op.DisplayPath, binarySniffLen)
	}

	tmpDir, tmpFile, err := writeTempFile(op.DisplayPath, currentBytes)
	if err != nil {
		return err
	}
	cleaned := false
	defer func() {
		if cleaned || o.keepTemp {
			return
		}
		_ = os.RemoveAll(tmpDir)
	}()

	editorBin, err := pickEditor(o.editor)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "edit: %s (editor: %s, %d byte%s)\n",
		op.DisplayPath, editorBin, len(currentBytes), pluralS(len(currentBytes)))

	if err := runEditor(ctx, editorBin, tmpFile, editorStdin, editorStdout, editorStderr); err != nil {
		// Editor failures are diagnostic — keep the temp file
		// regardless of --keep-temp so the user can recover work.
		fmt.Fprintf(out, "  ! editor exited non-zero; temp file retained at %s\n", tmpFile)
		o.keepTemp = true
		return fmt.Errorf("editor %q on %s: %w", editorBin, tmpFile, err)
	}

	newBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		o.keepTemp = true
		return fmt.Errorf("read edited %s: %w", tmpFile, err)
	}

	if bytesEqual(currentBytes, newBytes) {
		fmt.Fprintf(out, "  · no changes; nothing to upload\n")
		cleaned = true
		_ = os.RemoveAll(tmpDir)
		return nil
	}

	// Post-edit size cap: refuse to PUT a buffer that exceeds
	// --max-size. This is the second half of the cap (the
	// pre-edit check rides on Stat in statAndFetch). Keep the
	// temp file on this path regardless of --keep-temp — the
	// user has unsaved changes worth recovering, and pointing
	// them at the temp file lets them split the work into
	// chunks or use `files upload` if the edit really is
	// supposed to land a >cap blob.
	if o.maxSize > 0 && int64(len(newBytes)) > o.maxSize {
		o.keepTemp = true
		return fmt.Errorf(
			"edit %s: post-edit size %s exceeds --max-size %s; temp file retained at %s "+
				"(re-run with --max-size 0 to disable the cap or --max-size <bytes> to widen it; "+
				"`files upload %s %s` works regardless of cap)",
			op.DisplayPath,
			formatBytes(int64(len(newBytes))),
			formatBytes(o.maxSize),
			tmpFile,
			tmpFile, op.DisplayPath)
	}

	fmt.Fprintf(out, "uploading %d byte%s → %s\n",
		len(newBytes), pluralS(len(newBytes)), op.DisplayPath)
	if err := editClient.PutBytes(ctx, op, newBytes, o.contentType); err != nil {
		// Upload failure means the user's edits are NOT on the
		// server — keep the temp file regardless of --keep-temp
		// so they can salvage their work (e.g. with `files
		// upload <tmp> <remote>`).
		o.keepTemp = true
		fmt.Fprintf(out, "  ! upload failed; temp file retained at %s\n", tmpFile)
		return reformatEditHTTPErr(err, rp.OlaresID, "save", op.DisplayPath)
	}
	fmt.Fprintf(out, "  ✓ saved %s\n", op.DisplayPath)
	cleaned = true
	_ = os.RemoveAll(tmpDir)
	return nil
}

// frontendPathToEditTarget converts the user-supplied path into
// the edit package's Target shape. We refuse the volume root +
// directory paths up front in the same place ParseFrontendPath
// runs, so the user gets a clear error before we dial the server.
//
// `.` / `..` segments are blocked via ValidateNoDotSegments — same
// reasoning as mkdir / rename: ParseFrontendPath's path.Clean
// silently collapses them away (e.g. `drive/Home/foo/./bar` →
// `drive/Home/foo/bar`), which would let a typo land bytes on a
// different file than the user typed.
func frontendPathToEditTarget(raw string) (edit.Target, error) {
	if err := ValidateNoDotSegments(raw); err != nil {
		return edit.Target{}, fmt.Errorf("edit: %w", err)
	}
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return edit.Target{}, err
	}
	if strings.Trim(fp.SubPath, "/") == "" {
		return edit.Target{}, fmt.Errorf(
			"refusing to edit the root of %s/%s: pick a file path (e.g. %s/%s/notes.md)",
			fp.FileType, fp.Extend, fp.FileType, fp.Extend)
	}
	if strings.HasSuffix(fp.SubPath, "/") {
		return edit.Target{}, fmt.Errorf(
			"%s is a directory path (trailing '/'); edit only works on files",
			fp.String())
	}
	return edit.Target{
		FileType: fp.FileType,
		Extend:   fp.Extend,
		SubPath:  fp.SubPath,
	}, nil
}

// statAndFetch resolves "(does this file exist? if yes, give me
// its bytes; if no and --create, give me an empty buffer)" in a
// single helper so the cobra-runner doesn't have to track three
// failure modes (404 / dir / oversize).
//
// Returns:
//   - currentBytes: file contents, or empty []byte when the file
//     was missing and --create is set.
//   - isDir: true if the remote target is a directory; the cobra
//     layer surfaces a friendlier error than the wire 400.
//   - err: any non-recoverable failure.
//
// `maxSize > 0` activates the pre-edit size cap: when Stat reports
// the file is larger than maxSize, we error out BEFORE pulling
// bytes (one ListDir call wasted, but no multi-MB GET that we'd
// just throw away). `maxSize == 0` disables the check entirely
// (matches the --max-size 0 escape hatch on the cobra surface).
//
// The Stat call uses the parent-listing strategy (see
// internal/files/download/stat.go); this avoids the "GET on
// /api/resources/<file> returns 500" trap that the per-resource
// list handler hits for many real files.
func statAndFetch(
	ctx context.Context,
	statClient *download.Client,
	editClient *edit.Client,
	plain string,
	allowCreate bool,
	maxSize int64,
) (currentBytes []byte, isDir bool, err error) {
	st, statErr := statClient.Stat(ctx, plain)
	switch {
	case statErr == nil:
		if st.IsDir {
			return nil, true, nil
		}
		// Size cap: the cloud-drive listings populate Size via
		// the FileSize→Size flex-decoder in download/list.go, so
		// this check is uniform across drive / sync / cache /
		// external / awss3 / google / dropbox. A reported size
		// of 0 either means "really empty" or "the backend
		// didn't fill it in" — either way, 0 ≤ maxSize so we
		// safely fall through to the Fetch.
		if maxSize > 0 && st.Size > maxSize {
			return nil, false, fmt.Errorf(
				"edit %s: remote size %s exceeds --max-size %s; "+
					"`files edit` is meant for text editing — re-run with --max-size 0 to disable "+
					"the cap, or --max-size <bytes> to widen it (or use `files download` for binaries)",
				plain, formatBytes(st.Size), formatBytes(maxSize))
		}
	case download.IsNotFound(statErr):
		if !allowCreate {
			return nil, false, fmt.Errorf(
				"edit %s: not found on the server (HTTP 404); pass --create to start with an empty buffer",
				plain)
		}
		return []byte{}, false, nil
	default:
		return nil, false, statErr
	}

	body, err := editClient.Fetch(ctx, plain)
	if err != nil {
		// A 404 between Stat and Fetch is rare but possible — file
		// was deleted by another client. Treat it like the no-stat
		// 404 above so the user gets a uniform message.
		if edit.IsNotFound(err) {
			if !allowCreate {
				return nil, false, fmt.Errorf(
					"edit %s: not found on the server (HTTP 404); pass --create to start with an empty buffer",
					plain)
			}
			return []byte{}, false, nil
		}
		return nil, false, err
	}
	// Defense in depth: even when Stat said the size was OK,
	// the Fetch may return a different (larger) body — e.g. the
	// file was concurrently appended to between Stat and Fetch.
	// Surface this as the same cap error so the user gets one
	// consistent message instead of having to reason about the
	// race window.
	if maxSize > 0 && int64(len(body)) > maxSize {
		return nil, false, fmt.Errorf(
			"edit %s: fetched body %s exceeds --max-size %s (likely a concurrent write between stat and fetch); "+
				"re-run with --max-size 0 to disable the cap or --max-size <bytes> to widen it",
			plain, formatBytes(int64(len(body))), formatBytes(maxSize))
	}
	return body, false, nil
}

// writeTempFile creates a fresh $TMPDIR/olares-files-edit-XXXX/
// directory and drops `content` into a file whose basename matches
// the remote basename. Editors key syntax highlighting off the
// extension (vim's filetype= / VSCode's "associations"), so
// preserving the basename matters for the user experience.
//
// Returns the temp directory + the temp file path. The caller is
// responsible for cleanup; we don't `defer os.RemoveAll` here
// because the cobra layer wants to skip cleanup when --keep-temp
// is set or when an error path retains the file for recovery.
func writeTempFile(displayPath string, content []byte) (tmpDir, tmpFile string, err error) {
	tmpDir, err = os.MkdirTemp("", "olares-files-edit-*")
	if err != nil {
		return "", "", fmt.Errorf("mktemp: %w", err)
	}
	base := lastSegmentForEdit(displayPath)
	if base == "" {
		base = "file"
	}
	tmpFile = filepath.Join(tmpDir, base)
	if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("write temp %s: %w", tmpFile, err)
	}
	return tmpDir, tmpFile, nil
}

// pickEditor implements the editor cascade (--editor → $VISUAL →
// $EDITOR → fallback). On POSIX the fallback is "vi" (universally
// installed); on Windows it's "notepad". We resolve via PATH up
// front so a typo / missing binary fails before we even create
// the temp file — that's friendlier than seeing the editor command
// fail with a confusing exec error AFTER the user already typed.
//
// The editor string can carry arguments (`code --wait`, `emacs -nw`)
// — we split on whitespace before the lookup, mirroring `git`'s
// GIT_EDITOR handling. Quoting is intentionally NOT supported (no
// shell expansion); users with truly exotic editor commands can
// wrap them in a script and point --editor at that.
func pickEditor(flag string) (string, error) {
	candidate := flag
	if candidate == "" {
		candidate = os.Getenv("VISUAL")
	}
	if candidate == "" {
		candidate = os.Getenv("EDITOR")
	}
	if candidate == "" {
		if runtime.GOOS == "windows" {
			candidate = "notepad"
		} else {
			candidate = "vi"
		}
	}
	bin := strings.Fields(candidate)
	if len(bin) == 0 {
		return "", errors.New("editor cascade resolved to an empty command; set $EDITOR or pass --editor")
	}
	if _, err := exec.LookPath(bin[0]); err != nil {
		return "", fmt.Errorf(
			"editor %q not found in PATH: %w (set $EDITOR / $VISUAL or pass --editor)",
			bin[0], err)
	}
	return candidate, nil
}

// runEditor spawns the editor with the temp file as its only
// (positional) argument. Stdin/stdout/stderr are wired through to
// the caller-supplied streams (in production: the user's TTY) so
// curses-style editors like vim / nano / hx work without any
// extra plumbing.
//
// The editor is run as a foreground child that we wait on; ctx is
// honoured between attempts (a Ctrl-C inside vim itself is the
// editor's responsibility — we don't try to forward signals).
func runEditor(
	ctx context.Context,
	editor, tmpFile string,
	stdin io.Reader, stdout, stderr io.Writer,
) error {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return errors.New("runEditor: empty editor command")
	}
	args := append(append([]string{}, parts[1:]...), tmpFile)
	cmd := exec.CommandContext(ctx, parts[0], args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// reformatEditHTTPErr maps edit.HTTPError + download.HTTPError +
// the typed credential errors onto user-friendly messages. The
// pattern mirrors reformatHTTPErr (download), reformatRmHTTPErr
// (rm), reformatRenameHTTPErr (rename) — same status branches, so
// the user sees one consistent CTA across the verbs.
//
// The `op` argument is the human-readable verb suffix that lands
// in the message ("fetch" before the editor, "save" after); this
// makes a 401/403/404 self-describing without the caller having
// to re-format the message further.
func reformatEditHTTPErr(err error, olaresID, op, target string) error {
	if err == nil {
		return nil
	}
	var inv *credential.ErrTokenInvalidated
	if errors.As(err, &inv) {
		return inv
	}
	var nli *credential.ErrNotLoggedIn
	if errors.As(err, &nli) {
		return nli
	}
	// edit.HTTPError covers the PUT / GET-raw paths in the edit
	// package; download.HTTPError covers the Stat-via-listing
	// path used during pre-flight. Both are flat structs with the
	// same Status field shape, so two short branches is plenty.
	if status, ok := editStatus(err); ok {
		switch status {
		case 401, 403, 459:
			if olaresID != "" {
				return fmt.Errorf(
					"server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
					status, olaresID)
			}
			return fmt.Errorf(
				"server rejected the access token (HTTP %d); please re-run `olares-cli profile login`",
				status)
		case 404:
			return fmt.Errorf("%s %s: not found on the server (HTTP 404)", op, target)
		case 409:
			return fmt.Errorf(
				"%s %s: target conflict (HTTP 409); the file may have been changed concurrently — re-fetch and try again",
				op, target)
		case 413:
			return fmt.Errorf("%s %s: payload too large (HTTP 413); the server rejected the new contents", op, target)
		}
	}
	return err
}

// editStatus extracts the wire status code from either an
// edit.HTTPError or a download.HTTPError so reformatEditHTTPErr
// can branch uniformly. Returns (0, false) when the error is
// neither.
func editStatus(err error) (int, bool) {
	var eErr *edit.HTTPError
	if errors.As(err, &eErr) {
		return eErr.Status, true
	}
	var dErr *download.HTTPError
	if errors.As(err, &dErr) {
		return dErr.Status, true
	}
	return 0, false
}

// lastSegmentForEdit returns the file basename for use as the
// temp file's name. Display path always has at least one '/' (the
// fileType / extend separator), so an empty result here means the
// caller passed something unusual — fall back to "file" rather
// than producing a temp file with a suspicious name.
func lastSegmentForEdit(display string) string {
	s := strings.Trim(display, "/")
	if s == "" {
		return ""
	}
	if i := strings.LastIndex(s, "/"); i >= 0 {
		return s[i+1:]
	}
	return s
}

// bytesEqual is the no-change predicate the cobra layer uses to
// decide whether to skip the post-edit PUT. Pulled into a named
// helper rather than calling bytes.Equal inline so the call site
// at runEdit reads as the verb-level intent ("did the user
// actually change the file?") rather than an unspecific
// byte-comparison. A length pre-check short-circuits the common
// no-op path (vim's `:q` after a no-touch open writes the same
// length back to disk).
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return bytes.Equal(a, b)
}
