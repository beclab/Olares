# Configuring an external provider

Giving Olares a cloud vendor's models is two decisions, in this order: which upstream, then which of its models are offered. Credentials alone offer nothing — a provider with no model rows routes nothing, deliberately, so that one key does not silently expose a vendor's entire catalog.

Everything here is admin-only, reads included: even listing providers refuses a non-admin, for whom `router model list` is the view of what exists. Only the bare list of vendor kinds is open.

## 1. Choose the upstream kind

`router provider types` lists what Router can speak to, and for each: the provider type to pass to `--type`, how models arrive, which endpoint families it serves, how many predefined models Router already knows, and the credential fields it requires. `router provider types <vendor>` adds every optional field with its default and where to get a key, and `--models` lists the models it can import.

Two shapes matter more than the vendor's name:

- **`predefined-model`** — Router ships the vendor's catalog, so `model import` fills in capabilities and prices for you.
- **`customizable-model`** — the endpoint's catalog is its own. Either `provider sync-models` mirrors its live list, or `model add` names one model and describes it by hand.

`openai_api_compatible` is the generic escape hatch for any OpenAI-shaped endpoint. Its credential fields are declared per model upstream while Router stores them on the provider, which is why the summary row repeats a field once per mode; read `provider types openai_api_compatible` and supply the common ones plus `--base-url`.

## 2. Create it

```
olares-cli router provider create --type openai --name openai-main \
    --base-url https://api.openai.com/v1 \
    --credentials-json ~/secrets/openai.json --validate
```

- `--credentials-json` accepts a file or `-` for stdin, and is preferred over repeated `--credential key=value`, which leaves the secret in shell history.
- `--base-url` is required for every provider, including a vendor Router already has a catalog for. The catalog supplies the models, not the address.
- `--validate` probes the upstream once the row exists. Without it the provider is created unverified, and a wrong key is not discovered until the first call.
- A vendor that genuinely needs no credentials still needs the intent stated: pass `--credentials-json` with `{}`.

Credentials are encrypted at rest and never returned. `provider credentials` shows which fields are stored, with values masked.

## 3. Attach models

```
olares-cli router model import --provider openai-main gpt-4o gpt-4o-mini
olares-cli router provider sync-models ollama-box
olares-cli router model add llama-3.1-70b --provider my-endpoint \
    --mode chat --context-size 131072 --supports supports_vision=true
```

- `import` takes the capability flags, context window and prices from Router's catalog. Prefer it whenever the vendor is predefined.
- `sync-models` mirrors an upstream that publishes `/models` — Ollama, a local model application, most gateways. It is the whole list in one step, so it is also the way to pick up a model added upstream later.
- `add` is for an endpoint whose catalog Router cannot read. `--mode` defaults to `chat`; get it right, because a default is resolved per mode and a mislabelled row is only discovered at call time.
- `--supports` takes `supports_*` flags. Declaring one the model does not have makes a caller send a request the upstream will reject; omitting one it does have makes callers avoid a feature that works. Router stores a key it does not recognise instead of refusing it, so a misspelt flag is kept and never honoured — read a working row with `router model get <model>` and copy its spelling rather than guessing.

`router provider get <provider>` then shows the provider with every model it serves, their modes, sizes and headline capabilities. `router model list` shows every model across every provider, with the same capability summary in its SUPPORTS column.

## 4. Correct it later

| Change | Verb |
|---|---|
| Rename, re-point, disable the whole provider | `provider update --name/--title/--base-url/--status` |
| Rotate a key | `provider update --credentials-json - --credentials-note "why"` |
| Fix what one model claims | `model update <model> --supports ... --pricing ...` |
| Stop offering one model, keeping its row | `model update <model> --disable` |
| Detach a model | `model remove <model>` |
| Remove the provider and everything referencing it | `provider delete <provider>` |

`provider update` merges credentials: a field you omit keeps its stored value, so rotating one secret does not require resending the rest. `model update` merges the model's description the same way; `--replace-description` sends only what is on the command line and discards the rest.

Disabling a provider keeps its rows and its history and stops it answering. Deleting it takes the models, the quotas that referenced them and the keys' allowances with it — prefer `--status disabled` when the intent is reversible.

## Credential history

Every credential write is versioned. `provider history <provider>` lists the versions newest first, with who changed them and the note they left; `provider rollback <provider> <version>` restores one atomically. Values are never shown in either direction — a rollback is chosen by version and note, not by reading the old secret.

`provider validate <provider>` asks the upstream whether what is stored still works, and reports the verdict with the upstream's own message. Run it after a rotation and after a rollback; it is the only check that distinguishes a wrong credential from an unreachable endpoint.

## What does not belong here

A provider whose source is `olares` was created by installing a model application: see [local LLM applications](olares-router-local-llm.md). `provider delete` refuses it, `provider update` cannot move its address, and its credentials are the platform's. `provider register` is the repair path when such an application exists but its Router row does not.
