---
outline: [2, 3]
description: Learn how to install the Qwen3.6-27B-GGUF model app on Olares One, connect it to OpenCode, and use Olares CLI skills to manage your device through natural language.
head:
  - - meta
    - name: keywords
      content: OpenCode, Qwen3.6, llama.cpp, local LLM, Olares CLI skills, AI agent, Olares One
---

# Manage Olares with OpenCode and a local model <Badge type="tip" text="30 min" />

OpenCode is an AI coding agent. Pair it with a capable local model such as **Qwen3.6-27B-GGUF (llama.cpp)**, and you can manage your Olares device through natural language.

This guide walks you through installing the model app, connecting it to OpenCode, logging in to the Olares CLI from OpenCode, and completing a few tasks with Olares skills.

## Learning objectives

- Install the **Qwen3.6-27B-GGUF (llama.cpp)** model app and get its connection details.
- Install OpenCode and connect it to the local model.
- Log in to the Olares CLI from OpenCode and load Olares skills.
- Use natural language in OpenCode to manage Olares, from simple questions to installing and porting apps.

## Prerequisites

**Hardware** <br>
- Olares One connected to a stable network.
- Sufficient disk space to download the model.
- At least 16 GB of GPU VRAM is recommended for Qwen3.6 27B Q4_K_M.

**User permissions**
- Admin privileges to install shared apps from the Market and manage GPU resources.

## Step 1: Install the Qwen3.6-27B-GGUF model app and get the Base URL

1. Open Market, and search for **Qwen3.6-27B-GGUF (llama.cpp)**.

   <!-- ![Install Qwen3.6-27B-GGUF](/images/one/qwen3.6-27b-gguf-market.png#bordered){width=90%} -->

2. Click **Get**, then **Install**, and wait for the installation to finish.
3. Click **Open**. The model console opens automatically.
4. Wait for the model download to finish. The following means finish:
   - **Model** shows **READY**.
   - **Engine** shows **RUNNING**.

   ![Qwen3.6-27B-GGUF model console](/images/one/qwen3.6-27b-model-console.png#bordered)

5. Configure how OpenCode will reach the service:

   - **Connection source**: Select **Apps in Olares**.
   - **API format**: Select **OpenAI-Compatible**.
   - Note down the **Base URL**. For example, `https://b11a5b8a.laresprime.olares.com/v1`.
   - Note down the **Model name**, that is `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`.

## Step 2: Install OpenCode

1. Open Market, and search for "OpenCode".

   ![Install OpenCode](/images/manual/use-cases/opencode.png#bordered)

2. Click **Get**, then **Install**, and wait for the installation to finish.
3. Open **OpenCode** from the Launchpad. On first launch, OpenCode downloads dependency packages. This can take 10 to 30 minutes depending on your network.

   :::tip Track initialization progress
   To see the download progress, open Control Hub, select the OpenCode project, go to **Deployments** > **opencode**, click the running pod, and view the logs for the **init-packages** container.
   :::

## Step 3: Connect OpenCode to the model

1. In OpenCode, click <span class="material-symbols-outlined">settings</span> in the bottom-left corner.

   ![OpenCode settings](/images/manual/use-cases/opencode-settings.png#bordered)

2. Select **Providers**, scroll down, and then click **Connect** next to **Custom provider**.

   ![OpenCode custom provider](/images/manual/use-cases/opencode-custom-provider.png#bordered)

3. Enter the following details:
   - **Provider ID**: A unique identifier. For example, `olares-qwen3.6`.
   - **Display name**: The name shown in the provider list. For example, `local-llamacpp`.
   - **Base URL**: The **Base URL** you noted down in Step 1. Make sure it ends with `/v1`.
   - **Models**:
     - **Model ID**: The exact **Model name** you noted down from Step 1.
     - **Display Name**: The name shown for this model. For example, `Qwen3.6 27B`.

   ![OpenCode provider configuration](/images/one/opencode-provider-config.png#bordered){width=70%}

4. Click **Submit** to save the configuration.
5. Start a new chat.
6. Below the chat box, click **Big Pickle** to open the model selector, and then select the model you just added.

## Step 4: Log in to the Olares CLI and load skills

OpenCode can use the built-in Olares CLI Agent Skills to manage your device. Before the agent can run commands, you must log in to the Olares CLI from OpenCode.

1. In OpenCode, click **Search project** on the top of the page, and then select **Toggle terminal**.

   ![OpenCode terminal panel](/images/one/opencode-terminal.png#bordered)

2. Enter the following command to log in to your Olares account. Ensure you replace with your Olares ID.

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   For example:

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

3. Follow the prompt to enter your Olares login password. The password is hidden.

4. Verify the login:

   ```bash
   olares-cli profile list
   ```

   The output shows your profile with `logged-in` status.

## Step 5: Manage Olares through natural language

With the model connected and the Olares CLI authenticated, you can ask OpenCode to manage your device. The following examples progress from simple to complex.

### Ask a simple question

Start with a basic question to confirm the Olares skills are available:

```text
What apps are currently installed on my Olares?
```

Or ask about resource usage:

```text
Show me the CPU and memory usage of my Olares device.
```

### Install an app from Market

Ask OpenCode to install an app for you:

```text
Install Firefox from the Olares Market and tell me when it's ready.
```

OpenCode uses the `olares-market` skill to find, install, and confirm the app.

### Port an app to Olares

For a more advanced task, ask OpenCode to upload and deploy an existing project:

```text
Upload and deploy this GitHub repository as an Olares app:
https://github.com/chandruk4321/dockerize-static-web-project
```

Follow OpenCode's prompts and approvals until the app is deployed to **My Olares**.

## Troubleshooting

### The model console shows Model or Engine not ready

If **Model** is not **Ready** or **Engine** is not **Running**:

- Make sure you selected a GPU accelerator during installation.
- Check that your GPU has enough VRAM for Qwen3.6 27B Q4_K_M.
- In the model console, go to the **GPU residency** section, click **Detect**, and confirm the model is running on the GPU.

### OpenCode cannot connect to the model

- Make sure the **Base URL** includes `/v1` at the end.
- Make sure **API format** in the model console is set to **OpenAI-Compatible**.
- Make sure the **Model ID** in OpenCode matches the **Model name** shown in the model console exactly.

### Olares CLI commands fail inside OpenCode

- Make sure you ran `olares-cli profile login` and the profile shows `logged-in`.
- If a skill is not triggered, explicitly mention it in your prompt. For example: "Using the olares-market skill, install Firefox."

## Resources

- [Set up OpenCode as your AI coding agent](../use-cases/opencode.md): Full OpenCode setup guide.
- [Install and use Agent Skills](../developer/cli-agent-skills.md): Details about the Olares CLI skill bundles.
- [Switch GPU mode](gpu.md)
