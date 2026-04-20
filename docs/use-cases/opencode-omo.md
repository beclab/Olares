---
outline: [2, 3]
description: Enable oh-my-openagent (OMO) in OpenCode on Olares to orchestrate multiple AI agents. Trigger multi-agent collaboration with ultrawork, configure local or external models, and use built-in MCP servers.
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, oh-my-openagent, OMO, multi-agent, AI coding agent, ultrawork, MCP, self-hosted
app_version: "1.0.6"
doc_version: "1.0"
doc_updated: "2026-04-18"
---

# Orchestrate multi-agent workflows with oh-my-openagent

oh-my-openagent (OMO) is a multi-model agent orchestration plugin for OpenCode. Once enabled, you can trigger multi-agent collaboration in OpenCode with the keyword `ultrawork` (or the alias `ulw`). Specialized agents such as Sisyphus, Hephaestus, Oracle, and Atlas divide the work and handle complex coding tasks together.

:::warning
This guide focuses on a local-model setup. Running OMO purely on local models noticeably degrades orchestration quality and multi-agent collaboration speed compared to paid cloud models. For real work, we recommend a hybrid setup: a paid cloud model for the main agent and Ollama local models for subagents.
:::

## Learning objectives

By the end of this tutorial, you will learn how to:
- Enable OMO in OpenCode on Olares.
- Configure OMO to work with local Ollama models, cloud models, or a hybrid of both.
- Trigger multi-agent collaboration with the `ultrawork` keyword.
- Use the built-in context7, grep_app, and websearch MCP servers.
- Replace the built-in context7 MCP server with a self-hosted Olares deployment.

## Prerequisites

- [OpenCode installed](opencode.md) on Olares, chart version 1.0.6 or later.
- Local models that support tool use, [connected to OpenCode](opencode.md#connect-to-ollama). This guide uses Qwen3.5 27B Q4_K_M and Qwen3.5 9B Q4_K_M as an example. In Olares, each of these models is a separate single-model app, so you need to add them as two model providers.

  :::info Model capability requirements
  - Local models with fewer than 7B parameters usually can't handle `tool_use` or structured output correctly. Avoid them for any agent.
  - A context window of at least 32K tokens is recommended. Core agents (Sisyphus, Hephaestus, Prometheus, Atlas) work better with 64K or more.
  :::
- Your Olares device has internet access.

## Understand OMO on Olares

### How Olares manages OMO

When you install OpenCode, Olares pre-configures OMO. The pre-installed items live under the global OpenCode config directory, which maps to `Application/Data/opencode/.config/opencode/` in Olares Files. The rest of this guide uses `~/.config/opencode/` as shorthand for this path.

| Item | Path | Notes |
|:-----|:-----|:------|
| Plugin registration | `~/.config/opencode/`<br>`opencode.json`<br>(the `plugin` field) | Registers `"oh-my-openagent"`<br>in the global config. |
| Agent-to-model mapping | `~/.config/opencode/`<br>`oh-my-openagent.json` | Recommended model per agent<br>plus a runtime fallback chain.<br>Written on first install only. |
| Environment baseline instructions | `~/.config/opencode/`<br>`olares-baseline-instructions.md` | Web preview, package management,<br>and Olares domain conventions.<br>Refreshed on every startup. |
| Skill files | `~/.config/opencode/skills/`<br>`web-preview/` and<br>`system-admin/` | Force-updated on every startup. |
| npm package | `/usr/local/lib/node_modules/`<br>`oh-my-opencode/`<br>(inside the container) | Persisted in the system snapshot. |

All Olares-managed settings live in the global config directory. They are never written to your workspace `opencode.json`. If your workspace config still contains the old Olares-managed instructions, Olares removes them only when they exactly match the previous preset. Anything you added or modified yourself is preserved.

The `oh-my-openagent.json` file is only written on first install. If you later update it through the `install` command or by hand, your changes are not overwritten.

### Agents

OMO ships with a set of specialized agents, each with a prompt tuned for a specific role. The table below lists the main agents and how you access them.

| Agent | Role | How to access |
|:------|:-----|:--------------|
| Sisyphus | Main orchestrator.<br>Plans, delegates to subagents,<br>and drives parallel execution<br>in `ultrawork` mode. | Agent selector |
| Hephaestus | Autonomous deep worker.<br>Handles architectural reasoning<br>and long autonomous coding sessions. | Agent selector |
| Prometheus | Strategic planner.<br>Interviews you about requirements<br>and produces a detailed plan. | Agent selector |
| Atlas | Plan executor.<br>Takes over a Prometheus plan<br>and executes it. | Agent selector,<br>or `/start-work` after planning |
| Oracle | Read-only architecture consultant.<br>Advises on architecture decisions<br>and complex debugging<br>without modifying files. | Auto-dispatched in `ultrawork` |
| Explore | Code search agent.<br>Retrieves code across the repo. | Auto-dispatched in `ultrawork` |
| Librarian | Documentation search agent.<br>Searches docs and code. | Auto-dispatched in `ultrawork` |

Other supporting agents (such as Metis for plan gap analysis, Momus for plan review, and Multimodal Looker for vision input) are dispatched automatically by Sisyphus when needed.

### How models are selected

OMO splits model selection between you and the config file:

- **Main model**: The model for the agent you chat with directly. You pick it in the OpenCode UI's model selector. It is independent of `oh-my-openagent.json`.
- **Subagent models (`oh-my-openagent.json`)**: When the main agent calls `delegate_task` to hand work off to a subagent (such as Explore or Librarian), the subagent uses the model defined for that agent under the `agents` field in `~/.config/opencode/oh-my-openagent.json`.

The default `oh-my-openagent.json` ships with a multi-tier fallback chain for each subagent: a paid primary model, paid backup models, and free models provided by OpenCode as the last-resort fallback. If the first model is unavailable at runtime, `runtime_fallback` automatically walks down the chain to the next one. Roughly:

```text
oh-my-openagent.json
└── agents
    └── <agent name>
        └── model chain:  paid primary → paid backup → free fallback
                              (runtime_fallback auto-switches on failure)
```

This keeps subagent delegation working even when a provider is down.

To use OMO with a local Ollama setup, override the entries in this file to point at your local model.

### Default models in the configuration file

OMO tunes each agent's prompt for a specific model family. The default config file uses the following primary and fallback models:

| Agent | Recommended UI model (cloud) | Free fallback |
|:------|:-----------------------------|:--------------|
| Sisyphus | Claude Opus 4.6 | Big Pickle |
| Hephaestus | GPT-5.4 | Big Pickle |
| Prometheus | Claude Opus 4.6 | Big Pickle |
| Atlas | Claude Sonnet 4.6 | Big Pickle |
| Oracle | GPT-5.4 | Big Pickle |
| Explore | Claude Haiku 4.5 | GPT-5 Nano |
| Librarian | MiniMax M2.7 | GPT-5 Nano |

For the full list of recommended models per agent and per category, see the [Agent-to-model matching reference](https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/agent-model-matching.md).

## Configure OMO

### Step 1: Enable OMO

OMO is controlled by the `OPENCODE_OMO` environment variable:

- `false` (default): the plugin is not loaded. The npm package and config files stay on disk, so enabling it later doesn't require another download.
- `true`: the plugin is registered in the global config and the npm package is installed.

To enable OMO:

1. Open Settings and navigate to **Applications** > **OpenCode** > **Manage environment variables**.
   ![Locate the OPENCODE_OMO environment variable](/images/manual/use-cases/opencode-env-var.png#bordered)

2. Find the `OPENCODE_OMO` environment variable and click <i class="material-symbols-outlined">edit_square</i>.

3. In the **Value** drop down, select `true`, and click **Confirm**.
   ![Set OPENCODE_OMO to true](/images/manual/use-cases/opencode-enable-omo.png#bordered)

4. Click **Apply** to save your changes and wait for the app to restart.

On first launch, OMO needs time to finish its initial configuration. This might take a while.

To disable OMO later, repeat the steps above and set `OPENCODE_OMO` to `false`. The plugin registration and MCP servers stop, but the npm package and the `oh-my-openagent.json` config stay on disk for the next time you turn it back on.

### Step 2: Configure OMO to use local models

Edit `~/.config/opencode/oh-my-openagent.json` so OMO delegates subagent work to your local Ollama models. You can skip this step in two cases:

- **Cloud-only**: You only plan to use cloud models. Add your API keys under OpenCode **Providers** and move on.
- **Free fallback only**: You don't want to pay or self-host models. OMO uses the free fallback models shipped with OpenCode, but expect slower responses and quota limits.

:::tip Restart required
Restart OpenCode after every edit to `oh-my-openagent.json` to apply the changes.
:::

1. Open Olares Files and navigate to `Application/Data/opencode/.config/opencode/`.
   ![Locate oh-my-openagent.json](/images/manual/use-cases/opencode-config-file.png#bordered)

2. Edit `oh-my-openagent.json`. Under `agents`, point each agent's `model` field to your Ollama model and add `"stream": false`. The model name must include the provider prefix, which must match the provider names you defined when you added the Ollama apps (see `opencode.json`). For example, if your providers are named `ollama-27b` and `ollama-9b`, configure them as:

   ```json
   {
     "agents": {
       "sisyphus": { "model": "ollama-27b/qwen3.5:27b-q4_K_M", "stream": false },
       "hephaestus": { "model": "ollama-27b/qwen3.5:27b-q4_K_M", "stream": false },
       "prometheus": { "model": "ollama-27b/qwen3.5:27b-q4_K_M", "stream": false },
       "atlas": { "model": "ollama-27b/qwen3.5:27b-q4_K_M", "stream": false },
       "explore": { "model": "ollama-9b/qwen3.5:9b", "stream": false },
       "librarian": { "model": "ollama-9b/qwen3.5:9b", "stream": false }
     }
   }
   ```

   :::info Why `"stream": false`
   Ollama's streaming mode returns NDJSON, which the SDK can't parse. Agents that use tools (especially Librarian and Explore) silently fall back to the next model in the chain if this is missing. This is a known Ollama limitation.
   :::

3. In the `categories` section, make sure each category's `fallback_models` list points to local models first.

   ```json
   "categories": {
     "visual-engineering": {
       "model": "google/gemini-3.1-pro-preview",
       "variant": "high",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
     "ultrabrain": {
       "model": "openai/gpt-5.4",
       "variant": "xhigh",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
     "deep": {
       "model": "openai/gpt-5.4",
       "variant": "medium",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
     "artistry": {
       "model": "google/gemini-3.1-pro-preview",
       "variant": "high",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
     "quick": {
       "model": "openai/gpt-5.4-mini",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
     "unspecified-low": {
       "model": "anthropic/claude-sonnet-4-6",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
     "unspecified-high": {
       "model": "anthropic/claude-sonnet-4-6",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
     "writing": {
       "model": "google/gemini-3-flash-preview",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     }
   }
   ```

   If you'd rather replace the whole file at once, see the full sample below.

4. Add concurrency limits in the same file. A local Ollama server has limited resources, and multiple agents hitting it at once can exhaust VRAM or cause heavy slowdowns.

   ```json
   {
     "background_task": {
       "providerConcurrency": {
         "ollama-27b": 1,
         "ollama-9b": 1
       },
       "modelConcurrency": {
         "ollama-27b/qwen3.5:27b-q4_K_M": 1,
         "ollama-9b/qwen3.5:9b": 2
       }
     }
   }
   ```

   This caps requests to the large model to one at a time and prevents out-of-memory crashes.

   :::details Full oh-my-openagent.json

   ```json
   {
     "$schema": "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/oh-my-opencode.schema.json",
     "runtime_fallback": {
       "enabled": true,
       "max_fallback_attempts": 7
     },
     "agents": {
       "sisyphus": { "model": "ollama-27b/qwen3.5:27b-q4_K_M", "stream": false },
       "hephaestus": { "model": "ollama-27b/qwen3.5:27b-q4_K_M", "stream": false },
       "prometheus": { "model": "ollama-27b/qwen3.5:27b-q4_K_M", "stream": false },
       "atlas": { "model": "ollama-27b/qwen3.5:27b-q4_K_M", "stream": false },
       "explore": { "model": "ollama-9b/qwen3.5:9b", "stream": false },
       "librarian": { "model": "ollama-9b/qwen3.5:9b", "stream": false }
     },
     "categories": {
       "visual-engineering": {
         "model": "google/gemini-3.1-pro-preview",
         "variant": "high",
         "fallback_models": [
           { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
           { "model": "ollama-9b/qwen3.5:9b" },
           { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
           { "model": "anthropic/claude-opus-4-6", "variant": "max" },
           { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
           { "model": "opencode/big-pickle" }
         ]
       },
       "ultrabrain": {
         "model": "openai/gpt-5.4",
         "variant": "xhigh",
         "fallback_models": [
           { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
           { "model": "ollama-9b/qwen3.5:9b" },
           { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
           { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
           { "model": "anthropic/claude-opus-4-6", "variant": "max" },
           { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
           { "model": "opencode/big-pickle" }
         ]
       },
       "deep": {
         "model": "openai/gpt-5.4",
         "variant": "medium",
         "fallback_models": [
           { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
           { "model": "ollama-9b/qwen3.5:9b" },
           { "model": "github-copilot/gpt-5.4", "variant": "medium" },
           { "model": "anthropic/claude-opus-4-6", "variant": "max" },
           { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
           { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
           { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
           { "model": "opencode/big-pickle" }
         ]
       },
       "artistry": {
         "model": "google/gemini-3.1-pro-preview",
         "variant": "high",
         "fallback_models": [
           { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
           { "model": "ollama-9b/qwen3.5:9b" },
           { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
           { "model": "anthropic/claude-opus-4-6", "variant": "max" },
           { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
           { "model": "openai/gpt-5.4" },
           { "model": "github-copilot/gpt-5.4" },
           { "model": "opencode/big-pickle" }
         ]
       },
       "quick": {
         "model": "openai/gpt-5.4-mini",
         "fallback_models": [
           { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
           { "model": "ollama-9b/qwen3.5:9b" },
           { "model": "github-copilot/gpt-5.4-mini" },
           { "model": "anthropic/claude-haiku-4-5" },
           { "model": "github-copilot/claude-haiku-4.5" },
           { "model": "google/gemini-3-flash-preview" },
           { "model": "github-copilot/gemini-3-flash-preview" },
           { "model": "opencode/gpt-5-nano" }
         ]
       },
       "unspecified-low": {
         "model": "anthropic/claude-sonnet-4-6",
         "fallback_models": [
           { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
           { "model": "ollama-9b/qwen3.5:9b" },
           { "model": "github-copilot/claude-sonnet-4.6" },
           { "model": "openai/gpt-5.3-codex", "variant": "medium" },
           { "model": "google/gemini-3-flash-preview" },
           { "model": "github-copilot/gemini-3-flash-preview" },
           { "model": "opencode/big-pickle" }
         ]
       },
       "unspecified-high": {
         "model": "anthropic/claude-sonnet-4-6",
         "fallback_models": [
           { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
           { "model": "ollama-9b/qwen3.5:9b" },
           { "model": "github-copilot/claude-sonnet-4.6" },
           { "model": "openai/gpt-5.3-codex", "variant": "medium" },
           { "model": "google/gemini-3-flash-preview" },
           { "model": "github-copilot/gemini-3-flash-preview" },
           { "model": "opencode/big-pickle" }
         ]
       },
       "writing": {
         "model": "google/gemini-3-flash-preview",
         "fallback_models": [
           { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
           { "model": "ollama-9b/qwen3.5:9b" },
           { "model": "github-copilot/gemini-3-flash-preview" },
           { "model": "anthropic/claude-sonnet-4-6" },
           { "model": "github-copilot/claude-sonnet-4.6" },
           { "model": "opencode/big-pickle" }
         ]
       }
     },
     "background_task": {
       "providerConcurrency": {
         "ollama-27b": 1,
         "ollama-9b": 1
       },
       "modelConcurrency": {
         "ollama-27b/qwen3.5:27b-q4_K_M": 1,
         "ollama-9b/qwen3.5:9b": 2
       }
     }
   }
   ```

   :::

5. Restart OpenCode to apply the changes.

   a. Open Settings and navigate to **Applications** > **OpenCode**.

   b. Click **Stop**, then **Resume**.

### Step 3: Confirm the plugin is loaded

After you enable OMO, open the OpenCode web UI. You should see Sisyphus, Prometheus, Atlas, and Hephaestus in the agent selector at the top.

![Agent selector with OMO enabled](/images/manual/use-cases/opencode-agents.png#bordered){width=80%}

You can also click the status icon in the upper-right corner to confirm that `oh-my-openagent` appears under **plugin**.

![oh-my-openagent listed under plugin in the status panel](/images/manual/use-cases/opencode-omo-status.png#bordered){width=50%}

## Use OMO

### Single-agent mode

With OMO enabled, the agent selector changes from Plan / Build to Sisyphus, Prometheus, Atlas, and Hephaestus. You don't need to start multi-agent collaboration every time. Without the `ultrawork` keyword, each agent runs on its own and behaves like the original Build or Plan mode, with no extra resource usage.

| Previous usage | Current setup | Notes |
|:---------------|:--------------|:------|
| Build (direct coding) | Select Sisyphus and chat normally | Default agent, same experience as the original Build mode. |
| Plan (plan first, then execute) | Select Prometheus | Prometheus interviews you and produces a plan. Type `/start-work` and Atlas takes over to execute. |
| Deep autonomous coding | Select Hephaestus | Independent agent suited for large, complex tasks. |

:::tip
Without `ultrawork` or `ulw`, OMO will not be triggered. Using an agent on its own is equivalent to the original Build or Plan mode.
:::

### Multi-agent collaboration mode

Type `ultrawork` (or the alias `ulw`) followed by a task description in the chat box. For example:

```text
ulw Implement a REST API user registration feature with email verfication.
```

OMO analyzes the task, assigns it to specialized agents, and runs them in parallel until the task is complete. When `ultrawork` is active, the colors flashing above the chat box indicate which agent is currently working.

![ultrawork running multi-agent collaboration](/images/manual/use-cases/opencode-ulw.png#bordered){width=70%}

### Prometheus planning mode

For large, complex tasks, plan first:

1. Switch to Prometheus in the agent selector.
2. Describe your task. For example:

   ```text
   Build a CLI tool that scans a folder and generates a summary report of its contents.
   ```

3. Answer Prometheus's questions to clarify scope and details.

   ![Prometheus asking clarifying questions](/images/manual/use-cases/opencode-prometheus-ask.png#bordered){width=70%}

4. After Prometheus has generated a Plan file, type `/start-work` to start execution.

### Common commands

| Command | Type | Description |
|:--------|:-----|:------------|
| `ultrawork` / `ulw` | Keyword | Starts multi-agent collaboration mode. Type it in the chat box. |
| `/start-work` | Slash command | Executes the plan produced by Prometheus. |
| `/init-deep` | Slash command | Generates an `AGENTS.md` context file for each project directory. |
| `/ulw-loop` | Slash command | Runs `ultrawork` in a loop until the task reaches 100% completion. |
| `/handoff` | Slash command | Produces a context summary so you can continue the work in a new session. |

## Use the built-in MCP servers

OMO registers three remote MCP servers that give agents real-time access to external knowledge. These servers aren't native to OpenCode. OMO injects them through a hook. After you enable OMO, you can find them under the MCP tab in Settings.

![Built-in MCP servers](/images/manual/use-cases/opencode-omo-mcps.png#bordered){width=50%}

| MCP server | Provider | Remote endpoint | API key required | Use case |
|:-----------|:---------|:----------------|:-----------------|:---------|
| context7 | Upstash / Context7 | `https://mcp.context7.com/mcp` | No. Free to use. Sign up for a higher quota. | Real-time queries against official documentation. |
| grep_app | Vercel / grep.app | `https://mcp.grep.app` | No. | GitHub-wide code search. |
| websearch | Exa AI (default) or Tavily (configurable) | `https://mcp.exa.ai/mcp?tools=web_search_exa` | No for basic use. Set `EXA_API_KEY` or `TAVILY_API_KEY` for higher-quality results. | Real-time web search. |

The examples below explicitly tell the agent which MCP tool to use. If the agent detects your intent on its own, it triggers the matching tool automatically.

### context7: live official documentation

The AI's training data has a cutoff date. When you work with a newer library version, the AI might generate deprecated APIs or non-existent methods. Context7 pulls the latest content directly from the source docs and injects it into the conversation.

For example, enter:

```text
Implement a form submission with React 19's useActionState Hook. use context7
```

![context7 MCP in action](/images/manual/use-cases/opencode-omo-context7.png#bordered){width=70%}

### grep_app: GitHub-wide code search

When the AI is unsure about best practices or API usage, it can search real code from GitHub projects for reference.

For example, enter:

```text
Find real examples of Drizzle ORM with PostgreSQL migrations. use grep_app
```

![grep_app MCP searching GitHub code](/images/manual/use-cases/opencode-omo-grep_app.png#bordered){width=70%}

### websearch: real-time web search

For example, enter:

```text
What are the major changes in Kubernetes 1.32 released in 2026? use websearch
```

![websearch MCP returning real-time results](/images/manual/use-cases/opencode-omo-websearch.png#bordered){width=70%}

### Disable a specific MCP server

1. Edit `~/.config/opencode/oh-my-openagent.json` and add the `disabled_mcps` field. For example, to disable websearch:

   ```json
   {
     "disabled_mcps": ["websearch"]
   }
   ```

   Allowed values: `"websearch"`, `"context7"`, `"grep_app"`. When you disable OMO (`OPENCODE_OMO=false`), all three MCP servers stop with it.

2. Restart OpenCode to apply the change.

   When you refresh the OpenCode UI, the websearch MCP server no longer appears, and only the other two MCP servers are listed.

   ![Disable an MCP server in oh-my-openagent.json](/images/manual/use-cases/opencode-omo-disable-mcp.png#bordered){width=50%}

## Replace the built-in context7 with the Olares self-hosted version

After you enable OMO, OpenCode shows a built-in `context7` MCP server that connects to the Context7 official cloud service (`https://mcp.context7.com/mcp`). The Olares Market also offers a standalone [Context7 app](context7.md), which is a self-hosted Context7 MCP server.

:::tip Compatibility with existing self-hosted context7

If you added `context7` to your `opencode.json` before installing OMO (pointing to the Olares self-hosted version), upgrading to the OMO version still works.

Your `opencode.json` entry has higher priority than OMO's built-in version and overrides it automatically. Your self-hosted version keeps working without changes.

To switch back to OMO's cloud version later, delete the `context7` entry from `opencode.json`. The OMO built-in version takes effect automatically.
:::

If you want all doc queries to go to your self-hosted instance, follow these steps.

1. In `~/.config/opencode/oh-my-openagent.json`, add the following to disable the built-in context7:

   ```json
   {
     "disabled_mcps": ["context7"]
   }
   ```

2. In `~/.config/opencode/opencode.json`, point to your self-hosted Context7 endpoint:

   ```json
   {
     "mcp": {
       "context7-local": {
         "type": "remote",
         "url": "<your-context7-endpoint>/mcp",
         "enabled": true
       }
     }
   }
   ```

3. Restart OpenCode to apply the changes.

After the restart, refresh the page. The MCP tab should now show the self-hosted `context7-local` plus OMO's other two MCP servers.

## Optional optimizations

### Regenerate the config for your subscriptions

The pre-installed config enables all providers. If you only subscribe to some of them, ask the agent to regenerate the optimal config:

```text
Run: npx oh-my-opencode install --no-tui --claude=yes --openai=no --gemini=no --copilot=yes
```

Adjust `yes`/`no` based on which providers you actually have. This updates `~/.config/opencode/oh-my-openagent.json`.

:::info
This command requires OpenCode version 1.4.0 or later. This config only affects subagent delegation. The main agent's model is still selected in the UI.
:::

### Override an agent's model manually

To set a specific model for an agent, edit `~/.config/opencode/oh-my-openagent.json` and update the `model` field:

```json
{
  "agents": {
    "sisyphus": { "model": "kimi-for-coding/k2p5" },
    "oracle": { "model": "openai/gpt-5.4", "variant": "high" }
  }
}
```

:::warning
Do not assign expensive models to tool-type agents such as Explore and Librarian. It raises token usage without improving results.
:::

## FAQ

### `ultrawork` doesn't respond or throws an error

Ask OpenCode to run the diagnostic:

```text
Run: npx oh-my-opencode doctor
```

The most common cause is that no provider authentication has been completed.

### How do settings in `oh-my-openagent.json` relate to the model picked in the UI?

The UI selection controls the main agent (the one you chat with). `oh-my-openagent.json` controls subagents, which are background tasks dispatched by the main agent through `delegate_task`. The two are independent.

### Can I use multiple providers at the same time?

Yes, and it's recommended. The more providers you configure, the more models are available in the UI selector, and the more fallback options are available for subagent delegation. You still need to pick a model for the main agent manually.

### `use context7`, `use grep_app`, or `use websearch` doesn't work

Confirm OMO is enabled. The MCP tab in Settings should show a green dot next to each server. If a server is gray or disconnected, the remote endpoint might be temporarily unavailable. Try again later.

### Plugin installation fails on first launch

Check the Pod logs for the `init-packages` container and search for `oh-my-opencode`. If npm downloads time out, confirm your Olares can reach the internet.

### Will my authentication data persist after a restart?

Yes. Authentication is stored under `~/.config/opencode/` and reloaded automatically after a restart.

### Will an Olares upgrade overwrite my config?

No. `oh-my-openagent.json` is only written when the file doesn't exist. The global `opencode.json` only manages the Olares-owned `plugin` and `instructions` entries. Other entries you added are untouched. Skill files (`web-preview`, `system-admin`) are force-updated to the latest version on every startup.

### Will I know when a model switches automatically?

Yes. `runtime_fallback` has `notify_on_fallback` enabled by default. A toast notification appears when a model switch happens, showing which model is now in use.

## Learn more

- [Set up OpenCode as your AI coding agent](opencode.md): Install OpenCode and connect it to Ollama.
- [Extend OpenCode with skills and plugins](opencode-extensions.md): Add capabilities through skills and plugins.
- [Connect AI coding assistants to up-to-date docs with Context7](context7.md#opencode): Register Context7 as a remote MCP server in OpenCode.
- [OMO overview](https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/overview.md): Official introduction to OMO's architecture and agents.
- [Agent-to-model matching reference](https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/agent-model-matching.md): Recommended models for each agent and task category.
