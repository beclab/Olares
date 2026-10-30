# settings apps: per-entrance configuration

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md), the parent [`../SKILL.md`](../SKILL.md), and [post-install app configuration](olares-settings-apps.md) first — the verb table and the env vector live there.

## The per-entrance editing pipeline

Most per-entrance edits follow a 4-step pattern:

```bash
# 1. Discover entrances (ENTRANCE / STATE / AUTH LEVEL / DOMAIN columns).
olares-cli settings apps entrances list firefox
# 2. Inspect the entrance you want to edit.
olares-cli settings apps domain get firefox www
olares-cli settings apps policy get firefox www
# 3. RMW update (unspecified flags survive).
olares-cli settings apps domain set firefox www --third-level my-firefox
olares-cli settings apps policy set firefox www --default-policy two_factor
```

For the complete route ID workflow, enter the [`olares-chart`](../../olares-chart/SKILL.md) skill and follow its custom URL route.

## `domain set` — custom route ID

```bash
# Set a memorable route ID under the user's Olares domain.
olares-cli settings apps domain set firefox www --third-level my-firefox

# Remove the custom route ID.
olares-cli settings apps domain set firefox www --clear-third-level
```

- The update uses read-modify-write semantics and preserves unspecified settings.
- Changing the route ID makes app-service upgrade the app. The write is refused mid-operation, and the route is live only once the app runs again.
- Read the route ID back with `domain get <app> <entrance>` and confirm the new URL over HTTPS.

## `policy set` — replace sub-policies

```bash
# Any --sub-policy flag REPLACES the existing set. Bulk form: --sub-policies-file ./sub-policies.json
olares-cli settings apps policy set firefox www \
  --sub-policy "uri=/admin,policy=two_factor" \
  --sub-policy "uri=/api,policy=public"
```

- `--default-policy` values: `system` | `one_factor` | `two_factor` | `public`. That flag, `--one-time` and `--valid-duration` follow RMW semantics.
- **Sub-policy entries are REPLACED in full whenever any sub-policy flag is passed** — partial sub-policy edits don't compose safely, so this is intentional. `--clear-sub-policies` drops the set without adding new entries.

## `auth-level set` — no GET endpoint upstream

```bash
olares-cli settings apps auth-level set firefox www --level public
```

| Level | Reachability |
|---|---|
| `private` | Only the app's owner |
| `public` | Any authenticated user |
| `internal` | Intra-cluster traffic only |

> **There is no `auth-level get` verb** because no GET endpoint exists upstream. To inspect the current level, run `apps entrances list <app>` and read the `AUTH LEVEL` column.

