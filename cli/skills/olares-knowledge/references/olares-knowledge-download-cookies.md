# knowledge download cookies

> **Flags:** `olares-cli knowledge download cookies <verb> --help`.

> **Hidden command.** On a production cluster the integration provider is
> read-only and only implements `POST /api/cookie/retrieve`, so only
> `cookies retrieve` works there. `list` / `set` / `delete` / `health`
> target endpoints that exist only in the local `mock-integration` and
> return **404** against a real cluster. The whole `cookies` tree is
> therefore hidden from `--help` (still invokable for scripts / local
> dev). Do not surface `list` / `set` / `delete` as working prod verbs.

Provider cookies let download providers (e.g. yt-dlp) fetch gated content.
Cookies are stored per domain and supplied as a **Netscape cookies.txt** file.

## list

```bash
olares-cli knowledge download cookies list
olares-cli knowledge download cookies list -o json
```

`GET /api/integration/cookies`. Columns: `DOMAIN  PROVIDER  HAS_COOKIE
UPDATED`. The stored cookie text is never returned by list.

## set

```bash
olares-cli knowledge download cookies set --domain youtube.com --cookie-file ./cookies.txt
olares-cli knowledge download cookies set --domain youtube.com --provider yt-dlp --cookie-file ./cookies.txt
```

`PUT /api/integration/cookies`. `--cookie-file` is read locally and its full
text is uploaded. `--provider` defaults to the server default (yt-dlp) when
omitted.

## retrieve

```bash
olares-cli knowledge download cookies retrieve --domain youtube.com
olares-cli knowledge download cookies retrieve --domain youtube.com -o json
```

`POST /api/integration/cookies/retrieve`. Table output shows only whether a
cookie was found and its update time; the **plaintext cookie is printed only
with `-o json`**.

## delete

```bash
olares-cli knowledge download cookies delete --domain youtube.com
```

`DELETE /api/integration/cookies/<user>/<domain>`. The `:user` path segment is
a placeholder — the real identity is the gateway-injected `X-Bfl-User`.

## health

```bash
olares-cli knowledge download cookies health
```

`GET /api/integration/healthz`. Reports overall `Healthy` plus a per-provider
status line (`ok` / `error: …`).
