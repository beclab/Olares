# Deciding which layer is wrong

A model that does not answer can be failing in five places, and each has a different fix. Establish which before editing anything: three of the five are not configuration problems at all.

| Layer | Check |
|---|---|
| The CLI's reach into Router | any verb's own error; `olares-cli profile whoami`, `olares-cli market status router` |
| Router's own configuration | `router provider get`, `router model list`, `router route list --kind default` |
| Access control | `router key list`, `router quota list`, `router usage list --status failed` |
| The model application | `router model status`, `router model progress`, `router model diag` |
| Below the application | [`olares-doctor`](../../olares-doctor/SKILL.md) — pods, events, images, resources |

## Start at the top

There is no verb for the first layer. Every verb resolves Router before doing
anything else, so the one that just failed has already answered it — and there
is nothing a separate probe could add that would still be true a moment later.
Three failures look alike and are not:

- **"Router is not installed"** on an admin profile usually means it really is not, and `olares-cli market list --mine` will not name it either. Install it with `olares-cli market install router`.
- **The same message on a non-admin profile** means Router is installed and invisible to you: it is an admin-only application, so its entrance is not in your app list. `olares-cli profile whoami` is what tells the two apart, and nothing here works until an admin acts.
- **An entrance that resolves but does not answer** is Router itself being down. The error names the entrance it reached for; `olares-doctor` is where that goes, not this skill.

## A write that got no answer is not a write that did not happen

A failed read is nothing: run it again. A failed **write** has two histories the CLI cannot tell apart — Router never saw it, or Router applied it and the reply was lost — and only one of them makes "try again" right.

Router has no idempotency key, so a retry is a second request rather than the same one. Nothing the CLI sends changes that, which is why the error carries the consequence instead: it says the outcome is unknown, and what a second identical request would do to that particular route. Read it before acting. Three shapes:

- **Refused on repeat** — providers, routes, quotas, and adding a named model to a provider all have a uniqueness constraint. Retrying is safe and self-checking: a `409` means the first attempt landed, and is not a new problem.
- **Costly on repeat** — `key issue` mints a second key, `provider rollback` appends another credential version, and any `router call` may already have been billed. Look first: `key list`, `provider history`, `usage list --limit 5`.
- **Nothing said** — a route the CLI has no verdict for. Re-read the resource before retrying.

An error that does *not* carry that paragraph reached no connection at all, so nothing was applied and a retry is plain.

## Then ask what Router thinks it has

```
olares-cli router model list
olares-cli router provider get <provider>
olares-cli router route list --kind default
```

- **A provider missing from `provider list`** is usually not missing. An `olares`-sourced provider is listed through the install and while it runs, and stays listed once it is stopped, stopping or failed — those are all states somebody installed. What Router hides is a row nobody installed or nothing can reach: `available` (a catalog entry, one per installable application), `pending`, `unreachable`, `archived`. `router provider get <id>` shows the row anyway, and `olares-cli market status <app>` says which of those it is.
- **A provider with no models** means either that nothing was attached — `model import` or `provider sync-models` — or that Router could not read the application's live list just now, which is the model application layer below.
- **A model that exists but is not callable** has four different fixes, and the `CALLABLE` cell in `router model list` names which one. Read it as the verdict plus the one thing in the way:

| Cell | Whose problem | What to do |
|---|---|---|
| `no · provider disabled` | an admin switch | `provider update <provider> --enable` |
| `no · model disabled` | an admin switch | `model update <model> --enable`; `model list --disabled` finds them all |
| `no · app <phase>` | the application's container — `stopped`, `downloading`, `installing`, `initializing`, `failed` | `olares-cli market` territory; `market status <app>` |
| `no · fetching weights` / `engine loading` / `model service starting` | the model inside a running application | nothing — it gets there on its own; `model progress <model> --watch` |
| `no · model load failed` | the model inside a running application | `model retry <model>`, and `model status <model>` for the reason |

The last two rows are the axis that is easy to miss: a container reports `running` minutes before the model it serves can answer, so an application can be up and its model still unreachable. The two are spelled apart on purpose — `app downloading` is the platform fetching an image, `fetching weights` is the model fetching itself. A model stuck on `engine loading` is an engine that will not start, and checking disk and network finds nothing.

A model application owns its row from the moment it is installed, so a model that never ran is listed rather than absent.

- **A model `router model list` calls callable that `router call models` does not list** is down to the credential, and only when one is being presented: a key restricted to named models sees only those, and `router key list` shows the allowlist. Reaching a model the key may not call is a 404 on the name, which reads like a configuration problem and is not one. A keyless call has no allowlist, so dropping `--api-key` and `OLARES_ROUTER_API_KEY` is the quickest way to rule this out.
- **The reverse — `router call models` serving something `model list` says is not callable** is possible and narrow. Router only asks about weights while the loop that writes the phase is running; when it is not, its gate falls back to asking whether the container is up, and dispatch is more permissive than this list. The data plane is the authority in that disagreement.
- **A call refused with `no_default_model`** means that category has nothing behind it. `route list --kind default` names which do and which do not.
- **A model Router offers a capability for that it turns out not to have** is the projection trailing the application's own card. `router model spec show <model>` says which copy you are reading: `cache` is Router's, and an edit through `router model spec edit` corrects both at once.
- **A call refused for a mode mismatch, or a bare 404 on an audio route**, is the verb and the model disagreeing about what the model does. `router model list` prints the mode; for a local model `router model spec show` prints what the application declares, which is the copy to trust.

## Then whether the caller is allowed

A configuration that is right and a caller that is refused look identical from the outside. `router usage list --status failed --limit 20` separates them: a refused call that Router recorded has its error code in the row, and a call that produced no row at all never got past authentication.

- `invalid_api_key` with type `authentication_error` — Router refusing this key. Check `router key list` for revoked, expired, or an allowlist that excludes the model. Since a key is optional, not presenting one is also a fix.
- `missing_credentials` — the Router being called is older than v2.2.1 and does not read the platform's identity on `/v1`. Upgrade it, or pass `--api-key`.
- `unknown_bfl_user` — the platform vouched for somebody Router has no row for. Any console verb records them; `router model list` is the cheapest.
- An authentication error **without** that type — the vendor refusing Router's credential. `router provider validate <provider>` confirms it, and `provider history` shows what was last rotated.
- `quota_exceeded` — `router quota list` names which of the three ceilings bit.

## Then the model application

Router sees an unfinished local model as an unreachable provider, and reports it as one. The application's own console has the real state:

```
olares-cli router model status <model>
olares-cli router model progress <model>
```

- **A 5xx with an empty body from a call** is nothing listening behind Router: the application is stopped, or the engine has not come up. That empty body is the signature — Router's own refusals always carry a message.
- **`READY` false with a download unfinished** is the normal middle of an install. `router model progress <model> --watch` is the answer, not a fix.
- **A download that failed or stalled** — `router model retry <model>`.
- **An engine alive with no model loaded** — the weights or the card are wrong for this engine; `router model spec show <model>` reads the card and `router model spec edit` corrects it, relaunching the engine when the flags change.
- **An engine that was answering and stopped** — `router model restart <model>` relaunches the process on the card it already has, which is the fix when the configuration is right and the process is not.
- **A model answering, but slowly** — `router model diag gpu <model>` says how much is resident and `router model diag perf <model>` measures it. A model mostly on the CPU is the common answer.
- **An install that never finished** — `router provider get <app>` reports where the application stalled, and `olares-cli market status <app>` carries the Market's own reason.

## Then below it

If the application is stopped, crash-looping, cannot pull its image, or has no GPU binding, none of that is Router's and none of it is the console's. Route to [`olares-doctor`](../../olares-doctor/SKILL.md) for pods, events, logs, images and resources; to [`olares-market`](../../olares-market/SKILL.md) to stop, resume or reinstall; and to [`olares-settings`](../../olares-settings/SKILL.md) for compute bindings.

## Two mistakes worth naming

**Editing Router when the model application is the problem.** Changing a provider's base URL, re-importing models, or re-registering the provider does nothing for a model that has not finished downloading. Check `local progress` before touching the Router row.

**Treating a whole-subtree 404 as a missing resource.** `app` answering 404 means Router's Market proxy is not configured; a `local` route answering 404 can mean that route arrived in a later Model Console version, which `local endpoints` confirms. Neither is a row that went missing.
