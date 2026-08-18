# settings integration — cookie store

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings integration cookie --help`.

Same store as **Settings → Integration → Cookies** in the SPA, and the one `download-server` reads for yt-dlp / aria2. A cookie imported here is usable by [`olares-knowledge`](../../olares-knowledge/SKILL.md) downloads and Wise collection immediately.

## Sub-tree

| Verb | Floor | Notes |
|---|---|---|
| `cookie import --domain <d> --file <path>` | normal | `--file -` reads stdin; **replaces** the domain unless `--merge` |
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

A request-style header needs `--domain`. A `Set-Cookie:` line carries its own attributes.

## Importing without leaking the value

```bash
olares-cli settings integration cookie import --domain youtube.com --file cookies.txt
pbpaste | olares-cli settings integration cookie import --domain youtube.com --file - --format header
olares-cli settings integration cookie import --domain youtube.com --file extra.txt --merge
```

No flag takes the cookie value as an argument (`ps` / shell history). Always `--file` or `--file -`.

## Checking and clearing

```bash
olares-cli settings integration cookie list
olares-cli settings integration cookie validate youtube.com
olares-cli settings integration cookie rm youtube.com
```

`list` / `validate` never print names or values, in any `--output` format.

## Agent best practices

- Import **replaces the whole domain**. Use `--merge` to add without dropping existing records.
- Store under the bare host (`youtube.com`), not `.youtube.com` and not a full URL.
- After import, re-run `knowledge download inspect <url>` before `create`.
- Cookies expire — `cookie validate <domain>` before assuming something else broke.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `no cookies found in the input` | Wrong `--format`, or nothing for `--domain` | Explicit `--format`; confirm export is not empty |
| `input looks like JSON, not a Netscape cookies.txt file` | JSON with `--format netscape` | `--format json` or drop `--format` |
| `--domain is required` on a header import | Request-style `a=b; c=d` has no domain | Pass `--domain <host>` |
| Import OK but downloads still need login | Export dropped httpOnly cookies | Re-import via the `header` path above |
| `no cookies stored for <domain>` | Never imported, or different domain string | `cookie list` for exact domains |
