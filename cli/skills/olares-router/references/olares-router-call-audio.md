# Calling a model — audio

> **Prerequisite:** [the calling contract](olares-router-calling.md) — the credential, what `--model` takes, and how to read a refusal — holds for every verb here.

```
olares-cli router call transcribe meeting.m4a --language en
olares-cli router call speak "hello" --out hello.mp3
olares-cli router call vad meeting.m4a
olares-cli router call diarize meeting.m4a
olares-cli router call enhance noisy.wav --out clean.wav
olares-cli router call align meeting.m4a --text "what was said"
olares-cli router call clone me.wav "your build finished" --out done.mp3
olares-cli router call dialogue scene.json --out scene.mp3
olares-cli router call listen mic.pcm
olares-cli router call task get <task-id>
```

## One engine, one declared capability set

The current audio-engine applications each serve one declared capability set; legacy audio applications are outside this contract. A model that transcribes does not necessarily speak, and one that aligns cannot transcribe. Every verb therefore resolves its own default category. A bare 404 usually means that category reached an engine which does not mount the requested capability. `speaker-embed` is the speaker-vector command; do not substitute `call embed`. `speak --sound-fx` resolves `default-sound-fx`; `--model` is optional.

`speak`, `clone`, `dialogue` and `enhance` refuse to write audio to a terminal, so pass `--out` or redirect. **`--out` names the file and not the format.** These engines answer mp3 unless asked otherwise, so `--out say.wav` wrote an mp3 called `say.wav`; the CLI now says so on stderr, and refuses up front when `--out` and `--response-format` name different containers. Ask for a container with `--response-format`, whose values carry a sample rate — `wav_16000`, `mp3_44100_128`, `pcm_24000` — so a bare `wav` is refused with the list the engine does take. Get that wrong and the damage is downstream rather than here: an mp3 named `.wav` fed to `listen`, which reads raw PCM, is transcribed as noise.

`speak --voices` lists preset voices; a voice that persists under a name, a voice invented from a description, and the log of readings whose audio can be fetched again for free are a separate surface — [the voice library and the reading history](olares-router-voice.md). `align` takes a transcript from `--text` or standard input. `dialogue` reads a JSON script and turns local `ref_audio` paths into data URLs.

## Long audio decision tree

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

## Asynchronous tasks

Submit asynchronously only when the selected HTTP operation advertises async support, and preserve both the task id and exact model reference. WebSocket and HTTP chunked streams cannot be async. `task get` collects a text result; `task result` collects a binary one:

```
olares-cli router call transcribe meeting-16k-mono.flac --model default-stt --async
olares-cli router call task get <task-id> --model default-stt --wait -o json > transcript.json
olares-cli router call speak "read this" --model default-tts --async
olares-cli router call task result <task-id> --model default-tts --out speech.mp3
```

Each piece is a separate task and a separate billable call. One engine has one worker and a queue of 32 waiting tasks; a full queue refuses new work. Results expire after 1800 seconds, and tasks live only in memory, so a pod restart loses them. On a timeout, keep the id and model and use `task get` or `task list` before considering a resubmission — blindly submitting again can run and bill the work twice. Polls consume Router RPM quota; use `task get --wait` rather than a tight manual loop.

A poll now settles the call as soon as the task reads `succeeded` and names the seconds it moved, so a duration-based charge normally lands on a `task get` rather than waiting for the collection — deliberately, because a caller who polled until the work was done and then walked away used to leave minutes of GPU time in flight until a sweep wrote it off as abandoned, which cost the operator real money and recorded none of it. A finished task that measured nothing leaves the row open instead, since the collection can still price it. Fetching the same result twice is charged once: the second settlement loses a compare-and-swap on the task's binding, and only a Router restart or that binding expiring between the two fetches leaves the ledger's own duplicate check to catch it. JSON task results are already present in `task get`; fetch `task result` only when the result must be collected separately, especially binary audio.

When split STT input hits a queue-full 503, submit slices sequentially. Back off, then use `task get` or `task list` to check the accepted work before submitting the next slice; do not fan out requests or blindly resubmit the rejected slice. Consecutive 503 responses may temporarily trigger Router circuit open, so wait for the circuit and queue to recover instead of increasing concurrency.

`router call task get`, `router call task result` and `router call task cancel` each follow one task by id, at Router's canonical `/v1/tasks` prefix — the one a receipt names in its own `poll` and `result_url`. `/v1/audio/tasks` is the alias Router keeps for clients written before it. Router persists the backend and owner binding for new opaque `atask_*` ids, so a Router restart does not require `--model`; use the saved model for legacy upstream ids or another gateway. `router call task list` always needs `--model`, because every audio application owns a separate queue.

What the binding cannot repair is the engine losing the task itself: only the id was ever Router's, and a pod or engine restart takes the rest, which Router reports as `audio_task_lost` and does not resubmit — a different failure from a finished result ageing out, and the two read differently in [the calling contract's failure table](olares-router-calling.md#reading-a-failure). A binding Router cannot *read* is a third thing again and answers `audio_task_owner_unavailable` with a `Retry-After`: the task is somebody's and Router cannot say whose right now, so the read is worth repeating and the work is not worth resubmitting. `task list` decides the other way and drops a row whose owner it cannot establish, because a listing one row short is still an honest answer. The `model` printed in a receipt is a routing reference, not the engine's canonical model id; preserve it for legacy and cross-gateway recovery.

## Streaming

`listen` and `diarize --stream` are the two CLI verbs that open a WebSocket rather than uploading; Router's complete audio protocol surface also includes `/v1/audio/speech/stream` for streaming TTS. The two input verbs read 16-bit mono PCM from a file or standard input and default `--sample-rate` to 16000. The flag must match the actual PCM and a sample rate advertised by the selected model's operation catalogue. They print partial results as they arrive — on stderr, drawn in place, so partials are on by default only when stderr is a terminal — and need models declaring `supports_stt_stream` and `supports_diar_stream`, separate applications from their batch counterparts; [the diagnosis guide](olares-router-diagnosis.md) covers the empty answers a wrong input can produce.
