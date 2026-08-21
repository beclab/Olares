# Names, defaults and access control

Router decides three things about a call before it forwards it: which model the name in the request refers to, whether the caller may call at all, and whether it has room left under its ceiling. These are the verbs for each.

## What a caller may put in `model`

There are two shapes of name and no others.

A name **containing a slash** is a qualified reference, split at the first slash: `openai/gpt-5` is the model `gpt-5` on the provider `openai`, and `openrouter/openai/gpt-5` is the model `openai/gpt-5` on `openrouter`. Every configured model is callable this way with nothing set up — `router model list` shows them.

A name **without a slash** is a route, and has to exist. Three kinds do:

```
olares-cli router route list
olares-cli router route get <route>
olares-cli router route create fast --kind alias --model claude/claude-sonnet-4-5
olares-cli router route create house --kind group --mode chat
olares-cli router route add house openai/gpt-5 --priority 10 --weight 3
olares-cli router route add house olares/qwen3-8b --priority 20
olares-cli router route remove house olares/qwen3-8b
olares-cli router route rename fast quick
olares-cli router route disable fast
olares-cli router route delete fast --yes
```

- An **alias** is a second name for exactly one model. It takes `--model` and no `--mode`: it answers whatever its model answers.
- A **group** is one name served by several models. It takes `--mode`, is created empty, and is filled with `route add`. `--priority` orders the candidates lowest-first and only falls to the next tier when everything in the current one refused; `--weight` splits traffic in proportion within a tier. Every member has to answer the group's mode, and the mode cannot be changed afterwards.
- A **default** is a category Router maintains itself — see below.

Route names are lowercase letters, digits, `-` and `_`, up to 64 characters, and never contain a slash. That is what keeps the two shapes of name apart. Names beginning `default-` belong to Router.

Two states read alike and are not: a route can be **switched off**, and a route can be **on with nothing live behind it**. Both answer 404. `route list` separates them — `CALLABLE` is the caller's question, `BACKENDS` counts the live members against the total — and `route disable` is the reversible way to stop traffic to a name, where `route delete` gives the name up.

Reading routes is open to every console user, since the name is what a person types into their client. Every change is admin-only.

## The categories a caller can ask for instead of a model

```
olares-cli router route list --kind default
olares-cli router route disable chat
olares-cli router route enable chat
```

A caller that does not want to choose names a category: `default-chat`, `default-tts`, one per kind of request. **What a category answers with is not configured.** Router keeps the list of categories in its own code and points each one at an installed model that can serve it, so the answer moves as models are installed, enabled and disabled — that is the design, and `route list --kind default` reports where it currently stands.

A category with nothing behind it is refused rather than approximated. Installing or enabling a model of that kind is what fills it in; there is no setting to point it by hand, and no per-user override.

`route disable <category>` refuses that kind of request without uninstalling anything — the models keep running and stay callable by name. It is a different thing from disabling the model, which takes it away from every caller. A category cannot be renamed or deleted: Router owns the list and would create it again.

## Keys

An `sk-` key is how software calls Router when the platform cannot vouch for it: something running outside Olares, or a call that needs a model allowlist or a budget of its own. An Olares application, a browser and `olares-cli` all reach the data plane on the identity the edge injects, and need none.

```
olares-cli router key issue "ci-runner" --ttl 30d --model openai-main/gpt-4o
olares-cli router key list
olares-cli router key update ci-runner --disable
olares-cli router key revoke ci-runner
```

- The plaintext is printed **once**, at issue. There is no way to read it back; a lost key is replaced, not recovered.
- `--ttl` or `--expires-at` gives it an end. A key with neither never expires, which is rarely what a script wants.
- `--model`, repeatable, restricts what it may reach. Without it the key reaches every model, including one added next month.
- A `--model` entry is either a qualified `<provider>/<model>` or a route name, and the two grant different things. A qualified name grants one backend. A route name grants the name — whatever serves it today, whatever an admin attaches tomorrow, and for a category whatever Router repoints it to. Grant a route when the key should follow somebody else's decision, and a qualified name when it should not.
- `--for-user` issues on someone else's behalf, admin only. The key's calls are attributed to that person.
- `key update` renames, enables, disables, re-expires, or replaces the allowlist; `--clear-models` removes the restriction.
- Disabling and revoking are the same reversible state in Router: both stop the key working and keep its history, and `--enable` brings either back. Nothing is deleted, so past usage stays attributable.

A non-admin sees and manages their own keys. A machine that used an older olares-cli has one in this list that it issued for itself; calls no longer present it, and revoking it is what ends it — see [calling a model](olares-router-calling.md).

## Quotas

A quota is a ceiling on one of four things, and never on more than one at a time:

```
olares-cli router quota set --key ci-runner --budget 50 --warn-at 90
olares-cli router quota set --user alice --rpm 60
olares-cli router quota set --model openai-main/gpt-4o --tpm 100000
olares-cli router quota set --caller-app wise --budget 5
olares-cli router quota list
olares-cli router quota clear --key ci-runner --budget
```

- `--budget` is total spend in US dollars, for all time — not per month. It is the ceiling that stops runaway cost; the others shape load.
- `--rpm` and `--tpm` are requests and tokens per minute.
- `--concurrent` is how many calls the scope may have in flight at once, which is the one that protects a local model from being asked to do two things on one GPU.
- `--warn-at` is the percentage at which a warning is recorded, 80 by default.
- Quotas are always admin-only, including reading them.

A ceiling of `0` is meaningful and is not the same as no ceiling: it refuses everything in that scope, which is how a key or an application is switched off without deleting it. Removing the ceiling is `quota clear`, not `--budget 0`.

A quota on a **model** applies to everybody calling it, a quota on a **user** to everything that person's identity reaches, a quota on a **key** to that key alone, and a quota on a **caller app** to every call carrying that application's appid — which is the only control there is over an application, since it has no key here to revoke. When a call is refused for `quota_exceeded`, `quota list` is what identifies which one bit.

A key scope only binds calls that present that key. `router call` presents none unless one is named, so a `--key` ceiling does not constrain it: **`--user` is the scope that holds a person's `olares-cli` calls**, and the same is true of anyone calling from a browser.

`--caller-app` takes the application's title, its Olares application name or the appid itself. The name is hashed the way the platform hashes it and then matched against an application that exists, so a misspelling is refused rather than becoming a ceiling on nothing.

`quota clear` names the same target and, optionally, which ceiling to lift: `--budget`, `--rpm`, `--tpm` or `--concurrent` alone removes one and leaves the rest, and naming none removes the quota entirely.

## The people Router knows

The people on this Olares are [`olares-settings`](../../olares-settings/SKILL.md)'s
`settings users list`. Router has no separate list of its own to consult: the role
is Olares' — being an Olares admin is what makes you a Router admin — and Router
stores what the edge told it rather than deciding anything.

What Router adds is a row per person, created the first time that person's
identity reaches it, so a freshly created Olares account is unknown to Router
until it does. That is why `key issue --for-user` and `quota set --user` can only
name someone who has already arrived; naming anyone else is refused with the
names Router does know.

Restricting what somebody may call is a property of a key, not of a person:
`key issue --model` and `key update --model` are the allowlist, and a key
carrying one sees only those models in `router call models` and in `GET /v1/models`.
There is no per-person allowlist to set.

## The applications that call Router

```
olares-cli market list --mine
olares-cli router usage summary --by caller_app
olares-cli router quota set --caller-app wise --budget 5
```

**An application is not registered with Router and cannot be.** Olares vouches for it at its own edge, so the request arrives already carrying the application's identity — an `appid`, which is the application name hashed, or the name itself for a system application. There is no row to create, nothing to hand over, and correspondingly nothing to revoke: an application that may not call is one that is not installed, or one whose ceiling is zero.

That is why there are three different questions here and no single "callers" list:

- **what is installed** — `olares-cli market list --mine`, the whole machine, whatever each application does.
- **what has called, and what it cost** — `usage summary --by caller_app`. `--caller-app` filters to one, by title, application name or appid.
- **what it may spend** — a quota scoped to the appid. Setting it to zero is how an application is stopped.

Do not confuse an application that *calls* Router with a *model* application, which serves models rather than consuming them. One machine has both, and one application can be both.
