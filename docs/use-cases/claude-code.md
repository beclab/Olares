---
outline: [2, 3]
description: Run Anthropic's Claude Code on Olares to write, test, and manage code through natural language. Connect via OAuth or a local model.
head:
  - - meta
    - name: keywords
      content: Olares, Claude Code, Anthropic, AI coding, Ollama, terminal, TUI, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-05-15"
---

# Run Claude Code on Olares

Claude Code is Anthropic's official AI coding assistant CLI. It lets you write, test, and manage code through natural language in a terminal-based interface.

On Olares, Claude Code runs inside a browser-based terminal with a pre-configured Ubuntu development environment. This guide demonstrates both authentication methods. The local model example uses a single-model app.

## Prerequisites

- Admin privileges to install apps from Market

## Install Claude Code

1. Open Market and search for "Claude Code".
   <!-- ![Claude Code](/images/manual/use-cases/claude-code.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Initialize Claude Code

1. Open **Claude Code CLI** from Launchpad.
2. In the terminal, run:
   ```bash
   claude
   ```

### Authenticate with OAuth

Use this option if you have a Claude Pro or Max subscription.

1. Select the browser OAuth option when prompted.
2. Complete the sign-in flow in your browser.
3. Return to the terminal, select your workspace, and confirm trust settings.
4. The TUI opens automatically.
   <!-- ![Claude Code TUI](/images/manual/use-cases/claude-code-tui.png#bordered) -->

### Use a local model

Use this option if you want to run Claude Code with a local model. This example uses **Qwen3.5 27B Q4_K_M (Ollama)**.

1. Install **Qwen3.5 27B Q4_K_M (Ollama)** from Market.
2. Open the model app from Launchpad and wait for the download to complete. Note the model name displayed on the page (for example, `qwen3.5:27b-q4_K_M`).

   ![Model name on the model app page](/images/manual/use-cases/deerflow2-get-model-name.png#bordered)
3. Open **Settings**, then navigate to **Applications** > **Qwen3.5 27B Q4_K_M (Ollama)**.
4. Under **Shared entrances**, select the model app to view the endpoint URL.

   ![Qwen3.5 27B shared entrance](/images/manual/use-cases/deerflow2-shared-entrance.png#bordered){width=70%}
5. Copy the shared endpoint. For example:
   ```plain
   http://bd5355000.shared.olares.com
   ```
6. Open **Settings**, then navigate to **Applications** > **Claude Code**.
7. Add the following environment variables:
   - **ANTHROPIC_AUTH_TOKEN**: Enter `ollama`. The model app does not verify this value, but Claude Code requires an auth token to be set.
   - **ANTHROPIC_BASE_URL**: Enter the endpoint URL with `/v1` appended. For example:
     ```plain
     http://bd5355000.shared.olares.com/v1
     ```
   - **ANTHROPIC_MODEL**: Enter the model identifier from step 2. For example:
     ```plain
     qwen3.5:27b-q4_K_M
     ```
8. Save the changes and wait about 10 seconds for the container to restart.
9. Return to the Claude Code CLI terminal and run `claude`.

:::info Switching between modes
If you switch between remote and local models, run `/clear` in Claude Code before starting a new session. This prevents context from one model from affecting the other.
:::

## Use Claude Code

All work happens under `/opt/data`, which is `$HOME` in the container. This directory persists across pod restarts.

### Ask your first question

1. In the Claude Code CLI terminal, run:
   ```bash
   cd /opt/data
   claude
   ```
2. In the TUI, describe your task in natural language. For example:
   - "List the files in the current directory"
   - "Explain what main.py does"

### Build a full-stack project

Claude Code can scaffold multi-service projects, run tests, and verify integrations end to end.

The example below shows how to build a mini BFF (Backend For Frontend) stack inside `$HOME/tmp/mini-bff/` with two services:

- **Backend** (`backend/`, port 8801): A Python FastAPI app with a single endpoint `GET /internal/user/{user_id}` that returns fake user data. Tests use pytest.
- **Gateway** (`gateway/`, port 8802): A Node.js + TypeScript + Express app that proxies requests from `GET /user/:id` to the backend. Tests use Vitest + Supertest.

1. In the Claude Code TUI, enter the following prompt:

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

2. Claude Code creates both services, installs dependencies, runs tests, starts the servers, performs the live integration curl checks, and returns a report with the directory tree, test results, and curl transcripts.
   <!-- ![Claude Code mini BFF result](/images/manual/use-cases/claude-code-mini-bff.png#bordered) -->

## Security and environment

The Claude Code container runs with least-privilege settings. The main process and every command you run inside it use UID / GID 1000 (non-root), with `allowPrivilegeEscalation: false` and all Linux capabilities dropped. `sudo` and `apt install` are not available by design.

To install additional software, use project-level tools instead of system package managers:

**Python**

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install <package>
```

**Node.js**

```bash
npm install
```

**Other languages**

Install tools and dependencies to user-writable paths under `/opt/data`.

:::info
The environment variable `PIP_BREAK_SYSTEM_PACKAGES=1` is set. If needed, you can install packages to the system Python, but using a virtual environment is recommended.
:::

## Pre-installed development tools

The container image is based on Ubuntu 24.04 and comes with common development tools already installed:

| Category | Included tools |
|:---------|:---------------|
| Languages and runtimes | Python 3, Node.js, Go, Rust, Java (OpenJDK 21), Ruby, PHP 8.3, Lua, Perl, SQLite |
| Build tools | `build-essential`, `cmake`, `ninja-build`, `clang`, `pkg-config`, common `-dev` headers |
| CLI utilities | `git`, `git-lfs`, `curl`, `wget`, `jq`, `yq`, `openssh-client`, `unzip`, `zip`, `rsync`, `tmux`, `htop`, `shellcheck` |
| Database clients | `postgresql-client`, `mysql-client`, `redis-tools` |

:::warning
`ripgrep` (`rg`) is intentionally not pre-installed to avoid conflicts with Claude Code's built-in search behavior.
:::

## Troubleshooting

### `claude: command not found`

Wait for the init container to finish installing Claude Code. Confirm that `$HOME/.local/bin` is in your `PATH`.

### OAuth or install script fails

Check your cluster's outbound network. The init container downloads from `https://claude.ai/install.sh`.

### Missing language or library

Determine if it is a system-level dependency. If so, it must be added to the base image by the app maintainer. Otherwise, use `venv`, `npm install`, or similar user-level tools.

## Learn more

- [Ollama use case](./ollama.md): Host local models on Olares.
- [OpenCode use case](./opencode.md): Another AI coding agent for Olares.
- [Claude Code official documentation](https://code.claude.com/docs)
