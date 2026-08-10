# Local LLM applications

A local model is an Olares application. Installing it and telling Router about it are one step: `router app install` starts the Market install **and** creates the provider row that will route to it, addressed at the application's in-cluster shared entrance and marked pending until the Market reports it running.

Admin only, the catalog included.

## Choosing what to install

```
olares-cli router app catalog
olares-cli router app install llamacppqwen3627bggufv3 --watch
```

The catalog is the Market's own list of model applications, so two kinds appear in it:

- **A pinned model** — `Qwen3.6-27B (llama.cpp)`, `Gemma 4 26B (Ollama)`. The application knows which weights it serves; installing it is the whole decision.
- **An engine base** — `llama.cpp Engine Base`, `vLLM Engine Base`, `Ollama Engine Base`, `SGLang Engine Base`. The engine is fixed and the model is chosen when it is installed.

`app install` carries only the application's name. An engine base therefore arrives with nothing chosen: install it through [`olares-market`](../../olares-market/SKILL.md), where the install form exists, or install it here and then set its model card with `router local spec set` followed by `router local retry`.

The same application name can appear twice at different versions when more than one Market source offers it. Install names the application; the Market picks the version.

The engine also decides the weight format, which is not interchangeable: llama.cpp wants GGUF, vLLM and SGLang want Safetensors, Ollama wants a library model. A model card naming the wrong kind fails during download, not at launch.

## Following the install

An install of real weights takes minutes to hours. `--watch` follows it to the end; without it the command returns as soon as the task is accepted.

```
olares-cli router app tasks "Qwen3.6-27B (llama.cpp)"
olares-cli router app watch  "Qwen3.6-27B (llama.cpp)" --task 42
```

`app tasks <provider>` is where a failed install explains itself: the row keeps the error the Market reported. `app watch` replays a task's recorded events and then streams the live ones, so a task that finished long ago still reports what it did; `--since` resumes an interrupted watch.

Two states are not the same thing, and this is the most common confusion:

| Question | Where the answer is |
|---|---|
| Did the *application* install? | `app tasks`, or `olares-cli market status <app>` |
| Is the *model* downloaded and loaded? | `router local progress <app>` |

The Market can report an application running long before the model inside it is usable — the weights download after the container starts. A provider that is `active` with no models usually means exactly that.

## After the install

1. `router local progress <app>` until the download and load settle.
2. `router provider get <app-or-title>` to confirm Router now sees its models. A local application publishes its own list, so Router mirrors it rather than needing `provider models import`; `router provider sync-models <provider>` re-mirrors it if the card changed.
3. `router default set --mode chat=<provider>/<model>` if this should answer requests that name no model — see [defaults and access control](olares-router-governance.md).
4. `router call chat "hello" --model <provider>/<model>` to prove the whole path.

## Repairing

| Symptom | Step |
|---|---|
| Install failed | `app tasks <provider>` for the Market's reason; fix it, then `app install` again |
| Application installed, no Router provider | `router provider register <app>` creates the row for an application already on the machine |
| Download stuck or failed | `router local progress <app>`, then `router local retry <app>` |
| Model card wrong, or engine flags need changing | `router local spec set <app>`, then `router local restart <app>` |
| Provider exists, application does not answer | [deciding which layer is wrong](olares-router-diagnosis.md) |

`router app install` refuses an application that is already installed, naming it, rather than starting a second install that the Market would reject and that would leave a failed task on the existing provider. Use `app upgrade` for a newer version and `olares-cli market resume` for one that is stopped.

## Upgrading and removing

```
olares-cli router app upgrade   "Qwen3.6-27B (llama.cpp)" --watch
olares-cli router app uninstall "Qwen3.6-27B (llama.cpp)"
```

Both address the **provider**, not the application — Router looks up which application backs it. `app uninstall` removes the application and its provider together, which is the only way to remove an `olares`-sourced provider: `provider delete` refuses it.

A provider row survives a failed install on purpose, because it carries the task history that explains the failure. Removing that row means uninstalling the application.

Stopping, resuming, and binding an application to a GPU are not Router's: they belong to [`olares-market`](../../olares-market/SKILL.md) and [`olares-settings`](../../olares-settings/SKILL.md). Router only reports the application's state on the provider row, and hides an `olares` provider from `provider list` while its application is not running — `router provider get <id>` still shows it.

## Non-text models

Embedding, audio, OCR and CLIP applications install through exactly these verbs. What differs is the mode their models declare, the engine behind them and how they are called: see [local multimodal applications](olares-router-local-multimodal.md).
