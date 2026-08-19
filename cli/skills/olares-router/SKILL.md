---
name: olares-router
version: 1.2.0
description: "Olares models via olares-cli router — Router (the AI gateway) and the Model Console inside a locally installed model application. Configure cloud vendors and their models, install and manage local LLM / embedding / audio / OCR model applications, name models with aliases, groups and default categories, issue API keys and quotas, read usage, audit and traces, and call a model: chat, embed, transcribe, speak, OCR. Requires Olares 1.12.7+. Use for Router, llm-gateway, AI gateway, 模型, 模型网关, 本地模型, add an OpenAI/Anthropic/DeepSeek key, install a Qwen or Gemma model, sk- key, model quota, token spend, model alias, default-chat, which model answers by default, why a model call fails."
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
- Send work to a model from the command line: chat, embeddings, transcription, speech, OCR.

> **Mental model:** every call goes through **Router**, one AI gateway per Olares. Router holds providers, models, keys, quotas and the usage record; it runs no model itself. A local model runs inside a **model application**, whose own **Model Console** downloads the weights, launches the engine and serves the OpenAI-compatible endpoint Router forwards to. Router is the plane where access is decided; the Model Console is the plane where one model lives.

All verbs require Olares 1.12.7+ because Router ships as the `router` Market listing, which asks for that line. Router is an admin-only application: a non-admin profile cannot see its entrance, so every verb here reports it is not installed. Confirm with `router status` before concluding anything is missing.

## Verb index

| Family | Verbs | Read when triggered |
|---|---|---|
| where Router is, and who you are | `status`, `whoami` | [architecture and identity](references/olares-router-architecture.md) |
| cloud vendors and their models | `provider list/get/types/create/update/delete/validate/credentials/history/rollback/sync-models`, `provider models get/import/add/update/delete` | [configuring an external provider](references/olares-router-external.md) |
| local LLM applications | `app catalog/installed/install/upgrade/uninstall`, `provider register`, `local status/progress/spec/retry/restart` | [local LLM applications](references/olares-router-local-llm.md) |
| local embedding, audio, OCR, CLIP | the same verbs, different modes | [local multimodal applications](references/olares-router-local-multimodal.md) |
| what is configured | `list`, `capabilities` | [names, defaults and access control](references/olares-router-governance.md) |
| the names callers may send | `route list/get/create/rename/enable/disable/delete/add/remove`, `default show/enable/disable` | [names, defaults and access control](references/olares-router-governance.md) |
| access control | `key issue/list/update/revoke/local`, `quota set/list/clear`, `app installed`, `user list` | [names, defaults and access control](references/olares-router-governance.md) |
| what happened | `usage summary/list/export`, `audit list/get`, `trace list/get/capture` | [usage, audit and traces](references/olares-router-usage.md) |
| calling a model | `call chat/embed/transcribe/speak/ocr` | [calling a model](references/olares-router-calling.md) |
| inside one application | `local status/progress/spec/config/endpoints/gpu/perf/retry/restart` | [the Model Console](references/olares-router-console.md) |
| a call or a model that does not work | any of the above | [deciding which layer is wrong](references/olares-router-diagnosis.md) |

## Two planes, two credentials

Read [architecture and identity](references/olares-router-architecture.md) before the first write. In short:

- **Management** (`provider`, `app`, `key`, `quota`, `route`, `default`, `usage`, `audit`, `local`) travels on the active profile. Olares injects the identity; nothing has to be supplied.
- **Calling** (`router call`) needs a data-plane credential of its own. The CLI tries the platform's own identity first and mints an `sk-` key only if that is refused, keeping it in the keychain; `router key local` shows or forgets it.
- Most of the management plane is admin-only, reads included: providers, the vendor catalog's models, `capabilities`, market installs, quotas, users and audit all refuse a non-admin. What a non-admin can do is `router list`, `route list/get`, `default show`, `app installed`, their own keys, their own usage, their own traces, and `router call`. Reading the names and what is installed is deliberately open — a name is what a person types into their client, and an install is not a secret.

## Which layer owns the change

| Intent | Where it belongs |
|---|---|
| Use a vendor's hosted models | `provider create` + `provider models import`, or `provider sync-models` for an endpoint that publishes its own list |
| Run a model on this machine | `app install`, which installs the application **and** creates its provider |
| Change what a local model serves or how it is launched | `local spec set` — the model card inside the application, not the Router row |
| Change the address, credentials, or enabled state Router routes with | `provider update` |
| Repair an install that failed | `provider get <app>` then `local progress` / `local retry` |
| Stop, resume, or bind a model application to a GPU | [`olares-market`](../olares-market/SKILL.md) and [`olares-settings`](../olares-settings/SKILL.md) — Router does not own those |

A provider whose `source` is `olares` belongs to a Market application. Its address and lifecycle are the Market's; `provider delete` refuses it, and `app uninstall` is the way out.

## Naming

- A model is addressed as `<provider>/<model>` wherever ambiguity is possible — in `--model`, in a quota, in a key's allowed list. `router list` prints both halves. A name without a slash is a **route** — an alias, a group, or a `default-*` category — and has to exist.
- Every locally installed model application is a provider named `Olares`, so the qualified name is not unique for local models. `router list` shows the application in `SERVED BY`, and the model id is the only handle that always names one row.
- A provider is named by its title, its Olares app name, or its id. A model application is named by its Olares app id (`llamacppqwen3627bggufv3`), which is also what `provider register` and every `local` verb accept.
- An application that *calls* Router has no row here at all: Olares vouches for it at the edge and the call arrives carrying an `appid` — the app name hashed, or the name itself for a system app. So it cannot be registered or revoked; `app installed` says whether it is here, `usage --by caller_app` says what it spent, and `quota set --caller-app` is the only lever over it.

## Safety and escalation

- A named configuration request authorises the loop it implies: creating a provider, importing its models and validating it do not need re-confirmation one by one.
- Ask again before `provider delete`, `app uninstall`, `key revoke`, `quota clear`, `route delete` or `default disable` on something the user did not name — each one breaks callers that still depend on it. `route disable` and `default disable` are reversible and keep their membership; `route delete` gives the name up.
- **Never** put a credential in a shell argument where a file or stdin will do; `--credentials-json` reads either. Never print a plaintext `sk-` key into a transcript that will be shared: `key issue` shows it once, on purpose.
- Traces can contain whole prompts. `trace capture` is the switch; treat what it returns as the user's data even when the profile is an admin's.
- A model that is configured but does not answer is a diagnosis, not a configuration change. Route through [deciding which layer is wrong](references/olares-router-diagnosis.md) before editing anything.
- Stop for the shared auth gate on a persistent authentication failure, and stop when the target provider, model, application or user is ambiguous.
