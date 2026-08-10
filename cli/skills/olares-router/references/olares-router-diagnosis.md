# Deciding which layer is wrong

A model that does not answer can be failing in five places, and each has a different fix. Establish which before editing anything: three of the five are not configuration problems at all.

| Layer | Check |
|---|---|
| The CLI's reach into Router | `router status` |
| Router's own configuration | `router provider get`, `router list`, `router default show` |
| Access control | `router key list`, `router quota list`, `router usage list --status failed` |
| The model application | `router local status`, `router local progress` |
| Below the application | [`olares-doctor`](../../olares-doctor/SKILL.md) — pods, events, images, resources |

## Start at the top

`router status` answers the first layer completely: whether Router is installed, which listing and entrance it was reached at, whether it is healthy, and what role it gives you. Three failures look alike and are not:

- **"Router is not installed"** on an admin profile usually means it really is not, and `olares-cli router app catalog` will fail too. Install it from the Market.
- **The same message on a non-admin profile** means Router is installed and invisible to you: it is an admin-only application, so its entrance is not in your app list. Nothing here works until an admin acts.
- **An entrance that resolves but does not answer** is Router itself being down, and is `olares-doctor` territory rather than anything in this skill.

## Then ask what Router thinks it has

```
olares-cli router list
olares-cli router provider get <provider>
olares-cli router default show
```

- **A provider missing from `provider list`** is usually not missing. Router hides an `olares`-sourced provider while its application is not running; `router provider get <id>` shows it anyway, and `olares-cli market status <app>` says why.
- **A provider with no models** means either that nothing was attached — `provider models import` or `sync-models` — or that Router could not read the application's live list just now, which is the model application layer below.
- **A model that exists but is not offered** is disabled: `router list --disabled` finds it, `provider models update ... --enable` restores it.
- **A call that says no model was named** means that mode has no default. `default show` names the modes that have none.

## Then whether the caller is allowed

A configuration that is right and a caller that is refused look identical from the outside. `router usage list --status failed --limit 20` separates them: a refused call that Router recorded has its error code in the row, and a call that produced no row at all never got past authentication.

- `invalid_api_key` with type `authentication_error` — Router refusing this key. Check `router key list` for revoked, expired, or an allowlist that excludes the model.
- An authentication error **without** that type — the vendor refusing Router's credential. `router provider validate <provider>` confirms it, and `provider history` shows what was last rotated.
- `quota_exceeded` — `router quota list` names which of the three ceilings bit.

## Then the model application

Router sees an unfinished local model as an unreachable provider, and reports it as one. The application's own console has the real state:

```
olares-cli router local status <app>
olares-cli router local progress <app>
```

- **A 5xx with an empty body from a call** is nothing listening behind Router: the application is stopped, or the engine has not come up. That empty body is the signature — Router's own refusals always carry a message.
- **`READY` false with a download unfinished** is the normal middle of an install. `local progress --watch` is the answer, not a fix.
- **A download that failed or stalled** — `local retry`.
- **An engine alive with no model loaded** — the weights or the card are wrong for this engine; `local spec show` then `local restart`.
- **A model answering, but slowly** — `local gpu` says how much is resident and `local perf` measures it. A model mostly on the CPU is the common answer.
- **An install that never finished** — `router app tasks <provider>` carries the Market's own reason.

## Then below it

If the application is stopped, crash-looping, cannot pull its image, or has no GPU binding, none of that is Router's and none of it is the console's. Route to [`olares-doctor`](../../olares-doctor/SKILL.md) for pods, events, logs, images and resources; to [`olares-market`](../../olares-market/SKILL.md) to stop, resume or reinstall; and to [`olares-settings`](../../olares-settings/SKILL.md) for compute bindings.

## Two mistakes worth naming

**Editing Router when the model application is the problem.** Changing a provider's base URL, re-importing models, or re-registering the provider does nothing for a model that has not finished downloading. Check `local progress` before touching the Router row.

**Treating a whole-subtree 404 as a missing resource.** `trace` answering 404 means observability is off; `app` answering 404 means Router's Market proxy is not configured; a `local` route answering 404 can mean that route arrived in a later Model Console version, which `local endpoints` confirms. None of the three is a row that went missing.
