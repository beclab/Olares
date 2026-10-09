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

Each agent requires an API key for its underlying model provider. You configure these keys as environment variables in the Settings app.

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

4. Enter your company name and complete the onboarding process as prompted.

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

From the agent list, click the button to create an agent, enter its name, select an adapter, and complete its runtime configuration. 

Currently, Paperclip restricts the model provider for OpenCode to OpenRouter. To use OpenCode with a local model instead, follow the steps in [Use local models in Paperclip](#use-local-models-in-paperclip) and create an unbound agent using the CLI.

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

:::warning Use local models with caution
Paperclip is a fully autonomous multi-agent collaboration platform. Running it entirely on local models can cause workflow disruptions due to model capability limits, context overflow, or concurrency restrictions, potentially leading to cascading failures.

Consider using a hybrid configuration: assign high-performance cloud models to critical roles such as CEO or CTO, while using local models for execution-focused agents to reduce costs. Alternatively, designate local models as `cheap model` for less demanding tasks.

Carefully assess the capabilities of each model and your workflow requirements to determine the optimal setup.
:::

:::warning `opencode.json` is no longer supported in the latest Paperclip
In the latest version of Paperclip, the agent creation form connects **OpenCode to OpenRouter by default**, rather than using the local-model configuration from your existing OpenCode configuration file. When creating an agent with an OpenRouter AI connection, Paperclip creates an isolated configuration directory, injects your OpenRouter credentials, and explicitly assigns the agent’s selected model—overriding any default model set in your configuration file. **Simply editing `opencode.json` or `opencode.jsonc` is no longer sufficient for configuring these agents.**
:::

OpenCode can call an OpenAI-compatible local model through Olares Router. This OpenCode runs inside Paperclip and is separate from the OpenCode app in Market.

### Prepare the model and preserve existing settings

1. Install and start Qwen3.8-27B (llama.cpp) from Market.
2. Configure the model in Router and copy its connection details.

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

3. For the Router default model, use `default-chat`. Ensure Router actually routes that alias to your local model. For a direct model endpoint, use the model ID that endpoint accepts instead.

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
        "baseURL": "https://router.your_olares_id.olares.com/v1"
      }
    }
  }
}
```

Follow the Router connection instructions for authentication. If your endpoint requires a key, configure that endpoint's key; an OpenAI or OpenRouter cloud key is not a substitute. Do not publish credential files in support requests.

### Create an unbound OpenCode agent

1. Register and obtain instance-admin or organization access, then create an organization. You do not need a cloud-model key to configure OpenCode itself.
2. Use the bundled Paperclip CLI as your human administrator to create an agent without `runtimeConfig.aiConnection`.
3. Open **Paperclip CLI** from Launchpad. Replace the URL with your Paperclip application origin, without `/api` or a dashboard path:

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

Confirm `adapterType` is `opencode_local`, `adapterConfig.model` is `olares/default-chat`, and `runtimeConfig.aiConnection` is absent. Open **Agents** in Paperclip and select this agent. 

### Configure an existing agent

If you are upgrading from the previous version, you agents do not have the `runtimeConfig.aiConnection` by default. So you only need to merge the configuration above and set its model to `olares/default-chat`, preserving its other settings.

Use Market to restart Paperclip after changing shared configuration, once active work has stopped. Do not modify `/app`, replace the container image manually, or edit the database to detach a connection.

## FAQs

### Which agent adapters does Paperclip support?

Paperclip currently supports the following agent adapters. You configure specific API keys as environment variables depending on the adapter you choose:

- Claude Code: Requires `ANTHROPIC_API_KEY`.
- Codex: Requires `OPENAI_API_KEY`.
- OpenCode: Authentication depends on the model provider. Local models do not inherently require a key. For the OpenRouter provider, you will need to enter an OpenRouter api key in UI
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
