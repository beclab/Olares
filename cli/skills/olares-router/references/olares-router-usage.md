# Usage and audit

Two separate records, answering two different questions. Reaching for the wrong one is the usual reason a question looks unanswerable.

| Question | Record |
|---|---|
| What was called, by whom, and what did it cost? | `router usage` — one row per model call |
| Who changed Router, and to what? | `router audit` — one row per management write |

There is no third record. Router accepted OTLP spans from agent frameworks and served them back for a while; the tables were dropped and the routes withdrawn, because a usage row already carries the model, tokens, cost, latency, status and failure reason, and keeping request bodies to add to that bought compliance exposure rather than insight. A router `trace` verb in an older transcript or script no longer exists.

## Usage

```
olares-cli router usage summary --since 7d
olares-cli router usage summary --by user --since 30d
olares-cli router usage summary --by model,provider,caller_app --since 7d
olares-cli router usage list --status failed --limit 20
olares-cli router usage list --status in_progress
olares-cli router usage list --mode video_generation --sort-by cost_usd --limit 10
olares-cli router usage list --session task-4711
olares-cli router usage summary --by model --rank-by tokens --since 7d
olares-cli router usage list --mode audio,tts --limit 20
olares-cli router usage export --since 30d --out calls.csv
olares-cli router usage retention
olares-cli router usage apps
```

`summary` adds up; `list` explains a total by showing the individual calls behind it; `export` writes the same rows as CSV for a spreadsheet.

- `--by` groups a summary by `model`, `provider`, `user`, `caller_app`, `day` or `hour`. `day` and `hour` are how a spike gets located; `caller_app` is how it gets attributed.
- Several groupings, comma-separated, come back from one request: `--by model,provider,user` prints a table each and one set of totals, because every grouping counts the same calls. `hour` is the exception and is answered on its own — an hourly series grows with the window where the others are a bounded set of names.
- Filters compose across all three verbs: `--model`, `--provider`, `--key`, `--user`, `--caller-app`, `--status`, `--mode`, `--session`, `--tag`, `--since`, `--until`. `--since` takes an instant or a span like `24h` or `7d`. `--caller-app` names an application by its title, its Olares application name or the appid a row shows — the same spelling `quota set --caller-app` takes.
- `--status failed` is the one to reach for after a complaint: a failed call still carries the error code Router returned, so the reason is in the row.
- `--sort-by` is on `list` alone, and it ranks the whole matching set rather than the page: `--sort-by cost_usd` answers "the most expensive call this week" without exporting anything. It takes `created_at`, `cost_usd`, `total_tokens`, `latency_ms`, `ttft_ms` or `model_name`, with `--sort-order asc|desc`.
- `--rank-by` is the summary's counterpart and orders the buckets rather than the rows behind them: `--by model --rank-by tokens` is "which model ate the context window", which cost order does not answer when the expensive model is the one called twice. They are separate flags because they sort different things.
- `--mode` takes several, comma-separated. `--mode audio,tts` is one request over both halves of speech, which matters because synthesis is its own mode now and no longer counted with recognition.

Every accepted call becomes a row, including one the upstream then refused. Cost comes from the prices on the model row, so a model imported without prices records tokens and no money — that is a configuration gap in the model row, not missing usage.

### A row is not a token count

`--status in_progress` is the fourth status, and it is the same call as the row that follows it rather than a different kind of event: Router writes the row when it admits a call and prices it when the call ends. So a running row's cost reads `pending`, not `$0` — a video generation sits there for minutes, and a zero would be read as free. A call that stayed in that state long after the work should have finished is a call whose completion never arrived, which is a Router-side question rather than a billing one.

What a call is priced by depends on the mode, so the quantity column is whichever one the row filled in: tokens for text, seconds for audio and video, pictures for images, tracks and meshes for the two newest families, pages for OCR, objects for 3D, queries for search. Reading the row rather than the mode is deliberate — a mode this CLI has never heard of still reports the right figure. A dash means nothing measurable came back.

An asynchronous generation is counted by what it delivered rather than by the request that submitted it: the pictures the result actually carries, and one duration per song a track produced rather than one for the generation. Both were being read as a single unit, so a four-picture workflow and a three-song track reported the first one and gave the rest away.

**`TOOK` is not how long the request was open.** An asynchronous call — a diarization, a video, a long synthesis — is accepted, worked on, and collected later, so the request that returns the result is a JSON forward measured in milliseconds while the work took minutes. The row therefore carries both: `latency_ms` is the request and `job_ms` is the work, and `TOOK` shows the work when there is one. A row with no `job_ms` was synchronous and its request time is the whole story. This matters when reading a slow report: a 273-second diarization used to appear as five milliseconds, which made the engine look idle.

Three zeros mean three different things and `list` says which under the table:

- **still running** — priced when it ends, as above.
- **`unpriced`** — the quantity was measured and the model row carries no rate for it. Music and 3D generation are permanently here today: Router has no price list for either, so the traffic is real and the money is not. Fixing it means putting prices on the model row — a picture rate is per picture, so an image workflow that reads `unpriced` while its pictures are counted has only the tiered rates (resolution, quality) configured and not that one.
- **`audio_unmetered`** — the engine reported no duration, and audio is charged by the second, so there was nothing to multiply.

**`MODEL` is what answered, not always what was asked for.** A row records the name the caller wrote, and a call that named no model wrote a category — `default-tts-clone`, `default-stt`, `default-ocr`. That is a true record of the request and says nothing about which engine spent the time, so the column resolves the category to the model behind it and the page counts the rows it did that for. `-o json` carries both: `model_name` is what was asked for, `served_model_name` is what ran. A refusal that never reached a model has only the one name, and the column shows it.

Both names matter for different questions. `--model` and `usage summary --by model` group by the model that answered, so a category and the qualified name are one bucket; `model_name` is what somebody searches by when they remember what they typed.

`--session` follows one piece of work across the calls it took. An agent or a split audio job sends a session id, so the six calls that transcribed one recording are one filter apart instead of six rows to spot by timestamp.

A row carries a key only when the call presented one. `router call` presents none by default, so its rows have an empty key and are attributed to the person: **`--key` will not find them, and `--user` is how they are read.** A row with no key is the normal shape for a call made from `olares-cli` or from a browser, not a record that lost its attribution.

Scope follows the role. A non-admin sees only their own calls; `--user` and `--caller-app` are admin-only, because they are what makes another person's usage visible.

### What could have called

`usage apps` lists the applications installed on this Olares. It is what makes an empty `caller_app` bucket readable: an application that spent nothing is stopped, or idle, or was uninstalled and left its spend behind, and the summary cannot tell those apart. `APPLICATION` is the name `--caller-app` and `quota set --caller-app` take. `STATE` is a snapshot from the last directory sync rather than a live reading — an application can stop between two of them, so `olares-cli market list --mine` is the current answer. Unlike the rest of `usage`, any console user can read it: an install is not a secret and what it spent is.

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
