---
outline: [2, 3]
description: Set up Claude Code on Olares to write, test, and manage code through natural language. Connect via OAuth or a local model.
head:
  - - meta
    - name: keywords
      content: Olares, Claude Code, Anthropic, AI coding, Ollama, terminal, TUI, self-hosted
app_version: "0.1.3"
doc_version: "1.0"
doc_updated: "2026-05-19"
---

# Write code using Claude Code

Claude Code is Anthropic's official AI coding assistant command-line interface (CLI). Use it to write, test, and manage code through natural language directly in a terminal-based interface. On Olares, Claude Code runs inside a browser-based terminal equipped with a pre-configured Ubuntu development environment.

## Learning objectives

In this guide, you will learn how to:
- Install the Claude Code app from the Olares Market.
- Initialize the terminal and authenticate your account using an Anthropic subscription or a local model.
- Configure local environment variables for model connections.
- Execute basic and advanced natural language coding workflows.
- Manage dependencies and understand the secure container environment.

## Prerequisites

- A Claude Pro or Max subscription for remote model connectivity, or a compatible local model installed on your Olares device for local execution.

## Install Claude Code

1. Open Market, and search for "Claude Code".
   
   ![Claude Code](/images/manual/use-cases/claude-code.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Initialize and connect to a model

Start the Claude Code CLI and connect it to a language model. Authenticate using an external Anthropic subscription (OAuth) or configure the application to use a local model hosted on your Olares device.

### Authenticate using an Anthropic subscription

Use this method if you hold an active Claude Pro or Max subscription.

1. Open the Claude Code CLI from the Launchpad.
2. Enter the following command:

   ```bash
   claude
   ```
   
3. Select a terminal theme, for example, **Dark mode**.
4. Select **Claude account with subscription**. A browser window opens for sign-in. If the browser fails to open, select the provided URL to sign in manually.

   ![Claude Code sign in using subscription account](/images/manual/use-cases/claude-sign-subscription.png#bordered)

5. Complete the sign-in flow in your browser, and copy the authentication code.
6. Return to the terminal, paste the code, and select your workspace.
7. Confirm the trust settings. The Terminal User Interface (TUI) opens automatically.

   <!-- ![Claude Code TUI](/images/manual/use-cases/claude-code-tui.png#bordered) -->

### Connect to a local model

Use this method to run Claude Code locally. This example uses the Qwen3-Coder 30B (Ollama) model app.

1. Install the model app **Qwen3-Coder 30B (Ollama)** from Market.

   ![Qwen3-Coder 30B (Ollama)](/images/manual/use-cases/qwen3-coder-30b.png#bordered)

2. Open the model app from the Launchpad and wait for the download to complete.
3. Note down the exact model name displayed on the page. For example, `qwen3-coder:30b`.

   ![Model name on the model app page](/images/manual/use-cases/qwen3-coder-model-name.png#bordered){width=50%}

4. Open Settings, and then go to **Applications** > **Qwen3-Coder 30B (Ollama)** > **Shared entrances**.

   ![Model app endpoint in Settings](/images/manual/use-cases/qwen3-coder-30b-endpoint.png#bordered){width=70%}

5. Click **Qwen3-Coder 30B**, and then note down the endpoint URL. For example, `http://609c5d0c0.shared.olares.com`.
6. Go to **Applications** > **Claude Code** > **Manage environment variables**, and then specify the following environment variables:

   - **ANTHROPIC_AUTH_TOKEN**: Enter any text, such as `ollama`. The model app does not verify this value, but Claude Code requires a populated authentication token.
   - **ANTHROPIC_BASE_URL**: Enter the model app's endpoint URL. For example, `http://609c5d0c0.shared.olares.com`.
   - **ANTHROPIC_MODEL**: Enter the model name you noted down earlier. For example, `qwen3-coder:30b`.

   ![Claude Code environment variables settings](/images/manual/use-cases/claude-env-var.png#bordered){width=70%}  

7. Click **Apply**. Wait about 10 seconds for the container to restart.
8. Open the Claude Code CLI from the Launchpad, and then enter `claude` in the terminal to start your session.

:::info Switching between modes
If you switch between remote and local models, run `/clear` in Claude Code first before starting a new session. This prevents context from the previous model from affecting the new workspace.
:::

## Use Claude Code

All project work happens in the `/opt/data` directory, which serves as `$HOME` in the container. This directory persists your files across app restarts.

### Run basic queries

1. In the Claude Code CLI, run the `claude` command. The following security prompt is displayed:

   ```
   Accessing workspace:

   /opt/data

   Quick safety check: Is this a project you created or one you trust? (Like your own code, a well-known open source project, or work from your team). If not, take a moment to review what's in this folder first.

   Claude Code'll be able to read, edit, and execute files here.
   ```

2. Select **Yes, I trust this folder** to grant Claude Code read, edit, and execute permissions.
3. Press **Enter**. You enter the TUI.

   ![Claude Code first chat](/images/manual/use-cases/claude-first-chat.png#bordered)

4. (Optional) Run the `/clear` command to start a new session with empty context.

   :::info Switching between modes
   If you switch between remote and local models, run `/clear` in Claude Code first before starting a new session. This prevents context from the previous model from affecting the new workspace.
   :::

5. Describe your task in natural language. For example:

   ```text
   List the files in the current directory
   ```

   <!--![Claude Code first chat result](/images/manual/use-cases/claude-first-chat-result.png#bordered)-->

### Build a full-stack project

Claude Code creates multi-service projects, runs tests, and verifies end-to-end integrations. The following example demonstrates how to build a Backend For Frontend (BFF) stack with a Python FastAPI backend and a Node.js gateway.

1. In the Claude Code TUI, enter the following detailed prompt:

   ::: details Example prompt
   ```text
   You are in a Linux dev container. All work MUST stay under:

     $HOME/tmp/mini-bff/

   Do not use apt, do not modify system Python site-packages, and do not use Docker-in-Docker.

   ## Layout (create exactly this split)

   - backend/   → Python only (FastAPI). No package.json here.
   - gateway/   → Node + TypeScript + Express only. No Python venv here.

   Never run npm, npx, or node for the gateway while cwd is backend/.
   Never run pytest or .venv/bin/... while cwd is gateway/ unless you intentionally test nothing there.

   Prefer explicit paths every time, e.g.:

     cd "$HOME/tmp/mini-bff/backend" && .venv/bin/pytest ...
     cd "$HOME/tmp/mini-bff/gateway" && npm test

   ## backend/ (Python, port 8801)

   - FastAPI app: GET /internal/user/{user_id} → JSON with at least id and name (fake data is fine; seed user id 1 if helpful).
   - Use a Python venv inside backend/: python3 -m venv .venv, then install deps with pip.
   - Tests: pytest + httpx or Starlette TestClient under backend/tests/, cover success + unknown id (404).

   ## gateway/ (Node + TypeScript + Express, port 8802)

   - GET /user/:id proxies to http://127.0.0.1:8801/internal/user/:id using fetch (built-in) or axios.
   - Document and implement an error-mapping contract in code comments.
   - Tests: vitest + supertest. Mock/stub upstream fetch for error cases.
   - package.json scripts: at least build (tsc), test (vitest run).

   ## Live integration (mandatory)

   1) Start backend on 127.0.0.1:8801 in the background.
   2) Start gateway on 127.0.0.1:8802 in the background after npm run build.
   3) curl -sS -i http://127.0.0.1:8802/user/1 and curl -sS -i http://127.0.0.1:8802/user/999.
   4) Stop both background processes cleanly.

   ## Final report (required)

   Reply with:
   1) tree -L 3 rooted at mini-bff/
   2) Exact commands for backend tests and gateway build/test
   3) Pytest summary and Vitest summary
   4) The two curl transcripts
   5) One sentence about any benign warnings

   Execute everything yourself; do not ask me to run commands manually.
   ```
   :::

2. Wait for Claude Code to process the prompt. The assistant automatically creates both services, installs dependencies, runs tests, starts the servers, and performs live integration checks.
3. Review the final report returned by the assistant, which includes the directory tree, test summaries, and execution transcripts.

   <!-- ![Claude Code mini BFF result](/images/manual/use-cases/claude-code-mini-bff.png#bordered) -->

## Mange security and development environments

The Claude Code container operates under strict least-privilege settings to ensure security.

The main process and all executed commands use a non-root user (UID/GID 1000). The container disables `allowPrivilegeEscalation` and drops all Linux capabilities. Consequently, administrative commands like `sudo` and `apt install` are unavailable.

To install additional software, use project-level tools instead of system package managers:

<Tabs>
<template #Python>

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install <package>
```
</template>
<template #Node.js>

```bash
npm install
```
</template>
<template #Other-languages>

Install tools and dependencies to user-writable paths under `/opt/data`.
</template>
</Tabs>

:::info
The container preconfigures the environment variable `PIP_BREAK_SYSTEM_PACKAGES=1`. While the environment permits system-wide Python package installations, use a virtual environment to maintain a clean and reliable workspace.
:::

## Pre-installed development tools

The container image is based on Ubuntu 24.04 and comes with common development tools already installed:

| Category | Included tools |
|:---------|:---------------|
| Languages and runtimes | Python 3, Node.js, Go, Rust, Java (OpenJDK 21), Ruby, PHP 8.3, Lua, Perl, SQLite |
| Build tools | `build-essential`, `cmake`, `ninja-build`, `clang`, `pkg-config`, common `-dev` headers |
| CLI utilities | `git`, `git-lfs`, `curl`, `wget`, `jq`, `yq`, `openssh-client`, `unzip`, `zip`, `rsync`, `tmux`, `htop`, `shellcheck` |
| Database clients | `postgresql-client`, `mysql-client`, `redis-tools` |

:::info
The `ripgrep` (`rg`) utility is intentionally excluded to prevent conflicts with Claude Code's native search behavior.
:::

## Troubleshooting

### `claude: command not found`

Wait a few moments for the init container to finish installing Claude Code. Verify that the `$HOME/.local/bin` directory exists in your system `PATH`.

### OAuth or install script fails

Verify your Olares cluster's outbound network connection. The init container requires internet access to download dependencies from https://claude.ai/install.sh.

### Missing language or library

Determine if the missing tool is a system-level dependency. System-level dependencies require app maintainers to add them directly to the base image. For user-level dependencies, use virtual environments (`venv`), `npm install`, or similar local management tools.

## Learn more

- [Set up OpenCode as your AI coding agent](./opencode.md)
- [Claude Code official documentation](https://code.claude.com/docs)
