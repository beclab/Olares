# Usage and audit

Two separate records, answering two different questions. Reaching for the wrong one is the usual reason a question looks unanswerable.

| Question | Record |
|---|---|
| What was called, by whom, and what did it cost? | `router usage` — one row per model call |
| Who changed Router, and to what? | `router audit` — one row per management write |

There is no third record. Router accepted OTLP spans from agent frameworks and served them back for a while; the tables were dropped and the routes withdrawn, because a usage row already carries the model, tokens, cost, latency, status and failure reason, and keeping request bodies to add to that bought compliance exposure rather than insight. A `router trace` in an older transcript or script no longer exists.

## Usage

```
olares-cli router usage summary --since 7d
olares-cli router usage summary --by user --since 30d
olares-cli router usage summary --by model,provider,caller_app --since 7d
olares-cli router usage list --status failed --limit 20
olares-cli router usage export --since 30d --out calls.csv
olares-cli router usage retention
```

`summary` adds up; `list` explains a total by showing the individual calls behind it; `export` writes the same rows as CSV for a spreadsheet.

- `--by` groups a summary by `model`, `provider`, `user`, `caller_app`, `day` or `hour`. `day` and `hour` are how a spike gets located; `caller_app` is how it gets attributed.
- Several groupings, comma-separated, come back from one request: `--by model,provider,user` prints a table each and one set of totals, because every grouping counts the same calls. `hour` is the exception and is answered on its own — an hourly series grows with the window where the others are a bounded set of names.
- Filters compose across all three verbs: `--model`, `--provider`, `--key`, `--user`, `--caller-app`, `--status`, `--tag`, `--since`, `--until`. `--since` takes an instant or a span like `24h` or `7d`. `--caller-app` names an application by its title, its Olares application name or the appid a row shows — the same spelling `quota set --caller-app` takes.
- `--status failed` is the one to reach for after a complaint: a failed call still carries the error code Router returned, so the reason is in the row.

Every accepted call becomes a row, including one the upstream then refused. Cost comes from the prices on the model row, so a model imported without prices records tokens and no money — that is a configuration gap in the model row, not missing usage.

A row carries a key only when the call presented one. `router call` presents none by default, so its rows have an empty key and are attributed to the person: **`--key` will not find them, and `--user` is how they are read.** A row with no key is the normal shape for a call made from `olares-cli` or from a browser, not a record that lost its attribution.

Scope follows the role. A non-admin sees only their own calls; `--user` and `--caller-app` are admin-only, because they are what makes another person's usage visible.

### Retention

Usage is kept in two shapes, and only one of them expires. Daily totals per model, person, provider and application are kept for good; the individual calls behind them are deleted on a window `router usage retention` reports and `--days` changes.

So `summary` answering for a month whose `list` is empty is the setting working, not a gap. A shorter window applies at once — rows outside it are deleted rather than left for a nightly sweep — and `--days 0` is a real setting that keeps no per-call rows at all, leaving totals, quotas and the by-day export intact.

Admin only, the read included: how long records live is a property of the deployment.

## Audit

```
olares-cli router audit list --since 24h
olares-cli router audit list --target-type provider --failed
olares-cli router audit get <id>
```

An audit row records who changed what, when, the action (`provider.create`, `provider.update`, and so on), the target, the status code Router returned, and the state before and after.

- `audit list` filters by `--action`, `--actor`, `--target-type`, `--target-id`, `--status-class`, `--failed`, `--since`, `--until`.
- `audit get` shows one change with its before and after. For a **rejected** write the second block is labelled a refusal rather than an after state, because nothing changed: what is stored is the error Router sent back.
- `--failed` and `--status-class` merge two queries, so the count line reads differently from an unfiltered page. Narrow with `--action` or `--target-type` rather than paging through it.

Audit is admin-only, and it is the record to check first when a configuration is not what someone expected: it distinguishes "nobody changed it" from "it was changed and rejected" from "it was changed successfully by someone else".

Treat the record as complete for writes and partial for reads. A provider that was merely listed leaves no row, while some reads that reach out to the platform on your behalf — browsing the Market's model applications through Router, for one — do record `olares.model_apps.view`. An absent row is therefore evidence that nothing was *changed*, not that nobody looked.

Audit rows are not subject to the usage retention window; that setting governs per-call spend rows only.
