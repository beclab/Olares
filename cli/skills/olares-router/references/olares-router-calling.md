# Calling a model

`router call` sends work through Router's data plane, the same path an application uses. It is the fastest way to prove a configuration end to end, and the only verb here that needs a credential of its own.

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

`--model` takes `<provider>/<model>`, as `router list` prints it. Omit it and Router resolves the default for that call's mode — the tenant default, then the workspace one, then the oldest enabled model of that mode. `router default show` prints what would be chosen and from which layer.

A call whose mode has no default and no `--model` is refused by Router rather than guessed at. That is the usual meaning of a refusal on `embed`, `transcribe`, `speak` or `ocr` on a fresh install: only `chat` had a default.

## The five verbs

```
olares-cli router call chat "summarise this" --system "be terse"
cat notes.md | olares-cli router call chat --no-stream --quiet
olares-cli router call chat "what is in this picture" --image shot.png
olares-cli router call embed "text" --dimensions 512
cat lines.txt | olares-cli router call embed --per-line
olares-cli router call transcribe meeting.m4a --language en
olares-cli router call speak "hello" --out hello.mp3
olares-cli router call ocr invoice.pdf --pages 1-3
```

- **chat** streams by default and prints a model and token line after the answer; `--quiet` prints only the answer, `--no-stream` waits for the whole thing. A prompt comes from the arguments or from standard input. `--image` attaches a local file, which requires a model whose row declares `supports_vision`.
- **embed** prints a summary of each vector in table form and the whole vector in JSON. `--per-line` turns piped text into one input per line rather than a single input.
- **transcribe** and **speak** are the two directions of the `audio` mode, and a model that serves one does not necessarily serve the other. `speak` refuses to write audio to a terminal, before making the call, so pass `--out` or redirect.
- **ocr** submits a task and polls it. `--no-wait` prints the task id, `--task <id>` picks it up later, `--cancel` with `--task` drops it, and `--timeout` stops waiting without stopping the task.

`-o json` on any of them prints the upstream's own response, which is what to use when the shape matters more than the reading.

## Reading a failure

A `router call` failure comes from one of three places, and the message says which:

| What it looks like | What it means |
|---|---|
| `invalid_api_key` with type `authentication_error` | Router is refusing *your* key — revoked, expired, or not allowed this model |
| An authentication error without that type | The **vendor** is refusing Router's stored credential; `router provider validate <provider>` confirms it |
| `quota_exceeded` | A ceiling on the key, the user, or the model; `router quota list` shows which |
| A model-not-found or unsupported-endpoint refusal | The row's mode or capabilities do not match the call — check `router provider get` |
| A 5xx with an empty body | Nothing answered behind Router: the model application is stopped or still loading |

The last one is the common case for a local model, and it is a diagnosis rather than a configuration fix: continue in [deciding which layer is wrong](olares-router-diagnosis.md).

Every accepted call becomes a usage row, including one that failed upstream, so `router usage list --limit 5` immediately after a failure shows what Router recorded — status, model, tokens and cost. A call that produced no row was refused before Router routed it.
