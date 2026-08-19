# Names, defaults and access control

Router decides three things about a call before it forwards it: which model the name in the request refers to, whether the caller may call at all, and whether it has room left under its ceiling. These are the verbs for each.

## What a caller may put in `model`

There are two shapes of name and no others.

A name **containing a slash** is a qualified reference, split at the first slash: `openai/gpt-5` is the model `gpt-5` on the provider `openai`, and `openrouter/openai/gpt-5` is the model `openai/gpt-5` on `openrouter`. Every configured model is callable this way with nothing set up — `router list` shows them.

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
olares-cli router default show
olares-cli router default disable chat
olares-cli router default enable chat
```

A caller that does not want to choose names a category: `default-chat`, `default-tts`, one per kind of request. **What a category answers with is not configured.** Router keeps the list of categories in its own code and points each one at an installed model that can serve it, so the answer moves as models are installed, enabled and disabled — that is the design, and `default show` reports where it currently stands.

A category with nothing behind it is refused rather than approximated. Installing or enabling a model of that kind is what fills it in; there is no setting to point it by hand, and no per-user override.

`default disable` refuses that kind of request without uninstalling anything — the models keep running and stay callable by name. It is a different thing from disabling the model, which takes it away from every caller. A category cannot be renamed or deleted: Router owns the list and would create it again.

## Keys

An `sk-` key is how software that is not an Olares application calls Router.

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

A non-admin sees and manages their own keys. The key `router call` keeps for this machine is an ordinary key in this list — see [calling a model](olares-router-calling.md).

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
- `--warn-at` is the percentage at which a warning is recorded, 80 by default.
- Quotas are always admin-only, including reading them.

A quota on a **model** applies to everybody calling it, a quota on a **user** to everything that person's identity reaches, a quota on a **key** to that key alone, and a quota on a **caller app** to every call carrying that application's appid — which is the only control there is over an application, since it has no key here to revoke. When a call is refused for `quota_exceeded`, `quota list` is what identifies which one bit.

`--caller-app` takes the application's title, its Olares application name or the appid itself. The name is hashed the way the platform hashes it and then matched against an application that exists, so a misspelling is refused rather than becoming a ceiling on nothing.

`quota clear` names the same target and, optionally, which ceiling to lift: `--budget`, `--rpm` or `--tpm` alone removes one and leaves the rest, and naming none removes the quota entirely.

## The people Router knows

```
olares-cli router user list
```

A Router user row appears the first time that person's identity reaches Router, so a freshly created Olares account is absent until it does. That is why `key issue --for-user` and `quota set --user` can only name someone who has already arrived. Admin only.

The role is Olares': being an Olares admin is what makes you a Router admin, and Router stores what the edge told it rather than deciding anything. The user list and the disabled state are Router's own. `router status` reports the record it holds for you, including a model allowlist if somebody set one.

## The applications that call Router

```
olares-cli router app installed
olares-cli router usage summary --by caller_app
olares-cli router quota set --caller-app wise --max-budget 5
```

**An application is not registered with Router and cannot be.** Olares vouches for it at its own edge, so the request arrives already carrying the application's identity — an `appid`, which is the application name hashed, or the name itself for a system application. There is no row to create, nothing to hand over, and correspondingly nothing to revoke: an application that may not call is one that is not installed, or one whose ceiling is zero.

That is why there are three different questions here and no single "callers" list:

- **what is installed** — `app installed`, the whole machine, whatever it does. Readable by any console user.
- **what has called, and what it cost** — `usage summary --by caller_app`. `--caller` filters to one, by title, application name or appid.
- **what it may spend** — a quota scoped to the appid. Setting it to zero is how an application is stopped.

Do not confuse this with `router app catalog`, which is a *model* application — something that serves models rather than consumes them. One machine has both, and one application can be both.
