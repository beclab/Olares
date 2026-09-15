# Calling a model — the contract every verb shares

`router call` sends work through Router's data plane, the same path an application uses. It is the fastest way to prove a configuration end to end, and since Router v2.2.1 it needs no credential beyond the profile every other verb here already runs on.

What each family of verbs can express is its own read: [text and the web](olares-router-call-text.md), [translation](olares-router-call-translate.md), [pictures, clips, tracks and meshes](olares-router-call-media.md), [audio](olares-router-call-audio.md). This file is what holds for all of them — who the call is from, what may go in `--model`, and how to read a refusal.

## What this credential may call

`router call models` lists the models this credential may put in the `model` field, from the data plane's own point of view, which is a narrower list than the `router model list` a management-plane read produces. Beside each name it prints the mode, the capabilities the model card claims, and a `readiness` of `ready` or `unknown` — both of which mean "send it". `unknown` is an honest "nothing here can tell": it is what an application that runs its own engine, and so reports no phase for Router to read, looks like. A remote vendor has no weights to wait for and reads `ready`.

Route names are not in that list: an alias, a group or a `default-*` category is callable and describes no single model, so it has nothing to fill those columns with, and `router route list` is where those names live.

Two gates narrow the list against `router model list`, and a third applies only when a key is presented. The two are passed separately by a locally installed model application — its container has to be up, and its weights have to be loaded — and `router model list` reports both, in the `CALLABLE` cell and in the `readiness` field of its JSON. It owns a row from the moment it is installed whatever state it is in, so that list carries models of applications that are stopped, downloading or failed and the data plane admits none of them. The third is a key's own allowlist, which is the one `router model list` cannot see: that list is read over the console session, which has no allowlist. A keyless call has none either. So a name in `router model list` and not here is the weights, the application, or an allowlist — and the `CALLABLE` cell separates them: anything but `yes` is the weights or the application, while `yes` with the name still missing means the key being presented does not reach it. Drop `--api-key` and it should appear.

Two flags change the read. `--include-not-ready` widens it to the container gate alone, which is what to use while an install is running: a model still fetching or loading its weights appears as `warming` and turns `ready` under it, and one that could not load them appears as `failed` rather than being indistinguishable from a model nobody ever configured. It does not bring back an application that is not running — a stopped app has nothing to ask — so a name still absent under the flag is `olares-cli market` territory rather than a readiness problem. `--operations` asks a different question again: the capability flags say what a model can do, while the operation catalogue says which routes it answers, with the method, the path and whether the work can be submitted asynchronously. For audio those come apart — knowing an engine synthesises speech does not say whether it answers `/v1/audio/speech` or `/v1/text-to-speech/<voice>`, and no engine serves both — so the catalogue is what `speak` and `voices` route on, and where one is declared and still current Router refuses an operation it does not list. [The voice library and the reading history](olares-router-voice.md) reads its labels.

## The credential

**Calling needs no key.** Router's `/v1` reads three identities in order — an `sk-*` Bearer, a calling application's `x-caller-appid`, then a person's `X-BFL-USER` — and the last two are stamped by the Olares edge, which is the same edge, host and profile session the management verbs already travel on. A call sent with no `Authorization` is therefore not anonymous: it arrives as the profile.

So there are two steps, not five: pass `--api-key sk-...`, or `OLARES_ROUTER_API_KEY` for a script that should keep the key out of a process listing; otherwise pass no credential at all and the platform says who is calling.

Reach for a key when the call needs something the identity cannot carry: a model allowlist, a budget of its own, or an origin the platform cannot vouch for — anything outside Olares, which is where the header is added.

Two refusals are specific to this and mean different things:

- `missing_credentials` — the Router being called predates v2.2.1 and does not read `X-BFL-USER` on `/v1`. Upgrade the Router application, or pass a key.
- `unknown_bfl_user` — the platform knows this person, Router has no row for them. Router records a person the first time they use the console plane, so any management verb (`router model list` will do) creates it. Nothing creates it from the data plane, by design.

`router key current` says which of the two a call would present right now. **A machine that used an older olares-cli still has a key saved in its keychain**: calls no longer use it, and it is still a live unrestricted key in Router — `router key list` shows it and `router key revoke` is what ends it. `--forget` drops only the local copy, and since Router keeps just a hash the plaintext is gone for good afterwards, so revoke before forgetting rather than after.

### The identity is also the anchor

Router anchors a stored response and a media generation on `(user, key)`, so a job started with a key is not visible to a later keyless call — the answer is a 404, not a permission error. Whenever a job is created in one command and collected in another, **`--no-wait` and the follow-up `--id` have to run under the same credential**: both keyless, or both with the same key. OCR is unaffected; Router stores no task of its own for it.

## Choosing the model

`--model` takes a qualified `<provider>/<model>` as `router model list` prints it, or any route name — an alias, a group, or a `default-*` category.

Leaving `--model` off names the default category for that kind of work: `default-chat` for chat, `default-stt` for transcription, `default-tts` for speech, and so on for every verb. Router decides what a category answers with by reconciling it against what is installed; nothing is set by hand and nothing falls back per call. `router route list --kind default` prints where each category currently stands, and a category nothing serves is refused rather than approximated.

That refusal is the usual meaning of a failure on a fresh install: `chat`, `embedding` and whatever the configured vendor happens to publish have categories behind them, and the rest do not until a model of that kind exists.

Three verbs sit outside that. `router call translate` has no `--model` flag at all, because the translate routes resolve their own default per call. `router call responses`, `router call music` and `router call 3d` are the opposite: `--model` is required, since Router deliberately keeps no `default-responses`, `default-music-generation` or `default-model3d-generation` — the first because that mode is a different endpoint rather than a different model, the other two because one implementation apiece is not a choice worth dressing as one. `router model list --mode responses` names what the first can send, and `--mode music_generation` or `--mode model3d_generation` the other two.

## OCR

Always asynchronous, and the one family whose task Router does not store. The verb submits and polls; `--no-wait` prints the task id, `--task <id>` picks it up later, `--cancel` with `--task` drops it, `--timeout` stops waiting without stopping the task, and `--queue` lists what is outstanding.

```
olares-cli router call ocr invoice.pdf --pages 1-3
```

`-o json` on any `router call` verb prints the upstream's own response, which is what to use when the shape matters more than the reading.

## Reading a failure

A `router call` failure comes from one of a few places, and the message says which:

| What it looks like | What it means |
|---|---|
| `invalid_api_key` with type `authentication_error` | Router is refusing *your* key — revoked, expired, or not allowed this model |
| An authentication error without that type | The **vendor** is refusing Router's stored credential; `router provider validate <provider>` confirms it |
| `quota_exceeded` | A ceiling on the key, the user, the model or the calling application; `router quota list` shows which |
| `no_default_model` | The category for that kind of work has nothing behind it; `router route list --kind default` says which do |
| `model_route_disabled` | The name exists but is switched off; `router route enable <name>` |
| `model_not_allowed` | The key's allowed list does not include this model; `router key update` changes it |
| A mode mismatch or unsupported-endpoint refusal | The model's mode or capabilities do not match the call — `router model list` prints the mode, and for a local model `router model spec show <model>` prints what it declares |
| `audio_operation_not_supported`, or a bare 404 on an audio route | The same gap, either side of the operation catalogue. Router refuses before the engine is reached only where a catalogue was declared and has been seen within the last fifteen minutes; where none was declared, or where the declared one has aged past that window, the request is forwarded and the engine's own 404 comes back. `router call models --operations` prints what each model declares and `router route get default-<capability>` says which one the verb resolved; another model of the same mode may serve it, and leaving `--model` off lets Router pick one that does |
| `audio_task_owner_unavailable` with a 503 | Router could not read who owns an audio task, so it refused the read rather than answering a 404 that would read as "your submission is gone". Wait for the `Retry-After` and read again — `task get <id>`. **Never resubmit on this one**: the work is still running, and a second submission is a second transcription of the same recording |
| `model_not_ready` with a 503 | The model is real and its weights cannot answer yet; the fix is to wait, and `router call models --include-not-ready` shows whether it is `warming` or `failed` |
| `model_at_capacity` with a 503 | The engine is serving every request it was launched for and Router already waited for a slot — ten seconds for an interactive mode, sixty for a generation — before giving up. So this is a queue that stayed full, not a refusal to queue. Wait for the `Retry-After` the message names and retry the one request; more concurrency makes it worse. `router provider get <app>` shows what the engine is holding and `router model list` shows how wide it was launched |
| `kv_budget_exhausted` with a 503 | A slot was free and the KV cache was not. It is the engine's own admission gate, forwarded verbatim from llm-init, and it separates a full engine from a broken one — a spend row that read `upstream_error` used to hide the difference. Router fails the call over to the other members of the route but never retries the same backend, because that only waits through the same gate again. A shorter prompt gets through where a retry does not; `router model get <model>` prints the pool against the window and the width |
| `media_field_*` and `media_input_*` | Router refused a creative request before any provider saw it: a field the model has no parameter for, two flags describing the same thing, or an input this operation does not take. The message names the field, and `router model get <model>` lists what the row declares |
| A 5xx with an empty body | Nothing answered behind Router: the model application is stopped or still loading |

The last one is the common case for a local model, and it is a diagnosis rather than a configuration fix: continue in [deciding which layer is wrong](olares-router-diagnosis.md). Every accepted call becomes a usage row, including one that failed upstream, so `router usage list --limit 5` immediately after a failure shows what Router recorded — status, model, tokens and cost. A call that produced no row was refused before Router routed it.
