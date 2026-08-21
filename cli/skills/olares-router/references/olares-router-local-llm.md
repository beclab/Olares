# Local LLM applications

A local model is an Olares application. Installing it belongs to
[`olares-market`](../../olares-market/SKILL.md); Router's part starts once the
application is on the machine. Router's application directory notices it and
creates the provider row that routes to it, addressed at the application's
in-cluster shared entrance — so nothing here has to be told about an install.

Everything on this page is admin only.

## Choosing what to install

```
olares-cli market list -c AI
olares-cli market install llamacppqwen3627bggufv3 --watch
```

The Market publishes two kinds of model application:

- **A pinned model** — `Qwen3.6-27B (llama.cpp)`, `Gemma 4 26B (Ollama)`. The application knows which weights it serves; installing it is the whole decision.
- **An engine base** — `llama.cpp Engine Base`, `vLLM Engine Base`, `Ollama Engine Base`, `SGLang Engine Base`. The engine is fixed and the model is chosen while an instance is created from it.

`market install` takes a pinned model. An engine base is a template with no
installable form: `market clone <base> --title <name>` creates an instance from
it, and that is where the model, the engine arguments and the compute mode are
chosen — the per-engine values are in [`olares-chart`](../../olares-chart/SKILL.md)'s
LLM model workflow. Only `--title` is enforced, so a clone missing the rest of
the template's published environment is created and then fails to serve:
`MODEL_SOURCE`, `MODEL_NAME`, `MODEL_MODE`, `MODEL_SUPPORTS`, `ENGINE_ARGS` and
the engine's own `<ENGINE>_REQUIRED_GPU_MEMORY` all belong on that command.

The same application name can appear twice when more than one Market source
publishes it, usually at different versions, and only one copy can be installed —
one app name occupies one namespace. `market install -s <source>` picks the copy.

The engine also decides the weight format, which is not interchangeable: llama.cpp wants GGUF, vLLM and SGLang want Safetensors, Ollama wants a library model. A model card naming the wrong kind fails during download, not at launch.

## Following the install

An install of real weights takes minutes to hours. `market install --watch`
follows the Market's own task; without it the command returns as soon as the
request is accepted.

```
olares-cli market install llamacppqwen3627bggufv3 --watch
olares-cli router provider get llamacppqwen3627bggufv3
```

There is nothing to catch up on afterwards. Router's application directory keeps
the provider row current whether or not anyone is looking, so `provider get`
answers the same question at any later moment — including in the minutes right
after the install starts, when the row exists and the application does not
answer yet.

Two states are not the same thing, and this is the most common confusion:

| Question | Where the answer is |
|---|---|
| Did the *application* install? | `router provider get <app>`, or `olares-cli market status <app>` |
| Is the *model* downloaded and loaded? | `router model progress <model>`, or `--app <app>` |

The Market can report an application running long before the model inside it is usable — the weights download after the container starts. A provider that is `active` with no models usually means exactly that.

## After the install

1. `router model progress --app <app>` until the download and load settle. The model name works too once Router has the row; the app id works before it does.
2. `router provider get <app-or-title>` to confirm Router now sees its models. A local application publishes its own list, so Router mirrors it rather than needing `model import`; `router provider sync-models <provider>` re-mirrors it if the card changed.
3. `router route list --kind default` to see whether Router has pointed `default-chat` at the new model. It does that itself, against what the model says it can do — there is nothing to set. If the category is still empty, the model is not enabled or has not published a card yet. See [names, defaults and access control](olares-router-governance.md).
4. `router call chat "hello" --model <provider>/<model>` to prove the whole path.

## The model card

The card is where a local model's mode, capabilities, prices, context window and engine flags are declared. The application that runs the model owns it; Router keeps a projection to route and bill against, and refreshes that projection when the application reaches running. So the two can be a restart apart, and that gap is the answer to "why does Router offer something this model cannot do".

```
olares-cli router model spec show Olares/qwen3-4b
olares-cli router model spec edit Olares/qwen3-4b --mode chat
olares-cli router model spec edit Olares/qwen3-4b --engine-args "--ctx-size 8192 --n-gpu-layers 99"
olares-cli router model spec edit Olares/qwen3-4b --from card.json
olares-cli router model restart Olares/qwen3-4b
```

`model spec show` says where its answer came from: `live` means the application was asked, `cache` means it could not be and this is Router's stored copy — which is still the copy that routes and bills. `-o json` prints the card whole, including fields this CLI does not know about, which is what makes it the thing to edit and hand back to `--from`.

`model spec edit` merges. Keys you do not mention survive, one level deep: `pricing`, `supports` and `parameter_rules` are each replaced as a unit, so changing one price means sending the pricing object you want. Router writes the merged card to the application, then stores what the application confirms — which is not always what was sent, since the Model Console adds the flags a mode requires and files capability keys it does not know under `extensions`. That write-back is why the projection is right immediately instead of at the next restart.

Two consequences. Changing `--engine-args` relaunches the inference process, and the model does not answer until the weights have loaded again. And the application has to be running: a card cannot be written into something that is not there, which is when `model spec show` starts saying `cache`.

`router model spec set` reaches the same document at the application instead of through Router. It is a whole-document replace with no merge, so a field left out is gone — engine flags included. Use it for an application Router has no provider row for, and `model spec edit` otherwise.

An application whose engine is a sidecar — OCR, audio, embedding — takes no engine flags at all. There `--mode` is the field that matters, because it is the gate the data plane routes on, and `model restart` succeeds having changed nothing.

## Repairing

| Symptom | Step |
|---|---|
| Install failed | `olares-cli market status <app>` for the Market's own reason; fix it, then `olares-cli market uninstall <app>` and install again |
| Application installed, no Router provider | `router provider register <app>` creates the row for an application already on the machine |
| Download stuck or failed | `router model progress --app <app>`, then `router model retry --app <app>` |
| Model card wrong, or engine flags need changing | `router model spec edit <model>` |
| Engine wedged on a card that is right | `router model restart <model>` |
| Provider exists, application does not answer | [deciding which layer is wrong](olares-router-diagnosis.md) |

An application that is already installed is refused by the Market rather than
installed twice. `olares-cli market upgrade <app>` moves it to a newer version,
and `olares-cli market resume <app>` starts one that is stopped.

## Upgrading and removing

```
olares-cli market upgrade   llamacppqwen3627bggufv3 --watch
olares-cli market uninstall llamacppqwen3627bggufv3
```

These address the **application**, by its Olares app name. Uninstalling removes
the provider with it, which is the only way to remove an `olares`-sourced
provider: `provider delete` refuses it. `provider list`'s APP column is where to
read the app name of a provider you only know by title.

Every locally installed provider carries the same routing name, `Olares`, so that is not the handle to type. The application name is, and so is the display title an admin gave it; `provider list`'s APP column shows the first. A name that matches several rows is refused rather than resolved to whichever came back first.

A provider row survives a failed install on purpose: it is what records that an install was attempted and how it ended. Removing that row means uninstalling the application.

Stopping, resuming, and binding an application to a GPU are not Router's: they belong to [`olares-market`](../../olares-market/SKILL.md) and [`olares-settings`](../../olares-settings/SKILL.md). Router only reports the application's state on the provider row, and hides an `olares` provider from `provider list` while its application is not running — `router provider get <app>` still shows it.

## Non-text models

Embedding, audio, OCR and CLIP applications install the same way. What differs is the mode their models declare, the engine behind them and how they are called: see [local multimodal applications](olares-router-local-multimodal.md).
