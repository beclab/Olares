---
outline: [2, 3]
title: Run a self-hosted OpenClaw AI agent
description: Run OpenClaw on Olares as a self-hosted personal AI agent. Connect Discord or Slack while keeping the assistant and its data on your device.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, self-hosted ai agent, personal ai agent, local ai agent, openclaw on olares
app_version: "1.0.3"
doc_version: "2.1"
doc_updated: "2026-07-29"
---

# Run OpenClaw as your self-hosted personal AI agent

OpenClaw is a personal AI assistant that is designed to run on your local device. It connects directly to messaging apps like Discord and Slack, and allows you to interact with it right in the app. 

It acts as an "always-on" operator that can execute real tasks, such as searching and sending documents, managing calendars, or browsing webpages.

## Learning objectives

In this guide, you will learn how to:
- Install and initialize the OpenClaw environment.
- Integrate OpenClaw with channels like Discord.
- Optional: Enable the web search capability.
- Manage skills and plug-ins.
- Manage Olares with your OpenClaw agent. 
- Optional: Enable the sandbox for secure code execution.

## Prerequisites

Before you begin, you need:

- Discord account: Required to create the bot application.
- Discord server: A server where you have permissions to add bots.
- The following model:

  | Model type | Model | How to get it |
  | :--- | :--- | :--- |
  | Chat | Qwen3.6-27B (llama.cpp) | Install from Market |

  :::tip Model provider
  This tutorial uses Qwen3.6-27B through its OpenAI-compatible API. If you use a different provider or local proxy, see the [OpenClaw documentation on custom providers](https://docs.openclaw.ai/concepts/model-providers#providers-via-models-providers-custom%2Fbase-url).
  :::

## Upgrade notes

If you are upgrading an existing OpenClaw installation, review the version-specific changes and troubleshooting steps before proceeding. For more information, see [Upgrade OpenClaw](openclaw-upgrade.md).

## Install OpenClaw

1. From the Olares Market, search for "OpenClaw".

    ![Search for OpenClaw from Market](/images/manual/use-cases/find-openclaw1.png#bordered){width=90%}

2. Click **Get**, and then click **Install**. When the installation finishes, two shortcuts appear in the Launchpad:
    - **OpenClaw CLI**: The command line interface
    - **Control UI**: The graphical dashboard

    ![OpenClaw entry points](/images/manual/use-cases/openclaw-entry-points1.png#bordered){width=30%}

:::tip Run multiple OpenClaw agents
Olares supports app cloning. If you want to run multiple independent AI agents for different tasks, you can clone the OpenClaw app. For more information, see [Clone applications](../manual/olares/market/clone-apps.md).
:::

## Initialize OpenClaw

Run a quick setup for the agent.

### Step 1: Get model connection details

This tutorial uses Qwen3.6-27B (llama.cpp), a tool-capable model available from Market.

:::tip
OpenClaw requires a large "context window" (that is the AI's short-term memory) to handle complex tasks without forgetting your previous instructions. If you are using local models, it is recommended to select a model that natively supports a context window of at least 64K tokens.
:::

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

For Qwen3.6-27B (llama.cpp), OpenClaw uses the OpenAI-compatible API format:

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### Step 2: Verify model accessibility

Before configuring OpenClaw, verify that your model is accessible and responsive via the API.

1. Open the OpenClaw CLI app from the Launchpad.
2. Run the following command. Replace `{Model-Base-URL}` with the Base URL copied in [Step 1](#step-1-get-model-connection-details).

    ```bash
    curl {Model-Base-URL}/models
    ```

    The terminal returns the available models, indicating that the API is reachable.
 
3. Run the following command to test a chat response. Replace the placeholders with the Base URL and Model name copied in Step 1.

    ```bash
    curl {Model-Base-URL}/chat/completions \
      -H "Content-Type: application/json" \
      -d '{
        "model": "{Model-Name}",
        "messages": [{"role": "user", "content": "Say hello world"}]
      }'
    ```

    A successful response containing `Hello World!` indicates that the model is ready.

### Step 3: Run onboarding wizard

Set up OpenClaw using the step-by-step interactive wizard.

1. Open the OpenClaw CLI app from the Launchpad.
2. Enter the following command to start the onboarding wizard:
    ```bash
    openclaw onboard
    ```
3. The wizard guides you through a series of steps. Use the arrow keys to navigate and press **Enter** to confirm.

    :::tip Note on configurations
    To get you started quickly, this tutorial skips several advanced settings in the wizard. You can configure or modify them later.
    :::

    | Settings   | Option   |
    |:-----------|:---------|
    | I understand this is<br>personal-by-default and shared/multi-user use<br>requires lock-down. <br>Continue? | Select **Yes**.  |
    | Setup mode   | Select **QuickStart**.   |
    | Config handling  | Select **Keep current values**.    |
    | Model/auth provider  | Select **Custom Provider**. |
    | API Base URL | Enter the Base URL copied in [Step 1](#step-1-get-model-connection-details). |
    | How do you want to provide this API key? | Select **Paste API key now**. |
    | API Key | Enter a placeholder value, such as `local`. |
    | Endpoint compatibility | Select **OpenAI-compatible**. |
    | Model ID | Enter the Model name copied in Step 1. In this example, it is `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`. |
    | Endpoint ID | Enter a recognizable name, such as `qwen3.6-27b`. |
    | Model alias (optional) | Enter a short alias, such as `qwen3.6`. |
    | Select channel  | Select **Skip for now**.<br>(You can configure channels later)  |
    | Search provider | Select **Skip for now**.<br>(You can configure the search provider later) |
    | Configure skills now   | Select **No**. <br>(You can install skills later)       |
    | Enable hooks | Select **Skip for now**. <br>(Press **Space** to select and then press **Enter** to continue) |
    | How do you want to<br>hatch your bot   | Select **Hatch later**.   |

4. After you complete the onboarding wizard, scroll up to the **Control UI** section.
5. Find the **Web UI (with token)**, and then copy the token at the end of the URL (the text immediately following `#token=`). This is your Gateway Token. 

    In this case, it is `YrzY5wk1WYWIfcTHFodyO43Ge6n1JY4T`.

    ![Obtain gateway token](/images/manual/use-cases/obtain-gateway-token3.png#bordered)

6. Keep the OpenClaw CLI window open. You need it in the next step.
### Step 4: Pair device

Connect the Control UI to the OpenClaw CLI to use the graphical dashboard.

<Tabs>
<template #(Recommended)-Pair-device-automatically>

1. Open the Control UI app from the Launchpad. The **OpenClaw Gateway Dashboard** opens.

    ![Gateway dashboard](/images/manual/use-cases/gateway-dashboard3.png#bordered)

    The `Auth did not match` error appears. This is expected and means you have not provided your access token yet.

2. In the **Gateway Token** field, enter the token you copied in the previous step, and then click **Connect**.

    The `Device pairing required` error appears. This is expected and means the device connection is waiting for approval.

    ![Device pairing required](/images/manual/use-cases/gateway-device-pairing-required.png#bordered)

3. In the `Device pairing required` error message, locate the `Approve this request:` line, and then copy the command shown immediately after it. Do not copy the period at the end.

    In this case, copy `openclaw devices approve 673b3923-cb85-4a82-a8c2-f9f8327d0761`.

4. Return to the OpenClaw CLI window, and then run the copied command to authorize the Control UI.

    :::tip Note on timeout errors
    The approval command has a short time limit. If you receive an `unknown requestId` error, the request has expired. Refresh the Control UI, copy the newly generated command, and then run it in OpenClaw CLI again.

    ```text
    [openclaw] The CLI command failed.
    [openclaw] Reason: unknown requestId
    ```
    :::

5. When the terminal displays the approval message, return to the Control UI.

    ```text
    Approved 005748253152b66dc0f5f6a801f35617db043f107972f259630a6bd098d5f790 (967e3732-b3df-43e3-851a-d99b43198d8e)
    ```

6. Click **Connect** again. You will be logged in and directed to the **Chat** page by default.
7. From the left sidebar, click **Overview** to check the connection status. The **STATUS** in the **Snapshot** panel should now be **OK**.

    ![Health OK](/images/manual/use-cases/openclaw-connected3.png#bordered)
</template>
<template #(Optional)-Pair-device-manually>

:::tip When to use manual pairing
The quick setup in the automatic device pairing section approves the most recent pairing request. If you have multiple pending requests and need to manually select which device to approve, follow the steps in this section instead.
:::

1. Open the Control UI app from the Launchpad. The **OpenClaw Gateway Dashboard** opens.

    ![Gateway dashboard](/images/manual/use-cases/gateway-dashboard3.png#bordered)

    The `Auth did not match` error appears. This is expected and means you have not provided your access token yet.

2. In the **Gateway Token** field, enter the token you copied in the previous step, and then click **Connect**.

    The `Device pairing required` error appears. This is expected and means the device connection is waiting for approval.

3. Return to the OpenClaw CLI window and enter the following command:
    ```bash
    openclaw devices list
    ```
4. In the **Pending** table, find the **Request** ID associated with your current device.

    :::info
    The Request ID has a time limit. If the authorization fails, re-run `openclaw devices list` to obtain a new valid ID.
    :::

    ![View pending device request](/images/manual/use-cases/pending-request.png#bordered)
    
5. Authorize the device by entering the following command:

    ```bash
    openclaw devices approve {RequestID}
    ```

6. When the terminal displays the approval message, return to the Control UI. Now the **STATUS** in the **Snapshot** panel should be **OK**.

    ![Health OK](/images/manual/use-cases/openclaw-connected3.png#bordered)
</template>
</Tabs>

### Step 5: Configure context window

OpenClaw requires a large "context window" (that is the AI's short-term memory) to handle complex tasks without forgetting your previous instructions. 

1. Open the Files app, and then go to **Data** > **clawdbot** > **config**.
2. Double-click the `openclaw.json` file to open it.
3. Click <i class="material-symbols-outlined">edit_square</i> in the upper-right corner to enter the edit mode.
4. Find the `models` section and locate the configuration block for your model.
5. Update the `contextWindow` value to at least 65536 (64K). If your hardware VRAM permits, it is highly recommended to increase it to 204800 (200K).

    ![Configure context window in config file](/images/manual/use-cases/configure-context-win3.png#bordered)

6. Click <i class="material-symbols-outlined">save</i> in the upper-right corner.
7. Restart OpenClaw for the changes to take effect.

<!--1. In the Control UI, select **Config** from the left sidebar, and then switch to the **Raw** tab.
2. Click <i class="material-symbols-outlined">visibility_off</i> to reveal the configuration fields.

    ![Reveal configuration blocks](/images/manual/use-cases/click-hide-icon.png#bordered)

3. Find the `models` section and locate the configuration block for your model.
4. Add or update the `contextWindow` value. Set it to at least 64000 (64K). If your hardware VRAM permits, it is highly recommended to increase it to 200000 (200K).

    ![Configure context window](/images/manual/use-cases/configure-context-win2.png#bordered)
5. Click **Save** in the upper-right corner. The system validates the configuration and applies the change automatically.-->

### Step 6: Personalize OpenClaw

To make your OpenClaw bot more personalized, it is highly recommended to complete the persona setup process.

This process establishes the agent's identity, behavioral boundaries, and long-term memory through persona files. These files keep your agent's behavior consistent across all platforms and channels.

1. In the Control UI, select **Chat** from the left sidebar.
2. Ensure that <i class="material-symbols-outlined">neurology</i> at the upper-right corner is enabled. This allows you to watch the agent use tools and edit persona files in real time.
3. Type and send the following message to start:

    ```text
    Wake up please!
    ```
4. The agent responds and starts interviewing you. You can establish rules, personality traits, and preferences. For example:

    ```text
    - Call me Bella. I like simple language without technical jargon and 
    concise bulleted answers.
    - You are John, a witty assistant who uses emojis.
    - Never access my calendar without asking first, and never execute any 
    financial operations.
    ```

5. As you chat with the agent, you can see the agent is writing your preferences to its core persona files, such as `IDENTITY.md`, `USER.md`, and `SOUL.md`.

    ![Persona files editing by OpenClaw](/images/manual/use-cases/openclaw-persona-recording3.png#bordered)

    :::tip
    If you do not see the intermediate persona file operations, click <i class="material-symbols-outlined">refresh</i> at the upper-right corner or press **F5** to refresh the page.
    :::

6. (Optional) If the agent fails to update the persona files, explicitly instruct it to do so in the chat. 

    If the issue persists, resolve it using one of the following methods:
    - **Increase the context window**: Select **Config** from the left sidebar, switch to the **Raw** tab, find the `models` section, and then increase the `contextWindow` value to at least 64K (200K is recommended). 
    
        :::tip
        Note that a larger context window consumes more VRAM, so choose a value that your hardware can support.
        :::

    - **Change the model**: Switch to a model with better tool-calling and instruction-following capabilities.

7. Continue the conversation until the agent gathers enough information.
8. Verify that the persona files were successfully updated:

    a. Open the Files app from the Launchpad.
    
    b. Go to **Application** > **Data** > **clawdbot** > **config** > **workspace**.
    
    c. Check the modified time of the `.md` files to identify which ones were recently updated, such as `USER.md`, `IDENTITY.md`, and `SOUL.md`.

    ![Persona files generated by OpenClaw](/images/manual/use-cases/openclaw-persona-files.png#bordered){width=90%}

    d. (Optional) Download a file to view it in a supported text editor, and verify that it contains your newly established rules, such as your name, language style, and restrictions.

    e. If the temporary `BOOTSTRAP.md` file exist, delete it.

    :::tip Modify persona settings
    To change these settings in the future, use one of the following methods:
    - Ask the agent in the chat to update its rules.
    - Download the `.md` files from this folder, edit them in a text editor, and re-upload them to overwrite the old ones. 
    :::

## Next steps

1. [Integrate with Discord](openclaw-integration.md) to chat with your agent remotely.
2. [Enable web search](openclaw-web-access.md) to give your agent access to the live internet information.
3. [Install skills and plugins](openclaw-skills.md) to enhance your agent's capabilities.

## Troubleshooting and FAQs

Find solutions to common errors and behavioral issues in [Common issues](openclaw-common-issues.md).

## Learn more

- [How do I create a server in Discord](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)
