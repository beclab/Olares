---
outline: [2, 3]
description: Run Paperclip on Olares to coordinate multiple AI agents on the same set of tasks. Add agents backed by Claude Code, Codex, OpenCode, Cursor, or other providers, and file issues for them to work on.
head:
  - - meta
    - name: keywords
      content: Olares, Paperclip, AI agent, multi-agent, Claude Code, Codex, OpenCode, Cursor, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-22"
---

# Coordinate multiple AI agents with Paperclip

Paperclip is an open-source platform for coordinating multiple AI agents under one workspace. You create a company, add agents backed by Claude Code, Codex, OpenCode, Cursor, or other providers, and file issues for them to work on, whether that's coding, research, content, or anything else the agents can handle.

On Olares, Paperclip runs as a self-hosted app, so the API keys, task history, and agent output stay on your device.

## Learning objectives

In this guide, you will learn how to:
- Install Paperclip on Olares and create the first user.
- Configure API keys for the agents you plan to use.
- Set up your first company, agent, and task.
- File an issue and watch an agent work on it.
- Find your way around the Paperclip interface.

## Prerequisites

- Olares is installed and running.
- At least one API key for a supported agent provider, such as Anthropic, OpenAI, Google, or Cursor.

## Install Paperclip

1. Open Market and search for "Paperclip".
   <!-- ![Paperclip in Market](/images/manual/use-cases/paperclip.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Create the first user

Paperclip ships without a default user. The first time you open it, create the CEO account from the pod terminal:

1. Open Control Hub and navigate to **Browse** > **paperclip** > **Deployments** > **paperclip**.
2. Under **Pods**, click the pod name to view its containers, and then click <i class="material-symbols-outlined">terminal</i> next to the container to open the pod terminal.
   <!-- ![Open the Paperclip pod terminal from Control Hub](/images/manual/use-cases/paperclip-enter-container.png#bordered) -->

3. In the pod terminal, run:

    ```bash
    pnpm paperclipai auth bootstrap-ceo
    ```

    <!-- ![Run the bootstrap-ceo command](/images/manual/use-cases/paperclip-bootstrap-command.png#bordered) -->

4. Copy the invite URL from the command output.

    <!-- ![Bootstrap command result with invite URL](/images/manual/use-cases/paperclip-bootstrap-result.png#bordered) -->

5. Open the invite URL in your browser and register an account.

    <!-- ![Paperclip registration page](/images/manual/use-cases/paperclip-register.png#bordered) -->

6. After the account is created, click **Accept bootstrap invite**.

    <!-- ![Accept bootstrap invite](/images/manual/use-cases/paperclip-accept-invite.png#bordered) -->

7. Click **Open board** to enter your new workspace.

    <!-- ![Open board](/images/manual/use-cases/paperclip-open-board.png#bordered) -->
## Configure API keys

Each agent needs the API key for its underlying model provider. Set these as environment variables for Paperclip.

1. Go to **Settings** > **Applications** > **Paperclip** > **Manage environment variables**.
   <!-- ![Manage Paperclip environment variables](/images/manual/use-cases/paperclip-manage-env-vars.png#bordered) -->

2. Click <i class="material-symbols-outlined">edit_square</i> next to a variable and paste your API key. Paperclip supports these variables:

    | Variable | Used by |
    |:---------|:--------|
    | `ANTHROPIC_API_KEY` | Claude Code, OpenCode, Pi |
    | `OPENAI_API_KEY` | Codex, OpenCode, Pi |
    | `GEMINI_API_KEY` | Gemini CLI, Cursor |
    | `CURSOR_API_KEY` | Cursor |

    <!-- ![Input API key](/images/manual/use-cases/paperclip-input-api-key.png#bordered) -->

3. Click **Confirm** to save each value, then click **Apply** at the top of the page to save all changes.
   <!-- ![Apply environment variable changes](/images/manual/use-cases/paperclip-apply-env-vars.png#bordered) -->

4. Restart Paperclip so the new keys take effect. In Control Hub, navigate to **Browse** > **paperclip** > **Deployments** > **paperclip** and restart the deployment.
   <!-- ![Restart Paperclip](/images/manual/use-cases/paperclip-restart.png#bordered) -->

:::tip Add more keys later
You can come back and add or update keys at any time. Just repeat this procedure and restart Paperclip.
:::

## Set up your first company

A Paperclip workspace is organized around a company, which holds its agents, tasks, and issues.

1. Reopen Paperclip and fill in the company details.
   <!-- ![Set company basics](/images/manual/use-cases/paperclip-set-company.png#bordered) -->

2. On the **Create Agent** page, create your first agent. The default name is `CEO`, but you can rename it.

    a. Name the agent.

    b. Select an **Adapter type**. Claude Code and Codex are good starting points.

    c. Pick a model from the drop-down.

    d. Click **Test now** to check the configuration.

    <!-- ![Create an agent](/images/manual/use-cases/paperclip-create-agent.png#bordered) -->

    For what each adapter type requires, see [Agent adapter reference](#agent-adapter-reference).

3. Set up the first task. It becomes your first issue once setup completes. You can keep the defaults or customize the title and description.

    a. Enter a **Task title**.

    b. Enter a **Description**.

    <!-- ![Set up a task](/images/manual/use-cases/paperclip-set-task.png#bordered) -->

4. Click **Finish**. The **Issues** page opens.
   <!-- ![Issue page after setup](/images/manual/use-cases/paperclip-issue-page.png#bordered) -->

## File your first issue

In Paperclip, work happens through issues. You file an issue, Paperclip assigns it to an existing agent or hires a new one for it, and the agent takes it from there.

1. On the **Issues** page, click **New Issue** and describe the work. For example, ask Paperclip to hire a writer agent to draft an article.
   <!-- ![New Issue dialog](/images/manual/use-cases/paperclip-new-issue.png#bordered) -->

2. Click **Create**. Paperclip assigns the issue to an agent and starts work. Open **Inbox** to follow along.
   <!-- ![Inbox tracking the new issue](/images/manual/use-cases/paperclip-inbox.png#bordered) -->

3. Go to **Agents** to see the newly hired writer agent and its details.
   <!-- ![Writer Agent details](/images/manual/use-cases/paperclip-agent-details.png#bordered) -->

4. Open either the **Agents** page or the **Issues** page to read the agent's output.
   <!-- ![Agent output](/images/manual/use-cases/paperclip-agent-output.png#bordered) -->

## Navigate the Paperclip interface

### Main interface

<!-- ![Paperclip main interface](/images/manual/use-cases/paperclip-interface.png#bordered) -->

| Area | Description |
|------|-------------|
| Company workspace tabs | Paperclip groups work by company. Create a new company<br> from the lower-left corner and switch between companies using the tabs. |
| Quick actions | New issue, dashboard, inbox. |
| Workflow modules | Issues, routines, goals. |
| Project management | Plans, milestones, and related project views. |
| Agent management | Hire, configure, and monitor agents. |
| Company configuration | Organization, capabilities, costs, behaviors, settings. |
| Core information area | The board view for whichever module you opened. |
| Settings area | Documentation links, version info, demo settings, and display toggles. |

### Dashboard

<!-- ![Paperclip dashboard](/images/manual/use-cases/paperclip-dashboard.png#bordered) -->

The **Agents** panel shows the top-line KPIs for your company, such as overall health and current activity.

The charts below show 14-day trends so you can spot changes at a glance:

| Chart | Description |
|-------|-------------|
| Run Activity | Bar chart of execution counts over the past 14 days. |
| Issues by Priority | Stacked bar chart of issues grouped by priority (critical, high, medium, low). |
| Issues by Status | Stacked bar chart of issues grouped by status (in progress, completed, blocked). |
| Success Rate | Bar chart of the execution success rate over the past 14 days. |

The activity log captures system events and agent actions, giving you a running audit trail. It lists recent behaviors and recent tasks.

## Agent adapter reference

Paperclip currently supports these agent adapters. Each one depends on a specific API key.

| Adapter | Required environment variable | Notes |
|:--------|:------------------------------|:------|
| Claude Code | `ANTHROPIC_API_KEY` | |
| Codex | `OPENAI_API_KEY` | First-time setup needs a manual device-auth login. See [Codex login fails](#codex-login-fails). |
| Gemini CLI | `GEMINI_API_KEY` | Currently unavailable due to adapter compatibility issues. Waiting for an upstream fix. |
| Hermes Agent | N/A | Currently unavailable due to adapter compatibility issues. Waiting for an upstream fix. |
| OpenCode | `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` | |
| Pi | `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` | |
| Cursor | `CURSOR_API_KEY` | |
| OpenClaw | N/A | Not yet integrated. Waiting for upstream support. |

## Troubleshooting

### Codex login fails

After you update `OPENAI_API_KEY` and restart Paperclip, the Codex adapter still can't authenticate, because `codex login` opens a local browser for OAuth, which the container can't do. As a workaround, run a device-auth login manually:

1. In Control Hub, navigate to **Browse** > **paperclip** > **Deployments** > **paperclip**.
2. Under **Pods**, click the pod name to open its details, and then click <i class="material-symbols-outlined">terminal</i> next to the container to open the pod terminal.
   <!-- ![Open the Paperclip pod terminal for Codex login](/images/manual/use-cases/paperclip-codex-login.png#bordered) -->

3. In the pod terminal, run:

    ```bash
    codex login --device-auth
    ```

4. Open the device-auth link shown in the output to complete authorization.
   <!-- ![Codex device-auth result](/images/manual/use-cases/paperclip-codex-login-result.png#bordered) -->

Once login succeeds, retry the Codex adapter in Paperclip.

## Learn more

- [Paperclip documentation](https://docs.paperclip.ing): Official docs for concepts, features, and configuration.
- [Orchestrate multi-agent workflows with oh-my-openagent](opencode-omo.md): Run multi-agent collaboration inside a single OpenCode instance on Olares.
- [Set up OpenCode as your AI coding agent](opencode.md): Install and configure OpenCode, a common Paperclip adapter.
