---
outline: [2, 3]
description: Learn how to install the Qwen3.6-27B (llama.cpp) model app on Olares One, connect it to OpenCode, and use Olares CLI skills to manage your device through natural language.
head:
  - - meta
    - name: keywords
      content: OpenCode, Qwen3.6 27B, llama.cpp, local LLM, Olares CLI skills, AI agent, Olares One
---

# Manage Olares with OpenCode and a local model <Badge type="tip" text="30 min" />

OpenCode is an AI coding agent. Pair it with a capable local model such as Qwen3.6-27B (llama.cpp), and you can manage your Olares device through natural language.

This guide walks you through installing the model app, connecting it to OpenCode, logging in to the Olares CLI from OpenCode, and completing a few tasks with Olares skills.

## Learning objectives

- Install the Qwen3.6-27B (llama.cpp) model app and get its connection details.
- Install OpenCode and connect it to the local model.
- Log in to the Olares CLI from OpenCode and load Olares skills.
- Use natural language in OpenCode to manage Olares, from simple questions to installing and porting apps.

## Prerequisites

**Hardware** <br>
- Olares One connected to a stable network.
- Sufficient disk space to download the model.
- At least 23 GB of GPU VRAM is recommended for Qwen3.6 27B.

**User permissions**
- Admin privileges to install shared apps from the Market and manage GPU resources.

## Step 1: Install the model app and get the connection details

1. Open Market, and search for "Qwen3.6-27B (llama.cpp)".

   ![Install Qwen3.6-27B](/images/one/qwen3.6-27b-llamacpp-market.png#bordered)

2. Click **Get**, and then click **Install**.
3. Select a hardware accelerator, and then click **Confirm**.
4. When the installation finishes, click **Open**. The model console opens and the model download starts automatically.
5. Wait for the download to finish. You will see:
   - **Model**: **READY**
   - **Engine**: **RUNNING**

   ![Qwen3.6-27B model console](/images/one/qwen3.6-27b-model-console.png#bordered)

6. Configure how OpenCode will reach the service:

   - **Connection source**: Select **Apps in Olares**.
   - **API format**: Select **OpenAI-Compatible**.
   - Note down the **Base URL**. For example, `https://b11a5b8a.laresprime.olares.com/v1`.
   - Note down the **Model name**. For example, `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`.

## Step 2: Install OpenCode

1. Open Market, and search for "OpenCode".

   ![Install OpenCode](/images/manual/use-cases/opencode.png#bordered)

2. Click **Get**, and then click **Install**. When the installation finishes, two shortcuts appear in the Launchpad:

   - **OpenCode**: The graphical web interface for chatting with the agent and managing projects.
   - **OpenCode Terminal**: The terminal for running CLI commands or launching the TUI (Terminal User Interface).

3. Open **OpenCode**.

   On first launch, OpenCode downloads dependency packages. This can take 10 to 30 minutes depending on your network.

   :::tip Track initialization progress
   To see the download progress, open Control Hub, select the OpenCode project, go to **Deployments** > **opencode**, click the running pod, and view the logs for the **init-packages** container.
   :::

## Step 3: Connect OpenCode to the model

1. In OpenCode, click <span class="material-symbols-outlined">settings</span> in the bottom-left corner.

   ![OpenCode settings](/images/manual/use-cases/opencode-settings.png#bordered)

2. Select **Providers**, scroll down, and then click **Connect** next to **Custom provider**.

   ![OpenCode custom provider](/images/manual/use-cases/opencode-custom-provider.png#bordered)

3. Specify the following settings:
   - **Provider ID**: A unique identifier. For example, `olares-qwen3.6`.
   - **Display name**: The name shown in the provider list. For example, `local-llamacpp`.
   - **Base URL**: The **Base URL** you noted down in Step 1. Make sure it ends with `/v1`.
   - **Models**:
     - **Model ID**: The exact **Model name** you noted down in Step 1.
     - **Display Name**: The name shown for this model. For example, `Qwen3.6 27B`.

   ![OpenCode provider configuration](/images/one/opencode-provider-config.png#bordered){width=70%}

4. Click **Submit** to save the configuration.
5. Start a new chat.
6. Below the chat box, click **Big Pickle** to open the model selector, and then select the model you just added.

## Step 4: Log in to the Olares CLI and load skills

OpenCode can use the built-in Olares CLI Agent Skills to manage your device. Before the agent can run commands, you must log in to the Olares CLI from OpenCode.

1. In OpenCode, click **Search project** at the top of the page, and then select **Toggle terminal**.

   ![OpenCode terminal panel](/images/one/opencode-terminal.png#bordered)

2. Run the following command to log in to your Olares account. Replace `<your-olares-id>` with your actual Olares ID.

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   For example:

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

3. When prompted, enter your Olares login password. The password is hidden as you type.

4. Run the following command to verify that the profile is created and logged in:

   ```bash
   olares-cli profile list
   ```

   Example output:

   ```text
      NAME                   OLARES-ID              STATUS
   *  laresprime@olares.com  laresprime@olares.com  logged-in
   ```

## Step 5: Manage Olares through natural language

With the model connected and the Olares CLI authenticated, you can now manage your Olares device by chatting with OpenCode in natural language. The following examples cover some common scenarios.

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

### Port an app to Olares

For a more advanced task, ask OpenCode to deploy a project:

```text
Upload and deploy this GitHub repository as an Olares app:
https://github.com/chandruk4321/dockerize-static-web-project
```

Follow the prompts and approvals until the app is deployed to **My Olares**.

## Resources

- [Set up OpenCode as your AI coding agent](../use-cases/opencode.md): Full OpenCode setup guide.
- [Install and use Agent Skills](../developer/cli-agent-skills.md): Details about the Olares CLI skill bundles.
- [Switch GPU mode](gpu.md)
