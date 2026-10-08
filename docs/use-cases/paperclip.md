---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/paperclip
outline: [2, 3]
description: Run Paperclip on Olares to coordinate AI agents backed by Claude Code, Codex, OpenCode, Cursor, and other providers through shared issues.
head:
  - - meta
    - name: keywords
      content: Olares, Paperclip, AI agent, multi-agent, Claude Code, Codex, OpenCode, Cursor, self-hosted
app_version: "1.0.34"
doc_version: "1.2"
doc_updated: "2026-10-08"
---

# Coordinate multiple AI agents with Paperclip

<VersionRouteSelect />

Paperclip is an open-source platform for coordinating multiple AI agents under one unified workspace. By setting up a virtual company, you add AI agents powered by Claude Code, Codex, OpenCode, Cursor, or other providers, and assign them issues to work on. Whether the task involves coding, research, or content creation, Paperclip manages the workflow.

Running Paperclip as a self-hosted app on Olares ensures that your API keys, task history, and agent outputs remain entirely private on your device.

## Prerequisites

Before you begin, you need:

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- Qwen3.8-27B (llama.cpp) installed from Market, if you plan to use a local model.

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Learning objectives

In this guide, you will learn how to:

- Install Paperclip on Olares.
- Configure API keys for the agents you plan to use.
- Set up the initial admin user account.
- Set up your first company, agent, and task.
- Create an issue and track agent progress.
- Monitor operations and metrics from the dashboard.
- Optional: Use local models with Paperclip.

## Install Paperclip

1. Open Market and search for "Paperclip".

   ![Paperclip in Market](/images/manual/use-cases/paperclip.png#bordered)

2. Click **Get**, then click **Install**. When the installation finishes, two shortcuts appear in the Launchpad:

   ![Paperclip entry points](/images/manual/use-cases/paperclip-install-entry.png#bordered){width=35%}

## Initialize Paperclip

After installing Paperclip, create the first admin account to complete the initial setup. If you plan to use cloud models, configure your API keys first.

### Configure API keys

Cloud-model authentication depends on the provider. The following environment variables apply to adapters without an AI connection binding. Agents using AI Connections authenticate through their selected connection instead. Local models do not inherently require a cloud-model API key.

:::tip
Before creating an admin account, set up your API key in **Settings** so cloud models can be used immediately.
:::

1. Open Settings, then go to **Applications** > **Paperclip** > **Manage environment variables**.

   ![Manage Paperclip environment variables](/images/manual/use-cases/paperclip-manage-env-vars.png#bordered){width=70%}

2. Click <i class="material-symbols-outlined">edit_square</i> next to a variable, enter your API key in the **Value** field, then click **Confirm**.

   Paperclip supports the following variables:

   | Variable | Used by |
   |:---------|:--------|
   | `ANTHROPIC_API_KEY` | Claude Code, OpenCode, Pi |
   | `OPENAI_API_KEY` | Codex, OpenCode, Pi |
   | `GEMINI_API_KEY` | Gemini CLI, Cursor |
   | `CURSOR_API_KEY` | Cursor |

3. Click **Apply** to save and apply the new keys.

:::tip Add more API keys later
Return to this section to add or update keys at any time. Repeat this procedure and restart Paperclip to apply new configurations.
:::

### Create an admin account

Paperclip ships without a default user account. To access the platform for the first time, you need to create an admin account through the registration flow.

1. Open Paperclip from the Launchpad, and click **Create account**.

2. On the sign-up page, fill in the required information and submit.

3. After sign-up, click **Claim this instance** to bind the admin account.

   ![Claim this instance](/images/manual/use-cases/paperclip-claim-instance.png#bordered){width=60%}

4. Name your company. For cloud models, finish onboarding as instructed. For local models, continue with [Use local models in Paperclip](#use-local-models-in-paperclip).

## Complete onboarding

When you create your first organization, Paperclip opens onboarding to help you set up the organization and its first agent.

1. Enter the organization name and follow the prompts to describe its goal.
2. Name your first agent.
3. Select a model provider. The default onboarding currently **supports only OpenAI and Claude**. Follow the prompts to connect the corresponding account or provide an API key, select a model, and verify the connection.
4. Complete the guided agent setup to enter your organization's workspace.

### Create your first agent with another adapter

To use another adapter, such as **OpenCode or Hermes**, first create the organization. Then manually enter this URL in your browser's address bar:

```text
https://<your-paperclip-domain>/agents/all
```

From the agent list, click the button to create an agent, enter its name, select an adapter, and complete its runtime configuration. Avoid returning to Dashboard while the organization has no agents because it can reopen onboarding.

This entry point provides the full adapter selector, but does not remove OpenCode's OpenRouter connection restriction. To use OpenCode with a local model, continue with [Use local models in Paperclip](#use-local-models-in-paperclip) and create an unbound agent through the CLI.

## Create and track issues

In Paperclip, all work happens through issues. When you create an issue, Paperclip assigns it to an existing agent. If no suitable agent exists for the specific request, Paperclip automatically "hires" a new one to complete the job.

1. On the **Issues** page, click **New issue** in the left sidebar.
2. Specify details for the issue. For example:

   - **Issue title**: Write a guide on AI servers.
   - **Description**: Hire a writing agent to research and draft a 200-word article explaining the benefits of self-hosting AI models.
   - **Assignee**: Select **CEO** to evaluate the requirements.

   ![New issue dialog](/images/manual/use-cases/paperclip-new-issue.png#bordered){width=60%}

3. Click **Create Issue**. Paperclip assigns the issue and the agent starts working.
4. Click **Inbox** in the left sidebar to monitor incoming requests and execution progress. If the assigned agent decides to delegate the task, a hiring request appears in your inbox.

   ![Inbox tracking the new issue](/images/manual/use-cases/paperclip-inbox.png#bordered)

5. Go to the **Agents** section in the left sidebar to see the newly hired writer agent and its details.

   ![Writer Agent details](/images/manual/use-cases/paperclip-agent-details.png#bordered)

6. Review the agent's output:

   a. Go to the **Issues** page, then select the issue to view its details. Look for the output directly in the chat history.

   ![Agent output](/images/manual/use-cases/paperclip-agent-output.png#bordered)

   b. If the output is missing from the chat, enter a comment in the issue asking the agent for the file path. Open the Olares Files app, then go to the specified directory to retrieve your document.

   ![Agent output in Files app](/images/manual/use-cases/paperclip-agent-output-in-files.png#bordered)

## Monitor operations from the dashboard

As your agents complete issues, use the dashboard to track your company's overall performance, monitor API costs, and identify potential bottlenecks.

1. Click **Dashboard** in the left sidebar.

   ![Paperclip dashboard](/images/manual/use-cases/paperclip-dashboard.png#bordered)

2. Review the top agent cards to see the most recent tasks your agents worked on and their execution duration.
3. Check the high-level metrics to monitor operational health, such as the total API cost incurred for the current month.
4. Analyze the 14-day trend charts to spot performance changes:

   - **Run Activity**: Check the total number of agent executions.
   - **Issues by Priority**: View active issues grouped by urgency (critical, high, medium, and low).
   - **Issues by Status**: Track progress by reviewing which issues are in progress, completed, or blocked.
   - **Success Rate**: Monitor the execution success rate of your AI workforce.

5. Scroll to the activity log to audit recent system events and agent behaviors. This log provides a chronological trail of all recent tasks.

## Use local models in Paperclip

:::warning Follow the local-model setup below
The latest Paperclip agent creation form connects **OpenCode to OpenRouter by default**, instead of directly using the local-model settings in your existing OpenCode configuration file. For an agent bound to an OpenRouter AI connection, Paperclip uses an isolated configuration directory, injects OpenRouter credentials, and passes the agent's selected model explicitly, overriding the file's default model. **The previous approach of only editing `opencode.json` or `opencode.jsonc` no longer works for these agents.** Follow the procedure below to configure OpenCode and create an unbound agent.

This does not make all configuration files obsolete: OpenCode agents without an AI connection binding can still use them. Manually entering `olares/default-chat` does not remove an OpenRouter binding. **Environment: Local** only means the agent executes locally; it does not select a local model.

This behavior is confirmed in the upstream [agent creation form](https://github.com/paperclipai/paperclip/blob/v2026.1005.0/ui/src/components/new-agent/NewAgentSetup.tsx), [connection runtime](https://github.com/paperclipai/paperclip/blob/v2026.1005.0/server/src/services/ai-connection-runtime.ts), and [OpenCode execution code](https://github.com/paperclipai/paperclip/blob/v2026.1005.0/packages/adapters/opencode-local/src/server/execute.ts).
:::

:::warning Use local models with caution
Paperclip is a fully autonomous multi-agent collaboration platform. Running it entirely on local models can cause workflow disruptions due to model capability limits, context overflow, or concurrency restrictions, potentially leading to cascading failures.

Consider using a hybrid configuration: assign high-performance cloud models to critical roles such as CEO or CTO, while using local models for execution-focused agents to reduce costs. Alternatively, designate local models as `cheap model` for less demanding tasks.

Carefully assess the capabilities of each model and your workflow requirements to determine the optimal setup.
:::

OpenCode can call an OpenAI-compatible local model through Olares Router. This OpenCode runs inside Paperclip and is separate from the OpenCode app in Market.

### Prepare the model and preserve existing settings

1. Install and start Qwen3.8-27B (llama.cpp) from Market.
2. Configure the model in Router and copy its connection details.

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

3. For the Router default model, use `default-chat`. Ensure Router actually routes that alias to your local model. For a direct model endpoint, use the model ID that endpoint accepts instead.
4. In Files, back up **Data > paperclip > paperclip > .config > opencode** before editing. Also preserve the agent's current model and configuration. Do not delete credentials or application data to reset a model configuration.

### Configure OpenCode

The persistent config path inside Paperclip is `/paperclip/.config/opencode/opencode.json`.

- If `opencode.json` exists, merge the fields below into it, preserving other providers, plugins, MCP servers, and settings.
- If only `opencode.jsonc` exists, back it up and convert a copy to strict JSON, removing comments and trailing commas. Preserve its settings. After conversion, move the original JSONC outside the active config directory so two active files do not conflict.
- If neither exists, create `opencode.json`.

Use this example for Router's default model. Replace `<router-base-url-including-v1>` with the actual API Base URL, including `/v1`, not the model application's web page URL.

```json
{
  "$schema": "https://opencode.ai/config.json",
  "autoupdate": false,
  "model": "olares/default-chat",
  "small_model": "olares/default-chat",
  "provider": {
    "olares": {
      "name": "Olares local model",
      "npm": "@ai-sdk/openai-compatible",
      "models": {
        "default-chat": {
          "name": "Local model via Router"
        }
      },
      "options": {
        "baseURL": "<router-base-url-including-v1>"
      }
    }
  }
}
```

Follow the Router connection instructions for authentication. If your endpoint requires a key, configure that endpoint's key; an OpenAI or OpenRouter cloud key is not a substitute. Do not publish credential files in support requests.

`model` selects the default OpenCode model. `small_model` also routes OpenCode helper requests to the local model. Paperclip's separate **cheap model** setting, if enabled, must also use an available local model or be disabled. `autoupdate: false` disables OpenCode's in-app update attempts; update the bundled CLI through the Paperclip Market image instead.

OpenCode supports JSONC, but Paperclip's temporary runtime-config preparation reads `opencode.json` with a strict JSON parser. Using valid JSON avoids settings being ignored during agent runs.

:::tip Environment-variable alternative
Paperclip also reads `PAPERCLIP_OPENCODE_PROVIDERS` (a JSON object containing the entries inside `provider`, not the entire config) and `PAPERCLIP_OPENCODE_SMALL_MODEL`. These can supply runtime configuration instead of editing a file. This does not remove an OpenRouter AI connection binding. Only use Olares Settings for variables actually exposed by the installed chart; this guide does not assume these variables are available there.
:::

### Create an unbound OpenCode agent

:::info Validation scope
The procedure below uses the bundled CLI to create an unbound agent. Local inference through this path has not yet been validated end to end for this guide. Complete the validation steps before assigning production tasks.
:::

1. Register and obtain instance-admin or organization access, then create an organization. You do not need a cloud-model key to configure OpenCode itself.
2. After the organization exists, manually open `/agents/all` on your Paperclip domain to leave onboarding and access agent creation. For a local OpenCode model, use the CLI procedure below instead of completing the OpenRouter-bound form.
3. Use the bundled Paperclip CLI as your human administrator to create an agent without `runtimeConfig.aiConnection`.

Open **Paperclip CLI** from Launchpad. Replace the URL with your Paperclip application origin, without `/api` or a dashboard path:

```bash
export PAPERCLIP_LOCAL_API='https://your-paperclip-domain'
paperclipai auth login --api-base "$PAPERCLIP_LOCAL_API" --no-browser
```

Open the approval URL printed by the CLI in your browser and approve access using your own Paperclip account. This authenticates the CLI to Paperclip, not to a model provider. Then list organizations:

```bash
paperclipai company list --api-base "$PAPERCLIP_LOCAL_API" --json
```

Copy the intended organization's UUID from `id`, not its short prefix, and create one agent:

```bash
export PAPERCLIP_LOCAL_COMPANY='<organization-uuid>'
mkdir -p /paperclip/workspaces/local-model
paperclipai agent create \
  --api-base "$PAPERCLIP_LOCAL_API" \
  --company-id "$PAPERCLIP_LOCAL_COMPANY" \
  --payload-json '{
    "name": "Local model agent",
    "role": "general",
    "adapterType": "opencode_local",
    "adapterConfig": {
      "cwd": "/paperclip/workspaces/local-model",
      "model": "olares/default-chat"
    },
    "runtimeConfig": {
      "heartbeat": {
        "enabled": false,
        "maxConcurrentRuns": 1
      }
    },
    "permissions": {
      "canCreateAgents": false
    }
  }' --json
```

Keep the returned agent ID. This creates a new agent; rerunning the command creates another one. The disabled periodic heartbeat and hiring permission keep the initial validation controlled. `maxConcurrentRuns: 1` applies only to this agent, not to all agents or all clients of the model.

Verify the saved configuration:

```bash
paperclipai agent get '<agent-id>' --api-base "$PAPERCLIP_LOCAL_API" --json
```

Confirm `adapterType` is `opencode_local`, `adapterConfig.model` is `olares/default-chat`, and `runtimeConfig.aiConnection` is absent. Open **Agents** in Paperclip and select this agent. Keep its existing/unmanaged authentication configuration; do not adopt an OpenRouter connection. Do not ask a cloud-connected manager agent to perform this setup because agent-created hires may inherit a managed connection.

### Configure an existing agent

If an existing OpenCode agent has no `runtimeConfig.aiConnection`, merge the configuration above and set its model to `olares/default-chat`, preserving its other settings. If it is already bound to OpenRouter, create an unbound agent using the CLI procedure, validate it, and then reassign work. Removing `aiConnection` from an update payload does not detach it: the server preserves the binding.

Paperclip passes the agent's selected model to OpenCode with `--model`. Changing only the default `model` in `opencode.json` does not override an agent still configured with a cloud model.

Use Market to restart Paperclip after changing shared configuration, once active work has stopped. Do not modify `/app`, replace the container image manually, or edit the database to detach a connection.

### Validate and recover

1. In Paperclip CLI, check the bundled version and model discovery:

   ```bash
   opencode --version
   opencode models olares
   ```

   Confirm `olares/default-chat` appears. This checks discovery, not inference.
2. Send one short request explicitly to the local model:

   ```bash
   opencode run --model olares/default-chat 'Reply with OK only. Do not use tools.'
   ```

3. In the unbound agent's configuration, run its environment/connection test. CLI success alone does not prove the agent uses the same credentials, environment, and model.
4. Assign one short task manually and inspect the run log. Repeat five short runs sequentially before enabling scheduled work. Check for successful replies and absence of context, KV-cache, and timeout errors. Keep other model consumers idle during this validation; an agent's concurrency setting is not a model-wide capacity limit.

| Symptom | Next step |
| :--- | :--- |
| “This connection does not support the current harness and model” | Check `runtimeConfig.aiConnection`. Changing the OpenCode file cannot resolve an OpenRouter binding conflict. |
| “Method not allowed” | Check the actual request URL, `/v1` path, and OpenAI-compatible protocol. Model discovery does not validate the chat endpoint. |
| CLI works, but the agent reports “Internal server error” | Compare the agent's model, working directory, connection binding, and environment. Collect the matching server error with secrets redacted; the generic 500 message alone does not identify the cause. |
| OpenCode update fails | Use the Market-provided image update. Do not make image-owned directories writable to repair self-update. |
| Original JSONC was overwritten | Restore a backup if available. Otherwise reconstruct the required configuration; neither restart nor organization export restores the exact original file. |

To undo the change, pause the new agent or stop assigning it work, restore the backed-up OpenCode configuration and the previous agent model, and restart through Market after active runs stop. Keep existing agents until validation succeeds. Organization export is not a full backup of the database, local configuration, credentials, and workspace files; retain those separately before reinstalling.

## FAQs

### Which agent adapters does Paperclip support?

Paperclip currently supports the following agent adapters. You configure specific API keys as environment variables depending on the adapter you choose:

- Claude Code: Requires `ANTHROPIC_API_KEY`.
- Codex: Requires `OPENAI_API_KEY`.
- OpenCode: Authentication depends on the model provider. Local models do not inherently require an Anthropic or OpenAI key. For the OpenRouter-bound UI and local-model setup, see [Use local models in Paperclip](#use-local-models-in-paperclip).
- Pi: Requires `ANTHROPIC_API_KEY` or `OPENAI_API_KEY`.
- Cursor: Requires `CURSOR_API_KEY`.

### Can I use Hermes Agent as an agent runtime?

Yes, but it is not recommended for production environments. Note the following:

1. This Hermes Agent runs inside the Paperclip container and is separate from the Hermes Agent app installed from the Olares Market. You need to install and configure it separately in the Paperclip CLI (`pip install hermes-agent`).
2. The Hermes Agent and Paperclip adapter are still in development. Connection stability issues might require significant debugging.
3. Ensure the following configurations are correct, otherwise the agent might fail to work:

   - Turn off the manual approval setting to prevent Paperclip from timing out while waiting for approval when calling the agent:

   ```bash
   hermes config set approvals.mode "off"
   hermes config set approvals.cron_mode "approve"
   ```

   - Ensure your Hermes Agent has `PAPERCLIP_API_KEY` set and the Paperclip Skills installed so it can fetch and modify tasks in Paperclip.

### Codex adapter fails to authenticate

After you update `OPENAI_API_KEY` and restart Paperclip, the Codex adapter might still fail to authenticate. This happens because the `codex login` command attempts to open a local browser for OAuth, which the background container cannot do.

To resolve this issue, run a manual device-auth login:

1. Open Control Hub, then go to **Browse** > **paperclip-{username}** > **Deployments** > **paperclip**.
2. Under **Pods**, click the pod name to view its containers, then click <i class="material-symbols-outlined">terminal</i> next to the **paperclip** container to open the pod terminal.

   ![Open the Paperclip pod terminal from Control Hub](/images/manual/use-cases/paperclip-enter-container.png#bordered)

3. Type the following command, then press **Enter**:

   ```bash
   codex login --device-auth
   ```

4. Open the device-auth link in your browser and sign in to complete the authorization. Then retry the Codex adapter in Paperclip.

   ![Codex device-auth result](/images/manual/use-cases/paperclip-codex-login-result.png#bordered)

## Learn more

- [Paperclip documentation](https://docs.paperclip.ing): Official docs for concepts, features, and configuration.
- [Orchestrate multi-agent workflows with oh-my-openagent](opencode-omo.md): Run multi-agent collaboration inside a single OpenCode instance on Olares.
- [Set up OpenCode as your AI coding agent](opencode.md): Install and configure OpenCode, a common Paperclip adapter.
