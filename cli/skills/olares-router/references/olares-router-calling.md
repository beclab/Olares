# Calling a model

`router call` sends work through Router's data plane, the same path an application uses. It is the fastest way to prove a configuration end to end, and the only verb family here that needs a credential of its own.

`router models` is its companion: it lists the models this credential may put in the `model` field, from the data plane's own point of view, which is a narrower list than the `router list` a management-plane read produces. Beside each name it prints the mode, the capabilities the model card claims, and a `readiness` of `ready` or `unknown` — both of which mean "send it". `unknown` is an honest "nothing here can tell": it is what an application that runs its own engine, and so reports no phase for Router to read, looks like. A remote vendor has no weights to wait for and reads `ready`.

Route names are not in that list. An alias, a group or a `default-*` category is callable and describes no single model, so it has nothing to fill those columns with; `router route list` is where those names live.

Three things narrow the list against `router list`. The key's own allowlist is one. The other two are gates a locally installed model application passes separately: its container has to be up, and its weights have to be loaded. It owns a `router list` row from the moment it is installed whatever state it is in, so `router list` carries models of applications that are stopped, downloading or failed and the data plane admits none of them. A name in `router list` and not in `router models` is usually that, and the `STATE` cell says which phase it is stuck in.

`--include-not-ready` widens the read to the container gate alone, which is what to use while an install is running: a model still fetching or loading its weights appears as `warming` and turns `ready` under it, and one that could not load them appears as `failed` rather than being indistinguishable from a model nobody ever configured. It does not bring back an application that is not running — a stopped app has nothing to ask — so a name still absent under the flag is `olares-cli market` territory rather than a readiness problem.

## The credential

The management plane travels on the profile; the data plane does not accept it. `router call` resolves a data-plane credential in this order, without asking:

1. `--api-key sk-...`, when one is supplied.
2. `OLARES_ROUTER_API_KEY`, which is the way to supply one in a script without putting it in a process listing.
3. The key this machine already saved in the keychain.
4. No key at all, when running inside the cluster, where the platform supplies the calling application's identity.
5. A newly minted key, named after this host, saved to the keychain for next time.

`router key local` reports whether a key was saved and shows its prefix only — the plaintext stays in the keychain. `router key local --forget` drops the local copy, which stops *this machine* calling; the key itself keeps working until `router key revoke` ends it, and the next call mints a replacement.

The saved key is an ordinary one: it appears in `router key list`, it can be given a quota, and its calls are attributed to it in `router usage`.

## Choosing the model

`--model` takes a qualified `<provider>/<model>` as `router list` prints it, or any route name — an alias, a group, or a `default-*` category.

Leaving `--model` off names the default category for that kind of work: `default-chat` for chat, `default-stt` for transcription, `default-tts` for speech, and so on for every verb. Router decides what a category answers with by reconciling it against what is installed; nothing is set by hand and nothing falls back per call. `router route list --kind default` prints where each category currently stands, and a category nothing serves is refused rather than approximated.

That refusal is the usual meaning of a failure on a fresh install: `chat`, `embedding` and whatever the configured vendor happens to publish have categories behind them, and the rest do not until a model of that kind exists.

`router call translate` has no `--model` flag at all. The translate routes resolve their own default per call, so there is nothing for a caller to name.

## The verbs

```
olares-cli router call chat "summarise this" --system "be terse"
cat notes.md | olares-cli router call chat --no-stream --quiet
olares-cli router call chat "what is in this picture" --image shot.png
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
| A mode mismatch or unsupported-endpoint refusal | The model's mode or capabilities do not match the call — `router list` prints the mode, and for a local model `router model spec show <model>` prints what it declares |
| A bare 404 on an audio route | Router mounted the route and the engine behind the model does not serve it |
| `model_not_ready` with a 503 | The model is real and its weights cannot answer yet; the fix is to wait, and `router models --include-not-ready` shows whether it is `warming` or `failed` |
| A 5xx with an empty body | Nothing answered behind Router: the model application is stopped or still loading |

The last one is the common case for a local model, and it is a diagnosis rather than a configuration fix: continue in [deciding which layer is wrong](olares-router-diagnosis.md).

Every accepted call becomes a usage row, including one that failed upstream, so `router usage list --limit 5` immediately after a failure shows what Router recorded — status, model, tokens and cost. A call that produced no row was refused before Router routed it.
