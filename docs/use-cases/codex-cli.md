---
outline: [2, 3]
description: Run Codex CLI in a persistent workspace on Olares and access it remotely from any computer through a browser. Connect it to ChatGPT, an OpenAI API key, or a local model.
head:
  - - meta
    - name: keywords
      content: Olares, Codex CLI, OpenAI, AI coding agent, remote development, browser terminal, ChatGPT, local LLM, Qwen3.6, self-hosted
app_version: "1.0.18"
doc_version: "1.0"
doc_updated: "2026-09-18"
---

# Run Codex remotely on Olares

Codex CLI is OpenAI's open-source coding agent for the terminal. It can inspect a repository, edit files, run commands and tests, and help you understand unfamiliar code through natural-language requests.

On Olares, Codex runs in a persistent, browser-based development workspace on your own device. You can leave the environment on Olares and return to it from another computer, while connecting Codex to ChatGPT, an OpenAI API key, or a local model.

## Learning objectives

In this guide, you will learn how to:

- Run Codex CLI in a persistent Olares workspace.
- Access the workspace from another computer through a browser.
- Connect Codex to ChatGPT, an OpenAI API key, or a local model.
- Run a coding task and check the result.
- Switch model connections and clear stored credentials.
- Authenticate Olares CLI for Olares management tasks.

## Prerequisites

- Olares 1.12.6 or later.
- Qwen3.6-27B MTP (llama.cpp) installed from Market and ready in Model Console. This guide uses it for the local model connection.

To deploy a different local model, see [Host local large language models with Engine Base apps](llm-base-apps.md).

## Install Codex CLI

1. Open Market and search for "Codex CLI".

   <!-- ![Codex CLI in Market](/images/manual/use-cases/codex-cli.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Access Codex from another computer

The Codex CLI app opens as a browser terminal from Launchpad. Use the same Olares workspace from another computer over your local network or through LarePass VPN.

<tabs>
<template #Use-.local-domain-(LAN)>

Use this method when your computer and Olares are on the same local network.

:::info Windows users
On Windows, multi-level `.local` domains require additional setup. Try one of these:
- **Import hosts in LarePass**: Open the LarePass desktop app and use the built-in option to import Olares hosts to your system.
- **Use the single-level domain**: Open LarePass and use the local address it provides for your Olares desktop.

For details, see [Access Olares services locally](../manual/best-practices/local-access.md).
:::

1. Open a browser on your computer.
2. Go to your Olares desktop using its `.local` address. For example:

   ```text
   http://desktop.<username>.olares.local
   ```

3. Sign in to Olares.
4. Open Codex CLI from Launchpad. The Codex browser terminal opens on your computer.

</template>
<template #Use-.com-domain-(VPN)>

Use this method when your computer and Olares are on different networks.

1. Install LarePass on your computer and sign in with your Olares ID.
2. Enable LarePass VPN.

   ![Enable LarePass VPN on desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

3. Open your Olares desktop in a browser. For example:

   ```text
   https://desktop.<username>.olares.com
   ```

4. Sign in to Olares.
5. Open Codex CLI from Launchpad. The Codex browser terminal opens over the VPN connection.

For complete VPN setup instructions, see [Access Olares One services securely using LarePass VPN](/one/access-olares-via-vpn.md).

</template>
</tabs>

:::tip Keep work available between connections
Store projects under `/opt/data` so they remain available when the app restarts. For a command that must keep running after you close the browser, run it inside `tmux`.
:::

## Connect to a model

Choose one connection method. ChatGPT device code authentication is the simplest option in the browser-based Olares terminal.

### Connect with ChatGPT

Codex's standard browser login returns credentials to a callback service on `localhost:1455`. In the Olares container, your computer's browser cannot reach that callback directly. Use device code authentication instead.

1. Open Codex CLI from Launchpad.
2. Run:

   ```bash
   codex login --device-auth
   ```

3. Open the URL shown in the terminal, sign in to ChatGPT, and enter the one-time code.

   <!-- ![Codex device code login](/images/manual/use-cases/codex-cli-device-auth.png#bordered) -->

   If device code login is unavailable, enable it in your ChatGPT security settings or ask your workspace administrator to allow it.

4. Return to the terminal and check the active authentication method:

   ```bash
   codex login status
   ```

   <!-- ![Codex login status](/images/manual/use-cases/codex-cli-login-status.png#bordered) -->

::: details Fallback: Complete the browser callback inside the container
Use this method only when device code authentication is unavailable.

1. In the first Codex CLI terminal, run `codex login` and keep the process running.
2. Complete the browser sign-in. When the browser opens a URL beginning with `http://localhost:1455/auth/callback`, copy the complete URL from the address bar.
3. Open Codex CLI in a second browser tab. Replace `localhost` with `127.0.0.1`, keep the query string unchanged, and run the URL with `curl`:

   ```bash
   curl -i "http://127.0.0.1:1455/auth/callback?code=<code>&scope=<scope>&state=<state>"
   ```

4. Confirm that the response is `302 Found` and its `Location` header contains `/success`.
5. Return to the first terminal and run `codex login status`.

The authorization code and callback URL are short-lived credentials. Do not share them or include them in screenshots.

<!-- ![Codex browser callback success](/images/manual/use-cases/codex-cli-callback-success.png#bordered) -->
:::

### Connect with an OpenAI API key

Use this method to bill Codex usage through your OpenAI Platform account.

1. Navigate to **Settings** > **Applications** > **Codex CLI** > **Manage environment variables**.
2. Set **OPENAI_API_KEY** to your OpenAI API key.
3. Click **Apply** and wait for Codex CLI to restart.
4. Open Codex CLI from Launchpad and run `codex`.

### Connect with a local model

This example connects Codex CLI to Qwen3.6-27B MTP (llama.cpp) through its OpenAI-compatible API.

#### Get the model connection details

1. Open Qwen3.6-27B MTP (llama.cpp) from Launchpad. Its Model Console opens automatically.
2. Wait until **Model** shows **READY** and **Engine** shows **RUNNING**.
3. Under **Model**, copy the **Model name** exactly as shown.
4. Under **Engine**:

   a. For **Connection source**, select **Apps in Olares**.

   b. For **API format**, select **OpenAI-Compatible**.

   c. Copy the **Base URL** exactly as shown.

   <!-- ![Qwen3.6-27B MTP model connection details](/images/manual/use-cases/codex-cli-qwen36-mtp-model-console.png#bordered) -->

#### Configure Codex CLI

1. Navigate to **Settings** > **Applications** > **Codex CLI** > **Manage environment variables**.
2. Configure the following variables:

   - **OPENAI_BASE_URL**: Paste the Base URL copied from Model Console exactly as shown, including the trailing `/v1`.
   - **OPENAI_API_KEY**: Enter a non-empty placeholder value, such as `olares`. The local model app does not require a real OpenAI key for requests from another app in the same Olares cluster.
   - **CODEX_MODEL**: Enter the exact Model name copied from Model Console.

   <!-- ![Codex CLI local model environment variables](/images/manual/use-cases/codex-cli-local-model-env.png#bordered) -->

3. Click **Apply** and wait for Codex CLI to restart.
4. Open Codex CLI from Launchpad and run:

   ```bash
   codex
   ```

## Use Codex CLI

1. Check the installed CLI version:

   ```bash
   codex --version
   ```

2. Start an interactive session:

   ```bash
   codex
   ```

3. Describe a focused task. For example:

   ```text
   Read the current directory and summarize the project structure.
   ```

   Or ask Codex to create a small file:

   ```text
   Create a minimal README draft for this repository.
   ```

4. Review the proposed commands and file changes before approving them.
5. Inspect the result and ask a follow-up question or request another change.

   <!-- ![Codex CLI coding session](/images/manual/use-cases/codex-cli-session.png#bordered) -->

### Manage Olares with Olares CLI

Codex authentication only grants access to the selected model service. To ask Codex to install Olares apps, inspect cluster status, or perform other Olares management tasks, authenticate Olares CLI separately:

```bash
olares-cli profile login --olares-id <your-olares-id>
```

Complete two-factor authentication when prompted. For more details, see [Log in to Olares CLI](/developer/cli-log-in.md).

## Switch connections or sign out

### Switch from ChatGPT to an OpenAI API key

Set **OPENAI_API_KEY** in **Settings** > **Applications** > **Codex CLI** > **Manage environment variables**, click **Apply**, and wait for the app to restart.

### Switch from an OpenAI API key to ChatGPT

Clear **OPENAI_API_KEY**, click **Apply**, and wait for the app to restart. Then run `codex login --device-auth`.

### Switch from a local model to OpenAI

Clear **OPENAI_BASE_URL** and **CODEX_MODEL**. Keep a real **OPENAI_API_KEY** for API key authentication, or clear it and use ChatGPT authentication. Click **Apply** and wait for the app to restart.

### Sign out

1. Run:

   ```bash
   codex logout
   codex login status
   ```

2. Continue when the status shows that Codex is signed out.

Signing out clears stored Codex credentials, but it does not remove **OPENAI_API_KEY** from the app environment or sign out Olares CLI.

## FAQs

### Why does browser login stop at `localhost:1455`?

The callback service runs inside the Codex CLI container, so your computer's browser cannot reach it through its own `localhost`. Use `codex login --device-auth`. If device code authentication is unavailable, use the callback fallback described in [Connect with ChatGPT](#connect-with-chatgpt).

### Why does Codex still use the local model after I sign in to ChatGPT?

**OPENAI_BASE_URL** and **CODEX_MODEL** still route Codex to the local service. Clear both variables in the app settings, click **Apply**, and wait for the app to restart.

### Why does the local model return 404?

Open the model's Model Console and copy its Base URL again. Select **OpenAI-Compatible** and use the URL exactly as displayed, including `/v1`. Also confirm that **Model** shows **READY** and **Engine** shows **RUNNING**.

### Why does Codex report `model not found`?

Compare **CODEX_MODEL** with the **Model name** in Model Console. The values must match exactly.

### How do I remove a remaining login file?

First run `codex logout`. If `codex login status` still reports a stored login, remove the app's credential file and check again:

```bash
rm -f /opt/data/.codex/auth.json
codex login status
```

This command only removes Codex credentials stored in the app data. It does not clear **OPENAI_API_KEY** or affect Olares CLI authentication.

### What if Codex reports missing agentic execution capabilities?

Update the Codex binary in the app data directory:

```bash
curl -fsSL https://chatgpt.com/codex/install.sh | CODEX_INSTALL_DIR=/opt/data/.local/bin sh
```

Restart the Codex CLI session and try again.

## Learn more

- [Codex CLI documentation](https://developers.openai.com/codex/cli): Learn about Codex CLI workflows and commands.
- [Codex authentication](https://developers.openai.com/codex/auth): Review ChatGPT, API key, and headless login options.
- [Host local large language models with Engine Base apps](llm-base-apps.md): Deploy and manage models through Model Console.
- [Install Olares CLI](/developer/cli-install.md): Set up Olares management commands and agent skills.
