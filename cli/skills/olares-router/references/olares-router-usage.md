# Usage, audit and traces

Three separate records, answering three different questions. Reaching for the wrong one is the usual reason a question looks unanswerable.

| Question | Record |
|---|---|
| What was called, by whom, and what did it cost? | `router usage` — one row per model call |
| Who changed Router, and to what? | `router audit` — one row per management write |
| What did an agent framework do around a call? | `router trace` — the spans it reported |

## Usage

```
olares-cli router usage summary --since 7d
olares-cli router usage summary --by user --since 30d
olares-cli router usage list --status failed --limit 20
olares-cli router usage export --since 30d --out calls.csv
```

`summary` adds up; `list` explains a total by showing the individual calls behind it; `export` writes the same rows as CSV for a spreadsheet.

- `--by` groups a summary by `model`, `provider`, `user`, `caller_app`, `day` or `hour`. `day` and `hour` are how a spike gets located; `caller_app` is how it gets attributed.
- Filters compose across all three verbs: `--model`, `--provider`, `--key`, `--user`, `--caller`, `--status`, `--tag`, `--since`, `--until`. `--since` takes an instant or a span like `24h` or `7d`.
- `--status failed` is the one to reach for after a complaint: a failed call still carries the error code Router returned, so the reason is in the row.

Every accepted call becomes a row, including one the upstream then refused. Cost comes from the prices on the model row, so a model imported without prices records tokens and no money — that is a configuration gap in `provider models`, not missing usage.

Scope follows the role. A non-admin sees only their own calls; `--user` and `--caller` are admin-only, because they are what makes another person's usage visible.

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

Treat the record as complete for writes and partial for reads. A provider that was merely listed leaves no row, while some reads that reach out to the platform on your behalf — browsing the Market catalog through Router, for one — do record `olares.market.catalog_view`. An absent row is therefore evidence that nothing was *changed*, not that nobody looked.

## Traces

```
olares-cli router trace list --since 24h
olares-cli router trace get <trace-id>
olares-cli router trace capture
```

A trace is what an agent framework reported around a call — the spans, their timings, and optionally the prompt and completion content. Nothing appears here unless something instrumented sent it; an ordinary `router call` produces a usage row and no trace.

Two properties are not negotiable:

- **Traces are user-scoped even for an admin.** An admin sees their own traces, not everyone's, because the content is prompts.
- **Content capture is a preference under a policy.** `trace capture` with no flag reports your preference, the deployment's policy, and what the two add up to. `--on` and `--off` change the preference; a deployment that forbids capture refuses to turn it on, and that is the design rather than a limitation. The change applies to what is recorded from then on: it neither adds content to stored traces nor removes it.

If the whole `trace` subtree answers 404, this Router has observability switched off and stores no traces at all. That is a deployment setting, not a missing row, and `router usage` still works — one row per call rather than the spans around it.
