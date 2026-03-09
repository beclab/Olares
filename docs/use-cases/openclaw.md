---
outline: [2, 3]
description: Learn how to install, configure, personalize, and integrate OpenClaw with Discord.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning
---

# OpenClaw

OpenClaw is a personal AI assistant that is designed to run on your local device. It connects directly to the messaging apps like Discord and Slack, and allows you to interact with it right in the app. 

It acts as an "always-on" operator that can execute real tasks, such as searching and sending documents, managing calendars, or browsing webpages.

## Learning objectives

By the end of this guide, you are able to:
- Install and initialize the OpenClaw environment.
- Integrate OpenClaw with Discord.
- Optional: Enable the web search capability using Brave Search.
- Manage skills and plug-ins.

## Prerequisites

- Local model: Ensure Ollama is installed and running.
- Discord account: Required to create the bot application.
- Discord server: A server where you have permissions to add bots.
- (Optional) Brave search API key: Required for the agent to search the web for real-time information. 

    :::tip
    You can obtain a free API key from the [Brave Search API](https://brave.com/search/api/). The free tier of the "Data for Search" plan is usually sufficient for personal use.
    :::

## Upgrade notes

If you are upgrading an existing OpenClaw installation, review the version-specific changes and troubleshooting steps before proceeding. For more information, see [Upgrade OpenClaw](openclaw-upgrade.md).

## Install OpenClaw

1. From the Olares Market, search for "OpenClaw".

    ![Search for OpenClaw from Market](/images/manual/use-cases/find-openclaw.png#bordered){width=90%}

2. Click **Get**, and then click **Install**. When the installation finishes, two shortcuts appear in the Launchpad:
    - **OpenClaw CLI**: The command line interface
    - **Control UI**: The graphical dashboard

    ![OpenClaw entry points](/images/manual/use-cases/openclaw-entry-points.png#bordered){width=30%}

:::tip Run multiple OpenClaw agents
Olares supports app cloning. If you want to run multiple independent AI agents for different tasks, you can clone the OpenClaw app. For more information, see [Clone applications](../manual/olares/market/clone-apps.md).
:::

## Initialize OpenClaw

Run a quick setup for the agent in the OpenClaw CLI.

### Step 1: Prepare your model

Install a tool-capable model, such as `glm-4.7-flash`, `qwen3.5:27b`, and `gpt-oss:20b`. This tutorial uses `qwen3.5:27b`.

:::tip
OpenClaw requires a large "context window" (that is the AI's short-term memory) to handle complex tasks without forgetting your previous instructions. If you are using local models, it is recommended to select a model that natively supports a context window of at least 64K tokens.
:::

#### Download via Ollama

1. View the list of models that were installed by running the following command:

    ```bash
    ollama list
    ```
2. Copy and save the model name exactly as shown in the **Name** column.
3. If the model is not installed, download it. For more information, see [Ollama](ollama.md).
4. Obtain the Ollama API address from **Settings** > **Applications** > **Ollama** > **Shared Entrances** > **Ollama API**, and then copy the endpoint address. For example, `http://d54536a50.shared.olares.com`.

    ![Obtain Ollama API](/images/manual/use-cases/ollama-endpoint.png#bordered){width=50%}

#### Download from Market

1. From the Olares Market, search for "Qwen3.5 27B".

    ![Find model from Market](/images/manual/use-cases/find-model.png#bordered){width=90%}
2. Click **Get**, and then click **Install**. 
3. When the installation finishes, click **Open**. The model download is started automatically.
4. When the model download is completed, copy and save the **Model Name** and **API** address exactly as shown. You need the information in later configurations

    ![Note model detailed info](/images/manual/use-cases/obtain-model-details.png#bordered){width=50%}

### Step 2: Run onboarding wizard

1. Open the OpenClaw CLI app from the Launchpad.
2. Enter the following command to start the onboarding wizard:
    ```bash
    openclaw onboard
    ```
3. The wizard guides you through a series of steps. Use the arrow keys to navigate and press **Enter** to confirm.

    :::tip Note on configurations
    To get you started quickly, this tutorial skips several advanced settings in the wizard. You can configure or modify them later.
    :::

    | Settings | Option |
    |:---------|:-------|
    | I understand this is personal-by-default and shared/multi-user use requires lock-down. Continue? | Yes |
    | Onboarding mode | QuickStart |
    | Config handling |  Use existing values |
    | Model/auth provider | Custom Provider |
    | API Base URL | The API address appended with `/v1` from **Step 1**,<br>such as `https://37e62186.demo0002.olares.com/v1`|
    | How do you want to provide this API key? | Paste API key now |
    | API Key (leave blank if not required) | Leave it blank or enter any value |
    | Endpoint compatibility | OpenAI-compatible |
    | Model ID | The exact model name, <br>such as `qwen3.5:27b-q4_K_M` |
    | Endpoint ID | A name for this configuration, <br>such as `ollama-qwen3.5` |
    | Model alias (optional) | A short alias such as `qwen3.5` |
    | Select channel | Skip for now<br>(You can configure channels later) |
    | Configure skills now | No <br>(You can install later) |
    | Enable hooks | Select all | 
    | How do you want to hatch your bot | Do this later |

4. After you complete the onboarding wizard, scroll up to the **Control UI** section.
6. Find the **Web UI (with token)**, and then copy the token at the end of the URL (the text immediately following `#token=`). This is your Gateway Token.

    ![Obtain gateway token](/images/manual/use-cases/obtain-gateway-token1.png#bordered){width=70%}

### Step 3. Pair device

Connect the Control UI to the OpenClaw CLI to use the graphical dashboard.

1. Open the Control UI app from the Launchpad.
2. On the **Overview** page, in the **Gateway Access** panel, specify the following settings:
    - **Gateway Token**: Enter the token you copied in the previous step.
    - **Default Session Key**: Enter `agent:main:main`.
3. Click **Connect**.

    The connection error `pairing required` occurs. This is expected and means the device connection is waiting for approval.
4. Return to the OpenClaw CLI window and enter the following command:

    ```bash
    openclaw devices approve --latest
    ```
5. When the terminal displays the approval message, return to the Control UI.
    ![Pair sucess](/images/manual/use-cases/new-pair-success.png#bordered)

    Now the **STATUS** in the **Snapshot** panel should be **OK**.

    ![Health OK](/images/manual/use-cases/openclaw-connected1.png#bordered)

### (Optional) Step 4: Pair device manually

:::tip When to use manual pairing
The quick setup in the previous section uses the `openclaw devices approve --latest` command to automatically approve the most recent pairing request. If you have multiple pending requests and need to manually select which device to approve, follow the steps in this section instead.
:::

Connect the Control UI to the OpenClaw CLI to use the graphical dashboard.

1. Open the Control UI app from the Launchpad.
2. On the **Overview** page, in the **Gateway Access** panel, specify the following settings:
    - **Gateway Token**: Enter the token you copied in the previous step.
    - **Default Session Key**: Enter `agent:main:main`.
3. Click **Connect**. 

    The connection error `pairing required` occurs. This is expected and means the device connection is waiting for approval.
4. Return to the OpenClaw CLI window and enter the following command:
    ```bash
    openclaw devices list
    ```
5. In the **Pending** table, find the **Request** ID associated with your current device.

    :::info
    The Request ID has a time limit. If the authorization fails, re-run `openclaw devices list` to obtain a new valid ID.
    :::

    ![View pending device request](/images/manual/use-cases/pending-request.png#bordered)
    
6. Authorize the device by entering the following command:

    ```bash
    openclaw devices approve {RequestID}
    ```
7. When the terminal displays the approval message, return to the Control UI. Now the **STATUS** in the **Snapshot** panel should be **OK**.

    ![Health OK](/images/manual/use-cases/openclaw-connected1.png#bordered)

### Step 5: Personalize OpenClaw

To make your OpenClaw bot more personalized, it is highly recommended to complete the persona setup process. 

This process establishes the agent's identity, behavioral boundaries, and long-term memory through persona files. These files keep your agent's behavior consistent across all platforms and channels.

1. In the Control UI, select **Chat** from the left sidebar.
2. Ensure <i class="material-symbols-outlined">neurology</i> at the upper-right corner is enabled. This allows you to watch the agent think and edit persona files in real time.
3. Enter and send the following message to start:
    ```text
    Wake up please!
    ```
    The agent responds and starts interviewing you. You can establish rules, personality traits, and preferences. For example,

    ```text
    - Call me Bella. I like simple language without technical jargons and 
    concise bulleted answers.
    - You are John, a witty assistant who uses emojis.
    - Never access my calendar without asking first, and never execute any 
    financial operations.
    ```
4. As you chat with the agent, look for the **Edit** messages. These indicate the agent is successfully writing your preferences to its core persona files, such as `IDENTITY.md`, `USER.md`, and `SOUL.md`. 

    ![Persona files editing by OpenClaw](/images/manual/use-cases/openclaw-persona-recording.png#bordered){width=90%}

    :::tip
    If you do not see the intermediate persona file operations, refresh the page by clicking <i class="material-symbols-outlined">refresh</i> at the upper-right corner or by pressing F5.
    :::
5. Continue the conversation until the agent gathers enough information. Then, it automatically deletes the temporary `BOOTSTRAP.md` file to finish the personalization process.

    ![Finish hatch agent](/images/manual/use-cases/openclaw-hatch-finish.png#bordered){width=90%}

6. (Optional) If the agent fails to update the persona files or delete `BOOTSTRAP.md`, explicitly instruct it to do so in the chat. 

    If the issue persists, resolve it using one of the following methods:
    - **Increase the context window**: Select **Config** from the left sidebar, switch to the **Raw** tab, find the `models` section, and then increase the `contextWindow` value to at least 64K (200K is recommended). 
    
        :::tip
        Note that a larger context window consumes more VRAM, so choose a value that your hardware can support.
        :::

    - **Change the model**: Switch to a model with better tool-calling and instruction‑following capabilities.

7. Verify your agent's persona files are updated:

    a. Open Files from the Launchpad.
    
    b. Go to **Application** > **Data** > **clawdbot** > **config** > **workspace**.
    
    c. Check the modified time of the `.md` files to identify which ones were recently updated, such as `USER.md` and `IDENTITY.md`.

    ![Persona files generated by OpenClaw](/images/manual/use-cases/openclaw-persona-files.png#bordered){width=90%}

    d. (Optional) Double-click a file to verify that it contains your newly established rules such as name, language style, and restrictions.
      
    :::tip Modify persona settings
    To change these settings in the future, use one of the following methods:
    - Ask the agent in the chat to update its rules.
    - Download the `.md` files from this folder, edit them in a text editor, and re-upload them to overwrite the old ones. 
    :::

## Next

[Integrate with Discord](openclaw-integration.md) to chat with your agent remotely.

## FAQ

### Cannot restart OpenClaw in CLI

If you attempt to manually start, stop, or restart OpenClaw using commands like `openclaw gateway` or `openclaw gateway stop` in the OpenClaw CLI, you receive the following error messages:
- `Gateway failed to start: gateway already running (pid 1); lock timeout after 5000ms`
- `Gateway service check failed: Error: systemctl --user unavailable: spawn systemctl ENOENT`

#### Cause

OpenClaw is deployed as a containerized app in Olares, where the gateway runs as the primary container process `pid 1` and is always active. This environment does not use standard Linux system and service management tools such as `systemd` and `systemctl`, so these commands do not work. 

#### Solution

Do not use the OpenClaw CLI to manage the gateway service. Instead, restart OpenClaw using one of the following methods:
- **Restart OpenClaw from Settings or Market**: 
    - Open **Settings**, go to **Applications** > **OpenClaw**, click **Stop**, and then click then **Resume**.
    - Open **Market**, go to **My Olares**, find **OpenClaw**, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the operation button, select **Stop**, and then select **Resume**.
- **Restart the container**: Open **Control Hub**, click `clawdbot` under **Deployments**, and then click **Restart**.

### Why does my OpenClaw automatically stop during long tasks?

When you ask the OpenClaw agent to perform tasks that take a long time to process like massive web scrapes or deep analysis, the task is abruptly terminated before returning the result.

#### Cause

By default, OpenClaw sets a maximum runtime limit of 10 minutes per task. If a task exceeds this limit, the system forcefully terminates it to save resources.

#### Solution

Extend this timeout limit by modifying the configuration file as follows:
1. Open the Control UI, go to **Config** > **Raw**, and then find the `agents` section.
2. In the `defaults` block, add the `timeoutSeconds` field or modify the existing one in it. 

    To set it to 1 hour, specify `3600` for the value:

    ```json
    "agents": {
        "defaults": {
            "timeoutSeconds": 3600
        }
    }
    ```
3. Click **Save** to restart the gateway and apply the changes.

## Resources

- [How do I create a server in Discord](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)