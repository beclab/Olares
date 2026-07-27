# market list / categories / get / status (catalog + runtime)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) (especially "What apps do I have? routing" and the `-s` / `-a` matrix) first.
> **Flags & examples:** `olares-cli market <verb> --help` for each verb below.

The read-only family — catalog browsing, runtime status, and the `--mine` inventory view.

## `list` — catalog OR inventory

Two modes, selected by `--mine` / `-m`:

| Mode | Without `--mine` | With `--mine` / `-m` |
|---|---|---|
| Endpoint | `/market/data` | `/market/state` |
| Scope | Catalog | Active profile's apps |
| Default source scope | Auto-selected single source (usually `market.olares`) | **All sources** (matches Market UI "My Terminus" tab) |
| Columns | `NAME / TITLE / VERSION / CATEGORIES` | `NAME / TITLE / VERSION / STATE / SOURCE / CATEGORIES` |

```bash
olares-cli market list                          # catalog (auto-selected source)
olares-cli market list -s market.olares         # pin to a source
olares-cli market list -s cli                   # browse a local source
olares-cli market list -c AI                    # filter by category
olares-cli market list -a                       # every source the user has
olares-cli market list --mine                   # what apps does this user have (all sources)
olares-cli market list -m -s cli                # mine, narrowed to one source
olares-cli market list --no-headers             # table without column headers
olares-cli market list -q                       # exit code only
```

### `--mine` filter (matches SPA "My Terminus" exactly)

Hides only **6 SPA-hidden states** (`pendingCanceled`, `downloadingCanceled`, `downloadFailed`, `installFailed`, `installingCanceled`, `uninstalled`). Everything else stays visible — including in-flight installs, transitional states (`upgrading` / `stopping` / `resuming` / `applyingEnv` / `uninstalling`), and post-install failures (`upgradeFailed` / `stopFailed` / `resumeFailed` / `applyEnvFailed` / `uninstallFailed`).

> **"My apps" is intentionally broader than "completed installs only"** because the SPA's My Terminus tab is too. The user clicked something and wants to monitor / retry / cancel the row.

### Version on `--mine` rows

The version column reflects the **chart the user picked for this row**, NOT the catalog latest. If the user installed 1.0.10 and the marketplace catalog has since moved to 1.2.3, this listing surfaces 1.0.10. During an in-flight upgrade, the row may show the target version while `STATE=upgrading`. **Both are intentional — don't "correct" to the catalog version.**

Clones look up the catalog by `rawAppName`, not their per-instance `name` (e.g. `windowsefe992` → catalog row `windows`), so clones still surface the source app's title and categories.

## `categories`

```bash
olares-cli market categories                    # per-source counts
olares-cli market categories -s market.olares
olares-cli market categories -a -o json         # every source, JSON
olares-cli market categories --no-headers
olares-cli market categories -q
```

Returns category names with per-source app counts. Pure browsing — no state.

## `get <app>`

```bash
olares-cli market get firefox                   # curated summary
olares-cli market get firefox -o json           # full upstream payload
olares-cli market get firefox -s upload         # local-uploaded chart
```

- Table mode = curated summary.
- **JSON mode = full upstream payload.** Use it to fish out fields the table doesn't surface:
  - `cloneable: true` — whether `market clone` is supported for this app
  - `app_simple_info.app_labels` — the same suspend/remove labels `market upgrade` pre-flight checks
  - Declared env-var spec (needed by `market install --env KEY=VALUE`)

> `--no-headers` is deliberately NOT supported on `get` — the layout is key:value, not a row table. **For scripted access, use `-o json`.**

## `status` — runtime probe

```bash
olares-cli market status                        # every installed-app row (resolved source)
olares-cli market status -s market.olares       # pin to a source
olares-cli market status firefox                # single app, with cross-source fallback hint
olares-cli market status firefox -a             # search every source
olares-cli market status firefox --watch        # block until terminal
olares-cli market status firefox -q             # exit code only
```

Columns: `NAME / STATE / OPERATION / PROGRESS / SOURCE`.

### `status <app>` cross-source fallback

- If the row is missing in the resolved source AND in every other source the user has → `app 'X' is not installed (run 'olares-cli market install X' to install it)`
- If the row exists in source `Y` but the user passed `-s X` → stderr hint `App is installed under source 'Y' (not 'X')`, then renders the row anyway. **Agents do not need to retry blindly.**

### `status` (no app) does NOT support `--watch`

`runStatusAll` explicitly rejects `--watch`. Use `status <app> --watch` for a single app, or run `status` in a shell loop with a sleep for the all-apps view.

## "What apps do I have?" decision tree

```
user asks: "show me my apps" / "我的应用" / "list my Olares apps"
  → market list --mine          (broad inventory; SPA My Terminus parity)

user asks: "is <app> installed yet?" / "wait for <app> to be running"
  → market status <app> --watch  (single-app runtime; cross-source fallback)

user asks: "which apps are running right now"
  → market status [-a], then filter STATE=running
    (status is lifecycle inventory, not a running-only view)

user asks: "which installed app title matches <keyword>"
  → search app <keyword>         (visible installed apps by title)

user asks: "which apps are using resources"
  → dashboard applications       (resource ranking, not lifecycle inventory)

user asks: "browse the catalog" / "what apps are available"
  → market list                  (no --mine; /market/data, not /market/state)
```

`list --mine`, `status`, `search app`, and `dashboard applications` overlap but are NOT interchangeable — see the parent SKILL.md "What apps do I have?" table for the full distinction.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `no apps in source 'X'` | Unknown source id, or genuinely empty source | `market list -a` to enumerate; check spelling |
| `app 'X' is not installed` | `status <app>` couldn't find the row anywhere | Install first |
| `cannot --watch 'status' (no app argument)` | `--watch` on multi-app `status` | Use `status <app> --watch` |
| `--no-headers has no effect on this verb` (silent) | `--no-headers` on `get` / mutating verbs | Use `-o json` instead |
| Cross-source mismatch hint on `status -s X firefox` | Row in source Y, not X | Drop `-s` or pass the correct source |
