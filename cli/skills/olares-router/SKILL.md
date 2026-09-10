---
name: olares-router
version: 0.0.0-cli.0
description: "Olares models via olares-cli router — Router (the AI gateway) and the Model Console inside a locally installed model application. Configure cloud vendors and their models, manage local LLM / embedding / audio / OCR model applications, edit a local model's card, name models with aliases, groups and default categories, issue API keys and quotas, read usage and audit, and call a model: chat, embed, rerank, search, scrape, translate, images, video, music, 3D, transcribe, speak, diarize, OCR. Requires Olares 1.12.7+. Use for Router, llm-gateway, AI gateway, 模型, 模型网关, 本地模型, add an OpenAI/Anthropic/DeepSeek key, install a Qwen or Gemma model, sk- key, model quota, token spend, model alias, default-chat, engine_args, which model answers by default, why a model call fails."
compatibility: Requires olares-cli on PATH, active Olares profile, Olares >= 1.12.7
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# router (Router and Model Console)

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

Use `olares-cli router <verb> --help` for authoritative syntax.

## When to use

- Give Olares access to a cloud vendor's models, or inspect what it already has.
- Install, configure, or diagnose a model application running on this machine.
- Decide what a caller may put in the `model` field: an alias, a group of models, or a default category.
- Issue keys and quotas for software that calls models, and read what it spent.
- Send work to a model from the command line: chat, embeddings, reranking, web search and scraping, translation, images, video, speech in either direction, and OCR.

> **Mental model:** every call goes through **Router**, one AI gateway per Olares. Router holds providers, models, keys, quotas and the usage record; it runs no model itself. A local model runs inside a **model application**, whose own **Model Console** downloads the weights, launches the engine and serves the OpenAI-compatible endpoint Router forwards to. So **a local model has two states, not one**: the platform says whether the container is up, the application says whether the model can answer, and a container reports `running` minutes before the weights have loaded. Never read one as evidence for the other. [Architecture and identity](references/olares-router-architecture.md) is the long form, and is worth reading before the first write.

All verbs require Olares 1.12.7+ because Router ships as the `router` Market listing, which asks for that line. Router is an admin-only application: a non-admin profile cannot see its entrance, so every verb here reports it is not installed. Check `olares-cli profile whoami` for the role and `olares-cli market list --mine` for the application before concluding anything is missing.

## Fast paths

| Task | Read | First command |
|---|---|---|
| Prove a model answers | this file | `olares-cli router call chat "hello"` |
| See what this profile may call | this file | `olares-cli router call models` |
| Give Olares a cloud vendor's models | [configuring an external provider](references/olares-router-external.md) | `olares-cli router provider create --type openai --credentials-json @openai.json --validate` |
| Find out what a failed call recorded | this file | `olares-cli router usage list --limit 5` |

Calling needs no key: Router reads the caller from the platform, so `router call` arrives as the active profile. Adding a vendor is two writes, not one — `provider create` stores the credential, `model import` is what makes a model callable, and neither is optional.

## Which layer owns the change

| Intent | Where it belongs |
|---|---|
| Use a vendor's hosted models | `provider create` + `model import`, or `provider sync-models` for an endpoint that publishes its own list |
| Run a model on this machine | [`olares-market`](../olares-market/SKILL.md)'s `market install` for a pinned model, `market clone` for an engine base; Router creates the provider once the application runs |
| Change what a local model serves or how it is launched | `model spec edit <model>` — the model card, which the application owns; not the Router row |
| Change the address, credentials, or enabled state Router routes with | `provider update` |
| Repair an install that failed | `provider get <app>` then `model progress --app <app>` / `model retry --app <app>` |
| Restart an engine that has stopped behaving | `model restart <model>`, which relaunches the inference process without changing the card |
| Stop, resume, or bind a model application to a GPU | [`olares-market`](../olares-market/SKILL.md) and [`olares-settings`](../olares-settings/SKILL.md) — Router does not own those |

A provider whose `source` is `olares` belongs to a Market application. Its address and lifecycle are the Market's; `provider delete` refuses it, and `olares-cli market uninstall <app>` is the way out.

## Verb index

| Family | Verbs | Read when triggered |
|---|---|---|
| where Router is, and who you are | no verb of its own — every verb resolves it, and says so when it cannot | [architecture and identity](references/olares-router-architecture.md) |
| cloud vendors and their models | `provider list/get/types/create/update/delete/validate/credentials/history/rollback/sync-models`, `provider validate --draft` before a row exists, `model get/import/add/update/remove` | [configuring an external provider](references/olares-router-external.md) |
| local LLM applications | `provider register`, `model status/progress/retry/restart`, plus [`olares-market`](../olares-market/SKILL.md)'s `install` / `clone` | [local LLM applications](references/olares-router-local-llm.md) |
| local embedding, audio, OCR, CLIP | the same verbs, different modes | [local multimodal applications](references/olares-router-local-multimodal.md) |
| what a local model declares itself to be | `model spec show/edit/file/set`, `model restart` | [local LLM applications](references/olares-router-local-llm.md) |
| what is configured | `model list`, `model get` | [names, defaults and access control](references/olares-router-governance.md) |
| the names callers may send | `route list/get/create/rename/enable/disable/delete/add/remove`, the `default-*` categories via `route list --kind default`, and `route candidates/pin/unpin` to say which model one of them means | [names, defaults and access control](references/olares-router-governance.md) |
| access control | `key issue/list/update/revoke/current`, `quota set/list/clear` | [names, defaults and access control](references/olares-router-governance.md) |
| what happened | `usage summary/list/export/retention/apps`, `audit list/get` | [usage and audit](references/olares-router-usage.md) |
| what any call needs, and why one failed | `call models`, `key current`, `call ocr` | [the calling contract](references/olares-router-calling.md) |
| asking a model in words | `call chat/responses/count-tokens/embed/rerank/search/scrape/translate` | [text and the web](references/olares-router-call-text.md) |
| generating a picture, a clip, a track or a mesh | `call image/video/music/3d` — the four whose work outlives the request; `music` and `3d` require `--model`, and `call music format/draft/alignment/cancel` are music's own | [pictures, clips, tracks and meshes](references/olares-router-call-media.md) |
| speech in either direction | `call transcribe/listen/speak/clone/dialogue/vad/diarize/speaker-embed/enhance/align`, `call task get/result/list/cancel` | [audio](references/olares-router-call-audio.md) |
| a voice that outlives the call, and what was read out | `call voice list/get/add/design/settings/delete`, `call history list/get/download/delete` | [the voice library and the reading history](references/olares-router-voice.md) |
| inside one application | `model status/progress/retry/restart`, `model diag gpu/config/endpoints`, all taking `--app` | [the Model Console](references/olares-router-console.md) |
| a call or a model that does not work | any of the above | [deciding which layer is wrong](references/olares-router-diagnosis.md) |

## Naming

- A model is addressed as `<provider>/<model>` wherever ambiguity is possible — in `--model`, in a quota, in a key's allowed list. `router model list` prints both halves. A name without a slash is a **route** — an alias, a group, or a `default-*` category — and has to exist.
- Every locally installed model application is a provider named `Olares`, so the qualified name is not unique for local models. `<app_name>/<model>` names one of them — `llamacppqwen3v3/qwen3-8b` — and so does the application's display title, which `router model list` prints in `SERVED BY`. The model id is the only handle that always names one row, and an ambiguous reference is refused with the candidates rather than resolved to one.
- A provider is named by its title, its Olares app name, or its id. A model application is named by its Olares app id (`llamacppqwen3627bggufv3`), which is what `provider register` takes and what `--app` takes on the `model` verbs that reach a Model Console. Those verbs otherwise take a model name and ask Router which application serves it; `--app` skips Router, which is what keeps them usable when Router is the thing that is wrong.
- An application that *calls* Router has no row here at all: Olares vouches for it at the edge and the call arrives carrying an `appid` — the app name hashed, or the name itself for a system app. So it cannot be registered or revoked; `olares-cli market list --mine` says whether it is here, `usage --by caller_app` says what it spent, and `quota set --caller-app` is the only lever over it.

## Safety and escalation

- A named configuration request authorises the loop it implies: creating a provider, importing its models and validating it do not need re-confirmation one by one.
- Ask again before `provider delete`, `key revoke`, `quota clear` or `route delete` on something the user did not name — each one breaks callers that still depend on it. `route disable` is reversible and keeps its membership; `route delete` gives the name up.
- **Never** put a credential in a shell argument where a file or stdin will do; `--credentials-json` reads either. Never print a plaintext `sk-` key into a transcript that will be shared: `key issue` shows it once, on purpose.
- `usage retention --days` deletes per-call rows outside the new window immediately, and shortening it is not undoable. Daily totals survive, so confirm before shortening one somebody did not ask for.
- `model spec edit --engine-args` and `model restart` both relaunch the process serving a model, which stops answering until the weights have loaded again — minutes for a large one. Prefer `model spec edit` over `model spec set`: the first merges onto the card the application is serving and updates Router's copy, the second replaces the document at the application, and a field omitted from a replacement is gone.
- **Never retry a write that timed out without reading the error first.** Router has an idempotency key on the three music submissions and nowhere else, so a second attempt is a second request: `key issue` mints another key, `provider rollback` appends another version, a `router call` bills again. The error says the outcome is unknown and what a repeat would do to that route — most of them are refused on a duplicate, which makes retrying safe and a `409` proof the first one landed. A failed read can normally be run again, `task result` included: a second collection of the same result loses a compare-and-swap on the task's binding and settles once, so a retry there costs the request rather than the work.
- A model that is configured but does not answer is a diagnosis, not a configuration change. Route through [deciding which layer is wrong](references/olares-router-diagnosis.md) before editing anything.
- Stop for the shared auth gate on a persistent authentication failure, and stop when the target provider, model, application or user is ambiguous.
