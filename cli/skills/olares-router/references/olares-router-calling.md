# Calling a model

`router call` sends work through Router's data plane, the same path an application uses. It is the fastest way to prove a configuration end to end, and since Router v2.2.1 it needs no credential beyond the profile every other verb here already runs on.

`router call models` is its companion: it lists the models this credential may put in the `model` field, from the data plane's own point of view, which is a narrower list than the `router model list` a management-plane read produces. Beside each name it prints the mode, the capabilities the model card claims, and a `readiness` of `ready` or `unknown` — both of which mean "send it". `unknown` is an honest "nothing here can tell": it is what an application that runs its own engine, and so reports no phase for Router to read, looks like. A remote vendor has no weights to wait for and reads `ready`.

Route names are not in that list: an alias, a group or a `default-*` category is callable and describes no single model, so it has nothing to fill those columns with, and `router route list` is where those names live.

Two gates narrow the list against `router model list`, and a third applies only when a key is presented. The two are passed separately by a locally installed model application — its container has to be up, and its weights have to be loaded — and `router model list` reports both, in the `CALLABLE` cell and in the `readiness` field of its JSON. It owns a row from the moment it is installed whatever state it is in, so that list carries models of applications that are stopped, downloading or failed and the data plane admits none of them. The third is a key's own allowlist, which is the one `router model list` cannot see: that list is read over the console session, which has no allowlist. A keyless call has none either. So a name in `router model list` and not here is the weights, the application, or an allowlist — and the `CALLABLE` cell separates them: anything but `yes` is the weights or the application, while `yes` with the name still missing means the key being presented does not reach it. Drop `--api-key` and it should appear.

`--include-not-ready` widens the read to the container gate alone, which is what to use while an install is running: a model still fetching or loading its weights appears as `warming` and turns `ready` under it, and one that could not load them appears as `failed` rather than being indistinguishable from a model nobody ever configured. It does not bring back an application that is not running — a stopped app has nothing to ask — so a name still absent under the flag is `olares-cli market` territory rather than a readiness problem.

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
olares-cli router call music "a slow waltz" --model FlowStudio/<workflow> --out waltz.mp3
olares-cli router call 3d --model FlowStudio/<workflow> --image lantern.png --out lantern.glb
olares-cli router call transcribe meeting.m4a --language en
olares-cli router call speak "hello" --out hello.mp3
olares-cli router call vad meeting.m4a
olares-cli router call diarize meeting.m4a
olares-cli router call enhance noisy.wav --out clean.wav
olares-cli router call align meeting.m4a --text "what was said"
olares-cli router call clone me.wav "your build finished" --out done.wav
olares-cli router call dialogue scene.json --out scene.wav
olares-cli router call listen mic.pcm
olares-cli router call task get <task-id>
olares-cli router call ocr invoice.pdf --pages 1-3
```

**Text.** `chat` streams by default and prints a model and token line after the answer; `--quiet` prints only the answer, `--no-stream` waits for the whole thing. A prompt comes from the arguments or from standard input. A model running on this Olares serves a fixed number of requests at once and queues the rest, so `chat` allows ten minutes by default, says on stderr that it is waiting, and takes `--timeout` for a different budget; giving up stops the waiting and not the work, and the engine finishes an abandoned completion anyway. `--image` attaches a local file, which requires a model whose row declares `supports_vision`. `embed` prints a summary of each vector in table form and the whole vector in JSON, and `--per-line` turns piped text into one input per line rather than a single input. `rerank` takes a query and a repeatable `--document`, or the documents one per line on standard input, and prints them in the order the model put them.

`responses` sends one request to the Responses endpoint and prints the answer the way `chat` does. It exists so that a model configured with `--mode responses` can be checked at all: that mode is served on a different endpoint, so calling such a model with `chat` fails in a way that says nothing about the model. It is deliberately only that — one request, no streaming, no conversation carried across calls, nothing stored.

**The web.** `search` and `scrape` reach a provider that has one of those two modes; most do not, so both are commonly refused for want of a category rather than for anything wrong with the request.

**Translation.** `translate` translates, `--detect` identifies a language instead, and `--languages` lists the pairs the configured model serves. `--to` is required for a translation and `--from` is optional, since detection is the default.

**Creative work.** `image`, `video`, `music` and `3d` are the four verbs whose work outlives the request. All four submit, wait, and write the result to `--out`; `--no-wait` prints the generation id instead, `--id <id>` collects that generation later, and `--output-id` picks one when a generation produced several. The wait defaults follow how long the work takes — five minutes for an image, twenty for a video, ten for a track, fifteen for a mesh — and a `--timeout` only stops the waiting: the provider carries on and the id stays collectable. Router holds the bytes itself, so a result does not depend on a vendor link staying alive.

One contract, four shapes. Each verb offers only the fields its own family can express — an image has no `--fps`, a mesh has no `--lyrics` — so a field that could only be refused is not a flag there at all; all four share `--negative`, `--seed` and `--provider-option k=v`, the last carrying a vendor knob this contract has no field for. `--size` and `--aspect-ratio` describe the same shape, so giving both is refused before the request. `3d` is the one family that needs no words at all: most 3D workflows work from a picture, so `--image lantern.png` is a complete request, and a local file becomes a data URL while a data URL or a link is sent as written. Underneath, image and video ride the OpenAI-shaped routes they shipped with and the other two ride Router's unified one, which is not a distinction to reason about — the fields mean the same thing on both. It shows in one place only: an image provider that keeps no generations to poll answers inline, and `image` handles that as well as the polled kind, which is why it was not moved onto the unified route.

**Audio.** The current audio-engine applications each serve one declared capability set; legacy audio applications are outside this contract. A model that transcribes does not necessarily speak, and one that aligns cannot transcribe. Every verb therefore resolves its own default category. A bare 404 usually means that category reached an engine which does not mount the requested capability. `speaker-embed` is the speaker-vector command; do not substitute `call embed`. `speak --sound-fx` resolves `default-sound-fx`; `--model` is optional. `speak`, `clone`, `dialogue` and `enhance` refuse to write audio to a terminal, so pass `--out` or redirect. `speak --voices` lists preset voices. `align` takes a transcript from `--text` or standard input. `dialogue` reads a JSON script and turns local `ref_audio` paths into data URLs.

### Long audio decision tree

Use this sequence before uploading a recording:

1. If `ffprobe` is already installed, read duration, codec, channel count and sample rate. File size is not duration; never infer one from the other. If `ffprobe` is absent, do not install it or guess: tell the user duration is unknown and offer to submit the original asynchronously, provide a known duration, or use their preferred inspection tool.
2. Use a synchronous call only when the input is known to be short and the operation is expected to finish in about 30 seconds. Use `--async` by default for unknown duration, long audio, offline diarization and enhancement. Async removes inference waiting from the submission connection; it does not bypass upload size or upload time.
3. Check bytes separately. The CLI warns above 90 MiB. Router admits at most 96 MiB for the complete request body, including multipart fields and boundaries; the CLI therefore reserves 64 KiB and refuses an audio file above `96 MiB - 64 KiB`. The frontend admits `100m` only so Router can own the exact 96 MiB error.
4. If the upload is too large and `ffmpeg` is already installed, consider 16 kHz mono FLAC for WAV, high-sample-rate, multichannel or otherwise uncompressed/high-bitrate input, then re-check bytes. AAC, Opus and other already-compressed sources may become larger when transcoded to FLAC, so inspect and re-check rather than converting blindly. If `ffmpeg` is absent, do not install it silently: offer the original when it fits, ask the user for a converted file, or ask whether they want to use another tool.
5. Split offline STT only when compression still cannot meet the request byte budget, or when the business already has useful VAD or chapter boundaries. Pieces at or below 480 seconds are a product recommendation, not an engine limit. Qwen accepts longer requests and internally splits them at low-energy boundaries; this wrapper requests `return_time_stamps=false`. Prefer speech/silence boundaries; fixed pieces with a short overlap are the fallback. Submit each piece independently and merge its text while removing duplicated overlap. Offset timestamps only when the selected engine and response format actually return time fields; plain Qwen text has none.
6. Do not automatically split offline or streaming diarization, because speaker identity is not stable across independent pieces; do not split speaker embedding, because one whole clip produces one vector; and do not split voice cloning or dialogue. Split alignment only when a transcript is already divided into matching segments. Enhancement already windows and overlap-adds internally, so prefer one whole-file async task.

The inspection and whole-file conversion commands below are optional external tools, not prerequisites. Never add an installation command:

```
ffprobe -v error -show_entries format=duration:stream=codec_name,channels,sample_rate -of json meeting.m4a
ffmpeg -i meeting.m4a -vn -ar 16000 -ac 1 -c:a flac meeting-16k-mono.flac
```

Submit asynchronously and preserve both the task id and the exact model reference:

```
olares-cli router call transcribe meeting-16k-mono.flac --model default-stt --async
olares-cli router call task get <task-id> --model default-stt --wait -o json > transcript.json
```

Use `task result` to collect a binary audio result:

```
olares-cli router call speak "read this" --model default-tts --async
olares-cli router call task result <task-id> --model default-tts --out speech.wav
```

Each piece is a separate task and a separate billable call. One engine has one worker and a queue of 32 waiting tasks; a full queue refuses new work. Results expire after 1800 seconds, and tasks live only in memory, so a pod restart loses them. On a timeout, keep the id and model and use `task get` or `task list` before considering a resubmission — blindly submitting again can run and bill the work twice. Polls consume Router RPM quota; use `task get --wait` rather than a tight manual loop. A task `get` or status read may write a spend row, but it carries no duration and incurs no duration-based charge. Only `task result` can carry measured duration and be charged by duration; fetching the same result repeatedly may therefore charge it repeatedly. JSON task results are already present in `task get`; fetch `task result` only when the result must be collected separately, especially binary audio. When split STT input hits a queue-full 503, submit slices sequentially. Back off, then use `task get` or `task list` to check the accepted work before submitting the next slice; do not fan out requests or blindly resubmit the rejected slice. Consecutive 503 responses may temporarily trigger Router circuit open, so wait for the circuit and queue to recover instead of increasing concurrency.

`router call task get`, `router call task result` and `router call task cancel` each follow one task by id. Router normally remembers its backend; after a Router restart or when using another gateway, pass the model saved with the id. `router call task list` always needs `--model`, because every audio application owns a separate queue. The `model` printed in a task receipt, including JSON output, is a routing reference, not the engine's canonical model id. Preserve it with the task id so recovery after Router forgets the backend still reaches the engine that owns the task.

`listen` and `diarize --stream` are the two verbs that open a WebSocket rather than uploading: they read 16-bit mono PCM at 16 kHz from a file or standard input, print partial results as they arrive, and need models declaring `supports_stt_stream` and `supports_diar_stream` — which, again, are separate applications from their batch counterparts.

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
| A bare 404 on an audio route | Router mounted the route and the engine behind the model does not serve it; `router route get default-<capability>` says which model the verb resolved |
| `model_not_ready` with a 503 | The model is real and its weights cannot answer yet; the fix is to wait, and `router call models --include-not-ready` shows whether it is `warming` or `failed` |
| `model_at_capacity` with a 503 | The engine is serving every request it was launched for and Router already waited for a slot — ten seconds for an interactive mode, sixty for a generation — before giving up. So this is a queue that stayed full, not a refusal to queue. Wait for the `Retry-After` the message names and retry the one request; more concurrency makes it worse. `router provider get <app>` shows what the engine is holding and `router model list` shows how wide it was launched |
| `media_field_*` and `media_input_*` | Router refused a creative request before any provider saw it: a field the model has no parameter for, two flags describing the same thing, or an input this operation does not take. The message names the field, and `router model get <model>` lists what the row declares |
| A 5xx with an empty body | Nothing answered behind Router: the model application is stopped or still loading |

The last one is the common case for a local model, and it is a diagnosis rather than a configuration fix: continue in [deciding which layer is wrong](olares-router-diagnosis.md).

Every accepted call becomes a usage row, including one that failed upstream, so `router usage list --limit 5` immediately after a failure shows what Router recorded — status, model, tokens and cost. A call that produced no row was refused before Router routed it.
