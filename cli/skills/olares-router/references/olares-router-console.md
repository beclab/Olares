# The Model Console

Some `router model` verbs do not ask Router anything. They reach the console inside the application serving the model, which is where a local model's own story is: which weights, which engine, how far the download got, how much sits on the GPU, and why it is not answering yet. Router cannot tell you any of that — from its side an unfinished model application is simply an unreachable provider.

```
olares-cli router model status   qwen3-4b
olares-cli router model progress qwen3-4b --watch
```

## Naming the application

Every one of these verbs takes a model, and asks Router which application serves it. `--app` names the application directly instead — by its Olares app id (`llamacppqwen3627bggufv3`), or by the title of the Router provider that fronts it — and skips Router entirely.

```
olares-cli router model status --app llamacppqwen3627bggufv3
```

Prefer the model name; reach for `--app` when Router cannot resolve the model, when two applications answer to one name, or when Router is not answering at all. That last case is not hypothetical: a model that is not responding is exactly when the gateway in front of it may also be wrong, and `--app` is what keeps the diagnostic available then.

A model that resolves to a cloud provider is refused, with the reason. There is no console at the other end of one, and the whole of this page is inapplicable to it.

The console is reached through the application's own entrance on the active profile; no separate credential is involved. Two further refusals are the CLI checking before probing, and both are correct answers rather than errors to retry:

- An application that is stopped, suspended, or crash-looping has nothing listening. The message names the state and the `olares-cli market` verb that changes it.
- An application that is not a model application at all — Router itself, a workflow app — is refused because it serves no Model Console. Probing it would otherwise produce half-parsed nonsense.

## Reading the state

`model status` is the one-screen summary: the Olares state, the console's URL and version, the model and mode from the card, then health — `READY`, `ENGINE ALIVE`, `MODEL ON ENGINE`, `LAST VERIFY` — and, while a download is unfinished, how much has arrived, the retry counts and the last error.

`model progress` is the download and load snapshot on its own, with `--watch` to follow it. Speed and an estimate are shown only while a phase is still moving; a settled phase reports what it reached.

The distinction that matters: `ENGINE ALIVE` says a process is up, `MODEL ON ENGINE` says the engine has the model loaded, and `READY` says calls will be served. An engine alive without a model loaded is the normal middle of a start-up, and the normal state of a failed conversion.

## The model card

The card is the source of truth for what this application serves and how the engine is launched. Router mirrors it; it does not own it.

`model spec` carries both roads to it, and the difference is the write. `spec edit` goes through Router: it merges the change onto the card the application is serving and writes Router's own copy back, which is why it is the one to prefer. It is written up in [local LLM applications](olares-router-local-llm.md). `spec set` goes straight at the application and replaces the whole document.

```
olares-cli router model spec show qwen3-4b
olares-cli router model spec show qwen3-4b -o json > card.json
olares-cli router model spec set  qwen3-4b --from card.json
```

- `spec show` prints the model, mode, context window, maximum output, label and engine arguments. Through Router it also says whether the answer came from the application or from Router's stored copy; with `--app` it is always the application's own. `-o json` returns the card as stored, including fields this CLI does not know — which is why a round trip starts from the JSON and not from the table.
- `spec set` replaces the card. Edit a copy of the JSON and send it back; sending a hand-written subset silently drops whatever was left out, engine flags included, and an engine relaunched without its flags is a different engine.
- `spec file` shows the raw bytes on disk when the console serves that route. Older consoles do not, and the CLI says so rather than reporting a missing card.

The advertised context window must not exceed the one the engine was actually launched with. Nothing enforces that, and a card claiming more invites prompts the engine then truncates or rejects.

The values a model application was installed with — its model source and name, its capability groups, its engine arguments — seed the first boot and are not a second authority afterwards. This card is, which is why correcting a running model here beats reinstalling it with different environment values. Building or cloning such an application in the first place is [`olares-chart`](../../olares-chart/SKILL.md)'s, and the install form is [`olares-market`](../../olares-market/SKILL.md)'s.

A card written with `spec set` does not take effect by itself, and Router does not hear about it. `model restart` relaunches the engine with the new card; `model retry` re-enters the download and load loop, which is what a changed model source needs. After either, `router provider sync-models <provider>` re-mirrors the rows into Router — otherwise Router still advertises the old capabilities. None of those three steps is needed after `spec edit`, which is most of the reason to prefer it.

## Diagnostics

| Verb | Answers |
|---|---|
| `model diag config` | the effective configuration, secrets redacted — the resolved model source, engine kind and paths |
| `model diag endpoints` | which routes this deployment actually serves, and why a missing one is missing |
| `model diag gpu` | how much of the model is resident: layers on GPU, VRAM in use, CPU offload, KV cache, and what measured it |
| `model diag perf` | a real probe — time to first token and prefill/decode throughput, with the thinking variant when the card claims reasoning |

These are for a model that works but behaves oddly. Whether it works at all is `model status`.

`model diag endpoints` is the check to run before believing a 404 from anything in this family: consoles differ by version, and a route that never existed on this one is a different problem from a resource that is absent.

`model diag gpu` reports placement rather than a single number, because the engines disagree about what to report. For llama.cpp the layer count is the fact; some engines report no residency at all, and the report says so instead of implying zero.

`model diag perf` runs a real generation, so it costs GPU time and takes a few seconds. It is the honest answer to "is this model slow", where `model diag gpu` is the answer to "why".

## Retry and restart

```
olares-cli router model retry   <model>   # re-enter download and load now
olares-cli router model restart <model>   # relaunch the engine with the current card
```

`retry` is for a download that failed or stalled — it re-enters the lifecycle rather than waiting for the next backoff. `restart` is for an engine that is up with the wrong flags, or wedged. Neither reinstalls the application, and neither deletes weights already on disk.

`restart` is the one verb here with a Router road as well as a direct one: named by model it goes through Router, and `--app` signals the console itself. Both mean the same thing to a caller of the model.

If both leave the phase where it was, the problem is below the console — an image that will not pull, a node without the memory, a GPU binding that is missing. Continue in [deciding which layer is wrong](olares-router-diagnosis.md), which routes to [`olares-doctor`](../../olares-doctor/SKILL.md) for those.
