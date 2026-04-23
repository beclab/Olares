---
outline: [2, 3]
description: Learn how to install, configure, personalize, and integrate OpenClaw with Discord.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, OpenClaw installation
app_version: "1.0.9"
doc_version: "1.1"
doc_updated: "2026-04-23"
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
- Optional: Grant OpenClaw access to local files.
- Optional: Enable the sandbox for secure code execution.

## Prerequisites

- Local model: Ensure Ollama or another model provider is installed and running.

    :::tip Model provider
    This tutorial uses Ollama as the model provider. If you are using a different provider or a local proxy, see the [OpenClaw documentation on custom providers](https://docs.openclaw.ai/concepts/model-providers#providers-via-models-providers-custom%2Fbase-url) for configuration details.
    :::
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

Run a quick setup for the agent.

### Step 1: Install your model

Install a tool-capable model, such as `glm-4.7-flash`, `qwen3.5:35b`, and `gpt-oss:20b`. This tutorial uses `qwen3.5:35b`.

:::tip
OpenClaw requires a large "context window" (that is the AI's short-term memory) to handle complex tasks without forgetting your previous instructions. If you are using local models, it is recommended to select a model that natively supports a context window of at least 64K tokens.
:::

<Tabs>
<template #(Recommended)-Download-from-Market>

1. From the Olares Market, search for "Qwen3.5 35B A3B UD-Q4 (Ollama)".

    ![Find model app from Market](/images/manual/use-cases/find-model2.png#bordered)    
2. Click **Get**, and then click **Install**. 
3. When the installation finishes, click **Open**. The model download is started automatically.
4. When the model download is completed, copy and save the **Model Name** and **API** address exactly as shown. You need the information in later configurations

    ![Note model detailed info](/images/manual/use-cases/obtain-model-details1.png#bordered){width=45%}
</template>
<template #Download-via-Ollama>

1. View the list of models that were installed by running the following command:

    ```bash
    ollama list
    ```
2. Copy and save the model name exactly as shown in the **Name** column.
3. If the model is not installed, download and then run it. For more information, see [Ollama](ollama.md).
4. Obtain the Ollama API address from **Settings** > **Applications** > **Ollama** > **Shared entrances** > **Ollama API**, and then copy the endpoint address.

    ![Obtain Ollama API](/images/manual/use-cases/ollama-endpoint1.png#bordered){width=65%}
</template>
</Tabs>

### Step 2: Verify model accessibility

Before configuring OpenClaw, verify that your model is accessible and responsive via the API.

1. Open the OpenClaw CLI app from the Launchpad.
2. Enter the following command to verify your API address and retrieve the list of available models. Ensure you replaced `{Your-Model-API}` with the exact API endpoint you copied in [Step 1](#step-1-install-your-model).

    ```bash
    curl {Your-Model-API}/api/tags
    ```
    For example, 
    ```bash
    curl https://ab694c1c.laresprime.olares.com/api/tags
    ```

    The terminal returns the details of available models, indicating the API is reachable. For example,

    ```text
    {"models":[{"details":{"families":["qwen35moe"],"family":"qwen35moe","format":"gguf","parameter_size":"34.7B","parent_model":"","quantization_level":"Q8_0"},"digest":"ff81134b3a699cbc79d3a9e9ee439335fdcd6f43f4d296f31bf46986fa83e01a","model":"qwen3.5:35b-a3b-ud-q4_K_L","modified_at":"2026-03-24T09:59:18.969770729Z","name":"qwen3.5:35b-a3b-ud-q4_K_L","size":20205634377}]}
    ```    
 
3. Enter the following command to force the model to load into memory and test its response speed. Ensure you replaced `{Your-Model-API}` and `{Your-Model-Name}` with the exact details you copied in [Step 1](#step-1-install-your-model).

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
    curl https://ab694c1c.laresprime.olares.com/api/generate -d '{
    "model": "qwen3.5:35b-a3b-ud-q4_K_L",
    "prompt": "say hello world",
    "stream": false
    }'
    ```

    The terminal returns a successful response containing `Hello World!`, indicating your model is ready to use. For example,

    ```text
    {"model":"qwen3.5:35b-a3b-ud-q4_K_L","created_at":"2026-03-24T11:40:21.619815369Z","response":"Hello World","done":true,"done_reason":"stop","context":[248045,846,198,35571,23066,1814,593,26003,248046,198,248045,74455,198,248068,271,248069,271,9419,4196],"total_duration":47302041704,"load_duration":41637938018,"prompt_eval_count":13,"prompt_eval_duration":4645064174,"eval_count":7,"eval_duration":961633505}
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
    | I understand this is personal-by-default and <br>shared/multi-user use requires lock-down. Continue? | Yes  |
    | Setup mode   | QuickStart   |
    | Config handling  | Use existing values    |
    | Model/auth provider  | Ollama    |
    | Ollama base URL  | The API address from [Step 1](#step-1-install-your-model),<br>such as `https://37e62186.demo0002.olares.com` |
    | Ollama mode | Local |
    | Default model | Select your installed model |
    | Select channel  | Skip for now<br>(You can configure channels later)  |
    | Search provider | Skip for now |
    | Configure skills now   | No <br>(You can install skills later)       |
    | Enable hooks | Skip for now<br>(Press **Space** to select and then press **Enter** to continue) |
    | How do you want to hatch your bot   | Do this later   |

4. After you complete the onboarding wizard, scroll up to the **Control UI** section.
5. Find the **Web UI (with token)**, and then copy the token at the end of the URL (the text immediately following `#token=`). This is your Gateway Token. 

    In this case, it is `f8d86f68cd2457ddabc4e93a3e04a5f49aa9983104ea7be8`.

    ![Obtain gateway token](/images/manual/use-cases/obtain-gateway-token2.png#bordered)
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

    ![Gateway dashboard](/images/manual/use-cases/gateway-dashboard1.png#bordered){width=60%}

    The `unauthorized: gateway token mismatch` error appears. This is expected and means you have not provided your access token yet.

2. In **Gateway Token** field, enter the token you copied in the previous step, and then click **Connect**.

    The `pairing required` error appears. This is expected and means the device connection is waiting for approval.

3. Return to the OpenClaw CLI window and run the following command to view the pending connection request:

    ```bash
    openclaw devices approve --latest
    ```

    The terminal displays a `Selected pending device request` message similar to the following:

    ```text
    Selected pending device request 1174db8b-cad9-49af-b96a-e4ac634b7007
      Device: 005748253152b66dc0f5f6a801f35617db043f107972f259630a6bd098d5f790
      Requested: roles: operator; scopes: operator.admin, operator.approvals, operator.pairing, operator.read, operator.write
      Approved: roles: none; scopes: operator.admin, operator.approvals, operator.pairing, operator.read, operator.write
      Note:   First-time device pairing request.
    Approve this exact request with: openclaw devices approve 1174db8b-cad9-49af-b96a-e4ac634b7007
    ```

4. Locate the `Approve this exact request with` line at the very bottom of the message, and then copy and run that exact command to authorize the Control UI.

    In this example, run the following command as indicated:

    ```bash
    openclaw devices approve 1174db8b-cad9-49af-b96a-e4ac634b7007
    ```

5. When the terminal displays the approval message, return to the Control UI.

    ```text
    Approved 005748253152b66dc0f5f6a801f35617db043f107972f259630a6bd098d5f790 (1174db8b-cad9-49af-b96a-e4ac634b7007)
    ```

    ![Pair success](/images/manual/use-cases/new-pair-success2.png#bordered)

6. Click **Connect** again. You will be logged in and directed to the **Chat** page by default.
7. From the left sidebar, click **Overview** to check the connection status. The **STATUS** in the **Snapshot** panel should now be **OK**.
    ![Health OK](/images/manual/use-cases/openclaw-connected2.png#bordered)
</template>
<template #(Optional)-Pair-device-manually>

:::tip When to use manual pairing
The quick setup in the previous section uses the `openclaw devices approve --latest` command to automatically approve the most recent pairing request. If you have multiple pending requests and need to manually select which device to approve, follow the steps in this section instead.
:::

1. Open the Control UI app from the Launchpad. The **OpenClaw Gateway Dashboard** opens:

    ![Gateway dashboard](/images/manual/use-cases/gateway-dashboard1.png#bordered){width=60%}

    The `unauthorized: gateway token mismatch` error appears. This is expected and means you have not provided your access token yet.

2. In **Gateway Token** field, enter the token you copied in the previous step, and then click **Connect**.

    The `pairing required` error occurs. This is expected and means the device connection is waiting for approval.
    
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

    ![Health OK](/images/manual/use-cases/openclaw-connected2.png#bordered)
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
2. Ensure that <i class="material-symbols-outlined">neurology</i> at the upper-right corner is enabled. This allows you to watch the agent think and edit persona files in real time.
3. Type and send the following message to start:
    ```text
    Wake up please!
    ```
    The agent responds and starts interviewing you. You can establish rules, personality traits, and preferences. For example,

    ```text
    - Call me Bella. I like simple language without technical jargon and 
    concise bulleted answers.
    - You are John, a witty assistant who uses emojis.
    - Never access my calendar without asking first, and never execute any 
    financial operations.
    ```
4. As you chat with the agent, look for the **Tool output** messages. These indicate the agent is successfully writing your preferences to its core persona files, such as `IDENTITY.md`, `USER.md`, and `SOUL.md`. You can expand each tool output to view the details.

    ![Persona files editing by OpenClaw](/images/manual/use-cases/openclaw-persona-recording1.png#bordered)

    :::tip
    If you do not see the intermediate persona file operations, refresh the page by clicking <i class="material-symbols-outlined">refresh</i> at the upper-right corner or by pressing F5.
    :::
5. Continue the conversation until the agent gathers enough information. 
6. (Optional) If the agent fails to update the persona files, explicitly instruct it to do so in the chat. 

    If the issue persists, resolve it using one of the following methods:
    - **Increase the context window**: Select **Config** from the left sidebar, switch to the **Raw** tab, find the `models` section, and then increase the `contextWindow` value to at least 64K (200K is recommended). 
    
        :::tip
        Note that a larger context window consumes more VRAM, so choose a value that your hardware can support.
        :::

    - **Change the model**: Switch to a model with better tool-calling and instruction‑following capabilities.

7. Verify your agent's persona files are updated:

    a. Open the Files app from the Launchpad.
    
    b. Go to **Application** > **Data** > **clawdbot** > **config** > **workspace**.
    
    c. Check the modified time of the `.md` files to identify which ones were recently updated, such as `USER.md` and `IDENTITY.md`.

    ![Persona files generated by OpenClaw](/images/manual/use-cases/openclaw-persona-files.png#bordered){width=90%}

    d. (Optional) Double-click a file to verify that it contains your newly established rules such as name, language style, and restrictions.
      
    :::tip Modify persona settings
    To change these settings in the future, use one of the following methods:
    - Ask the agent in the chat to update its rules.
    - Download the `.md` files from this folder, edit them in a text editor, and re-upload them to overwrite the old ones. 
    :::
8. Right-click the temporary `BOOTSTRAP.md` file and select **Delete** to finish the personalization process.

## Next steps

1. [Integrate with Discord](openclaw-integration.md) to chat with your agent remotely.
2. [Optional: Enable web search](openclaw-web-access.md) to give your agent access to the live internet information.
3. [Install skills and plugins](openclaw-skills.md) to enhance your agent's capabilities.

## Troubleshooting and FAQs

Find solutions to common errors and behavioral issues in [Common issues](openclaw-common-issues.md).

## Learn more

- [How do I create a server in Discord](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)