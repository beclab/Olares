# settings appearance

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings appearance --help` and `olares-cli settings appearance <noun> <verb> --help`.

Mirror the Settings → Appearance page: locale, desktop widget preferences, wallpaper, and the desktop layout reset. Every verb is per-user; none of them require admin or owner.

**There is no theme verb.** A desktop's light/dark state is a browser cookie the SPA never sends upstream, so no CLI can change it. Say that plainly instead of reaching for `settings advanced env user set --var OLARES_USER_THEME=...`, which writes a variable no Olares interface reads.

## Sub-tree

| Verb | Floor | Needs | Status | Purpose |
|---|---|---|---|---|
| `get` | normal | any | VERIFIED | Read the whole page: locale + widget + wallpaper |
| `language set <locale>` | normal | any | VERIFIED | Set the system language |
| `widget set [flags]` | normal | >= 1.12.6 | VERIFIED | Set desktop widget preferences |
| `wallpaper list <surface>` | normal | any | VERIFIED | Show the built-in range and uploaded images |
| `wallpaper set <surface> <number\|url>` | normal | any | VERIFIED | Select a wallpaper |
| `wallpaper style set <surface> <Fill\|Stretch\|Tile>` | normal | any | VERIFIED | Set the wallpaper fill mode |
| `wallpaper upload <surface> --file` | normal | any | VERIFIED | Upload a local image and select it |
| `wallpaper delete <surface> <url>` | normal | any | VERIFIED | Remove an uploaded image |
| `layout reset` | normal | >= 1.12.6 | VERIFIED | Restore the default desktop layout |

`<surface>` is always `desktop` or `login`.

## Version gate (Olares >= 1.12.6)

Two verbs need Olares 1.12.6+, where their upstream arrived: `widget set` (the widget preferences API) and `layout reset` (the desktop layout reset route). `get`, `language set` and every `wallpaper` verb work on any supported backend.

On an older backend the two fail up front with the shared message naming the verb and the detected version (see [Common errors](#common-errors)); nothing is sent. `layout reset` is gated **before** its confirmation prompt, so an older backend never asks the user to confirm a reset it cannot perform. Treat daily builds by their `major.minor.patch` base (`1.12.6-20260203` is the 1.12.6 line). Follow the shared auth/version gate when the version cannot be established.

`get` is the one verb that degrades instead of failing: below 1.12.6 it reads locale and wallpaper — which exist everywhere — and renders the widget section as `requires Olares >= 1.12.6`. With `-o json` that section is `null`, the key still present so a caller can tell a gated section from one this CLI does not know about. On 1.12.5 the locale section also returns an empty `timezone`, which renders as `-`.

## `get` is the only read verb

```bash
olares-cli settings appearance get
olares-cli settings appearance get -o json
```

There is no `widget get` or `wallpaper get`. One `get` reads all three upstream endpoints (`/api/wallpaper/config/system`, `/api/widget`, `/api/wallpaper`) — the widget one only on 1.12.6+, see the [version gate](#version-gate-olares--1126).

A section that fails for any other reason — auth, 500, a malformed body — fails the whole command, because a zeroed section would be read as real configuration.

`-o json` is three sections, each holding the upstream field names verbatim:

```json
{
  "locale":    { "language": "en-US", "location": "", "timezone": "UTC" },
  "widget":    { "showWeight": true, "is24HourFormat": true, "dateFormat": "YYYY/MM/DD", "showDashboard": true },
  "wallpaper": { "desktop": "/bg/0.jpg", "desktopStyle": "fill", "login": "/bg/0.jpg", "loginStyle": "fill",
                 "upload_desktop_backgrounds": [], "upload_login_backgrounds": [] }
}
```

Read the language from `.locale.language`, not `.language`. `showWeight` is the upstream name of the widget master switch despite what it reads like.

**JSON carries the stored values; the table renames them for the reader** — `"desktop": "/bg/3.jpg"` prints as `built-in 3` and `"desktopStyle": "cover"` as `Stretch`, matching how `wallpaper set` and Settings name them. Script against `-o json`.

## `widget set` changes only the flags you pass

```bash
olares-cli settings appearance widget set --show-widgets=false
olares-cli settings appearance widget set --24-hour=false --date-format M/D/YY
olares-cli settings appearance widget set --show-dashboard=true
```

The upstream POST replaces the whole preferences object, so the CLI reads the current values first and sends the unnamed ones back untouched. Passing no flag at all is an error rather than a no-op write.

`--date-format` is checked against the SPA's list (`YYYY/MM/DD`, `D/M/YY`, `M/D/YY`, `DD/MM/YYYY`, `DD.MM.YYYY`, `DD-MM-YYYY`, `YYYY.MM.DD`, `YYYY-MM-DD`, `YY/MM/DD`, `YY-M-D`, `YY.M.D`) because the upstream stores any string it is handed and an unlisted one silently fails to render in the desktop clock.

## `wallpaper`

```bash
olares-cli settings appearance wallpaper list desktop
olares-cli settings appearance wallpaper set desktop 3
olares-cli settings appearance wallpaper style set desktop Stretch
olares-cli settings appearance wallpaper upload desktop --file ./bg.jpg
olares-cli settings appearance wallpaper delete desktop https://files.example/bg.jpg --yes
```

**A built-in is selected by number**, as in the Settings grid — that grid is unnamed thumbnails, so a number is the only handle either side has. `wallpaper list <surface>` prints the range and the active value; it needs the surface because the two ship a different number of images.

```
Built-in desktop wallpapers
  0-27, currently 3

Uploaded desktop wallpapers
  none
```

When an uploaded image is active, the built-in line drops `currently` and that URL is marked `*` instead, so no built-in is ever reported as active when it is not.

Ranges (mirroring the SPA's `picturesCount`, the only place they exist): desktop `0-27`, login `0-28` — `28` is valid for login and invalid for desktop. `set` also takes an uploaded `https://` URL, and the stored `/bg/<n>.jpg` spelling so a value read from `get -o json` can be fed back in.

The value is stored verbatim and **not validated upstream** — a bad one is accepted and leaves a broken background with no error. The CLI therefore validates first; `--force` drops the range check for a built-in a newer release adds, still storing a number as `/bg/<n>.jpg`. Uploaded URLs are judged by shape, not looked up, so `set` stays a single request.

**Never offer `/login/<n>.jpg`.** Both surfaces store `/bg/<n>.jpg`: the greeter resolves the stored value under its own asset root (`LoginPage.vue` renders `"auth/" + value`), and the Settings page swaps the prefix to `/login/` only to preview the login variant. A stored `/login/<n>.jpg` resolves nowhere.

Fill modes are named as Settings names them: `Fill`, `Stretch`, `Tile` (case-insensitive). The stored values `fill`, `cover`, `repeat` are accepted too, and `get` prints the label.

`upload` posts the image, registers the returned URL in the surface's gallery, then selects it; `--no-set` stops after registering. **The upload goes to `<SettingsURL>/images/upload/v1`, not the desktop ingress** — `/images` is only proxied on the settings host, so an error naming that origin is expected rather than a misrouted request. Login uploads carry a public access policy because the greeter fetches them before anyone is signed in. Accepted types: png, jpeg, jpg, gif.

`delete` takes an uploaded URL, prompts for confirmation, and if the image being removed is the active wallpaper the upstream falls back to a built-in default. **It only removes the image from the gallery — the file stays in the user's Drive** (`Home/Pictures/Upload`), same as the SPA's delete button. Do not tell the user it freed disk space; removing the file is a separate `files rm`.

## `layout reset`

```bash
olares-cli settings appearance layout reset
olares-cli settings appearance layout reset --yes
```

Drops the launchpad ordering, folders and dock arrangement for the current user and broadcasts the change to open sessions. **There is no undo.** Needs Olares >= 1.12.6, checked before the prompt. The output names the surfaces the upstream reset — observed to be `launchpad, dock` every time, including on a layout that was already default, so it is not a signal that anything had been customized.

## Agent best practices

- **For "switch to dark mode"** → report that it can only be changed in Settings → Appearance in the browser, because the value is a local cookie. Do not hunt for a verb or write the user env instead.
- **For "change my wallpaper"** → ask which surface, since `desktop` and `login` are separate settings and users rarely mean both, then run `wallpaper list <surface>` rather than guessing a number. Refer to built-ins by number, never as `/login/<n>.jpg`.
- **Confirm separately for `layout reset` and `wallpaper delete`.** Both are irreversible for the user's arrangement or gallery, and neither is implied by a request to change the wallpaper.
- **Do not use this subtree for per-app icons or entrance titles** — those are [post-install app configuration](olares-settings-apps.md).

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `unknown surface "..."` | First positional is not `desktop` / `login` | Pass one of the two |
| `<surface> has no built-in wallpaper <n>` | Number past this surface's range | Run `wallpaper list <surface>`; desktop stops at 27, login at 28 |
| `"..." is not a <surface> wallpaper` | Neither a number, an `http(s)://` URL, nor a `/bg/<n>.jpg` value — e.g. a `/login/<n>.jpg` path | Pass a number from `wallpaper list <surface>`; `--force` only for a value a newer release adds |
| `unsupported fill mode "..."` | Mode outside `Fill` / `Stretch` / `Tile` | Use one of the three, as spelled in Settings |
| `unsupported --date-format "..."` | Format not in the SPA's list | Pick from the list above |
| `widget set requires at least one of ...` | Called with no flags | Name the preference to change |
| `read locale: ...` / `read wallpaper: ...` | One section of `get` failed | Retry, then check `settings advanced status` |
| `requires Olares >= 1.12.6 ...` on `widget set` / `layout reset` | Backend predates the verb's upstream | Upgrade Olares; nothing was sent, and no layout was reset |
| `requires Olares >= 1.12.6 ..., but the backend version could not be determined` | Version undetectable | Log in, then `olares-cli profile list --refresh-version` |
| `upload succeeded but the image service returned no imageUrl` | Image service accepted the file but returned an unexpected body | Retry; if it persists the image service is misconfigured |
| `stdin is not a terminal — pass --yes to confirm` | Destructive verb in a script | Pass `--yes` after confirming with the user |
