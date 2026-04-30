---
outline: [2, 3]
description: Run NemoClaw on Olares with a local LLM such as Qwen3.5. Set up an always-on AI agent backed by the NVIDIA OpenShell runtime, with no cloud API required.
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, NVIDIA, OpenShell, Nemotron, OpenClaw, local LLM, AI assistant, self-hosted AI, Qwen
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-30"
---

# Run NemoClaw with a local LLM

NemoClaw is an open-source reference stack from NVIDIA that runs OpenClaw always-on assistants with a single command. It bundles the NVIDIA OpenShell runtime for policy-based privacy and security guardrails, giving you more control over your agent's behavior and data handling.

This guide walks you through running NemoClaw on Olares with the Qwen3.5 27B Q4_K_M model app as the backend LLM.

## Prerequisites

- A local model app installed and running on your Olares device.
- Admin privileges to install apps from Market and edit application settings.

## Get the model name and endpoint URL

NemoClaw needs the model name and its shared endpoint URL during installation.

1. Open your model app from Launchpad and note the model name shown on the page. In this example, it's `qwen3.5:27b-q4_K_M`.

   ![Model name shown in the model app](/images/one/qwen3.5-27b-downloaded.png#bordered)

2. Open Settings, then go to **Applications** > **Qwen3.5 27B Q4_K_M (Ollama)**.
3. In **Shared entrances**, select **Qwen3.5 27B Q4_K_M** to view the endpoint URL.

   ![Get shared endpoint](/images/manual/use-cases/deerflow2-shared-entrance.png#bordered){width=90%}

4. Note the shared endpoint. For example:

   ```plain
   http://94a553e00.shared.olares.com
   ```

   :::tip Why the shared endpoint?
   The URL on the model app's main page is user-specific and routes through the browser. The shared endpoint is reachable from other apps on Olares without sign-in or CORS issues, which is what NemoClaw needs.
   :::

## Install NemoClaw

1. Open Market and search for "NemoClaw".

   ![NemoClaw](/images/manual/use-cases/nemoclaw.png#bordered)

2. Click **Get**, then **Install**.
3. When prompted, set the environment variables:

   - **NEMOCLAW_ENDPOINT_URL**: Enter or paste the shared endpoint URL, such as `http://94a553e00.shared.olares.com`.
   - **NEMOCLAW_MODEL**: Enter or paste the model name, such as `qwen3.5:27b-q4_K_M`.

   ![Set environment variables for NemoClaw](/images/manual/use-cases/nemoclaw-set-environment-variables.png#bordered){width=70%}

4. Click **Confirm** and wait for installation to complete.

   Installation takes about 10 minutes, depending on your network. During this time, NemoClaw installs the NVIDIA OpenShell runtime and runs the initial agent onboarding.

When the installation finishes, two shortcuts appear on Launchpad:

- **NemoClaw CLI**: The terminal interface for running NemoClaw commands.
- **NemoClaw Web UI**: The browser-based dashboard.

## Optional: Keep the model loaded

By default, the local LLM unloads from memory after 5 minutes of inactivity, and the next reply has to wait for the model to reload. For an always-on agent, enable the keep-alive setting on the model app to keep it resident in memory.

1. Open Settings and go to **Applications** > **Qwen3.5 27B Q4_K_M (Ollama)** > **Manage environment variables**.
2. Find **KEEP_ALIVE**, click <i class="material-symbols-outlined">edit_square</i>, set the value to **true**, and click **Confirm**.

   ![Enable KEEP_ALIVE for the model app](/images/manual/use-cases/keep-alive-enable.png#bordered){width=80%}

3. Click **Apply**.

:::tip When to leave KEEP_ALIVE unset
Keeping the model loaded consumes VRAM continuously. If you only use the agent occasionally and don't mind the cold-start delay, leave **KEEP_ALIVE** unset.
:::

## Start your first chat

NemoClaw uses the OpenClaw TUI for chat. Because the model and endpoint were configured during installation, you can skip the manual onboarding and go straight to a session.

1. Open the NemoClaw CLI app from Launchpad.
2. Run the following command to connect to the sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

   Wait until the terminal shows the sandbox prompt.

   ![Sandbox connected](/images/manual/use-cases/nemoclaw-connect.png#bordered)

3. Launch the OpenClaw TUI:

   ```bash
   openclaw tui
   ```

   ![Launch the OpenClaw TUI in the sandbox shell](/images/manual/use-cases/nemoclaw-tui.png#bordered)

4. Send a test message, such as `Hi`.

   The first response can take about 30 seconds while the model loads into memory. Subsequent replies are much faster. Once the model is loaded, you can ask the agent about itself to confirm the configuration. For example:

   ```text
   How are you, and what model are you running on?
   ```

   ![NemoClaw chat session](/images/manual/use-cases/nemoclaw-chat-test.png#bordered)

## Learn more

- [NVIDIA NemoClaw](https://build.nvidia.com/nemoclaw): Official reference stack and documentation from NVIDIA.
- [OpenClaw](openclaw.md): Set up OpenClaw features such as persona setup, Discord integration, and skills.
- [Common issues](openclaw-common-issues.md): Troubleshoot model timing and context window settings.
