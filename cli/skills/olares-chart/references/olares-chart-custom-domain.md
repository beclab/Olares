# Custom URL: route ID (post-deploy)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. The app must be installed and `running`.
> **Verb mechanics** live in the [`olares-settings`](../../olares-settings/SKILL.md) skill under [per-entrance configuration](../../olares-settings/references/olares-settings-apps-entrance.md).

A custom route ID gives an application a memorable URL within the user's Olares domain. Olares manages DNS and TLS for that URL.

## Set a custom route ID

This workflow runs against a live app through `olares-cli settings apps domain`, so it requires login and an entrance name. Discover the entrance first:

```bash
olares-cli settings apps entrances list <app>
olares-cli settings apps domain get <app> <entrance>
```

The user manual calls this the **custom route ID**; the CLI flag and wire field call the same thing `--third-level` / `third_level_domain`.

```bash
olares-cli settings apps domain set <app> <entrance> --third-level <prefix>
```

The old URL keeps working; the new one is `https://<prefix>.<zone>`.

`domain set` triggers an app upgrade. The call is refused while the app is mid-operation. Wait for the app to read `running` again, then read the route ID back and confirm the new URL over HTTPS.

## Constraints

- `auth`, `desktop`, and `wizard` are reserved prefixes.
- A prefix must not duplicate another entrance's custom route ID or default domain in the same user's zone. The write rejects conflicts. Use `doctor thirdleveldomain` from the [`olares-doctor`](../../olares-doctor/SKILL.md) skill to audit existing routes.
- Configuration applies to one entrance at a time. Repeat the procedure for each entrance that needs a custom route ID.
- Updates use read-modify-write semantics. To remove the custom route ID, pass `--clear-third-level`.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `third_level_domain "x" is reserved and cannot be used` | The prefix is `auth`, `desktop`, or `wizard` | Pick another prefix |
| `... is already used by entrance "e" of app "a"` / `... conflicts with the default domain of ...` | The prefix is taken in this zone | Pick another prefix; `doctor thirdleveldomain` lists conflicts |
| `<op> operation is not allowed for <state> state` | `domain set` arrived while the app was mid-operation | Wait until the app reads `running`, then retry |
