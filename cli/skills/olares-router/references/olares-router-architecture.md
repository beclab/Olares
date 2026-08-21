# Architecture and identity

Read this before the first write. It explains what the two systems own, how each is addressed, and which credential reaches which one.

## Router

One AI gateway per Olares, installed from the Market. It holds:

- **providers** — an upstream it can route to: a cloud account, or a model application on this machine.
- **models** — rows attached to a provider, each declaring a mode (`chat`, `embedding`, `rerank`, `moderation`, `audio`, `translate`, `image_generation`, `video_generation`, `search`, `scrape`, `responses`, `ocr`) and what it supports. The vendor catalog names the same families differently (`llm`, `text-embedding`, `speech2text`, `tts`); the mode on the row is Router's. For a locally installed model the row is a projection of the application's own model card, which `router model spec` reads and edits.
- **keys, quotas, defaults** — who may call what, up to what ceiling, and which model answers when a request names none.
- **the record** — one usage row per call, an audit row per management change, and optionally the spans an agent framework reported.

Router runs no model. Every call it accepts is forwarded to a provider, and every failure it reports either happened inside Router or came back from that provider.

## The Model Console

A locally installed model application does not simply "contain a model". Inside it, a Model Console:

- resolves the model source and downloads the weights, with progress and a verification step;
- launches the engine — llama.cpp, vLLM, SGLang, Ollama, an embedding server, an audio engine, an OCR adapter — with flags taken from a **model card** it stores;
- serves an OpenAI-compatible endpoint, which is what Router's provider row points at;
- answers `/api/*` for its own lifecycle: phase, progress, effective configuration, GPU residency, a performance probe, retry, engine restart.

The `router model` verbs that read a lifecycle — `status`, `progress`, `retry`, `restart`, `spec`, `diag` — address that console directly. They are the only way to see why a local model is not answering yet, because Router sees a provider that is simply unreachable.

## Addressing

Router is a Market application, not a system service, so it has no fixed subdomain the way `files.<olares-id>` does — app-service gives each install its own host. Every verb resolves the entrance at runtime from the installed-app list, where Router is the `router` application, and names that entrance in any error it reports from the far end.

A model application is reached the same way, by its own entrance, which is why those verbs accept `--app <app_id>` as well as a model name.

Inside the cluster, Router addresses a local model application at its shared entrance rather than its user-facing one, so a provider's base URL is not a URL a browser can open.

## Identity

Two planes, and since Router v2.2.1 the same identity reaches both.

| Plane | Routes | Credential |
|---|---|---|
| Management | `/console/api/*` — everything except `router call` | The active profile. Olares injects the user identity at its edge; nothing is supplied by hand. |
| Data | `/v1/*` — `router call` | Three, tried in order: an `sk-` Bearer, the calling application's identity inside the cluster, then the same profile identity the edge injects. A key is optional. |

`router call` sends the key named by `--api-key` or `OLARES_ROUTER_API_KEY`, and otherwise no `Authorization` at all, letting the edge speak for the caller. `router key current` says which of the two it would be. Nothing issues a key on its own — a machine that used an older olares-cli has one saved in its keychain that calls no longer present, and it stays valid in Router until `router key revoke` ends it.

The identity a call arrives with is also the anchor Router files a stored response or media generation under, so a job created with a key is a 404 to a later keyless call. `--no-wait` and the `--id` that collects the result belong under the same one.

## Roles

Router is installed as an admin-only application. Two consequences:

- A non-admin profile cannot see Router's entrance at all, so every verb reports that Router is not installed. `olares-cli profile whoami` is what distinguishes "not installed" from "not visible to you": on a non-admin profile the second is the likelier reading of the same message.
- Within Router, the role is the Olares role: Router reads it from the identity the edge injects rather than keeping one of its own, so `olares-cli profile whoami` answers for both and there is nothing here that could disagree with it. Most of the management plane requires `admin`, and not only for writes: providers, the vendor catalog's models, quotas, audit and the usage retention window all refuse a non-admin read. What a non-admin identity is left with is the model list, the routes and default categories, their own keys, their own usage, and calling.

  That second layer is real but it is not reachable from here. Because the entrance is admin-only, a non-admin never gets far enough to be told `forbidden_admin_required` — the request does not arrive. The identities that do meet Router's RBAC are the ones the platform routes internally: applications calling with `x-caller-appid` and users whose `X-BFL-USER` the edge injects. Treat the paragraph above as a description of Router, and the bullet above it as the description of this CLI.

A Router user is not an Olares user: the row appears the first time that person's identity reaches Router, which is why a freshly created Olares account is unknown to Router until it makes a call. The people themselves are [`olares-settings`](../../olares-settings/SKILL.md)'s `settings users list`; `key issue --for-user` and `quota set --user` name one of the arrived ones, and refuse anybody else with the names Router does hold.

## Reading an error

Three different systems can answer, and their envelopes differ:

- **Router's own refusal** carries a code — `provider_not_found`, `quota_exceeded`, `invalid_api_key`, `market_app_not_found`. The CLI turns the ones with a known remedy into a sentence naming it.
- **An upstream's refusal** is passed through unchanged. An `invalid_api_key` with type `authentication_error` is Router refusing your key; the same message without it is usually the *vendor* refusing Router's credential, and `provider validate` settles which.
- **An empty body** with a 5xx status was written by a proxy in front of Router — normally the model application is stopped or still starting, not Router failing.

A 404 on a whole subtree means a feature is switched off rather than a missing row: the Market proxy for `app`, or a route that arrived in a later Model Console version for `local`.
