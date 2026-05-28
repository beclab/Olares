---
outline: [2, 3]
description: Learn how to install, configure, personalize, and integrate OpenClaw with Discord.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, OpenClaw installation
app_version: "1.0.2"
doc_version: "2.0"
doc_updated: "2026-05-28"
---

# OpenClaw

OpenClaw is a personal AI assistant that is designed to run on your local device. It connects directly to the messaging apps like Discord and Slack, and allows you to interact with it right in the app. 

It acts as an "always-on" operator that can execute real tasks, such as searching and sending documents, managing calendars, or browsing webpages.

## Learning objectives

In this guide, you will learn how to:
- Install and initialize the OpenClaw environment.
- Integrate OpenClaw with channels like Discord.
- Optional: Enable the web search capability.
- Manage skills and plug-ins.
- Manage Olares through natural language. 
- Optional: Grant OpenClaw access to local files.
- Optional: Enable the sandbox for secure code execution.

## Prerequisites

- Local model: Ensure Ollama or another model provider is installed and running.

    :::tip Model provider
    This tutorial uses Ollama as the model provider. If you are using a different provider or a local proxy, see the [OpenClaw documentation on custom providers](https://docs.openclaw.ai/concepts/model-providers#providers-via-models-providers-custom%2Fbase-url) for configuration details.
    :::
- Discord account: Required to create the bot application.
- Discord server: A server where you have permissions to add bots.

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

### Step 1: Install your model

Install a tool-capable model, such as `glm-4.7-flash`, `qwen3.5:35b`, and `gpt-oss:20b`. This tutorial uses `qwen3.5:35b`.

:::tip
OpenClaw requires a large "context window" (that is the AI's short-term memory) to handle complex tasks without forgetting your previous instructions. If you are using local models, it is recommended to select a model that natively supports a context window of at least 64K tokens.
:::

<Tabs>
<template #(Recommended)-Download-from-Market>

1. From the Olares Market, search for "Qwen3.5 35B A3B UD-Q4 (Ollama)".

    ![Find model app from Market](/images/manual/use-cases/qwen35b.png#bordered)    
2. Click **Get**, and then click **Install**. 
3. When the installation finishes, click **Open**. The model download is started automatically.
4. When the model download is completed, note down the **Model Name** exactly as shown. For example, `qwen3.5:35b-a3b-ud-q4_K_L`. You need the name in later configurations.

    ![Note model detailed info](/images/manual/use-cases/obtain-model-details2.png#bordered)

5. Open Settings, and then go to **Applications** > **Qwen3.5 35B A3B UD-Q4 (Ollama)** > **Shared entrances**.

    ![Get model shared endpoint in Settings](/images/manual/use-cases/obtain-model-details3.png#bordered){width=70%}

6. Click ****Qwen3.5 35B A3B UD-Q4_K_L****, and then note down the endpoint URL. For example, `http://026076110.shared.olares.com`.

:::tip Why not use the URL shown on the model page?
The URL shown on the model app page is user-specific and relies on browser-based frontend calls. If your device and Olares are not on the same local network, those calls might trigger Olares sign-in and you might encounter cross-origin restrictions (CORS). To avoid these issues, use the shared endpoint URL.
:::
</template>
<template #Download-via-Ollama>

1. View the list of models that were installed by running the following command:

    ```bash
    ollama list
    ```
2. Copy and save the model name exactly as shown in the **NAME** column. For example, `qwen3.5:27b`.
3. If the model is not installed, download and then run it. For more information, see [Ollama](ollama.md).
4. Open Settings, and then go to **Applications** > **Ollama** > **Shared entrances**.

   ![Ollama endpoint in Settings](/images/manual/use-cases/bifrost-ollama-endpoint.png#bordered){width=80%}

5. Click **Ollama API**, and then note down the endpoint URL. For example, `http://d54536a50.shared.olares.com`.
</template>
</Tabs>

### Step 2: Verify model accessibility

Before configuring OpenClaw, verify that your model is accessible and responsive via the API.

1. Open the OpenClaw CLI app from the Launchpad.
2. Enter the following command to verify your API address and retrieve the list of available models. Ensure you replaced `{Your-Model-API}` with the exact API endpoint you copied in [Step 1](#step-1-install-your-model).

    ```bash
    curl {Your-Model-API}/api/tags
    ```

    For example:
    ```bash
    curl http://026076110.shared.olares.com/api/tags
    ```

    The terminal returns the details of available models, indicating the API is reachable. For example:

    ```text
    {"models":[{"capabilities":["completion","tools","thinking"],"details":{"context_length":262144,"embedding_length":2048,"families":["qwen35moe"],"family":"qwen35moe","format":"gguf","parameter_size":"34.7B","parent_model":"qwen3.5:35b-a3b-ud-q4_K_L-base","quantization_level":"Q8_0"},"digest":"e8cb37adef5d1325d7fed17ec8124d37cb6ba5f2f357887811d75a139ddb79dc","model":"qwen3.5:35b-a3b-ud-q4_K_L","modified_at":"2026-05-27T07:07:06.19481654Z","name":"qwen3.5:35b-a3b-ud-q4_K_L","size":20205634377}]}
    ```    
 
3. Enter the following command to force the model to load into memory and test its response speed. Ensure you replaced `{Your-Model-API}` and `{Your-Model-Name}` with the exact details you noted down in [Step 1](#step-1-install-your-model).

    :::info Why do this before onboarding?
    Ollama unloads models from memory after 5 minutes of inactivity by default. Reloading large models takes time and can cause the onboarding verification in the next step to time out and fail. This command "wakes" the model to ensure a smooth setup.
    ::: 

    ```bash
    curl {Your-Model-API}/api/generate -d '{
    "model": "{Your-Model-Name}",
    "prompt": "say hello world",
    "stream": false
    }'
    ```
    For example,

    ```bash
    curl http://026076110.shared.olares.com/api/generate -d '{
    "model": "qwen3.5:35b-a3b-ud-q4_K_L",
    "prompt": "say hello world",
    "stream": false
    }'
    ```

    The terminal returns a successful response containing `Hello World!`, indicating your model is ready to use. For example,

    ```text
    {"model":"qwen3.5:35b-a3b-ud-q4_K_L","created_at":"2026-05-27T07:17:52.542888337Z","response":"Hello, World! 🌍","done":true,"done_reason":"stop","context":[248045,846,198,35571,23066,1814,593,26003,248046,198,248045,74455,198,248068,271,248069,271,9419,11,4196,0,10838,234,235],"total_duration":22384074696,"load_duration":197437036,"prompt_eval_count":13,"prompt_eval_duration":22064969000,"eval_count":12,"eval_duration":73444000}
    ```

### Step 3: Run onboarding wizard

Set up OpenClaw using the step-by-step interactive wizard, or bypass the prompts using direct commands.

<Tabs>
<template #Interactive-setup>

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
    | Config handling  | Select **Use existing values**.    |
    | Model/auth provider  | Select **More**, and then select **Ollama**.    |
    | Ollama mode | Select **Local only**. |
    | Ollama base URL  | Remove the default placeholder text, and then enter<br>the shared endpoint URL from [Step 1](#step-1-install-your-model).<br>For example, `http://026076110.shared.olares.com`. |
    | Default model | Select your installed model.<br>For example, **ollama/qwen3.5:35b-a3b-ud-q4_K_L**. |
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
</template>
<template #Command-setup>

1. Open the OpenClaw CLI app from the Launchpad.
2. Enter the following command to open the onboarding wizard. Ensure you replaced `{Your-Model-API}` and `{Your-Model-Name}` with the exact details you copied in [Step 1](#step-1-install-your-model).

    ```bash
    openclaw onboard --non-interactive \
    --auth-choice ollama \
    --custom-base-url "{Your-Model-API}" \
    --custom-model-id "{Your-Model-Name}" \
    --accept-risk
    ```

    For example:

    ```text
    openclaw onboard --non-interactive \
    --auth-choice ollama \
    --custom-base-url "http://026076110.shared.olares.com" \
    --custom-model-id "qwen3.5:35b-a3b-ud-q4_K_L" \
    --accept-risk
    ```

    A success message displaying your agent's information will appear in the terminal. For example:

    ```text
    Agents: main (default)
    Heartbeat interval: 30m (main)
    Session store (main): /home/node/.openclaw/agents/main/sessions/sessions.json (0 entries)
    Tip: run `openclaw configure --section web` to store your Brave API key for web_search. Docs: https://docs.openclaw.ai/tools/web
    ```

3. Enter the following command to verify that your model is correctly configured:

    ```bash
    openclaw models status --probe
    ```

    The **Status** column in the **Auth probes** table shows `ok`, indicating the model is successfully connected and ready to use.
    
4. Enter the following command to obtain the gateway dashboard access token:
    ```bash
    openclaw dashboard --no-open
    ```

    The dashboard information is displayed. For example,
    ```
    Dashboard URL: http://127.0.0.1:18789/#token=489bad6c7dbe1f49ace62bf647ca66d6f7d78c76d1ba5d0b
    Copy to clipboard unavailable.
    Browser launch disabled (--no-open). Use the URL above.
    ```

5. Open Olares Files, go to **Application** > **Data** > **clawdbot** > **config**, open the `openclaw.json` file, locate the `gateway` section, and then note down the token in `auth`. For example, `YrzY5wk1WYWIfcTHFodyO43Ge6n1JY4T`.

    ![Obtain gateway token in config file](/images/manual/use-cases/obtain-gateway-token-in-config.png#bordered)

6. Keep the OpenClaw CLI window open. You need it in the next step.
</template>
</Tabs>

<!--### Step 3: Run onboarding wizard

1. Open the OpenClaw CLI app from the Launchpad.
2. Enter the following command to open the onboarding wizard. Ensure you replaced `{Your-Model-API}` and `{Your-Model-Name}` with the exact details you copied in [Step 1](#step-1-install-your-model).
    ```bash
    openclaw onboard --non-interactive \
    --auth-choice ollama \
    --custom-base-url "{Your-Model-API}" \
    --custom-model-id "{Your-Model-Name}" \
    --accept-risk
    ```
    For example,
    ```text
    openclaw onboard --non-interactive \
    --auth-choice ollama \
    --custom-base-url "https://ab694c1c.laresprime.olares.com" \
    --custom-model-id "qwen3.5:35b-a3b-ud-q4_K_L" \
    --accept-risk
    ```

    A success message displaying your agent's information will appear in the terminal. For example,
    ```text
    Agents: main (default)
    Heartbeat interval: 30m (main)
    Session store (main): /home/node/.openclaw/agents/main/sessions/sessions.json (0 entries)
    Tip: run `openclaw configure --section web` to store your Brave API key for web_search. Docs: https://docs.openclaw.ai/tools/web
    ```

3. Enter the following command to verify that your model is correctly configured:

    ```bash
    openclaw models status --probe
    ```

    The **Status** column in the **Auth probes** table shows `ok`, indicating the model is successfully connected and ready to use.
    
4. Enter the following command to obtain the gateway dashboard access token:
    ```bash
    openclaw dashboard --no-open
    ```

    The dashboard information is displayed. For example,
    ```
    Dashboard URL: http://127.0.0.1:18789/#token=489bad6c7dbe1f49ace62bf647ca66d6f7d78c76d1ba5d0b
    Copy to clipboard unavailable.
    Browser launch disabled (--no-open). Use the URL above.
    ```

5. Find the **Dashboard URL**, and then copy the token at the end of the URL (the text immediately following `#token=`). This is your Gateway Token. 

    For example, in the output above, the token you need to copy is `489bad6c7dbe1f49ace62bf647ca66d6f7d78c76d1ba5d0b`.

6. Keep the OpenClaw CLI window open. You need it in the next step.-->

<!--### Step 2: Run onboarding wizard

1. Open the OpenClaw CLI app from the Launchpad.
2. Enter the following command to start the onboarding wizard:
    ```bash
    openclaw onboard
    ```
3. The wizard guides you through a series of steps. Use the arrow keys to navigate and press **Enter** to confirm.

    :::tip Note on configurations
    To get you started quickly, this tutorial skips several advanced settings in the wizard. You can configure or modify them later.
    :::

    | Settings                                                                                         | Option                                                                                                    |
    |:-------------------------------------------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------|
    | I understand this is personal-by-default and shared/multi-user use requires lock-down. Continue? | Yes                                                                                                       |
    | Onboarding mode                                                                                  | QuickStart                                                                                                |
    | Config handling                                                                                  | Use existing values                                                                                       |
    | Model/auth provider                                                                              | Custom Provider                                                                                           |
    | API Base URL                                                                                     | The API address appended with `/v1` from **Step 1**,<br>such as `https://37e62186.demo0002.olares.com/v1` |
    | How do you want to provide this API key?                                                         | Paste API key now                                                                                         |
    | API Key (leave blank if not required)                                                            | Leave it blank or enter any value                                                                         |
    | Endpoint compatibility                                                                           | OpenAI-compatible                                                                                         |
    | Model ID                                                                                         | The exact model name from **Step 1**, <br>such as `qwen3.5:27b-q4_K_M`                                    |
    | Endpoint ID                                                                                      | A name for this configuration, <br>such as `ollama-qwen3.5`                                               |
    | Model alias (optional)                                                                           | A short alias such as `qwen3.5`                                                                           |
    | Select channel                                                                                   | Skip for now<br>(You can configure channels later)                                                        |
    | Configure skills now                                                                             | No <br>(You can install skills later)                                                                     |
    | Enable hooks                                                                                     | Select all                                                                                                | 
    | How do you want to hatch your bot                                                                | Do this later                                                                                             |

4. After you complete the onboarding wizard, scroll up to the **Control UI** section.
5. Find the **Web UI (with token)**, and then copy the token at the end of the URL (the text immediately following `#token=`). This is your Gateway Token.

    ![Obtain gateway token](/images/manual/use-cases/obtain-gateway-token1.png#bordered){width=70%}-->
### Step 4: Pair device

Connect the Control UI to the OpenClaw CLI to use the graphical dashboard.

<Tabs>
<template #(Recommended)-Pair-device-automatically>

1. Open the Control UI app from the Launchpad. The **OpenClaw Gateway Dashboard** opens.

    ![Gateway dashboard](/images/manual/use-cases/gateway-dashboard2.png#bordered)

    The `unauthorized: device token mismatch` error appears. This is expected and means you have not provided your access token yet.

2. In the **Gateway Token** field, enter the token you copied in the previous step, and then click **Connect**.

    The `device pairing required` error appears. This is expected and means the device connection is waiting for approval.

3. Return to the OpenClaw CLI window and run the following command to view the pending connection request:

    ```bash
    openclaw devices approve --latest
    ```

    The terminal displays a `Selected pending device request` message similar to the following:

    ```text
    Selected pending device request 301f6c63-b5ce-4465-952f-76363d0cb116
      Device: 005748253152b66dc0f5f6a801f35617db043f107972f259630a6bd098d5f790
      Requested: roles: operator; scopes: operator.admin, operator.approvals, operator.pairing, operator.read, operator.write
      Note:   First-time device pairing request.
    Approve this exact request with: openclaw devices approve 301f6c63-b5ce-4465-952f-76363d0cb116
    ```   

4. Locate the `Approve this exact request with` line at the very bottom of the message.
5. Copy and run the entire command shown after `Approve this exact request with:` to authorize the Control UI.

    In this example, run the following command as indicated in the message line:

    ```bash
    openclaw devices approve 301f6c63-b5ce-4465-952f-76363d0cb116
    ```

6. When the terminal displays the approval message, return to the Control UI.

    ```text
    Approved 005748253152b66dc0f5f6a801f35617db043f107972f259630a6bd098d5f790 (301f6c63-b5ce-4465-952f-76363d0cb116)
    ```

    ![Pair success](/images/manual/use-cases/new-pair-success3.png#bordered)

7. Click **Connect** again. You will be logged in and directed to the **Chat** page by default.
8. From the left sidebar, click **Overview** to check the connection status. The **STATUS** in the **Snapshot** panel should now be **OK**.
    ![Health OK](/images/manual/use-cases/openclaw-connected3.png#bordered)
</template>
<template #(Optional)-Pair-device-manually>

:::tip When to use manual pairing
The quick setup in the previous section uses the `openclaw devices approve --latest` command to automatically approve the most recent pairing request. If you have multiple pending requests and need to manually select which device to approve, follow the steps in this section instead.
:::

1. Open the Control UI app from the Launchpad. The **OpenClaw Gateway Dashboard** opens.

    ![Gateway dashboard](/images/manual/use-cases/gateway-dashboard2.png#bordered)

    The `unauthorized: device token mismatch` error appears. This is expected and means you have not provided your access token yet.

2. In the **Gateway Token** field, enter the token you copied in the previous step, and then click **Connect**.

    The `device pairing required` error appears. This is expected and means the device connection is waiting for approval.

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

To make your OpenClaw bot truly yours, you must complete the initial Bootstrap workflow.

This guided process establishes the agent's identity, behavioral boundaries, and long-term memory by writing them into core persona files, including `IDENTITY.md`, `USER.md`, and `SOUL.md`. This ensures your agent's behavior remains consistent across all chats and platforms.

1. In the Control UI, select **Chat** from the left sidebar.
2. Ensure that <i class="material-symbols-outlined">neurology</i> at the upper-right corner is enabled. This allows you to watch the agent use tools and edit persona files in real time.
3. Type and send the following message to start the initialization:

    ```text
    Wake up please!
    ```

    The agent will reply with a `Bootstrap pending` warning, stating that it cannot reply normally until it handles the `BOOTSTRAP.md` workflow.

4. Instruct the agent to read the file and begin the setup by sending:

    ```text
    Please read BOOTSTRAP.md and follow the instructions to proceed.
    ```

5. The agent will read the file and ask you for basic preferences. Reply with your rules. For example:

    ```text
    - Call me Bella. I like simple language without technical jargon and 
    concise bulleted answers.
    - You are John, a witty assistant who uses emojis.
    - Never access my calendar without asking first, and never execute any 
    financial operations.
    ```

    The agent will generate a **Tool call** to write these details into the `IDENTITY.md` and `USER.md` files.

    ![Persona files editing by OpenClaw](/images/manual/use-cases/openclaw-persona-recording2.png#bordered)

    :::tip Monitor the tool calls
    As you chat, look for the **Tool call** blocks. You can expand them to see exactly what the agent is writing to your core files. If you do not see the intermediate persona file operations, click <i class="material-symbols-outlined">refresh</i> at the upper-right corner or press **F5** to refresh the page.
    :::

6. When the agent asks what to include in `SOUL.md`, which contains its core directives and day-to-day behaviors, reply to finalize its personality. For example:

    ```text
    Just be witty, keep things concise, always ask before taking major
    actions outside this chat, and never use corporate jargon.
    ```

    The agent will generate a **Tool call** to create the `SOUL.md` file.

7. (Optional) If the agent fails to update the persona files, explicitly instruct it to do so in the chat. 

    If the issue persists, resolve it using one of the following methods:
    - **Increase the context window**: Select **Config** from the left sidebar, switch to the **Raw** tab, find the `models` section, and then increase the `contextWindow` value to at least 64K (200K is recommended). 
    
        :::tip
        Note that a larger context window consumes more VRAM, so choose a value that your hardware can support.
        :::

    - **Change the model**: Switch to a model with better tool-calling and instruction-following capabilities.

8. When the agent asks if you want to set up external connections like WhatsApp, choose to skip for now by sending:

    ```text
    Skip for now.
    ```

9. Continue the conversation until the agent gathers enough information and asks to delete the `BOOTSTRAP.md` file. Agree to finish the bootstrap workflow by sending:

    ```text
    Yes please go ahead and delete BOOTSTRAP.md.
    ```

    The agent will delete the temporary bootstrap file and officially come online, ready to assist you.

10. Verify that the persona files were successfully updated:

    a. Open the Files app from the Launchpad.
    
    b. Go to **Application** > **Data** > **clawdbot** > **config** > **workspace**.
    
    c. Check the modified time of the `.md` files to identify which ones were recently updated, such as `USER.md`, `IDENTITY.md`, and `SOUL.md`.

    ![Persona files generated by OpenClaw](/images/manual/use-cases/openclaw-persona-files.png#bordered){width=90%}

    d. (Optional) Download a file to view it in a supported text editor and verify that it contains your newly established rules, such as your name, language style, and restrictions.
      
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