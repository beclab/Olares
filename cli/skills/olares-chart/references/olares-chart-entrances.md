# Entrances, ports and how the app opens

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This is §4 of the [Manifest refinement areas](olares-chart-manifest.md): what the user clicks on the Desktop, how it opens, and what it looks like. Re-run `olares-cli chart lint ./<app>` after every change.

## The entrance block

`from-compose` writes one entrance with `authLevel: private`, `openMethod: default` and the default icon. Every field below is one the user notices:

```yaml
entrances:
- name: myapp            # matches .Values.domain.<name> in templates
  host: myapp-svc        # an existing Service name in templates/
  port: 8080             # the Service port
  title: My App          # 1-30 characters; the Desktop label
  icon: https://raw.githubusercontent.com/<owner>/<repo>/<commit-sha>/<path>.png
  authLevel: private     # public | private | internal
  openMethod: window     # window = its own browser tab; default/iframe/"" = a window inside the Desktop
  invisible: false       # true for internal-only services
```

- **One entrance per user-facing HTTP service.** Add entries for additional UIs; set `invisible: true` (or omit the entrance) for internal-only services.
- **No web UI at all** (a CLI tool, an MCP server, an API daemon)? Don't force a fake entrance — apply the [headless archetype](olares-chart-archetype-headless.md): a web terminal as the visible entrance + the service port as an invisible internal entrance.

## `openMethod` — where clicking the icon opens the app

The request usually arrives in the user's words, not the field's. Map it, then write the value; `lint` rejects anything outside the four values.

| The user asks for | Value | What the Desktop does |
|---|---|---|
| "open in a new tab", "separate browser tab", "own window", "standalone", "full screen", 新标签页, 独立窗口 | `window` | opens the entrance URL in a new browser tab, outside the Desktop |
| "open inside Olares", "embedded", "in the Desktop", or nothing said | `default` (same as `""` and `iframe`) | opens the app in a window inside the Desktop |

Most third-party web apps in the catalog use `window`, so their own navigation and login redirects behave as in a normal browser tab. Keep `default` for apps designed to live inside the Desktop. `tab`, `newTab`, `browser` and similar guesses do not exist.

## `icon` — what the Desktop shows

`metadata.icon` is required and is the app's icon; `entrances[].icon` is what the Desktop shows for that entrance, and falls back to `metadata.icon` when absent. Set both to the same URL unless the entrances genuinely differ. Both must be `http(s)` URLs — `lint` checks only the form, not that the URL serves an image.

- **No icon asked for:** the scaffolded default (`https://app.cdn.olares.com/appstore/default/defaulticon.webp`) is fine for a local deploy.
- **"Use the official icon":** see [Using the upstream icon](#using-the-upstream-icon) below.

## Using the upstream icon

The upstream repository already hosts its icon; point at it instead of re-hosting.

1. Find the file the project itself uses: the image in the README header, a `logo` / `icon` / `brand` asset, or the PWA / Tauri / Electron app icon. Prefer a square PNG of 256 px or more.
2. Pin it to a commit, never to a branch: `https://raw.githubusercontent.com/<owner>/<repo>/<commit-sha>/<path>`. Take the SHA from `git ls-remote https://github.com/<owner>/<repo> HEAD`. A branch URL changes or 404s when upstream reorganises its assets.
3. Confirm it serves an image: `curl -sIL <url>` returns `200` with a `content-type: image/*`.
4. Write that URL into `metadata.icon` and every visible `entrances[].icon`.

**Do not upload the icon anywhere for a local deploy.** `imghost.olares.com` is a web page a person uses during [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md); scripting it — reading its bundle for an endpoint or an embedded key — is outside the task and is the developer's call, not the agent's. If the upstream has no usable icon, keep the default and say so.

## Non-HTTP services and outbound ports

- **Non-HTTP services** (game server, SMTP, RDP, …) are exposed via `ports[]`, not entrances:
  ```yaml
  ports:
  - name: game
    host: game-svc
    port: 7777
    protocol: udp
    exposePort: 47777    # cluster-unique; avoid reserved 22/80/81/443/444/2379/18088
  ```
- **Outbound non-HTTP** (e.g. the app sends SMTP): `options.allowedOutboundPorts: [465, 587]`.

## Long-running requests: `options.apiTimeout`

The per-app **entrance proxy** caps every request at `options.apiTimeout` **seconds** — **default 15s**, so anything slower (LLM streaming, big uploads, slow report generation) is cut at the entrance (504 / closed connection) regardless of the app or browser. Set `options.apiTimeout: 0` to disable the cap, or a large value (e.g. `3600`) for a bounded one. **Only `0` disables it** — a *negative* value is not "no timeout"; the sidecar treats it (like an unset field) as the 15s default. It is an install-time **manifest** field (not an install-time env), so for `templateOnly` env-driven charts you must edit it in the chart manifest and re-package.
