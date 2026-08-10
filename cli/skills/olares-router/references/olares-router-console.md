# The Model Console

`router local` addresses the console inside one model application, not Router. It is where a local model's own story is: which weights, which engine, how far the download got, how much sits on the GPU, and why it is not answering yet. Router cannot tell you any of that — from its side an unfinished model application is simply an unreachable provider.

Every verb takes an application reference: the Olares app id (`llamacppqwen3627bggufv3`), or the title of the Router provider that fronts it.

```
olares-cli router local status  llamacppqwen3627bggufv3
olares-cli router local progress llamacppqwen3627bggufv3 --watch
```

The console is reached through the application's own entrance on the active profile; no separate credential is involved. Two refusals are the CLI checking before probing, and both are correct answers rather than errors to retry:

- An application that is stopped, suspended, or crash-looping has nothing listening. The message names the state and the `olares-cli market` verb that changes it.
- An application that is not a model application at all — Router itself, a workflow app — is refused because it serves no Model Console. Pointing `local` at it would otherwise produce half-parsed nonsense.

## Reading the state

`local status` is the one-screen summary: the Olares state, the console's URL and version, the model and mode from the card, then health — `READY`, `ENGINE ALIVE`, `MODEL ON ENGINE`, `LAST VERIFY` — and, while a download is unfinished, how much has arrived, the retry counts and the last error.

`local progress` is the download and load snapshot on its own, with `--watch` to follow it. Speed and an estimate are shown only while a phase is still moving; a settled phase reports what it reached.

The distinction that matters: `ENGINE ALIVE` says a process is up, `MODEL ON ENGINE` says the engine has the model loaded, and `READY` says calls will be served. An engine alive without a model loaded is the normal middle of a start-up, and the normal state of a failed conversion.

## The model card

The card is the source of truth for what this application serves and how the engine is launched. Router mirrors it; it does not own it.

```
olares-cli router local spec show llamacppqwen3627bggufv3
olares-cli router local spec show llamacppqwen3627bggufv3 -o json > card.json
olares-cli router local spec set llamacppqwen3627bggufv3 --from card.json
```

- `spec show` prints the model, mode, context window, maximum output, label and engine arguments. `-o json` returns the card as stored, including fields this CLI does not know — which is why a round trip starts from the JSON and not from the table.
- `spec set` replaces the card. Edit a copy of the JSON and send it back; sending a hand-written subset silently drops whatever was left out.
- `spec file` shows the raw bytes on disk when the console serves that route. Older consoles do not, and the CLI says so rather than reporting a missing card.

A card change does not take effect by itself. `local restart` relaunches the engine with the new card; `local retry` re-enters the download and load loop, which is what a changed model source needs. After either, `router provider sync-models <provider>` re-mirrors the rows into Router — otherwise Router still advertises the old capabilities.

## Diagnostics

| Verb | Answers |
|---|---|
| `local config` | the effective configuration, secrets redacted — the resolved model source, engine kind and paths |
| `local endpoints` | which routes this deployment actually serves, and why a missing one is missing |
| `local gpu` | how much of the model is resident: layers on GPU, VRAM in use, CPU offload, KV cache, and what measured it |
| `local perf` | a real probe — time to first token and prefill/decode throughput, with the thinking variant when the card claims reasoning |

`local endpoints` is the check to run before believing a 404 from anything in this family: consoles differ by version, and a route that never existed on this one is a different problem from a resource that is absent.

`local gpu` reports placement rather than a single number, because the engines disagree about what to report. For llama.cpp the layer count is the fact; some engines report no residency at all, and the report says so instead of implying zero.

`local perf` runs a real generation, so it costs GPU time and takes a few seconds. It is the honest answer to "is this model slow", where `local gpu` is the answer to "why".

## Retry and restart

```
olares-cli router local retry   <app>   # re-enter download and load now
olares-cli router local restart <app>   # relaunch the engine with the current card
```

`retry` is for a download that failed or stalled — it re-enters the lifecycle rather than waiting for the next backoff. `restart` is for an engine that is up with the wrong flags, or wedged. Neither reinstalls the application, and neither deletes weights already on disk.

If both leave the phase where it was, the problem is below the console — an image that will not pull, a node without the memory, a GPU binding that is missing. Continue in [deciding which layer is wrong](olares-router-diagnosis.md), which routes to [`olares-doctor`](../../olares-doctor/SKILL.md) for those.
