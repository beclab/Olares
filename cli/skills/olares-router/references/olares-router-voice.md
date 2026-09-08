# The voice library and the reading history

`call speak` and `call clone` produce audio and keep nothing. A reference recording is used for one reading and is not a voice afterwards, and the bytes exist only where `--out` put them. This surface is the durable half: a voice that persists under a name, and a log of every reading with its audio still attached.

Read this when a user wants a voice they can reuse, wants a voice invented rather than recorded, or wants back a file a synthesis call produced and they no longer have. For the synthesis verbs themselves, and for the async task machinery all audio shares, read [calling a model](olares-router-calling.md).

```
olares-cli router call voice list
olares-cli router call voice get <voice>
olares-cli router call voice add "Night Host" --sample me.wav --ref-text "the words in the clip"
olares-cli router call voice design "a tired night-shift radio host" --out preview.wav
olares-cli router call voice design "a bright children's narrator" --save "Narrator"
olares-cli router call voice settings
olares-cli router call voice delete <voice> --yes
olares-cli router call speak "hello" --voice <voice> --out hello.mp3

olares-cli router call history list --model <tts-model>
olares-cli router call history get <reading> --model <tts-model>
olares-cli router call history download <reading> --model <tts-model> --out again.wav
olares-cli router call history delete <reading> --model <tts-model> --yes
```

## Two ways to make a voice, and they are different capabilities

`voice add` needs a recording of somebody. `voice design` needs only a description. They are not two spellings of one thing: a model that clones may not design, and the categories they resolve are different for exactly that reason.

| Verb | Falls back to | The model must declare |
|---|---|---|
| `voice list/get/settings/delete` | `default-tts` | nothing beyond speaking |
| `voice add` | `default-tts-clone` | `supports_tts_clone` |
| `voice design` | `default-tts-design` | `supports_tts_design` |

Reading the table is `default-tts` rather than a category of its own because a clone and a design land in that same table, and an engine with no listable voices of its own still has whatever was added to it.

`voice list` printing nothing is an answer, not a fault. A model built to speak from a reference recording has no library, and `call clone` is how it is used.

`CATEGORY` in the list says where a voice came from — `premade` ships with the weights, `cloned` was made from a recording, `generated` was designed. All three are spoken with the same way, by passing the id to `call speak --voice`.

## Designing is two steps, and the first one keeps nothing

`voice design "a description"` answers with previews: short samples of what that voice sounds like reading a line. Nothing is stored. `--out` writes the first preview to a file so it can be listened to, and `--text` sets the line it reads, which is worth setting when the voice is for a specific piece of writing.

`--save <name>` is the second step: it keeps the preview as a real voice with an id. **A preview is not addressable afterwards** — running the same description again produces a different voice, so a design worth keeping has to be saved in the same invocation that produced it. Tell the user this before they run the first one, not after they liked a preview they can no longer reach.

Every design and every clone is a billed synthesis. Designing four descriptions to compare them costs four calls.

## Cloning here versus `call clone`

`call clone me.wav "your build finished" --out done.wav` uploads a recording, speaks one line with it, and forgets the recording. `call voice add "Night Host" --sample me.wav` uploads the same recording and keeps it, so `call speak --voice <id>` reaches it again without another upload.

Prefer `call clone` for a one-off. Prefer `voice add` when the same voice will be used more than once, or when the recording lives somewhere the caller will not have next time. `--ref-text` is what is said in the sample; engines that align a clone against a transcript use it, and the rest ignore it harmlessly.

The upload limits are the ones every audio verb has: Router admits at most 96 MiB for the whole request body, and the CLI refuses a sample above `96 MiB - 64 KiB`. The [long-audio decision tree](olares-router-calling.md) applies to a sample as much as to a recording being transcribed.

## Settings are the engine's, and are printed as it states them

`voice settings` shows the defaults every voice falls back to; `voice settings <voice>` shows what one voice speaks with. Stability, similarity, speed — the knobs are the engine's own and differ between engines, so nothing here interprets them or claims which matter. Report what came back rather than translating it.

## History has no default category

Every `history` verb requires `--model`. This is not an oversight to route around: Router registers no category for the history routes, because a reading exists inside the engine that performed it and there is no engine "the history" could sensibly resolve to. `router model list --mode tts` names the candidates.

A reading is visible to the caller that made it. Another key's readings are not missing, they are not yours to read — so history under `--api-key` and history under the profile are different lists, and a user who cannot find a reading may simply have made it with the other credential.

**Fetching a reading again is free.** The engine kept the bytes when it spoke them, so `history download` copies a file: no model runs, nothing is billed, and no spend row appears. This is the answer to "I lost the audio" and to "I forgot `--out`", and it is worth offering before re-synthesising anything.

`state` is `processing` while the engine is still speaking a long piece of text. Such a reading can already be downloaded — the audio arrives as it is produced.

`history delete` on a reading still being spoken stops it, which is how to abandon a long piece of text partway through. The spend row stays: what was billed happened, and throwing the result away does not unbill it.

`--voice` narrows the list to one voice, which is how to find the take that came out right when several were tried; `--after <id>` continues past the first page.

## Where this shows up in spend

Synthesis is `tts` in the usage record, its own mode since ADR-63 rather than part of `audio`. So `router usage list --mode tts` is what shows a design, a clone, an `add`, and every reading; a download shows nothing at all. See [usage and audit](olares-router-usage.md) for the columns.
