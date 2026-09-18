# settings appearance

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings appearance --help` and `olares-cli settings appearance <noun> <verb> --help`.

Mirror the Settings → Appearance page: locale, desktop widget preferences, wallpaper, and the desktop layout reset. Every verb is per-user; none of them require admin or owner.

**There is no theme verb.** A desktop's light/dark state is a browser cookie the SPA never sends upstream, so no CLI can change it. Say that plainly instead of reaching for `settings advanced env user set --var OLARES_USER_THEME=...`, which writes a variable no Olares interface reads.

## Sub-tree

| Verb | Floor | Needs | Status | Purpose |
|---|---|---|---|---|
| `get` | normal | any | VERIFIED | Read the whole page: locale + widget + wallpaper |
| `language set <locale>` | normal | any (5 of 7 locales >= 1.12.7) | VERIFIED | Set the system language |
| `widget set [flags]` | normal | >= 1.12.6 | VERIFIED | Set desktop widget preferences |
| `wallpaper list <surface>` | normal | any | VERIFIED | Show the built-in range and uploaded images |
| `wallpaper set <surface> <number\|url>` | normal | any | VERIFIED | Select a wallpaper |
| `wallpaper style set <surface> <Fill\|Stretch\|Tile>` | normal | any | VERIFIED | Set the wallpaper fill mode |
| `wallpaper upload <surface> --file` | normal | any | VERIFIED | Upload a local image and select it |
| `wallpaper delete <surface> <url>` | normal | any | VERIFIED | Remove an uploaded image |
| `layout reset` | normal | >= 1.12.6 | VERIFIED | Restore the default desktop layout |

`<surface>` is always `desktop` or `login`. Every fixed value in this sub-tree — surfaces, fill modes, date patterns, locales — is matched without regard to case and stored in its canonical spelling, so `Desktop`, `stretch`, `yy.m.d` and `en-us` all work. `language set` keeps a `--force` for a locale shipped ahead of this CLI, and a forced value is passed through as typed.

## Version gates

Two verbs need Olares 1.12.6+, where their upstream arrived: `widget set` (the widget preferences API) and `layout reset` (the desktop layout reset route). Every `wallpaper` verb works on any supported backend, and so does `get`.

`language set` has a gate of its own on the value rather than the verb: `en-US` and `zh-CN` work everywhere, while `de-DE`, `es-ES`, `it-IT`, `fr-FR` and `ja-JP` need Olares >= 1.12.7, which is where their translations shipped. The version is only resolved when one of those five is asked for, so the two universal locales still work on a backend whose version cannot be established. Warn the user that those five change LarePass only: the browser desktop still ships `en-US` and `zh-CN` bundles alone and falls back to English for the others, and both read this one field.

On an older backend the gated verbs and locales fail up front with a message naming the requirement and the detected version (see [Common errors](#common-errors)); nothing is sent. `layout reset` is gated **before** its confirmation prompt, so an older backend never asks the user to confirm a reset it cannot perform. Treat daily builds by their `major.minor.patch` base (`1.12.6-20260203` is the 1.12.6 line). Follow the shared auth/version gate when the version cannot be established.

`get` is the one verb that degrades instead of failing: below 1.12.6 it reads locale and wallpaper — which exist everywhere — and renders the widget section as `requires Olares >= 1.12.6`. With `-o json` that section is `null`, the key still present so a caller can tell a gated section from one this CLI does not know about. On 1.12.5 the locale section also returns an empty `timezone`, which renders as `-`.

## `get` is the only read verb

```bash
olares-cli settings appearance get
olares-cli settings appearance get -o json
```

There is no `widget get` or `wallpaper get`. One `get` reads all three upstream endpoints (`/api/wallpaper/config/system`, `/api/widget`, `/api/wallpaper`) — the widget one only on 1.12.6+, see the [version gates](#version-gates).

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

**JSON carries the stored values; the table renames them for the reader** — `"desktop": "/bg/3.jpg"` prints as `built-in 3`, `"desktopStyle": "cover"` as `Stretch`, `"dateFormat": "M/D/YY"` as `M/D/YY (8/27/26)`, and `"language": "zh-CN"` as `zh-CN (简体中文)`, matching how the verbs and Settings name them. Script against `-o json`.

## `widget set` changes only the flags you pass

```bash
olares-cli settings appearance widget set --show-widgets=false
olares-cli settings appearance widget set --24-hour=false --date-format M/D/YY
olares-cli settings appearance widget set --show-dashboard=true
```

The upstream POST replaces the whole preferences object, so the CLI reads the current values first and sends the unnamed ones back untouched. Passing no flag at all is an error rather than a no-op write.

The four preferences are the whole Widget section of the page: `--show-widgets` is the master switch, and `--24-hour` / `--date-format` / `--show-dashboard` are the ones the desktop displays only while it is on (their values are still stored while it is off). Boolean flags need the equals sign — `--24-hour false` sets it to true and then rejects the leftover word, naming the correct form.

What the user sees: the widget is the clock block on the desktop — time, date, and CPU/memory rings beneath them. `--show-widgets=false` hides the whole block; `--show-dashboard=false` hides only the rings and keeps the clock. So a request to hide resource or usage readings is `--show-dashboard`, not the master switch.

`--date-format` is checked against the SPA's list (`YYYY/MM/DD`, `D/M/YY`, `M/D/YY`, `DD/MM/YYYY`, `DD.MM.YYYY`, `DD-MM-YYYY`, `YYYY.MM.DD`, `YYYY-MM-DD`, `YY/MM/DD`, `YY-M-D`, `YY.M.D`) because the upstream stores any string it is handed and an unlisted one silently fails to render in the desktop clock. Case does not matter and the canonical spelling is what gets stored, so `--date-format yy.m.d` sets `YY.M.D`. Rejections print the whole list with each pattern rendered as today's date, which is how Settings shows it — `YY-M-D` and `YY.M.D` are otherwise indistinguishable. There is no `--force`: a pattern a newer Settings adds needs a newer CLI to render it.

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
| `unsupported locale "..."` | Not one of the seven the SPA carries | Pick from the list the error prints; `--force` only for a locale a newer release adds |
| `locale de-DE (Deutsch) needs Olares >= 1.12.7` | One of the five 1.12.7 locales on an older backend | Upgrade Olares, or stay on `en-US` / `zh-CN`; nothing was sent |
| `unsupported --date-format "..."` | Pattern not in the SPA's list | Pick from the list the error prints, each shown as today's date |
| `widget set requires at least one of ...` | Called with no flags | Name the preference to change |
| `read locale: ...` / `read wallpaper: ...` | One section of `get` failed | Retry, then check `settings advanced status` |
| `requires Olares >= 1.12.6 ...` on `widget set` / `layout reset` | Backend predates the verb's upstream | Upgrade Olares; nothing was sent, and no layout was reset |
| `requires Olares >= 1.12.6 ..., but the backend version could not be determined` | Version undetectable | Log in, then `olares-cli profile list --refresh-version` |
| `upload succeeded but the image service returned no imageUrl` | Image service accepted the file but returned an unexpected body | Retry; if it persists the image service is misconfigured |
| `stdin is not a terminal — pass --yes to confirm` | Destructive verb in a script | Pass `--yes` after confirming with the user |
