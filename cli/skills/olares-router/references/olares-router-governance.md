# Defaults and access control

Router decides three things about a call before it forwards it: which model answers when none was named, whether the caller may call at all, and whether it has room left under its ceiling. These are the verbs for each.

## Which model answers by default

```
olares-cli router default show
olares-cli router default set --mode chat=<provider>/<model>
olares-cli router default set --mine --mode chat=<provider>/<model>
olares-cli router default clear --mode chat
```

Defaults are resolved per mode, through three layers:

1. **The user's own override** — set with `--mine`, wins for that person only, and is not admin-only.
2. **The workspace default** — the admin's choice for everybody without an override.
3. **The fallback** — the oldest enabled model of that mode, when no default was ever set.

`default show` prints what is in effect and which layer it came from, and names the modes that have no default at all. A call in one of those modes must name its model.

`default clear` drops one layer and falls back to the one beneath it; it does not disable the mode.

Setting several modes at once is atomic — all of them land or none do. The CLI also refuses a model whose mode does not match the mode being set, which Router itself would accept and then fail on at call time.

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
- `--model <provider>/<model>`, repeatable, restricts what it may reach. Without it the key reaches every model, including one added next month.
- `--for-user` issues on someone else's behalf, admin only. The key's calls are attributed to that person.
- `key update` renames, enables, disables, re-expires, or replaces the allowlist; `--clear-models` removes the restriction.
- Disabling and revoking are the same reversible state in Router: both stop the key working and keep its history, and `--enable` brings either back. Nothing is deleted, so past usage stays attributable.

A non-admin sees and manages their own keys. The key `router call` keeps for this machine is an ordinary key in this list — see [calling a model](olares-router-calling.md).

## Quotas

A quota is a ceiling on one of three things, and never on more than one at a time:

```
olares-cli router quota set --key ci-runner --budget 50 --warn-at 90
olares-cli router quota set --user alice --rpm 60
olares-cli router quota set --model openai-main/gpt-4o --tpm 100000
olares-cli router quota list
olares-cli router quota clear --key ci-runner --budget
```

- `--budget` is total spend in US dollars, for all time — not per month. It is the ceiling that stops runaway cost; the others shape load.
- `--rpm` and `--tpm` are requests and tokens per minute.
- `--warn-at` is the percentage at which a warning is recorded, 80 by default.
- Quotas are always admin-only, including reading them.

A quota on a **model** applies to everybody calling it, a quota on a **user** to everything that person's identity reaches, and a quota on a **key** to that key alone. When a call is refused for `quota_exceeded`, `quota list` is what identifies which of the three bit.

`quota clear` names the same target and, optionally, which ceiling to lift: `--budget`, `--rpm` or `--tpm` alone removes one and leaves the rest, and naming none removes the quota entirely.

## The people Router knows

```
olares-cli router user list
```

A Router user row appears the first time that person's identity reaches Router, so a freshly created Olares account is absent until it does. That is why `key issue --for-user` and `quota set --user` can only name someone who has already arrived. Admin only.

Router's roles are its own: being an Olares admin is what makes you a Router admin, but the user list, the roles and the disabled state live in Router. `router whoami` reports what it decided about you.

## The applications that call Router

```
olares-cli router caller list
olares-cli router caller archive <app>
```

Router discovers a **caller** the first time an application calls its data plane, and creates a row and an internal key for it. Nothing has to be registered in advance — that is deliberate, so an application shipped with no configuration works.

`caller archive` stops that application calling. It does not uninstall it, and an archived application does not revive itself by calling again: the next call is refused rather than re-creating the row.

Do not confuse this with `router app`, which is a model application from the Market. A caller is anything that consumes models; a model application is something that serves them. One machine can have both, and the same application can appear in both lists for good reason.
