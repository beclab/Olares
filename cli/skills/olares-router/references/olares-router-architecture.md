# Architecture and identity

Read this before the first write. It explains what the two systems own, how each is addressed, and which credential reaches which one.

## Router

One AI gateway per Olares, installed from the Market. It holds:

- **providers** — an upstream it can route to: a cloud account, or a model application on this machine.
- **models** — rows attached to a provider, each declaring a mode (`chat`, `completion`, `embedding`, `rerank`, `moderation`, `audio`, `translate`, `image_generation`, `responses`, `ocr`) and what it supports. The vendor catalog names the same families differently (`llm`, `text-embedding`, `speech2text`, `tts`); the mode on the row is Router's.
- **keys, quotas, defaults** — who may call what, up to what ceiling, and which model answers when a request names none.
- **the record** — one usage row per call, an audit row per management change, and optionally the spans an agent framework reported.

Router runs no model. Every call it accepts is forwarded to a provider, and every failure it reports either happened inside Router or came back from that provider.

## The Model Console

A locally installed model application does not simply "contain a model". Inside it, a Model Console:

- resolves the model source and downloads the weights, with progress and a verification step;
- launches the engine — llama.cpp, vLLM, SGLang, Ollama, an embedding server, an audio engine, an OCR adapter — with flags taken from a **model card** it stores;
- serves an OpenAI-compatible endpoint, which is what Router's provider row points at;
- answers `/api/*` for its own lifecycle: phase, progress, effective configuration, GPU residency, a performance probe, retry, engine restart.

`router local` addresses that console directly. It is the only way to see why a local model is not answering yet, because Router sees a provider that is simply unreachable.

## Addressing

Router is a Market application, not a system service, so it has no fixed subdomain the way `files.<olares-id>` does — app-service gives each install its own host. Every verb resolves the entrance at runtime from the installed-app list, preferring the newer `router` listing over `llmgatewayv3`. `router status` prints which id and which entrance it settled on.

A model application is reached the same way, by its own entrance, which is why `router local` accepts an Olares app id.

Inside the cluster, Router addresses a local model application at its shared entrance rather than its user-facing one, so a provider's base URL is not a URL a browser can open.

## Identity

Two planes, two credentials, and they are not interchangeable.

| Plane | Routes | Credential |
|---|---|---|
| Management | `/console/api/*` — everything except `router call` | The active profile. Olares injects the user identity at its edge; nothing is supplied by hand. |
| Data | `/v1/*` — `router call` | An `sk-` key, or the platform's own application identity when the caller runs inside the cluster. The profile's session alone is **not** accepted here. |

`router call` resolves that second credential itself, in order: an explicit `--api-key`, the key this machine already saved, a keyless attempt when running inside the cluster, and finally minting a key named after the host and storing it in the keychain. `router key local` shows or forgets it; `router key list` shows it alongside every other key, because it is an ordinary key.

## Roles

Router is installed as an admin-only application. Two consequences:

- A non-admin profile cannot see Router's entrance at all, so every verb reports that Router is not installed. `router status` is the check that distinguishes "not installed" from "not visible to you".
- Within Router, `whoami` reports the role. Most of the management plane requires `admin`, and not only for writes: providers, the vendor catalog's models, market installs, quotas, users, callers and audit all refuse a non-admin read. What is left to a non-admin is the model list, the resolved defaults and their own override, their own keys, their own usage, their own traces, and calling.

A Router user is not an Olares user: the row appears the first time that person's identity reaches Router. `router user list` shows the ones it knows, which is why a freshly created Olares account is absent until it makes a call.

## Reading an error

Three different systems can answer, and their envelopes differ:

- **Router's own refusal** carries a code — `provider_not_found`, `quota_exceeded`, `invalid_api_key`, `market_app_not_found`. The CLI turns the ones with a known remedy into a sentence naming it.
- **An upstream's refusal** is passed through unchanged. An `invalid_api_key` with type `authentication_error` is Router refusing your key; the same message without it is usually the *vendor* refusing Router's credential, and `provider validate` settles which.
- **An empty body** with a 5xx status was written by a proxy in front of Router — normally the model application is stopped or still starting, not Router failing.

A 404 on a whole subtree means a feature is switched off rather than a missing row: observability for `trace`, the Market proxy for `app`, or a route that arrived in a later Model Console version for `local`.
