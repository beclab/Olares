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
| Speech to text | `audio` | an audio engine | `router call transcribe`, and `align` against a transcript |
| Text to speech | `audio` | an audio engine | `router call speak`, `speak --voices` |
| Analysing a recording | `audio` | an audio engine | `router call vad`, `diarize`, `enhance` |
| Translation | `translate` | a translation engine | `router call translate` |
| Document OCR | `ocr` | an OCR adapter in front of llama.cpp | `router call ocr` |

The rest of the vocabulary — `responses`, `image_generation`, `video_generation`, `search`, `scrape`, `moderation` — is served by cloud providers rather than by anything installable here today. `router model list --mode <mode>` is what says which of them this Olares actually has.

Two of those cannot be called from this CLI, for different reasons. `moderation` has no data plane endpoint in Router at all: a row can declare the mode and a default category exists for it, but there is no `/v1/moderations` to send anything to, from here or from any other client. `responses` does have an endpoint — `router call responses` — but it is the one mode Router resolves no default for, so that verb requires `--model` and there is no `default-responses` to fall back to.

`audio` is one mode covering six different jobs, and which of them a row actually serves is in its capability flags rather than its mode: `supports_stt` and `supports_stt_stream` for transcription, `supports_tts`, `supports_tts_clone` and `supports_tts_dialogue` for speech, and `supports_vad`, `supports_diar`, `supports_enhance`, `supports_speaker_embed` for the surrounding steps. Each of those is a separate engine image, so a model that transcribes genuinely cannot speak. `router model list` names them in its SUPPORTS column, `router provider get <provider>` shows which ones each row of a provider declares, `router model get <model>` prints a row's flags in full, and `router model spec show <model>` shows what the application itself says.

This is the one place where a mislabelled row is expensive: a speech model that ends up answering for transcription makes every transcription request fail with a bare 404 from the engine. Router points each audio category — `default-stt`, `default-tts`, `default-vad`, `default-diar`, `default-enhance`, `default-sound-fx` — at a model whose flags match, so the fix for a mismatch is the card rather than the category.

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
olares-cli router call ocr invoice.pdf --pages 1-3
```

Details, including how each call resolves a model when `--model` is omitted, are in [calling a model](olares-router-calling.md). Three properties are specific to these modes:

- **Every audio verb and OCR upload a file**, so they fail on a path before any model is reached — that error is the CLI's, not Router's.
- **OCR is asynchronous.** Router accepts a task and the CLI polls it; `--no-wait` returns the task id instead, which is what to use for a long PDF, and `--queue` lists what is outstanding.
- **A bare 404 from an audio route** means Router mounted it and the engine behind the model does not serve it. That is a mismatch between the verb and the model, not a missing route.

## Changing what one serves

The model card inside the application decides the weights, the engine flags, and the capabilities it advertises to Router:

```
olares-cli router model spec show Olares/embeddinggemma-300m
olares-cli router model spec edit Olares/embeddinggemma-300m --mode embedding
```

`model spec edit` merges the change onto the card the application is serving and stores what the application confirms, so Router's own copy is corrected in the same step. That matters more here than for an LLM: a change of dimension, or of which job an audio model does, is what Router routes on.

These applications' engines are sidecars, so they take no engine flags. The CLI does not know that and does not check: `model spec edit --engine-args` sends the value and the application decides what to do with it, which for a sidecar is nothing useful. `--mode` is the field that matters. See [local LLM applications](olares-router-local-llm.md) for the card in full and [the Model Console](olares-router-console.md) for reaching it at the application instead.
