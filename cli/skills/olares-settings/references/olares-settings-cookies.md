# settings integration — cookie store

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings integration cookie --help`.

Same store as **Settings → Integration → Cookies** in the SPA, and the one `download-server` reads for yt-dlp / aria2. A cookie imported here is usable by [`olares-knowledge`](../../olares-knowledge/SKILL.md) downloads and Wise collection immediately.

## Sub-tree

| Verb | Floor | Notes |
|---|---|---|
| `cookie import [--domain <d>] --file <path>` | normal | `--file -` reads stdin; **replaces** each written domain unless `--merge` |
| `cookie list` | normal | Domains, record counts, next expiry. Never prints names or values |
| `cookie rm <domain>` | normal | Deletes every cookie for the domain |
| `cookie validate <domain>` | normal | Non-zero exit when the domain has no cookies, or all have expired |

## When to import

Import when a download or collection fails for a login reason. The CLI already prints the exact command; signals include:

- `knowledge download inspect` → `error_code` 501 / 507 / 511 / 512, or `error_category` `authorization_failed` / `private_resource` / `bot_detected`
- `knowledge download create` fails with 501, or a waited task fails with one of those categories

## Choosing a format

`--format` defaults to `auto`. Name it when auto-detection is wrong.

| Format | Source | Prefer when |
|---|---|---|
| `netscape` | `cookies.txt` (extensions, yt-dlp, curl) | User has a file |
| `json` | Cookie-Editor / EditThisCookie | User has a JSON array |
| `header` | One header line, copied by hand | **Rescue path** when a `cookies.txt` export is missing the login |

**Prefer `header` when a `cookies.txt` export is missing the login.** The browser *request* `Cookie:` header carries httpOnly cookies (httpOnly blocks JS, not the browser). YouTube login lives in httpOnly cookies such as `__Secure-3PSID` / `SID` / `HSID`. DevTools → Network → reload → document request → Request Headers → copy the `Cookie:` line.

A request-style header needs `--domain`. A `Set-Cookie:` line carries its own attributes. Auto-detect recognises a bare `a=b; c=d` line as `header`; without `--domain` the import still fails asking for one — that is expected.

## Domain flag

Omit `--domain` for Netscape / JSON: every host in the file is written to its own store key (same as the SPA paste). Pass `--domain` to **filter** to matching hosts only (e.g. `youtube.com` keeps `.youtube.com`, drops `.google.com`). It never rewrites foreign hosts onto another key.

Download failure hints include `--domain` so a full browser export only updates the site that failed.

## Importing without leaking the value

```bash
olares-cli settings integration cookie import --file cookies.txt
olares-cli settings integration cookie import --domain youtube.com --file cookies.txt
olares-cli settings integration cookie import --domain youtube.com --file www.youtube.com_cookies.txt
pbpaste | olares-cli settings integration cookie import --domain youtube.com --file - --format header
olares-cli settings integration cookie import --domain youtube.com --file extra.txt --merge
```

`--file www.youtube.com_cookies.txt` works when that file is a single-line request Cookie header (`a=b; c=d`); auto-detect picks `header`, but `--domain` is still required.

No flag takes the cookie value as an argument (`ps` / shell history). Always `--file` or `--file -`.

## Checking and clearing

```bash
olares-cli settings integration cookie list
olares-cli settings integration cookie validate youtube.com
olares-cli settings integration cookie rm youtube.com
```

`list` / `validate` never print names or values, in any `--output` format.

## Agent best practices

- Import **replaces each written domain**. Use `--merge` to add without dropping existing records.
- Prefer the failure hint's `--domain` when fixing one URL; omit it only when the user wants a full multi-host paste like the SPA.
- After import, re-run `knowledge download inspect <url>` before `create`.
- Cookies expire — `cookie validate <domain>` before assuming something else broke.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `no cookies found in the input` | Wrong `--format`, or empty export | Explicit `--format`; confirm export is not empty |
| `input looks like JSON, not a Netscape cookies.txt file` | JSON with `--format netscape` | `--format json` or drop `--format` |
| `could not tell which cookie format this is` | Ambiguous paste | Pass `--format netscape\|json\|header`; a `;`+`=` line that still fails auto is almost always a Cookie header → `--format header --domain <host>` |
| `this input is a browser Cookie header, which carries no domain; re-run with --domain <host> (for example --domain youtube.com)` | Request-style `a=b; c=d` has no domain | Pass `--domain <host>` |
| `no cookies for domain X in the input` | Filter matched no host in the file | Drop `--domain`, or use a host that appears in the export |
| Import OK but downloads still need login | Export dropped httpOnly cookies | Re-import via the `header` path above |
| `no cookies stored for <domain>` | Never imported, or different domain string | `cookie list` for exact domains |
