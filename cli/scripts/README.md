# cli/scripts

Auxiliary scripts for working with `olares-cli`. Not built into the
binary — these are developer / CI tooling.

## test-files-cmd.sh

End-to-end smoke tests for `olares-cli files <verb>`.

Drives every `files` verb against a live Olares instance and reports
per-test pass/fail. Includes a regression test for the
"directory upload to namespace root" bug
(`olares-cli files upload ./mydir/ drive/Home/` previously failed
with `local "./mydir/" is a directory; remote "" must end with '/'`).

### Quick start

```bash
# Make sure you have a profile selected first.
olares-cli profile login --olares-id <id>

# Minimal run (drive/Home + drive/Data only).
./cli/scripts/test-files-cmd.sh

# Full namespace coverage (sync + cache + external).
./cli/scripts/test-files-cmd.sh \
    --sync-repo-name cli-test-repo \
    --cache    cache/<node> \
    --external external/<node>/<volume>

# Add cloud drives.
./cli/scripts/test-files-cmd.sh \
    --awss3   awss3/<account>/<bucket> \
    --google  google/<account> \
    --dropbox dropbox/<account>

# Use a locally-built binary instead of the system install.
./cli/scripts/test-files-cmd.sh --cli ./cli/olares-cli

# Keep the remote test directories after the run (for debugging).
./cli/scripts/test-files-cmd.sh --keep --sync-repo-name cli-test-repo

# See every flag.
./cli/scripts/test-files-cmd.sh --help
```

### CLI flags vs environment variables

Both interfaces are accepted; **CLI flags win on conflict** so a
one-off override doesn't require unsetting a shell-wide export.

| Flag | Env var | Description |
|---|---|---|
| `--cli` / `-b <path>` | `OLARES_CLI` | Use this binary |
| `--sync-repo-name <name>` | `OLARES_TEST_SYNC_REPO_NAME` | Provision a temp Sync repo and run sync tests |
| `--cache cache/<node>` | `OLARES_TEST_CACHE_PATH` | Run cache tests under this base path |
| `--external external/<node>/<volume>` | `OLARES_TEST_EXTERNAL_PATH` | Run external tests under this base path |
| `--awss3 awss3/<account>/<bucket>` | `OLARES_TEST_AWSS3_PATH` | Cloud drive: AWS S3 |
| `--google google/<account>` | `OLARES_TEST_GOOGLE_PATH` | Cloud drive: Google Drive |
| `--dropbox dropbox/<account>` | `OLARES_TEST_DROPBOX_PATH` | Cloud drive: Dropbox |
| `--keep` | `OLARES_TEST_KEEP=1` | Skip cleanup; preserve test dirs + temp repo |
| `--help` / `-h` | — | Print the flag table and exit |

### What it covers

| Verb       | Cases                                                      |
|------------|------------------------------------------------------------|
| `ls`       | drive/Home, drive/Data, --json, invalid namespace (negative)|
| `mkdir`    | drive/Home + drive/Data unconditional;                    |
|            | sync + cache + external when configured.                  |
|            | Each namespace gets the same triple: per-run `-p`,        |
|            | single-level (no -p), and multi-level `-p` (auto-rename   |
|            | quirk regression: server silently auto-renames `Foo` to   |
|            | `Foo (1)` on collision; -p mode's parent-listing skip     |
|            | side-steps this).                                         |
| `upload`   | single file, rename-on-upload, empty file (CreateEmptyFile),|
|            | directory tree, **directory to namespace root (regression)**|
| `download` | single file, recursive directory                          |
| `cat`      | text file, empty file, directory (negative)               |
| `cp`       | file → dir, rename mode, recursive, multi-source          |
| `mv`       | file → dir, rename mode (drive only — see sync notes)     |
| `rename`   | in-place, slash in new-name (negative)                    |
| `rm`       | single (`-f`), directory without `-r` (negative),         |
|            | recursive (`-rf`), volume root (negative)                 |
| `share`    | list, smb-users list, public lifecycle if password works  |
| `repos`    | list (table + JSON), get / rename if `--sync-repo-name` set |
| sync       | full files-verb suite on `sync/<repo_id>/` if `--sync-repo-name` set: |
|            | mkdir, upload (single / rename / empty / **`--parallel 1`** dir / **root-regression**), |
|            | download (single / dir), cat, cp (intra + drive→sync),    |
|            | mv (move-into-dir only — see notes), rename (incl. negative |
|            | slash test), rm (incl. repo-root protection), public share |
|            | on sync file                                              |
| cloud      | upload/ls/cat/rm round-trip per cloud-drive flag/env      |

#### Sync (Seafile) backend caveats the test bakes in

Two real Seafile-backend behaviors that the script accommodates so
the suite stays green on Olares deployments:

- **Directory upload uses `--parallel 1`.** Seafile's
  `/seafhttp/upload-aj/` relies on the chunk's `relative_path`
  form field to auto-create intermediate directories. With the
  CLI's default `--parallel 2`, parallel POSTs for files at
  different depths (e.g. `testU/sub/level1.txt` and
  `testU/sub/deep/level2.txt`) race the auto-create and the
  deeper file can land before its parent dir exists, returning
  HTTP 500 with an empty body. Drive isn't affected. (Tracking
  task: have the uploader pre-mkdir every dir in `plan.Files`,
  not just `plan.EmptyDirs`, when `DriveType==Sync`.)
- **`mv` on sync is move-into-dir only — no rename mode.**
  Seafile's `/api/paste/<node>/` honors the destination directory
  but drops the destination basename, always using the source's
  basename. So `mv x/foo.txt y/bar.txt` lands as `y/foo.txt`
  (rename dropped); `mv x/foo.txt x/bar.txt` triggers Seafile's
  self-conflict resolver and lands as `x/foo (1).txt`. To rename
  a sync-backed file, use `files rename` (synchronous PATCH
  `/api/resources` → `seafile_api.rename_file`, which respects
  the new name).

### Exit code

`0` if every test passed, `1` if any failed. The summary at the end
lists exactly which tests failed.

### Cleanup

Per-run subdirectories under each configured namespace hold every
test artifact, so cleanup is one recursive `files rm -rf` per
namespace in the `EXIT` trap (plus `repos rm -y` for the temp Sync
library). The user-supplied `--cache` / `--external` *base* paths
are never removed — only the per-run-id subdir under them, since
the base might be node-local scratch shared with other workloads.
Pass `--keep` (or `OLARES_TEST_KEEP=1`) to skip cleanup for
post-mortem debugging.
