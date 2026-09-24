---
connectionVersion: "1.12.6"
connectionLatestPath: /use-cases/codex-cli
noindex: true
search: false
outline: [2, 3]
description: Run Codex CLI on Olares to inspect repositories, edit code, execute commands, and test changes. Connect with ChatGPT, an OpenAI API key, or a local model.
head:
  - - meta
    - name: keywords
      content: Olares, Codex CLI, OpenAI, AI coding agent, browser terminal, ChatGPT, local LLM, Qwen3.6, self-hosted
app_version: "1.0.18"
doc_version: "1.0"
doc_updated: "2026-09-18"
---

# Run Codex CLI on Olares — Olares 1.12.6

<VersionRouteSelect />

Codex CLI is OpenAI's open-source coding agent for the terminal. It can inspect a repository, edit files, run commands and tests, and help you understand unfamiliar code through natural-language requests.

On Olares, Codex CLI runs inside a browser-based terminal with a preconfigured development environment. You can connect it to ChatGPT, use an OpenAI API key, or route it to a local OpenAI-compatible model.

## Learning objectives

In this guide, you will learn how to:

- Install Codex CLI from Market.
- Connect Codex to ChatGPT, an OpenAI API key, or a local model.
- Run coding tasks and review the results.
- Switch model connections and clear stored credentials.
- Authenticate Olares CLI for Olares management tasks.

## Prerequisites

Before you begin, you need:

- An Olares device running Olares 1.12.6 or later, with sufficient disk space and memory.
- A ChatGPT account with Codex access, if you plan to connect using ChatGPT.
- An OpenAI API key, if you plan to connect through OpenAI Platform.
- A local model optimized for coding running on your Olares device, if you plan to use local execution. This guide uses the following model:

   | Model type | Model | How to get it |
   | :--- | :--- | :--- |
   | Chat | Qwen3.6-27B MTP (llama.cpp) | Install from Market |

<!--@include: ../reusables/ai-service-connections-1.12.6.md#use-different-model-->

## Install Codex CLI

1. Open Market and search for "Codex CLI".
2. Click **Get**, then **Install**, and wait for installation to complete.

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

   ![Codex device code login](/images/manual/use-cases/codex-cli-device-auth.png#bordered)

   If device code login is unavailable, enable it in your ChatGPT security settings or ask your workspace administrator to allow it.

4. Return to the terminal and check the active authentication method:

   ```bash
   codex login status
   ```

   ![Codex login status](/images/manual/use-cases/codex-cli-login-status.png#bordered)

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

#### Configure Codex CLI

1. Navigate to **Settings** > **Applications** > **Codex CLI** > **Manage environment variables**.
2. Configure the following variables:

   - **OPENAI_BASE_URL**: Paste the Base URL copied from Model Console exactly as shown, including the trailing `/v1`.
   - **OPENAI_API_KEY**: Enter a non-empty placeholder value, such as `olares`. The local model app does not require a real OpenAI key for requests from another app in the same Olares cluster.
   - **CODEX_MODEL**: Enter the exact Model name copied from Model Console.

   ![Codex CLI local model environment variables](/images/manual/use-cases/codex-cli-local-model-env.png#bordered)

3. Click **Apply** and wait for Codex CLI to restart.
4. Open Codex CLI from Launchpad and run:

   ```bash
   codex
   ```

## Use Codex CLI

All project work happens in the `/opt/data` directory. Files stored there persist across app restarts.

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

   ![Codex CLI coding session](/images/manual/use-cases/codex-cli-session.png#bordered)

## Manage Olares with Olares CLI

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

`OPENAI_BASE_URL` and `CODEX_MODEL` still route Codex to the local service. Clear both variables in the app settings, click **Apply**, and wait for the app to restart.

### Why does the local model return 404?

Open the model's Model Console and copy its Base URL again. Select **OpenAI-Compatible** and use the URL exactly as displayed, including `/v1`. Also confirm that **Model** shows **READY** and **Engine** shows **RUNNING**.

### Why does Codex report `model not found`?

Compare `CODEX_MODEL` with the **Model name** in Model Console. The values must match exactly.

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
- [Host local large language models with Engine Base apps](llm-base-apps-1.12.6.md): Deploy and manage models through Model Console.
- [Install Olares CLI](/developer/cli-install.md): Set up Olares management commands and agent skills.
