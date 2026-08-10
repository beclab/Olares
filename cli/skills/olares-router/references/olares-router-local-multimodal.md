# Local multimodal applications

Embedding, CLIP, audio and OCR models install and are managed with exactly the verbs in [local LLM applications](olares-router-local-llm.md): `app catalog`, `app install`, `app tasks`, `app watch`, `provider register`, and the `local` family. What differs is the mode their model rows declare, the engine behind them, and how a call reaches them.

## What runs behind each mode

A model application's Model Console launches one engine, chosen by the kind of model it serves. That choice decides which endpoints exist inside the application — and therefore which Router endpoint can reach it.

| Kind | Router mode | Engine inside the application | Called with |
|---|---|---|---|
| Text generation | `chat`, `completion` | llama.cpp, vLLM, SGLang, Ollama | `router call chat` |
| Text embeddings | `embedding` | the embedding server | `router call embed` |
| Image + text embeddings (CLIP) | `embedding` | the embedding server, two towers | `router call embed`; image input goes through the same endpoint |
| Speech to text | `audio` | an audio engine | `router call transcribe` |
| Text to speech | `audio` | an audio engine | `router call speak` |
| Document OCR | `ocr` | an OCR adapter in front of llama.cpp | `router call ocr` |

`audio` is one mode covering both directions. Which direction a row actually supports is in its capability flags, not its mode: `supports_stt` and `supports_stt_stream` for transcription, `supports_tts`, `supports_tts_clone` and `supports_tts_dialogue` for speech, and `supports_vad`, `supports_diar`, `supports_enhance`, `supports_speaker_embed` for the surrounding steps. `router capabilities` lists every flag; `router provider get <provider>` shows which ones a row declares.

This is the one place where a mislabelled row is expensive: a `tts` model set as the `audio` default makes every transcription request fail with the upstream refusing the endpoint. Set defaults per direction only when the row's flags match, and name the model explicitly otherwise.

## Installing

```
olares-cli router app catalog
olares-cli router app install embeddinggemmav3 --watch
olares-cli router local progress embeddinggemmav3
```

Non-text models are usually small — hundreds of megabytes rather than tens of gigabytes — so the install finishes in a fraction of the time an LLM takes, and `--watch` is normally enough on its own.

An embedding or OCR application that is running but answers nothing is nearly always still verifying or converting weights: `router local progress` names the phase, and `router local status` reports the last verification.

## Confirming what arrived

A local application publishes its own model list, so Router mirrors it rather than needing `provider models import`:

```
olares-cli router provider get embeddinggemmav3
olares-cli router list --mode embedding
```

Check three things on the row before relying on it:

1. **The mode** is the one you expect. An OCR application whose row says `chat` was registered before its card declared the mode; `provider sync-models <provider>` re-mirrors it.
2. **The capability flags** cover the direction you need, per the table above.
3. **The dimension**, for embeddings, matches whatever already holds vectors. Changing the embedding model changes the vector space: existing vectors do not become wrong, they become incomparable. `router call embed --model <provider>/<model>` prints the dimension it got.

## Calling them

```
olares-cli router call embed "some text" --model embeddinggemmav3/embeddinggemma-300m
olares-cli router call transcribe meeting.m4a --language en
olares-cli router call speak "hello" --voice alloy --out hello.mp3
olares-cli router call ocr invoice.pdf --pages 1-3
```

Details, including how each call resolves a model when `--model` is omitted, are in [calling a model](olares-router-calling.md). Two properties are specific to these modes:

- **Transcription and OCR upload a file**, so they fail on a path before any model is reached — that error is the CLI's, not Router's.
- **OCR is asynchronous.** Router accepts a task and the CLI polls it; `--no-wait` returns the task id instead, which is what to use for a long PDF.

## Changing what one serves

The model card inside the application decides the weights, the engine flags, and the capabilities it advertises to Router:

```
olares-cli router local spec show embeddinggemmav3 -o json > card.json
# edit
olares-cli router local spec set embeddinggemmav3 --from card.json
olares-cli router local restart embeddinggemmav3
olares-cli router provider sync-models embeddinggemmav3
```

The last step matters more here than for an LLM: a change of dimension, or of which direction an audio model serves, is invisible to Router until its rows are re-mirrored. See [the Model Console](olares-router-console.md) for what the card contains and which fields are safe to change.
