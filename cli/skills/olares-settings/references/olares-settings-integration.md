# settings integration

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings integration accounts --help`, `olares-cli settings integration cookie --help`, and `olares-cli settings integration <group> <verb> --help`.

Two sub-trees: external integration accounts (S3 / Tencent COS / Google Drive / Dropbox / Olares Space / NFT cloud binding), and the per-domain **cookie** store that downloads and Wise collection use to fetch content behind a login.

## What's covered by the CLI

| Account type | CLI verb | Why |
|---|---|---|
| AWS S3 (or S3-compatible endpoint) | `accounts add awss3` | Direct credentials — no OAuth |
| Tencent COS | `accounts add tencent` | Direct credentials — no OAuth |
| Google Drive | (SPA only) | OAuth-bound browser session |
| Dropbox | (SPA only) | OAuth-bound browser session |
| Olares Space / NFT cloud binding | (SPA only) | Browser- and wallet-bound by design |

> **OAuth and wallet flows stay in the SPA.** The access tokens they produce are scoped to a browser session and have no useful one-shot CLI capture. If the user asks to add Google Drive / Dropbox / Olares Space / NFT integrations via CLI, direct them to the Settings → Integration page in the SPA.

## Sub-tree

| Verb | Floor | Notes |
|---|---|---|
| `accounts list` | normal | `accountMini` shape (no `raw_data`) |
| `accounts list-by-type <type>` | normal | Filter by account type |
| `accounts get <type> [name]` | normal | `accountFull` shape (includes `raw_data`); `name` optional for single-tenant types |
| `accounts add awss3 [flags]` | normal | AWS S3 / S3-compatible |
| `accounts add tencent [flags]` | normal | Tencent COS |
| `accounts delete <type> [name]` | normal | `name` optional for single-tenant types |

## `accounts add awss3`

```bash
olares-cli settings integration accounts add awss3 \
  --access-key-id     "$AWS_ACCESS_KEY_ID" \
  --access-key-secret "$AWS_SECRET_ACCESS_KEY" \
  --endpoint          "https://s3.amazonaws.com" \
  --bucket            "my-bucket"
```

- `--bucket` is **optional** — omit for "any bucket the credentials can reach"; provide for "scope to this bucket".
- `--endpoint` accepts any S3-compatible endpoint (MinIO, Backblaze B2 via S3 API, etc.) — not just AWS.

## `accounts add tencent`

```bash
olares-cli settings integration accounts add tencent \
  --access-key-id     "$TENCENT_SECRET_ID" \
  --access-key-secret "$TENCENT_SECRET_KEY" \
  --endpoint          "https://cos.ap-shanghai.myqcloud.com"
```

- Region is encoded in the endpoint URL — e.g. `cos.ap-beijing.myqcloud.com`, `cos.ap-shanghai.myqcloud.com`.
- Tencent COS is **single-tenant**: no `--bucket` flag, no `<name>` argument on add. There is at most one Tencent account per profile.

## `accounts get` / `accounts delete` — name handling

```bash
# Multi-tenant types (S3, Drive, Dropbox) — need a name.
olares-cli settings integration accounts get awss3 my-bucket
olares-cli settings integration accounts delete awss3 my-bucket

# Single-tenant types (Tencent, Space, NFT) — name omitted.
olares-cli settings integration accounts get tencent
olares-cli settings integration accounts delete tencent
```

The store key is composed as `integration-account:<type>:<name>` (or `integration-account:<type>` when no name is supplied), matching the SPA's `getStoreKey`.

## Secret handling — agent rules

**The single most important rule in this sub-tree: NEVER paste secret-key values into the agent transcript.**

- **Always recommend env vars or stdin pipes**: `--access-key-secret "$AWS_SECRET_ACCESS_KEY"`.
- Bash history retention is the user's responsibility, but the agent's default phrasing should make it easy to keep secrets out of the transcript / scrollback.
- For the agent's own suggestions: write `--access-key-secret "$AWS_SECRET_ACCESS_KEY"` (placeholder), NOT `--access-key-secret "AKIA..."` (real-looking).

## `accounts get` JSON shape

```bash
olares-cli settings integration accounts get awss3 my-bucket -o json
```

Returns the `accountFull` shape including the un-redacted `raw_data` field. **The secret-key value WILL appear in the output** — pipe to `jq` and select only the fields you need rather than dumping the whole payload.

## Agent best practices

- For "add my S3 credentials" → prompt the user to **set the secret in an env var FIRST**, then construct the command using `"$VAR"` interpolation.
- For "show me my integration accounts" → `accounts list` (no secrets) instead of `accounts get` (with secrets) unless the user specifically needs the credential payload.
- For "delete my old S3 account" → `accounts list-by-type awss3` first to confirm the right name, then `accounts delete awss3 <name>`.
- If the user asks for Google Drive / Dropbox / Olares Space / NFT cloud binding, **redirect to the SPA** — explain that OAuth tokens are browser-scoped and can't be captured by one-shot CLI calls.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `account 'X' of type '<type>' not found` | Wrong name / never added | `accounts list-by-type <type>` to enumerate |
| `missing required flag: --access-key-id` (or `--access-key-secret` / `--endpoint`) | Mandatory flag omitted | Provide all three for S3 / Tencent |
| `account already exists` | Single-tenant type (Tencent / Space) already has one | `accounts delete <type>` first, then add |
| Secret value shows up in shell history | The user (or agent) embedded the secret on the command line directly | Re-issue from env var; clear shell history manually |

# cookie store

This is the same store as **Settings → Integration → Cookies** in the SPA, and the one `download-server` reads when it runs yt-dlp or aria2 on the user's behalf. A cookie imported here is usable by [`olares-knowledge`](../../olares-knowledge/SKILL.md) downloads and Wise collection immediately — there is no second place to upload one.

## Sub-tree

| Verb | Floor | Notes |
|---|---|---|
| `cookie import --domain <d> --file <path>` | normal | `--file -` reads stdin; **replaces** the domain unless `--merge` |
| `cookie list` | normal | Domains, record counts, next expiry. Never prints names or values |
| `cookie rm <domain>` | normal | Deletes every cookie for the domain |
| `cookie validate <domain>` | normal | Non-zero exit when the domain has no cookies, or all have expired |

## When to import a cookie

Import when a download or collection fails for a reason a login fixes. The CLI already prints the exact command in that case; these are the signals:

- `knowledge download inspect` reports `error_code` 501 / 507 / 511 / 512, or an `error_category` of `authorization_failed`, `private_resource`, or `bot_detected`.
- `knowledge download create` fails with 501, or `create --wait` settles into a failure with one of those categories.

## Choosing a format

`--format` defaults to `auto`. Name it explicitly when auto-detection is wrong.

| Format | Where it comes from | When to prefer it |
|---|---|---|
| `netscape` | `cookies.txt` — what browser extensions export and what yt-dlp / curl / wget consume | Default choice when the user has a file |
| `json` | Extension exports such as Cookie-Editor / EditThisCookie | The user has a JSON array instead of a text file |
| `header` | One header line, copied by hand | **The rescue path** — see below |

**Prefer `header` when a `cookies.txt` export is missing the login.** The browser's *request* `Cookie:` header carries httpOnly cookies (httpOnly blocks JavaScript, not the browser itself), and the login for YouTube and similar sites lives entirely in httpOnly cookies such as `__Secure-3PSID` / `SID` / `HSID`. To get it: DevTools → Network → reload the page → click the document request → Request Headers → copy the whole `Cookie:` line.

A request-style header carries no domain, so `--domain` is mandatory there. A `Set-Cookie:` line carries its own attributes and is parsed in full.

## Importing without leaking the value

```bash
# From a file exported by a browser extension.
olares-cli settings integration cookie import --domain youtube.com --file cookies.txt

# From the clipboard, without ever touching disk.
pbpaste | olares-cli settings integration cookie import --domain youtube.com --file - --format header

# Add to what is already stored instead of replacing it.
olares-cli settings integration cookie import --domain youtube.com --file extra.txt --merge
```

There is deliberately no flag that takes the cookie value as an argument: command-line arguments are readable by any process via `ps` and land in shell history. Always use `--file`, or `--file -` with a pipe.

## Checking and clearing

```bash
olares-cli settings integration cookie list
olares-cli settings integration cookie validate youtube.com
olares-cli settings integration cookie rm youtube.com
```

`list` and `validate` report counts, expiry and staleness only — no cookie name or value is ever printed, in any `--output` format.

## Agent best practices

- Import is a **whole-domain replace**. If the user is adding one site's cookie to a domain that already has some, use `--merge` or you will silently drop the rest.
- `--domain` also decides where the records land: every cookie in the file is stored under it, and the download side looks a cookie up by the URL's host. Use the bare host (`youtube.com`), not `.youtube.com` and not a full URL.
- After importing, re-run `knowledge download inspect <url>` to confirm the failure cleared before creating the task.
- Cookies expire. When a download that used to work starts failing on a login, `cookie validate <domain>` before assuming anything else broke.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `no cookies found in the input` | Wrong `--format`, or the file has no record for `--domain` | Re-run with an explicit `--format`; confirm the export is not empty |
| `input looks like JSON, not a Netscape cookies.txt file` | JSON export passed with `--format netscape` | Use `--format json` (or drop `--format` for auto) |
| `--domain is required` on a header import | Request-style `a=b; c=d` carries no domain | Pass `--domain <host>` |
| Import succeeds but downloads still need a login | The export dropped httpOnly cookies | Re-import via the `header` path described above |
| `no cookies stored for <domain>` | Never imported, or imported under a different domain string | `cookie list` to see the exact domains stored |
