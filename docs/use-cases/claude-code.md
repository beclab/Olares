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

Claude Code is an AI coding assistant that helps you write, test, and manage code using natural language. On Olares, this command-line interface runs inside a browser-based terminal equipped with a pre-configured Ubuntu development environment.

## Learning objectives

In this guide, you will learn how to:
- Install the Claude Code app from the Olares Market.
- Connect the Claude Code CLI to a model using an Anthropic subscription or a local model.
- Execute basic and advanced natural language coding workflows.
- Manage software dependencies securely.

## Prerequisites

- An Olares device with sufficient disk space and memory.
- An active Claude Pro or Max subscription, if you plan to use remote model connectivity.
- A local model optimized for coding running on your Olares device, if you plan to use local execution.

   You can install local models using one of the following methods:
   - **Ollama application**: One app that hosts multiple models. Ensure [Ollama is installed](ollama.md) with at least one model downloaded, such as `qwen3-coder:30b`.
   - **Single-model application**: Runs one specific model as a standalone application. Ensure the model app is installed from Market with the model fully downloaded. This guide uses **Qwen3-Coder 30B (Ollama)**.

## Install Claude Code

1. Open Market, and search for "Claude Code".
   
   ![Claude Code](/images/manual/use-cases/claude-code.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Connect to a model

Start the Claude Code CLI and connect it to a language model. Choose one of the following connection methods.

### Connect using an Anthropic subscription

Use this method if you hold an active Claude Pro or Max subscription.

1. Open the Claude Code CLI from the Launchpad.
2. Enter the following command:

   ```bash
   claude
   ```
   
3. Select a terminal theme such as **Dark mode**, and then press **Enter**.
4. Select **Claude account with subscription** as the login method, and then press **Enter**. A browser window opens for sign-in. If the browser fails to open, click the provided URL to sign in manually.

   ![Claude Code sign in using subscription account](/images/manual/use-cases/claude-sign-subscription.png#bordered)

5. Complete the sign-in flow in your browser, and then copy the authentication code.
6. Return to the terminal, paste the code, and then press **Enter** to complete the login.
7. Review the **Accessing workspace: /opt/data** security prompt, and then select **Yes, I trust this folder**. 

   The Terminal User Interface (TUI) opens automatically.

   ![Claude Code TUI](/images/manual/use-cases/claude-code-tui.png#bordered)

### Connect using a local model

Use this method to run Claude Code locally. This example uses the model app **Qwen3-Coder 30B (Ollama)**.

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

The following examples demonstrate how to interact with Claude Code to complete everyday development tasks.

### Run basic queries

1. In the Claude Code CLI, enter the following command:

   ```bash
   claude
   ```

2. Review the **Accessing workspace: /opt/data** security prompt, and then select **Yes, I trust this folder** to grant Claude Code read, edit, and execute permissions.
3. (Optional) In the TUI, run the `/clear` command to start a new session with empty context.

   ![Claude Code first chat](/images/manual/use-cases/claude-first-chat.png#bordered)

   :::info Switching between modes
   If you switch between remote and local models, run `/clear` in Claude Code first before starting a new session. This prevents context from the previous model from affecting the new workspace.
   :::

4. Describe your task in natural language. For example:

   ```text
   List the files in the current directory
   ```

   The assistant automatically executes the necessary internal commands to explore the directory and returns a detailed list of your files.

   ![Claude Code first chat result](/images/manual/use-cases/claude-first-chat-result.png#bordered)

5. Review the results.

### Build a full-stack project

Claude Code creates multi-service projects, runs tests, and verifies end-to-end integrations.

The following example demonstrates how to build a lightweight "Hello Olares" web application using a single Node.js Express server to handle both the backend API and the frontend display.

1. In the Claude Code TUI, enter the following prompt:

   ```text
   Create a simple full-stack "Hello Olares" application in a new directory called `hello-olares`.
   
   Please do the following:
   1. Initialize a Node.js project and install the `express` package.
   2. Create a backend API (`server.js`) that runs on port 3000 and has a single endpoint `/api/message` returning `{"message": "Hello Olares!"}`.
   3. Create a frontend (`public/index.html`) with vanilla JavaScript that fetches the message from the API and displays it on the screen. Configure the server to serve this static directory.
   4. Start the server in the background, use `curl` to verify the `/api/message` endpoint works, and then stop the server cleanly.
   ```

2. Wait for Claude Code to process the prompt. The assistant automatically initializes the project, installs Express, writes the code, starts the server, and performs a live curl integration check.
3. When the assistant prompts you for permission to proceed, select **Yes, and don't ask again...**. You might need to approve several prompts for different types of actions.
4.	Review the final summary report returned by the assistant. It outlines the newly created project structure, the configured backend API, the frontend setup, and the successful curl test results.

   ![Claude Code coding project result](/images/manual/use-cases/claude-code-report.png#bordered)

<!--
The following example demonstrates how to build a Backend For Frontend (BFF) stack with a Python FastAPI backend and a Node.js gateway.

1. In the TUI, enter the following detailed prompt:

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

2. When the assistant prompts you for permission to proceed with the directory creation, select **Yes, and always allow access to mini-bff/ from this project** (or press `2`). This grants the assistant the necessary permissions to build the rest of the project without asking for approval for every new file.
3. Wait for Claude Code to process the prompt. The assistant automatically creates both services, installs dependencies, runs tests, starts the servers, and performs live integration checks.
4. Review the final report returned by the assistant, which includes the directory tree, test summaries, and execution transcripts.

   ![Claude Code mini BFF result](/images/manual/use-cases/claude-code-mini-bff.png#bordered)
-->

## Manage security and development environments

The Claude Code container operates under strict least-privilege settings to ensure security.

The main process and all executed commands use a non-root user (UID/GID 1000). The container disables `allowPrivilegeEscalation` and drops all Linux capabilities. Consequently, administrative commands like `sudo` and `apt install` are unavailable.

### Review pre-installed development tools

Before you install additional software, review the tools already included in your workspace. The container image is based on Ubuntu 24.04 and comes with the following common development tools pre-installed:

| Category | Included tools |
|:---------|:---------------|
| Languages and runtimes | Python 3, Node.js, Go, Rust, Java (OpenJDK 21), Ruby, PHP 8.3, Lua, Perl, SQLite |
| Build tools | `build-essential`, `cmake`, `ninja-build`, `clang`, `pkg-config`, common `-dev` headers |
| CLI utilities | `git`, `git-lfs`, `curl`, `wget`, `jq`, `yq`, `openssh-client`, `unzip`, `zip`, `rsync`, `tmux`, `htop`, `shellcheck` |
| Database clients | `postgresql-client`, `mysql-client`, `redis-tools` |

:::info
The `ripgrep` (`rg`) utility is intentionally excluded to prevent conflicts with Claude Code's native search behavior.
:::

### Install additional software

If your project requires tools or libraries beyond the pre-installed ones, you must manage them within the container's security boundaries.

#### What you cannot install yourself

If your project requires a system‑level library (e.g., `libpq-dev`, `ffmpeg`, `libssl-dev`), you cannot install it directly. These dependencies must be added to the base container image by the application maintainer.

#### What you can install in your workspace

Inside your writable directories (primarily `/opt/data`), you can install project‑level dependencies without root privileges using common tools:

- **Python**: Create a virtual environment and use `pip`. For example:

   ```bash
   python3 -m venv .venv
   source .venv/bin/activate
   pip install <package>
   ```

- **Node.js**: Use `npm` inside your project folder. For example:

   ```bash
   npm install <package>
   ```

- **Rust/Go** (or other compiled languages): Install binaries to a user‑writable path. For example:

   ```bash
   cargo install --root ~/.local <package>   # Rust
   go install <package>@latest               # Go (installs to ~/go/bin)
   ```

:::info
The container preconfigures the environment variable `PIP_BREAK_SYSTEM_PACKAGES=1`. While this permits system‑wide Python package installations, using a virtual environment is recommended to keep your workspace clean and reliable.
:::

## FAQs

### `claude: command not found`

Wait a few moments for the init container to finish installing Claude Code. Verify that the `$HOME/.local/bin` directory exists in your system `PATH`.

### OAuth or install script fails

Verify your Olares cluster's outbound network connection. The init container requires internet access to download dependencies from https://claude.ai/install.sh.

### Missing language or library

Determine if the missing tool is a system-level dependency. System-level dependencies require app maintainers to add them directly to the base image. For user-level dependencies, use virtual environments (`venv`), `npm install`, or similar local management tools.

## Learn more

- [Set up OpenCode as your AI coding agent](opencode.md)
- [Claude Code official documentation](https://code.claude.com/docs)
