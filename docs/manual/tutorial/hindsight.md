---
outline: [2, 3]
description: Configure Hindsight on Olares, connect your agents, and store, retrieve, and reason over persistent memories.
---

# Using Hindsight on Olares

Use this guide to connect an agent to an existing Hindsight installation on Olares. It covers Hindsight 0.10.2 with the Olares Chart 0.0.16 configuration.

Hindsight gives your agent a persistent memory service. A bank groups memories under one ID. Your agent can retain experiences, recall relevant facts, and reflect over those memories to answer questions across conversations.

:::warning Before connecting an agent
Configure working chat and embedding models in Olares Router first. Hindsight checks model connectivity during startup and can fail to start when the embedding endpoint fails.
:::

## How the services connect

Hindsight uses two separate connections:

| Connection | Purpose | URL format |
| --- | --- | --- |
| Agent → Hindsight | Store and query memories | Your installation's Hindsight API entrance URL |
| Hindsight → Router | Run chat and embedding inference | `https://router.<your-olares-domain>/v1` |

Do not enter the Router URL as your agent's Hindsight API URL. Do not append `/v1` to the Hindsight API entrance URL.

The web control plane lets you inspect banks and memory data. The API serves agents. Olares system PostgreSQL stores the data in a dedicated database with the `vector` extension. The package does not deploy a separate PostgreSQL server or include model weights.

## Preparing the installation

1. Open Router and check that `default-chat` points to your chat model and `default-embedding` points to your embedding model.
2. Test both models with a short input. A model application marked Running does not prove that inference works.
3. Install Hindsight from the source available to you in Market. While the listing is under testing, use the uploaded Chart in Local Sources → Upload.
4. Configure the installation inputs below. Keep the default values when using Router's default routes.

| Input | Default behavior | When to change it |
| --- | --- | --- |
| `HINDSIGHT_API_LLM_BASE_URL` | Leave empty to generate `https://router.<your-olares-domain>/v1` | Use another OpenAI-compatible endpoint serving chat and embeddings |
| `HINDSIGHT_API_LLM_MODEL` | `default-chat` | Use a specific chat model or route |
| `HINDSIGHT_API_EMBEDDINGS_MODEL` | `default-embedding` | Use a specific embedding model or route |
| `HINDSIGHT_API_LLM_API_KEY` | Leave empty to use the non-empty placeholder `not-required` | Supply a valid key when Router or the provider checks API keys |

The placeholder satisfies Hindsight's non-empty key check. It does not authenticate with a provider that requires a valid key. The same configured key and base URL are used for chat and embeddings.

5. Wait for Hindsight to reach Running. Open its web control plane.
6. Check the API health endpoint from the agent's environment:

   ```bash
   curl -fsS "${HINDSIGHT_API_URL}/health"
   ```

   Set `HINDSIGHT_API_URL` to the actual API entrance URL first. A successful response contains `"status":"healthy"` and `"database":"connected"`.

:::tip Existing configuration
An upgrade may preserve a previously entered base URL. Clear `HINDSIGHT_API_LLM_BASE_URL` and apply the change if you want to use the dynamic Router default.
:::

## Getting the API URL and choosing a bank

1. Open Hindsight's application settings in Olares Settings and locate its API entrance. The web control-plane entrance and API entrance are different.
2. Copy the API entrance URL. Keep its HTTPS scheme and remove any trailing path such as `/health` before using it as a client base URL.
3. Check the entrance access policy for the agent you want to connect. The Chart declares the API entrance internal by default. Test `/health` from that agent; browser access alone is not enough.
4. Create a bank in the Hindsight control plane and copy its ID. Use a name such as `hermes-personal` or `opencode-project-a`.

Use separate banks for unrelated projects or users. Agents that use the same bank can share its memories. A bank ID organizes data; it does not replace access controls.

## Connecting Hermes

Run these commands inside the Hermes environment. The Hermes image used in this setup already includes `hindsight-client` 0.10.2. The client package and the Hermes memory-provider plugin are separate components.

1. If the Hindsight provider is missing, install it from the Hermes catalog:

   ```bash
   hermes plugins install hindsight
   ```

2. Start the memory setup wizard:

   ```bash
   hermes memory setup
   ```

3. Select Hindsight, then Local External (`local_external`). Enter the Hindsight API URL and your bank ID. Supply an API token only if the Hindsight endpoint requires one.
4. If offered a starter memory template, choose one that fits your agent or select Blank. Avoid replacing an existing bank's settings unintentionally.
5. Start a new Hermes conversation:

   ```bash
   hermes
   ```

Local External uses the server you installed on Olares. Local Embedded (`local_embedded`) starts another server and requires additional packages such as `hindsight-all`.

### Using memory only when requested

To reduce per-turn model work, use the Hindsight provider settings on the Plugins page first.

Using the UI:

1. Open Plugins in the Hermes Web UI and check that the active profile is the one you want to configure.
2. In Runtime Provider Plugins, check that Memory Provider is `hindsight` and its status is ready / active. Keep Mode, API URL and Bank ID unchanged.
3. Find Auto Recall in the Hindsight settings and turn it off. This stops automatic memory retrieval before each conversation turn.
4. Find Auto Retain and turn it off to stop automatic conversation retention. If your plugin version does not show Auto Retain, use the file method below to set `auto_retain` to `false`; do not use Retain Indicator as a substitute.
5. Click Save, refresh the page to check the saved switch states, then start a new Hermes session.

:::info Automatic calls and status indicators
Recall Indicator and Retain Indicator only control the display of recalled-memory and “saving to memory” messages. Turning an indicator off does not stop memory reads or writes. Recall Sync controls whether automatic recall runs synchronously; it does not disable automatic recall.

Memory Enabled and User Profile Enabled on the general Memory page control Hermes' built-in memory, not Hindsight's automatic retention or recall.
:::

If your UI does not expose the automatic-memory switches, edit the configuration file directly. You can also access it through Files, download a backup, edit it on your computer, and upload it to the same directory. Uploading a file with the same name overwrites the original; keep other fields and check the directory first. If Files cannot access the configuration directory, use the direct-edit method.

Editing the file directly: open `~/.hermes/hindsight/config.json` in a standard Hermes environment, or `/opt/data/hindsight/config.json` with this Olares Chart's default configuration. Merge these fields into the existing file:

```json
{
  "auto_retain": false,
  "auto_recall": false
}
```

Keep the existing connection and bank settings, then start a new Hermes session. These options belong to the Hindsight provider; the general Memory switches control Hermes' built-in memory.

Explicit memory tools remain available when the provider is active and the tools are enabled. Ask for the tool you want to use:

| Goal | Example prompt |
| --- | --- |
| Store a fact | Use `hindsight_retain` to save this fact: I develop and deploy applications on Olares. |
| Retrieve a fact | Use `hindsight_recall` to find which platform I use to develop and deploy applications. |
| Analyze past decisions | Use `hindsight_reflect` to compare the recorded embedding configurations and the reasons for changing them. Separate current facts, historical facts, and items that still need checking. |

Check the tool output and inspect the bank. A chat reply such as “saved” does not prove that memory processing finished. For a reliable test, start another conversation and recall the fact from the same bank.

Recent Hermes plugin versions can limit recall to consolidated observations. If a newly retained fact is not found while consolidation is pending, check the provider's `recall_types` setting. The official plugin supports `"observation,world,experience"` to include raw facts as well.

## Connecting other agents

The following are upstream-supported integrations. Hermes has been exercised in this Olares setup; OpenClaw and OpenCode still need testing in your own environment.

### OpenClaw

1. Install the plugin and start its setup wizard in the OpenClaw environment:

   ```bash
   openclaw plugins install @vectorize-io/hindsight-openclaw
   npx --package @vectorize-io/hindsight-openclaw hindsight-openclaw-setup
   ```

2. Select External API, then enter your Hindsight API URL and the optional Hindsight API token.
3. Set a bank ID if you want a specific shared or isolated bank. Start or restart your OpenClaw gateway after configuration.

This plugin occupies OpenClaw's memory slot. Its automatic memory behavior is configured separately from Hermes. See the [official OpenClaw integration](https://github.com/vectorize-io/hindsight/blob/v0.10.2/hindsight-integrations/openclaw/README.md) for version requirements and settings.

### OpenCode

Add the official plugin to the existing `opencode.json` configuration:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["@vectorize-io/opencode-hindsight"]
}
```

Merge it with your existing plugins. Before launching OpenCode from a Linux shell, configure the self-hosted server:

```bash
export HINDSIGHT_API_URL="https://your-hindsight-api-host"
export HINDSIGHT_BANK_ID="opencode-project-a"
opencode
```

Replace the example host with the actual API entrance URL. Set `HINDSIGHT_API_TOKEN` only if that endpoint requires a token. Without an API URL override, this plugin defaults to Hindsight Cloud.

The plugin offers retain, recall, reflect and automatic memory hooks. For a broader set of coding harnesses, see the [official integrations hub](https://hindsight.vectorize.io/integrations). Client plugins have their own versions and configuration.

## Troubleshooting

### Startup reports that an LLM API key is required

The process received an empty key. Chart 0.0.16 supplies `not-required` when the installation field is empty. Check that your installed Chart includes this behavior. For an authenticated endpoint, enter its valid API key instead.

### Startup fails with `upstream_circuit_open`

Router has temporarily stopped forwarding requests after upstream failures. Check which model serves `default-embedding`, test it with a short text, and inspect that model application's logs.

In the tested environment, EmbeddingGemma reported `infer failed on device GPU.0: general error`. That inference failure triggered Router's circuit breaker and caused Hindsight's embedding startup check to fail. Restore working model inference before retrying Hindsight. If the GPU error persists, test the same embedding model on CPU. Increasing Hindsight's resources does not fix a separate model service.

### “Saving to memory” remains after a reply

Retain can run asynchronously and includes extraction, embeddings and background processing. Check the bank and the operation result before deciding it is stuck. Inspect the Hindsight and model-service logs if saving repeatedly fails. Turn off automatic retain if you want explicit writes only.

### Retain or reflect is slow

These operations use model inference. Check whether chat requests and background memory processing compete for a low-concurrency model route. Start with short facts and manual calls. Avoid repeatedly submitting the same long transcript while processing is still pending.

The API container has limits of 1 CPU core and 2Gi memory; the control plane has 0.5 core and 512Mi. Increase them when measurements show resource pressure. Chat models, embedding models and system PostgreSQL have separate resource allocations.

## Further reading

- [Hindsight 0.10.2 release](https://github.com/vectorize-io/hindsight/releases/tag/v0.10.2)
- [Hermes integration and provider settings](https://github.com/vectorize-io/hindsight/blob/v0.10.2/hindsight-integrations/hermes/README.md)
- [OpenCode integration](https://github.com/vectorize-io/hindsight/blob/v0.10.2/hindsight-integrations/opencode/README.md)
- [Olares documentation](https://docs.olares.com)
