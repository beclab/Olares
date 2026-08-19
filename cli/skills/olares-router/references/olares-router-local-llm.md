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
- **An engine base** — `llama.cpp Engine Base`, `vLLM Engine Base`, `Ollama Engine Base`, `SGLang Engine Base`. The engine is fixed and the model is chosen while an instance is created from it.

`app install` carries only the application's name, so it can install a pinned model and nothing else. An engine base is a template: it has no installable form, and the Market refuses a direct install. `app catalog`'s TAKES column says which of the two a row is, and `app install` refuses a template rather than letting the Market refuse it a moment later.

Creating an instance from a base is [`olares-market`](../../olares-market/SKILL.md)'s `market clone`, where the model, the engine arguments and the compute mode are chosen; the workflow is in [`olares-chart`](../../olares-chart/SKILL.md). Router discovers the instance on its own once it is running, and `router spec edit` changes what it serves after that.

The same application name can appear twice when more than one Market source publishes it, usually at different versions, and only one copy can be installed — one app name occupies one namespace. `app install` refuses such a name rather than guessing; `--source` picks the copy. A row whose name is already held by another source says so in the catalog's STATE column and cannot be installed until that copy goes.

The engine also decides the weight format, which is not interchangeable: llama.cpp wants GGUF, vLLM and SGLang want Safetensors, Ollama wants a library model. A model card naming the wrong kind fails during download, not at launch.

## Following the install

An install of real weights takes minutes to hours. `--watch` follows it to the end; without it the command returns as soon as the Market accepts the request.

```
olares-cli router app install llamacppqwen3627bggufv3 --watch
olares-cli router provider get llamacppqwen3627bggufv3
```

There is nothing to catch up on afterwards and no separate watch command, because what `--watch` follows is the provider row's own status: Router's application directory keeps it current whether or not anyone is looking, so `provider get` answers the same question at any later moment. Interrupting a watch stops the watch, not the install.

Two states are not the same thing, and this is the most common confusion:

| Question | Where the answer is |
|---|---|
| Did the *application* install? | `router provider get <app>`, or `olares-cli market status <app>` |
| Is the *model* downloaded and loaded? | `router local progress <app>` |

The Market can report an application running long before the model inside it is usable — the weights download after the container starts. A provider that is `active` with no models usually means exactly that.

## After the install

1. `router local progress <app>` until the download and load settle.
2. `router provider get <app-or-title>` to confirm Router now sees its models. A local application publishes its own list, so Router mirrors it rather than needing `provider models import`; `router provider sync-models <provider>` re-mirrors it if the card changed.
3. `router default show` to see whether Router has pointed `default-chat` at the new model. It does that itself, against what the model says it can do — there is nothing to set. If the category is still empty, the model is not enabled or has not published a card yet. See [names, defaults and access control](olares-router-governance.md).
4. `router call chat "hello" --model <provider>/<model>` to prove the whole path.

## The model card

The card is where a local model's mode, capabilities, prices, context window and engine flags are declared. The application that runs the model owns it; Router keeps a projection to route and bill against, and refreshes that projection when the application reaches running. So the two can be a restart apart, and that gap is the answer to "why does Router offer something this model cannot do".

```
olares-cli router spec show Olares/qwen3-4b
olares-cli router spec edit Olares/qwen3-4b --mode chat
olares-cli router spec edit Olares/qwen3-4b --engine-args "--ctx-size 8192 --n-gpu-layers 99"
olares-cli router spec edit Olares/qwen3-4b --from card.json
olares-cli router spec restart Olares/qwen3-4b
```

`spec show` says where its answer came from: `live` means the application was asked, `cache` means it could not be and this is Router's stored copy — which is still the copy that routes and bills. `-o json` prints the card whole, including fields this CLI does not know about, which is what makes it the thing to edit and hand back to `--from`.

`spec edit` merges. Keys you do not mention survive, one level deep: `pricing`, `supports` and `parameter_rules` are each replaced as a unit, so changing one price means sending the pricing object you want. Router writes the merged card to the application, then stores what the application confirms — which is not always what was sent, since the Model Console adds the flags a mode requires and files capability keys it does not know under `extensions`. That write-back is why the projection is right immediately instead of at the next restart.

Two consequences. Changing `--engine-args` relaunches the inference process, and the model does not answer until the weights have loaded again. And the application has to be running: a card cannot be written into something that is not there, which is when `spec show` starts saying `cache`.

`router local spec set` reaches the same document at the application instead of through Router. It is a whole-document replace with no merge, so a field left out is gone — engine flags included. Use it for an application Router has no provider row for, and `spec edit` otherwise.

An application whose engine is a sidecar — OCR, audio, embedding — takes no engine flags at all. There `--mode` is the field that matters, because it is the gate the data plane routes on, and `spec restart` succeeds having changed nothing.

## Repairing

| Symptom | Step |
|---|---|
| Install failed | `olares-cli market status <app>` for the Market's own reason; fix it, then `app uninstall` the failed row and `app install` again |
| Application installed, no Router provider | `router provider register <app>` creates the row for an application already on the machine |
| Download stuck or failed | `router local progress <app>`, then `router local retry <app>` |
| Model card wrong, or engine flags need changing | `router spec edit <model>` |
| Engine wedged on a card that is right | `router spec restart <model>` |
| Provider exists, application does not answer | [deciding which layer is wrong](olares-router-diagnosis.md) |

`router app install` refuses an application that is already installed, naming it, rather than starting a second install that the Market would reject and that would leave a failed task on the existing provider. Use `app upgrade` for a newer version and `olares-cli market resume` for one that is stopped.

## Upgrading and removing

```
olares-cli router app upgrade   "Qwen3.6-27B (llama.cpp)" --watch
olares-cli router app uninstall "Qwen3.6-27B (llama.cpp)"
```

Both address the **provider**, not the application — Router looks up which application backs it. `app uninstall` removes the application and its provider together, which is the only way to remove an `olares`-sourced provider: `provider delete` refuses it.

Every locally installed provider carries the same routing name, `Olares`, so that is not the handle to type. The application name is, and so is the display title an admin gave it; `provider list`'s APP column shows the first. A name that matches several rows is refused rather than resolved to whichever came back first.

A provider row survives a failed install on purpose: it is what records that an install was attempted and how it ended. Removing that row means uninstalling the application.

Stopping, resuming, and binding an application to a GPU are not Router's: they belong to [`olares-market`](../../olares-market/SKILL.md) and [`olares-settings`](../../olares-settings/SKILL.md). Router only reports the application's state on the provider row, and hides an `olares` provider from `provider list` while its application is not running — `router provider get <app>` still shows it, and so does `app catalog`.

## Non-text models

Embedding, audio, OCR and CLIP applications install through exactly these verbs. What differs is the mode their models declare, the engine behind them and how they are called: see [local multimodal applications](olares-router-local-multimodal.md).
