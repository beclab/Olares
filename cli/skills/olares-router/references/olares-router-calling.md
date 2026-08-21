# Calling a model

`router call` sends work through Router's data plane, the same path an application uses. It is the fastest way to prove a configuration end to end, and since Router v2.2.1 it needs no credential beyond the profile every other verb here already runs on.

`router call models` is its companion: it lists the models this credential may put in the `model` field, from the data plane's own point of view, which is a narrower list than the `router model list` a management-plane read produces. Beside each name it prints the mode, the capabilities the model card claims, and a `readiness` of `ready` or `unknown` — both of which mean "send it". `unknown` is an honest "nothing here can tell": it is what an application that runs its own engine, and so reports no phase for Router to read, looks like. A remote vendor has no weights to wait for and reads `ready`.

Route names are not in that list. An alias, a group or a `default-*` category is callable and describes no single model, so it has nothing to fill those columns with; `router route list` is where those names live.

Two gates narrow the list against `router model list`, and a third applies only when a key is presented. The two are passed separately by a locally installed model application — its container has to be up, and its weights have to be loaded — and `router model list` reports both, in the `CALLABLE` cell and in the `readiness` field of its JSON. It owns a row from the moment it is installed whatever state it is in, so that list carries models of applications that are stopped, downloading or failed and the data plane admits none of them. The third is a key's own allowlist, which is the one `router model list` cannot see: that list is read over the console session, which has no allowlist. A keyless call has none either.

So a name in `router model list` and not here is the weights, the application, or an allowlist — and the `CALLABLE` cell separates them: anything but `yes` is the weights or the application, while `yes` with the name still missing means the key being presented does not reach it. Drop `--api-key` and it should appear.

`--include-not-ready` widens the read to the container gate alone, which is what to use while an install is running: a model still fetching or loading its weights appears as `warming` and turns `ready` under it, and one that could not load them appears as `failed` rather than being indistinguishable from a model nobody ever configured. It does not bring back an application that is not running — a stopped app has nothing to ask — so a name still absent under the flag is `olares-cli market` territory rather than a readiness problem.

## The credential

**Calling needs no key.** Router's `/v1` reads three identities in order — an `sk-*` Bearer, a calling application's `x-caller-appid`, then a person's `X-BFL-USER` — and the last two are stamped by the Olares edge, which is the same edge, host and profile session the management verbs already travel on. A call sent with no `Authorization` is therefore not anonymous: it arrives as the profile.

So there are two steps, not five:

1. `--api-key sk-...`, or `OLARES_ROUTER_API_KEY` for a script that should keep the key out of a process listing.
2. Otherwise no credential at all, and the platform says who is calling.

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

`router call translate` has no `--model` flag at all. The translate routes resolve their own default per call, so there is nothing for a caller to name.

`router call responses` is the opposite exception: `--model` is required. Router resolves a default for every mode except this one, deliberately, so there is no `default-responses` to leave the flag off for. `router model list --mode responses` is where the names are.

## The verbs

```
olares-cli router call chat "summarise this" --system "be terse"
cat notes.md | olares-cli router call chat --no-stream --quiet
olares-cli router call chat "what is in this picture" --image shot.png
olares-cli router call responses "summarise this" --model openai/gpt-4o
olares-cli router call embed "text" --dimensions 512
olares-cli router call rerank "who wrote it" --document "…" --document "…"
olares-cli router call search "olares release notes" --limit 5
olares-cli router call scrape https://example.com/post
olares-cli router call translate "hello" --to zh
olares-cli router call image "a red bicycle" --out bike.png
olares-cli router call video "a bicycle rolling downhill" --out clip.mp4
olares-cli router call transcribe meeting.m4a --language en
olares-cli router call speak "hello" --out hello.mp3
olares-cli router call vad meeting.m4a
olares-cli router call diarize meeting.m4a
olares-cli router call enhance noisy.wav --out clean.wav
olares-cli router call align meeting.m4a --text "what was said"
olares-cli router call ocr invoice.pdf --pages 1-3
```

**Text.** `chat` streams by default and prints a model and token line after the answer; `--quiet` prints only the answer, `--no-stream` waits for the whole thing. A prompt comes from the arguments or from standard input. `--image` attaches a local file, which requires a model whose row declares `supports_vision`. `embed` prints a summary of each vector in table form and the whole vector in JSON, and `--per-line` turns piped text into one input per line rather than a single input. `rerank` takes a query and a repeatable `--document`, or the documents one per line on standard input, and prints them in the order the model put them.

`responses` sends one request to the Responses endpoint and prints the answer the same way `chat` does. It exists so that a model configured with `--mode responses` can be checked at all: that mode is served on a different endpoint, so calling such a model with `chat` fails in a way that says nothing about the model. It is deliberately only that — one request, no streaming, no conversation carried across calls, nothing stored — so it answers "does this model work" and not much else.

**The web.** `search` and `scrape` reach a provider that has one of those two modes; most do not, so both are commonly refused for want of a category rather than for anything wrong with the request.

**Translation.** `translate` translates, `--detect` identifies a language instead, and `--languages` lists the pairs the configured model serves. `--to` is required for a translation and `--from` is optional, since detection is the default.

**Images and video.** These are the two verbs whose work can outlive the request. Both submit, wait, and write the result to `--out`; `--no-wait` prints the generation id instead, and `--id <id>` collects that generation later. Video defaults to waiting twenty minutes and images five, and a `--timeout` only stops the waiting — the provider carries on, and the id is still collectable. An image provider with no persistent generations API answers inline instead, and the verb handles both without the caller choosing.

**Audio.** Six verbs over one upstream, and which of them a model serves depends on the engine behind it rather than on the mode: recognition, synthesis, voice activity, diarization, enhancement and alignment are separate engine images. A model that transcribes does not necessarily speak, and a bare 404 from one of these routes usually means the model does the other thing. `speak` and `enhance` refuse to write audio to a terminal, before making the call, so pass `--out` or redirect. `speak --voices` lists what the chosen model can sound like. `align` takes the transcript from `--text` or standard input, and defaults to the transcription category rather than one of its own.

**OCR.** Always asynchronous. The verb submits and polls; `--no-wait` prints the task id, `--task <id>` picks it up later, `--cancel` with `--task` drops it, `--timeout` stops waiting without stopping the task, and `--queue` lists what is outstanding.

`-o json` on any of them prints the upstream's own response, which is what to use when the shape matters more than the reading.

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
| A bare 404 on an audio route | Router mounted the route and the engine behind the model does not serve it |
| `model_not_ready` with a 503 | The model is real and its weights cannot answer yet; the fix is to wait, and `router call models --include-not-ready` shows whether it is `warming` or `failed` |
| A 5xx with an empty body | Nothing answered behind Router: the model application is stopped or still loading |

The last one is the common case for a local model, and it is a diagnosis rather than a configuration fix: continue in [deciding which layer is wrong](olares-router-diagnosis.md).

Every accepted call becomes a usage row, including one that failed upstream, so `router usage list --limit 5` immediately after a failure shows what Router recorded — status, model, tokens and cost. A call that produced no row was refused before Router routed it.
