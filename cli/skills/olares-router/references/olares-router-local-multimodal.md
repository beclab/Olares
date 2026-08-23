# Local multimodal applications

Embedding, CLIP, audio and OCR models install and are managed exactly as in [local LLM applications](olares-router-local-llm.md): [`olares-market`](../../olares-market/SKILL.md)'s `install` / `upgrade` / `uninstall`, then `provider register` and the `model status` / `progress` / `retry` / `restart` / `diag` / `spec` verbs here. What differs is the mode their model rows declare, the engine behind them, and how a call reaches them.

## What runs behind each mode

A model application's Model Console launches one engine, chosen by the kind of model it serves. That choice decides which endpoints exist inside the application — and therefore which Router endpoint can reach it.

| Kind | Router mode | Engine inside the application | Called with |
|---|---|---|---|
| Text generation | `chat` | llama.cpp, vLLM, SGLang, Ollama | `router call chat` |
| Text embeddings | `embedding` | the embedding server | `router call embed` |
| Image + text embeddings (CLIP) | `embedding` | the embedding server, two towers | `router call embed`; image input goes through the same endpoint |
| Reordering candidates | `rerank` | the embedding server | `router call rerank` |
| Speech to text | `audio` | an audio engine | `router call transcribe`, `listen` for a live stream, `align` against a transcript |
| Text to speech | `audio` | an audio engine | `router call speak`, `speak --voices`, `clone` from a recording, `dialogue` for several speakers |
| Analysing a recording | `audio` | an audio engine | `router call vad`, `diarize`, `diarize --stream`, `enhance` |
| Translation | `translate` | a translation engine | `router call translate` |
| Document OCR | `ocr` | an OCR adapter in front of llama.cpp | `router call ocr` |

The rest of the vocabulary — `responses`, `image_generation`, `video_generation`, `search`, `scrape`, `moderation` — is served by cloud providers rather than by anything installable here today. `router model list --mode <mode>` is what says which of them this Olares actually has.

Two of those cannot be called from this CLI, for different reasons. `moderation` has no data plane endpoint in Router at all: a row can declare the mode and a default category exists for it, but there is no `/v1/moderations` to send anything to, from here or from any other client. `responses` does have an endpoint — `router call responses` — but it is the one mode Router resolves no default for, so that verb requires `--model` and there is no `default-responses` to fall back to.

`audio` is one mode covering different jobs, and which one a row serves is in its capability flags rather than its mode. The 13 entries below are the current `terminus-apps` staging directories, not a claim that all 13 are formally released. An application declares only its listed capability or pair, so a model that transcribes genuinely cannot speak, and one that aligns cannot transcribe:

| Application | Flag | Verb |
|---|---|---|
| `audioqwen3asrv3`, `audiofwwhispersttv3`, `audiofwsystransttv3` | `supports_stt` | `router call transcribe` |
| `audioqwen3asrv3` | `supports_stt_stream` | `router call listen` |
| `audioqwenalignerv3` | `supports_align` | `router call align` |
| `audioqwen3ttsv3` | `supports_tts` | `router call speak` |
| `audioqwen3ttsclonev3` | `supports_tts_clone` | `router call clone` |
| `audiosoulxdialogv3` | `supports_tts_dialogue` | `router call dialogue` |
| `audiosilerovadv3` | `supports_vad` | `router call vad` |
| `audiopyannotediarv3` | `supports_diar` | `router call diarize` |
| `audiosortformerdiarsv3` | `supports_diar_stream` | `router call diarize --stream` |
| `audiopyannoteembedv3` | `supports_speaker_embed` | `router call speaker-embed` |
| `audiospeechbrainenhv3` | `supports_enhance` | `router call enhance` |
| `audiodashengsfxv3` | `supports_sound_fx` | `router call speak --sound-fx` |

`supports_audio_llm` and `supports_audio_s2s` are the two flags with no engine behind them yet; a row can declare either and nothing will answer. `router model list` names the flags in its SUPPORTS column, `router provider get <provider>` shows which ones each row of a provider declares, `router model get <model>` prints a row's flags in full, and `router model spec show <model>` shows what the application itself says.

Router keeps one default category per capability — `default-stt`, `default-stt-stream`, `default-align`, `default-tts`, `default-tts-clone`, `default-tts-dialogue`, `default-vad`, `default-diar`, `default-diar-stream`, `default-speaker-embed`, `default-enhance`, `default-sound-fx` — because one category per mode would have to pick a single engine for twelve jobs it cannot all do. **A bare 404 from an audio verb is most often a category pointing at another engine**, not a route Router failed to mount: the request reached a running model that has no such endpoint. `router route get default-align` says which model a category resolved, and the fix is the card that mislabelled the row rather than the category.

`tts_dialogue` and `sound_fx` share `/v1/audio/speech` with plain synthesis, so the verb supplies the distinction: `router call dialogue` resolves `default-tts-dialogue`, while `router call speak --sound-fx` resolves `default-sound-fx`. `--model` is optional for both.

`router model diag endpoints` reads the Model Console's `/api/endpoints`. For audio, Model Console internally fetches the engine's `/api/engine-spec` and merges those rows; `/api/engine-spec` is not an application route. Do not infer routes from the staging table above. Audio rows include `ASYNC` when the engine declares it:

```
olares-cli router model diag endpoints --app audioqwen3asrv3
```

## Installing

```
olares-cli market list -c AI
olares-cli market install embeddinggemmav3 --watch
olares-cli router model progress --app embeddinggemmav3
```

Non-text models are usually small — hundreds of megabytes rather than tens of gigabytes — so the install finishes in a fraction of the time an LLM takes, and `--watch` is normally enough on its own.

An embedding or OCR application that is running but answers nothing is nearly always still verifying or converting weights: `router model progress` names the phase, and `router model status` reports the last verification.

## Confirming what arrived

A local application publishes its own model list, so Router mirrors it rather than needing `model import`:

```
olares-cli router provider get embeddinggemmav3
olares-cli router model list --mode embedding
```

Check three things on the row before relying on it:

1. **The mode** is the one you expect. An OCR application whose row says `chat` was registered before its card declared the mode; `provider sync-models <provider>` re-mirrors it.
2. **The capability flags** cover the direction you need, per the table above.
3. **The dimension**, for embeddings, matches whatever already holds vectors. Changing the embedding model changes the vector space: existing vectors do not become wrong, they become incomparable. `router call embed --model <provider>/<model>` prints the dimension it got.

## Calling them

```
olares-cli router call embed "some text" --model embeddinggemmav3/embeddinggemma-300m
olares-cli router call rerank "who wrote it" --document "…" --document "…"
olares-cli router call transcribe meeting.m4a --language en
olares-cli router call speak "hello" --voice alloy --out hello.mp3
olares-cli router call diarize meeting.m4a
olares-cli router call clone me.wav "your build finished" --out done.wav
olares-cli router call transcribe keynote.m4a --async
olares-cli router call ocr invoice.pdf --pages 1-3
```

Details, including how each call resolves a model when `--model` is omitted, are in [calling a model](olares-router-calling.md). Four properties are specific to these modes:

- **File-taking audio verbs and OCR read local paths before reaching a model.** A missing path is the CLI's error, not Router's. `dialogue` also reads local `ref_audio` paths, while `listen` and `diarize --stream` read a local path or standard input into a WebSocket; none of those three is a multipart upload. Plain JSON synthesis reads no local path.
- **OCR is asynchronous.** Router accepts a task and the CLI polls it; `--no-wait` returns the task id instead, which is what to use for a long PDF, and `--queue` lists what is outstanding.
- **Audio uploads need a duration and byte-budget decision before the call.** Follow the executable [long audio decision tree](olares-router-calling.md#long-audio-decision-tree); unknown or long input, offline diarization and enhancement default to `--async`.
- **The two streaming verbs send PCM, not a container.** `router call listen` and `diarize --stream` read 16-bit mono PCM at 16 kHz from a file or standard input; they open a WebSocket rather than uploading, so a `.wav` header would arrive as audio and be heard as a click.

## Changing what one serves

The model card inside the application decides the weights, the engine flags, and the capabilities it advertises to Router:

```
olares-cli router model spec show Olares/embeddinggemma-300m
olares-cli router model spec edit Olares/embeddinggemma-300m --mode embedding
```

`model spec edit` merges the change onto the card the application is serving and stores what the application confirms, so Router's own copy is corrected in the same step. That matters more here than for an LLM: a change of dimension, or of which job an audio model does, is what Router routes on.

These applications' engines are sidecars, so they take no engine flags. The CLI does not know that and does not check: `model spec edit --engine-args` sends the value and the application decides what to do with it, which for a sidecar is nothing useful. `--mode` is the field that matters. See [local LLM applications](olares-router-local-llm.md) for the card in full and [the Model Console](olares-router-console.md) for reaching it at the application instead.
