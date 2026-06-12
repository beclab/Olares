---
outline: [2, 3]
description: Run Paperclip on Olares to coordinate multiple AI agents on the same set of tasks. Add agents backed by Claude Code, Codex, OpenCode, Cursor, or other providers, and file issues for them to work on.
head:
  - - meta
    - name: keywords
      content: Olares, Paperclip, AI agent, multi-agent, Claude Code, Codex, OpenCode, Cursor, self-hosted
app_version: "1.0.0"
doc_version: "1.1"
doc_updated: "2026-06-12"
---

# Coordinate multiple AI agents with Paperclip

Paperclip is an open-source platform for coordinating multiple AI agents under one unified workspace. By setting up a virtual company, you add AI agents powered by Claude Code, Codex, OpenCode, Cursor, or other providers, and assign them issues to work on. Whether the task involves coding, research, or content creation, Paperclip manages the workflow.

Running Paperclip as a self-hosted app on Olares ensures that your API keys, task history, and agent outputs remain entirely private on your device.

## Learning objectives

In this guide, you will learn how to:

- Install Paperclip on Olares.
- Set up the initial admin user account.
- Configure API keys for the agents you plan to use.
- Set up your first company, agent, and task.
- Create an issue and track agent progress.
- Monitor operations and metrics from the dashboard.
- Use local models with Paperclip (optional).

## Upgrade notice

When upgrading to version 1.0.22 or later, please be aware that this update introduces new environment dependencies and switches to the official image. As a result, your runtime environment may have changed. We recommend the following:

- If you have not yet initialized Paperclip or there is no important data, uninstall the app and delete all related data before reinstalling.
- If Paperclip has already been initialized, your existing data will be preserved. However, you may need to reconfigure certain local agent settings and environment dependencies. If you encounter any issues after upgrading, please contact support for assistance.

## Install Paperclip

1. Open Market and search for "Paperclip".

   ![Paperclip in Market](/images/manual/use-cases/paperclip.png#bordered)

2. Click **Get**, then click **Install**. Wait for the installation to finish.

After installation, two entry points are available from the Launchpad.

![Paperclip entry points](/images/manual/use-cases/paperclip-install-entry.png#bordered){width=30%}

## Initialize Paperclip

### Configure API keys first

Before creating an admin account, set up your API key in **Settings** so cloud models can be used immediately. See [Configure API keys](#configure-api-keys) for the full procedure.

### Create an admin account

Paperclip ships without a default user account. To access the platform for the first time, you need to create a admin account through the registration flow.

1. Open Paperclip from the Launchpad. Click **Register account**.

2. On the sign-up page, fill in the required information and submit.

3. After sign-up, click **Claim this instance** to bind the admin account.

   ![Claim this instance](/images/manual/use-cases/paperclip-claim-instance.png#bordered){width=60%}

4. Name your Name your company and finish the onboarding process as instructed.

### Create your first company

A Paperclip workspace is organized around a virtual company structure. This company organizes your agents and tasks.

1. On the **Company** tab, configure the basic information:

   a. Specify a name for the company.

   b. (Optional) Specify the mission or goal.

   ![Set company basics](/images/manual/use-cases/paperclip-set-company.png#bordered)

   c. Click **Next**.

2. On the **Agent** tab, create your first agent:

   a. Specify the following settings:

      - **Agent name**: Use the default name **CEO** or a custom one.
      - **Adapter type**: Select the underlying framework. For example, select **Claude Code (Local Claude agent)** or **Codex (Local Codex agent)**.
      - **More Agent Adapter Types**: Expand to select alternatives like **OpenCode** or **Cursor**.
      - **Model**: Select a specific AI model from the drop-down list.

   b. Click **Test now** to verify the configuration works with your API key.

   c. Click **Next**.

   ![Create an agent](/images/manual/use-cases/paperclip-create-agent.png#bordered)

   :::tip
   To review the API key requirements for each adapter, see [Which agent adapters does Paperclip support](#which-agent-adapters-does-paperclip-support).
   :::

3. On the **Task** tab, define your first task:

   a. **Task title** and **Description**: Specify the title and description, or keep the defaults.

   b. Click **Next**. This task automatically converts into your first issue after the setup is completed.

   ![Set up a task](/images/manual/use-cases/paperclip-set-task.png#bordered)

4. Click **Create & Open Issue**. The **Issues** page opens, displaying your active workspace.

   ![Issue page after setup](/images/manual/use-cases/paperclip-issue-page.png#bordered)

## Configure API keys

Each agent requires an API key for its underlying model provider. You configure these keys as environment variables in the Settings app.

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

:::warning Risk note
Paperclip is a fully autonomous multi-agent collaboration platform. Running it entirely on local models can cause workflow disruptions due to model capability limits, context overflow, or concurrency restrictions, potentially leading to cascading failures. Use a hybrid setup: configure powerful cloud models for CEO/CTO roles, local models for other execution roles, or use local models as a cheap mode to save token costs. Evaluate model capabilities and your use case before deciding on the best configuration.
:::

### Use OpenCode to call local models

:::info
This OpenCode runs inside the Paperclip container and is separate from the OpenCode app installed from the Olares Market. You need to configure it independently.
:::

1. Install a model from the Market. For this example, we’ll use `gemma4:26b`.

2. When creating an agent, choose **OpenCode** as the runtime. For the model, you may begin with a free option such as **big-pickle** to complete the initial setup.

3. After initialization, go to **Data/paperclip/paperclip/.config/opencode/**:

   a. Rename `opencode.jsonc` to `opencode.json` and open it in edit mode.

   b. Replace the default configuration with the following example (using gemma4:26b as the model):

   ```json {wrap}
   {
     "$schema": "https://opencode.ai/config.json",
     "model": "olares/gemma4:26b",
     "provider": {
       "olares": {
         "name": "Gemma4:26b (Ollama)",
         "npm": "@ai-sdk/openai-compatible",
         "models": {
           "gemma4:26b": {
             "name": "gemma4:26b"
           }
         },
         "options": {
           "baseURL": "https://<your-olares-id>.olares.com/v1"
         }
       }
     }
   }
   ```

   Update `<your-olares-id>` in the configuration with your specific Olares ID. Ensure that the `baseURL` ends with `/v1` to provide OpenAI-compatible API access through Ollama.

4. Restart the Paperclip container.

5. In Paperclip, go to **Agents** > **Configuration** > **Permissions & Configuration** to verify the newly added local model.

   ![OpenCode local model in agent config](/images/manual/use-cases/paperclip-opencode-model-config.png#bordered)

   :::warning
   If you do not plan to use the default `openai/gpt-5.1-codex-mini` as the cheap model, be sure to turn this feature off or switch to another available model.
   :::

### Test the workflow
:::info Create a task to have the CEO hire a new agent
**Task title:** Hire a CMO
**Task description:** Hire a content generation agent that uses opencode as the runtime and olares/gemma4:26b as the model.
:::

You can monitor the execution process and result in the task's **Activity** > **Continuation Summary**.

![Agent run activity](/images/manual/use-cases/paperclip-agent-run-activity.png#bordered)

:::info Let the CMO to write a brand story
**Task title:** Write a brand story
**Task description:** Output in md format and upload the final result as an attachment to the task.
:::

![Brand story output](/images/manual/use-cases/paperclip-brand-story-output.png#bordered)

## FAQs

### Which agent adapters does Paperclip support?

Paperclip currently supports the following agent adapters. You configure specific API keys as environment variables depending on the adapter you choose:

- Claude Code: Requires `ANTHROPIC_API_KEY`.
- Codex: Requires `OPENAI_API_KEY`.
- OpenCode: Requires `ANTHROPIC_API_KEY` or `OPENAI_API_KEY`.
- Pi: Requires `ANTHROPIC_API_KEY` or `OPENAI_API_KEY`.
- Cursor: Requires `CURSOR_API_KEY`.

### Can I use Hermes Agent as an agent runtime?

Yes, but it is not recommended for production environments. Note the following:

1. This Hermes Agent runs inside the Paperclip container and is separate from the Hermes Agent app installed from the Olares Market. You need to install and configure it separately in the Paperclip CLI (`pip install hermes-agent`).

2. The Hermes Agent and Paperclip adapter are still in development. Connection stability issues may require significant debugging.

3. Ensure the following configurations are correct, otherwise the agent may fail to work:

   - Turn off the manual approval setting to prevent Paperclip from timing out while waiting for approval when calling the agent:

   ```json {wrap}
   hermes config set approvals.mode "off"
   hermes config set approvals.cron_mode "approve"
   ```

   - Ensure your Hermes Agent has obtained `PAPERCLIP_API_KEY` and has the Paperclip Skills installed to correctly fetch and modify Tasks in Paperclip.

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