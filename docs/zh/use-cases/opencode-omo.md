---
outline: [2, 3]
description: 在 Olares 上的 OpenCode 中启用 oh-my-openagent (OMO) 以编排多个 AI 代理。使用 ultrawork 触发多代理协作，配置本地或外部模型，并使用内置 MCP 服务器。
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, oh-my-openagent, OMO, multi-agent, AI coding agent, ultrawork, MCP, self-hosted
app_version: "1.0.10"
doc_version: "1.0"
doc_updated: "2026-04-21"
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/opencode-omo.md)为准。
:::

# 使用 oh-my-openagent 编排多代理工作流

oh-my-openagent (OMO) 是 OpenCode 的多模型代理编排插件。启用后，你可以在 OpenCode 中使用关键词 `ultrawork`（或别名 `ulw`）触发多代理协作。Sisyphus、Hephaestus、Oracle 和 Atlas 等专业代理分工协作，共同处理复杂的编码任务。

:::warning
本指南侧重于本地模型设置。与付费云模型相比，完全在本地模型上运行 OMO 会明显降低编排质量和多代理协作速度。对于实际工作，我们推荐混合设置：为主代理使用付费云模型，为子代理使用 Ollama 本地模型。
:::

## 学习目标

在本教程结束时，你将学习如何：
- 在 Olares 上的 OpenCode 中启用 OMO。
- 将 OMO 配置为与本地 Ollama 模型、云模型或两者混合使用。
- 使用 `ultrawork` 关键词触发多代理协作。
- 使用内置的 context7、grep_app 和 websearch MCP 服务器。
- 将文档查询路由到自托管的 Context7， alongside OMO。

## 前提条件
- 你的 Olares 设备必须具有互联网访问权限。
- 在 Olares 上[安装 OpenCode](opencode.md)，chart 版本 1.0.6 或更高。
- 支持工具使用的本地模型，[已连接到 OpenCode](opencode.md#connect-to-a-custom-provider)。本指南以 Qwen3.5 27B Q4_K_M 和 Qwen3.5 9B Q4_K_M 为例。在 Olares 中，这些模型中的每一个都是单独的单一模型应用，因此你需要将它们添加为两个模型提供方。

  :::details 模型提供方配置

  ```json
  {
    "$schema": "https://opencode.ai/config.json",
    "disabled_providers": [],
    "provider": {
      "ollama-27b": {
        "name": "Olares Ollama Qwen3.5 27B",
        "npm": "@ai-sdk/openai-compatible",
        "models": {
          "qwen3.5:27b-q4_K_M": {
            "name": "Qwen3.5 27B"
          }
        },
        "options": {
          "baseURL": "http://94a553e00.shared.olares.com/v1"
        }
      },
      "ollama-9b": {
        "name": "Olares Ollama Qwen3.5 9B",
        "npm": "@ai-sdk/openai-compatible",
        "models": {
          "qwen3.5:9b": {
            "name": "Qwen3.5 9B"
          }
        },
        "options": {
          "baseURL": "http://bd5355000.shared.olares.com/v1"
        }
      }
    }
  }
  ```

  :::

  :::info 模型能力要求
  - 参数少于 7B 的本地模型通常无法正确处理 `tool_use` 或结构化输出。避免将它们用于任何代理。
  - 建议上下文窗口至少为 32K token。核心代理（Sisyphus、Hephaestus、Prometheus、Atlas）使用 64K 或更多时效果更好。
  - 本地模型，尤其是 Qwen 系列，有时无法正确生成 `write` 工具调用。这是已知的 Ollama 限制。
  :::

## 了解 Olares 上的 OMO

### 模型如何选择

OMO 将模型选择在你和配置文件之间分配：

- **主模型**：你直接与之聊天的代理的模型。你在 OpenCode UI 的模型选择器中挑选它。它独立于 `oh-my-openagent.json`。
- **子代理模型**：当主代理调用 `delegate_task` 将工作交给子代理（如 Explore 或 Librarian）时，该子代理使用 `~/.config/opencode/oh-my-openagent.json` 中 `agents` 字段下为该代理定义的模型。

默认的 `oh-my-openagent.json` 为每个子代理配备了多层回退链：付费主模型、付费备用模型，以及 OpenCode 提供的免费模型作为最后的回退。如果第一个模型在运行时不可用，`runtime_fallback` 会自动沿链向下走到下一个。大致如下：

```text
oh-my-openagent.json
└── agents
    └── <agent name>
        └── model chain:  paid primary → paid backup → free fallback
                              (runtime_fallback auto-switches on failure)
```

这使子代理委托即使在提供方宕机时也能继续工作。

### 配置文件中的默认模型

OMO 为每个代理的提示调整特定的模型系列。默认配置文件使用以下主模型和回退模型：

| 代理 | 推荐的 UI 模型（云） | 免费回退 |
|:------|:-----------------------------|:--------------|
| Sisyphus | Claude Opus 4.6 | Big Pickle |
| Hephaestus | GPT-5.4 | Big Pickle |
| Prometheus | Claude Opus 4.6 | Big Pickle |
| Atlas | Claude Sonnet 4.6 | Big Pickle |
| Oracle | GPT-5.4 | Big Pickle |
| Explore | Claude Haiku 4.5 | GPT-5 Nano |
| Librarian | MiniMax M2.7 | GPT-5 Nano |

有关每个代理和每个类别的推荐模型的完整列表，请参阅 [Agent-to-model 匹配参考](https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/agent-model-matching.md)。

### Olares 为 OMO 管理的文件

安装 OpenCode 时，Olares 会预配置 OMO。预安装项目位于全局 OpenCode 配置目录下，该目录映射到 Olares Files 中的 `Application/Data/opencode/.config/opencode/`。本指南的其余部分使用 `~/.config/opencode/` 作为此路径的简写。

:::details 文件清单

| 路径 | 作用 |
|:-----|:-------------|
| `~/.config/opencode/opencode.json`（`plugin` 字段） | 在全局配置中注册 `"oh-my-openagent"`。 |
| `~/.config/opencode/oh-my-openagent.json` | 每个代理的推荐模型以及运行时回退链。仅在首次安装时写入。 |
| `~/.config/opencode/olares-baseline-instructions.md` | Web 预览、包管理和 Olares 域约定。每次启动时刷新。 |
| `~/.config/opencode/skills/web-preview/` 和 `system-admin/` | 技能文件。每次启动时强制更新。 |
| `/usr/local/lib/node_modules/oh-my-opencode/`（容器内） | npm 包。持久化在系统快照中。 |

:::

所有 Olares 管理的设置都位于全局配置目录中。它们永远不会写入你的工作空间 `opencode.json`。如果你的工作空间配置仍然包含旧的 Olares 管理的指令，Olares 只会在它们与之前的预设完全匹配时才删除它们。你添加或修改的任何内容都会被保留。

`oh-my-openagent.json` 文件仅在首次安装时写入。如果你稍后通过 `install` 命令或手动更新它，你的更改不会被覆盖。

## 配置 OMO

### 启用 OMO

OMO 由 `OPENCODE_OMO` 环境变量控制：

- `false`（默认）：插件不加载。npm 包和配置文件保留在磁盘上，因此稍后启用它不需要再次下载。
- `true`：插件在全局配置中注册，并安装 npm 包。

要启用 OMO：

1. 打开 Settings 并导航到 **Applications** > **OpenCode** > **Manage environment variables**。
   ![Locate the OPENCODE_OMO environment variable](/images/manual/use-cases/opencode-env-var.png#bordered)

2. 找到 `OPENCODE_OMO` 环境变量并点击 <i class="material-symbols-outlined">edit_square</i>。

3. 在 **Value** 下拉菜单中，选择 `true`，然后点击 **Confirm**。
   ![Set OPENCODE_OMO to true](/images/manual/use-cases/opencode-enable-omo.png#bordered)

4. 点击 **Apply** 保存更改，并等待应用重启。

首次启动时，OMO 需要时间来完成其初始配置。这可能需要几分钟。

要稍后禁用 OMO，请重复上述步骤并将 `OPENCODE_OMO` 设置为 `false`。插件注册和 MCP 服务器停止，但 npm 包和 `oh-my-openagent.json` 配置保留在磁盘上，供下次重新启用。

### 配置本地模型

编辑 `~/.config/opencode/oh-my-openagent.json`，使 OMO 将子代理工作委托给你的本地 Ollama 模型。你可以在以下两种情况下跳过此步骤：

- **仅云**：你只计划使用云模型。在 OpenCode **Providers** 下添加你的 API key，然后继续。
- **仅免费回退**：你不想付费或自托管模型。OMO 使用 OpenCode 附带的免费回退模型，但预计响应会更慢且有限额。

:::tip 需要重启
每次编辑 `oh-my-openagent.json` 后重启 OpenCode 以应用更改。
:::

1. 打开 Olares Files，导航到 `Application/Data/opencode/.config/opencode/`，并找到 `oh-my-openagent.json`。
   ![Locate oh-my-openagent.json](/images/manual/use-cases/opencode-config-file.png#bordered)

2. 打开 `oh-my-openagent.json` 并点击 <i class="material-symbols-outlined">edit_square</i> 打开编辑器。

3. 更新 `agents` 部分，使子代理委托使用你的本地 Ollama 模型：

   a. 将每个代理的 `model` 字段指向你的 Ollama 模型。模型名称必须包含提供方前缀，该前缀必须与你在 `opencode.json` 中定义的提供方名称匹配。

   b. 为每个代理添加 `"stream": false`。

   例如，如果你的提供方名为 `ollama-27b` 和 `ollama-9b`：

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

   :::info `"stream": false` 要求
   Ollama 的流式模式返回 NDJSON，SDK 无法解析。使用工具的代理（尤其是 Librarian 和 Explore）如果缺少 `"stream": false`，会静默回退到链中的下一个模型。这是已知的 Ollama 限制。
   :::

4. 在 `categories` 部分，更新每个类别的 `model` 和 `fallback_models` 列表，使本地模型排在前面。例如：
  ```json
   "categories": {
     "visual-engineering": {
       "model": "ollama-27b/qwen3.5:27b-q4_K_M",
       "fallback_models": [
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
     "ultrabrain": {
       "model": "ollama-27b/qwen3.5:27b-q4_K_M",
       "fallback_models": [
         { "model": "ollama-9b/qwen3.5:9b" }
         // ... keep the remaining default fallback entries unchanged
       ]
     },
   }
   ```

  :::details `categories` 部分
   ```json
   "categories": {
     "visual-engineering": {
       "model": "ollama-27b/qwen3.5:27b-q4_K_M",
       "fallback_models": [
         { "model": "ollama-9b/qwen3.5:9b" },
         { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
         { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
         { "model": "anthropic/claude-opus-4-6", "variant": "max" },
         { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
         { "model": "opencode/big-pickle" }
       ]
     },
     "ultrabrain": {
       "model": "ollama-27b/qwen3.5:27b-q4_K_M",
       "fallback_models": [
         { "model": "ollama-9b/qwen3.5:9b" },
         { "model": "openai/gpt-5.4", "variant": "xhigh" },
         { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
         { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
         { "model": "anthropic/claude-opus-4-6", "variant": "max" },
         { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
         { "model": "opencode/big-pickle" }
       ]
     },
     "deep": {
       "model": "ollama-27b/qwen3.5:27b-q4_K_M",
       "fallback_models": [
         { "model": "ollama-9b/qwen3.5:9b" },
         { "model": "openai/gpt-5.4", "variant": "medium" },
         { "model": "github-copilot/gpt-5.4", "variant": "medium" },
         { "model": "anthropic/claude-opus-4-6", "variant": "max" },
         { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
         { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
         { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
         { "model": "opencode/big-pickle" }
       ]
     },
     "artistry": {
       "model": "ollama-27b/qwen3.5:27b-q4_K_M",
       "fallback_models": [
         { "model": "ollama-9b/qwen3.5:9b" },
         { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
         { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
         { "model": "anthropic/claude-opus-4-6", "variant": "max" },
         { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
         { "model": "openai/gpt-5.4" },
         { "model": "github-copilot/gpt-5.4" },
         { "model": "opencode/big-pickle" }
       ]
     },
     "quick": {
       "model": "ollama-9b/qwen3.5:9b",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "openai/gpt-5.4-mini" },
         { "model": "github-copilot/gpt-5.4-mini" },
         { "model": "anthropic/claude-haiku-4-5" },
         { "model": "github-copilot/claude-haiku-4.5" },
         { "model": "google/gemini-3-flash-preview" },
         { "model": "github-copilot/gemini-3-flash-preview" },
         { "model": "opencode/gpt-5-nano" }
       ]
     },
     "unspecified-low": {
       "model": "ollama-27b/qwen3.5:27b-q4_K_M",
       "fallback_models": [
         { "model": "ollama-9b/qwen3.5:9b" },
         { "model": "anthropic/claude-sonnet-4-6" },
         { "model": "github-copilot/claude-sonnet-4.6" },
         { "model": "openai/gpt-5.3-codex", "variant": "medium" },
         { "model": "google/gemini-3-flash-preview" },
         { "model": "github-copilot/gemini-3-flash-preview" },
         { "model": "opencode/big-pickle" }
       ]
     },
     "unspecified-high": {
       "model": "ollama-27b/qwen3.5:27b-q4_K_M",
       "fallback_models": [
         { "model": "ollama-9b/qwen3.5:9b" },
         { "model": "anthropic/claude-sonnet-4-6" },
         { "model": "github-copilot/claude-sonnet-4.6" },
         { "model": "openai/gpt-5.3-codex", "variant": "medium" },
         { "model": "google/gemini-3-flash-preview" },
         { "model": "github-copilot/gemini-3-flash-preview" },
         { "model": "opencode/big-pickle" }
       ]
     },
     "writing": {
       "model": "ollama-9b/qwen3.5:9b",
       "fallback_models": [
         { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
         { "model": "google/gemini-3-flash-preview" },
         { "model": "github-copilot/gemini-3-flash-preview" },
         { "model": "anthropic/claude-sonnet-4-6" },
         { "model": "github-copilot/claude-sonnet-4.6" },
         { "model": "opencode/big-pickle" }
       ]
     }
   }
   ```
  :::
5. 在同一文件中添加并发限制。本地 Ollama 服务器资源有限，多个代理同时访问它可能会耗尽 VRAM 或导致显著减速。

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

   这将大型模型限制为一个并发请求，并防止内存不足崩溃。

  :::details 本地模型的 `oh-my-openagent.json`

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
        "model": "ollama-27b/qwen3.5:27b-q4_K_M",
        "fallback_models": [
          { "model": "ollama-9b/qwen3.5:9b" },
          { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
          { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
          { "model": "anthropic/claude-opus-4-6", "variant": "max" },
          { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
          { "model": "opencode/big-pickle" }
        ]
      },
      "ultrabrain": {
        "model": "ollama-27b/qwen3.5:27b-q4_K_M",
        "fallback_models": [
          { "model": "ollama-9b/qwen3.5:9b" },
          { "model": "openai/gpt-5.4", "variant": "xhigh" },
          { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
          { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
          { "model": "anthropic/claude-opus-4-6", "variant": "max" },
          { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
          { "model": "opencode/big-pickle" }
        ]
      },
      "deep": {
        "model": "ollama-27b/qwen3.5:27b-q4_K_M",
        "fallback_models": [
          { "model": "ollama-9b/qwen3.5:9b" },
          { "model": "openai/gpt-5.4", "variant": "medium" },
          { "model": "github-copilot/gpt-5.4", "variant": "medium" },
          { "model": "anthropic/claude-opus-4-6", "variant": "max" },
          { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
          { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
          { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
          { "model": "opencode/big-pickle" }
        ]
      },
      "artistry": {
        "model": "ollama-27b/qwen3.5:27b-q4_K_M",
        "fallback_models": [
          { "model": "ollama-9b/qwen3.5:9b" },
          { "model": "google/gemini-3.1-pro-preview", "variant": "high" },
          { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "high" },
          { "model": "anthropic/claude-opus-4-6", "variant": "max" },
          { "model": "github-copilot/claude-opus-4.6", "variant": "max" },
          { "model": "openai/gpt-5.4" },
          { "model": "github-copilot/gpt-5.4" },
          { "model": "opencode/big-pickle" }
        ]
      },
      "quick": {
        "model": "ollama-9b/qwen3.5:9b",
        "fallback_models": [
          { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
          { "model": "openai/gpt-5.4-mini" },
          { "model": "github-copilot/gpt-5.4-mini" },
          { "model": "anthropic/claude-haiku-4-5" },
          { "model": "github-copilot/claude-haiku-4.5" },
          { "model": "google/gemini-3-flash-preview" },
          { "model": "github-copilot/gemini-3-flash-preview" },
          { "model": "opencode/gpt-5-nano" }
        ]
      },
      "unspecified-low": {
        "model": "ollama-27b/qwen3.5:27b-q4_K_M",
        "fallback_models": [
          { "model": "ollama-9b/qwen3.5:9b" },
          { "model": "anthropic/claude-sonnet-4-6" },
          { "model": "github-copilot/claude-sonnet-4.6" },
          { "model": "openai/gpt-5.3-codex", "variant": "medium" },
          { "model": "google/gemini-3-flash-preview" },
          { "model": "github-copilot/gemini-3-flash-preview" },
          { "model": "opencode/big-pickle" }
        ]
      },
      "unspecified-high": {
        "model": "ollama-27b/qwen3.5:27b-q4_K_M",
        "fallback_models": [
          { "model": "ollama-9b/qwen3.5:9b" },
          { "model": "anthropic/claude-sonnet-4-6" },
          { "model": "github-copilot/claude-sonnet-4.6" },
          { "model": "openai/gpt-5.3-codex", "variant": "medium" },
          { "model": "google/gemini-3-flash-preview" },
          { "model": "github-copilot/gemini-3-flash-preview" },
          { "model": "opencode/big-pickle" }
        ]
      },
      "writing": {
        "model": "ollama-9b/qwen3.5:9b",
        "fallback_models": [
          { "model": "ollama-27b/qwen3.5:27b-q4_K_M" },
          { "model": "google/gemini-3-flash-preview" },
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
6. 保存文件。
7. 重启 OpenCode 以应用更改。

   a. 打开 Settings 并导航到 **Applications** > **OpenCode**。

   b. 点击 **Stop**，然后 **Resume**。


### 确认插件已加载

启用 OMO 后，打开 OpenCode Web UI。你应该会在顶部的代理选择器中看到 Sisyphus、Prometheus、Atlas 和 Hephaestus。

![Agent selector with OMO enabled](/images/manual/use-cases/opencode-agents.png#bordered){width=80%}

你也可以点击右上角的状态图标来确认 `oh-my-openagent` 出现在 **plugin** 下。

![oh-my-openagent listed under plugin in the status panel](/images/manual/use-cases/opencode-omo-status.png#bordered){width=50%}

## 使用 OMO

OMO 附带一组专业代理，每个代理的提示都针对特定角色进行了调整。你从代理选择器中选择 Sisyphus、Prometheus、Atlas 或 Hephaestus 作为主代理。其余的代理在多代理协作期间自动调度。

| 代理 | 角色 | 如何访问 |
|:------|:-----|:--------------|
| Sisyphus | 主编排器。r>在 `ultrawork` 模式下规划、委托给子代理并驱动并行<br>执行。 | 代理选择器 |
| Hephaestus | 自主深度工作者。r>处理架构推理和长时间<br>自主编码会话。 | 代理选择器 |
| Prometheus | 战略规划师。r>就需求采访你并生成<br>详细计划。 | 代理选择器 |
| Atlas | 计划执行者。r>接管 Prometheus 计划并执行它。 | 代理选择器，或在规划后使用 `/start-work` |
| Oracle | 只读架构顾问。r>就架构决策和复杂<br>调试提供建议，不修改文件。 | 在 `ultrawork` 中自动调度 |
| Explore | 代码搜索代理。r>检索整个仓库的代码。 | 在 `ultrawork` 中自动调度 |
| Librarian | 文档搜索代理。r>搜索文档和代码。 | 在 `ultrawork` 中自动调度 |

其他支持代理（如用于计划差距分析的 Metis、用于计划审查的 Momus，以及用于视觉输入的 Multimodal Looker）在需要时由 Sisyphus 自动调度。

### 单代理模式

启用 OMO 后，代理选择器从 Plan/Build 变为 Sisyphus、Prometheus、Atlas 和 Hephaestus。你不需要每次都启动多代理协作。没有 `ultrawork` 关键词时，每个代理独立运行，行为类似于原始的 Build 或 Plan 模式，没有额外的资源使用。

| 任务 | 如何操作 |
|:-----|:-------------|
| 直接编码<br>（以前的 Build） | 选择 Sisyphus 并正常聊天。与<br>原始的 Build 模式体验相同。 |
| 先规划，再执行<br>（以前的 Plan） | 选择 Prometheus。它采访你并生成<br>计划。输入 `/start-work` 交给 Atlas。 |
| 深度自主编码 | 选择 Hephaestus。适合大型、复杂任务，可<br>自主运行。 |

### 多代理协作模式

在聊天框中输入 `ultrawork`（或别名 `ulw`）后跟任务描述。例如：

```text
ulw Implement a REST API user registration feature with email verification.
```

OMO 分析任务，将其分配给专业代理，并并行运行它们直到任务完成。当 `ultrawork` 激活时，聊天框上方闪烁的颜色指示当前正在工作的代理。

![ultrawork running multi-agent collaboration](/images/manual/use-cases/opencode-ulw.png#bordered){width=70%}

### Prometheus 规划模式

:::info 意外的技能触发
任务描述中的某些关键词可能会触发 OpenCode 调用 `/web-preview` 技能。如果发生这种情况，请要求代理忽略它并继续规划。
:::

对于大型、复杂任务，先规划：

1. 在代理选择器中切换到 Prometheus。
2. 描述你的任务。例如：

   ```text
   Build a CLI tool that scans a folder and generates a summary report of its contents.
   ```

3. 回答 Prometheus 的问题以明确范围和细节。

   ![Prometheus asking clarifying questions](/images/manual/use-cases/opencode-prometheus-ask.png#bordered){width=70%}

4. Prometheus 在 `.sisyphus/plans/` 下生成计划文件后，输入 `/start-work` 将执行交给 Atlas。如果没有计划文件，`/start-work` 没有内容可传递给 Atlas，因此不会执行任何操作。

   :::tip 规划期间请耐心等待
   如果 Prometheus 没有立即产生输出，或代理状态显示 `background output`，代理仍在工作。等待它完成后再发送另一条消息。
   :::

### 常用命令

| 命令 | 描述 |
|:--------|:------------|
| `ultrawork` / `ulw` | 启动多代理协作模式。在聊天框中作为关键词输入。 |
| `/start-work` | 执行 Prometheus 生成的计划。 |
| `/init-deep` | 为每个项目目录生成 `AGENTS.md` 上下文文件。 |
| `/ulw-loop` | 循环运行 `ultrawork` 直到任务达到 100% 完成。 |
| `/handoff` | 生成上下文摘要，以便你可以在新会话中继续工作。 |

## 使用内置 MCP 服务器

OMO 注册了三个远程 MCP 服务器，为代理提供对外部知识的实时访问。这些服务器不是 OpenCode 原生的。OMO 通过钩子注入它们。启用 OMO 后，你可以在 Settings 的 MCP 标签页下找到它们。

![Built-in MCP servers](/images/manual/use-cases/opencode-omo-mcps.png#bordered){width=50%}

| MCP 服务器 | 使用场景 | API key |
|:-----------|:---------|:--------|
| context7 | 针对官方<br>文档的实时查询。 | 基本使用不需要。<br>添加 key 以获得更高配额。 |
| grep_app | GitHub 全站代码搜索。 | 不需要。 |
| websearch | 实时网页搜索。 | 基本使用不需要。<br>添加 key 以获得更高配额，<br>或从 Exa 切换到 Tavily。 |

下面的示例显式命名 MCP 工具。你也可以让代理在检测到意图时自动选择。

要提高 context7 或 websearch 的配额，或将 websearch 提供方从 Exa 切换到 Tavily，请参阅 [配置 API key](#configure-api-keys-and-switch-providers)。

### context7：实时官方文档

AI 的训练数据有截止日期。当你使用较新的库版本时，AI 可能会生成已弃用的 API 或不存在的方法。Context7（由 Upstash 在 `https://mcp.context7.com/mcp` 提供）直接从源文档中提取最新内容并将其注入对话中。

例如，输入：

```text
Implement a form submission with React 19's useActionState Hook. use context7
```

![context7 MCP in action](/images/manual/use-cases/opencode-omo-context7.png#bordered){width=70%}

### grep_app：GitHub 全站代码搜索

当 AI 不确定最佳实践或 API 用法时，它可以搜索 GitHub 项目中的真实代码作为参考。此 MCP 服务器由 Vercel 在 `https://mcp.grep.app` 提供。

例如，输入：

```text
Find real examples of Drizzle ORM with PostgreSQL migrations. use grep_app
```

![grep_app MCP searching GitHub code](/images/manual/use-cases/opencode-omo-grep_app.png#bordered){width=70%}

### websearch：实时网页搜索

此 MCP 服务器默认使用 Exa AI (`https://mcp.exa.ai/mcp?tools=web_search_exa`)，可以通过配置切换到 Tavily。

例如，输入：

```text
What are the major changes in Kubernetes 1.32 released in 2026? use websearch
```

![websearch MCP returning real-time results](/images/manual/use-cases/opencode-omo-websearch.png#bordered){width=70%}

### 配置 API key 和切换提供方

context7 和 websearch 都可以在没有 key 的情况下工作，但添加 key 会提高你的配额并解除速率限制。你也可以将 websearch 从 Exa（默认）切换到 Tavily。grep_app 没有 key。

要配置任何这些，请在 `~/.config/opencode/oh-my-openagent.json` 的 `mcp` 字段下添加内置 MCP 定义的覆盖。覆盖会完全替换内置定义，因此 `type`、`url` 和 `enabled` 都必须存在。如果你设置多个覆盖，请将它们放在同一个 `mcp` 对象下。

:::tip 需要重启
每次编辑 `oh-my-openagent.json` 后重启 OpenCode，以使新定义生效。
:::

#### 为 context7 添加 API key

从 [context7.com](https://context7.com/) 获取 key，然后将其添加到 `oh-my-openagent.json`：

```json
{
  "mcp": {
    "context7": {
      "type": "remote",
      "url": "https://mcp.context7.com/mcp",
      "enabled": true,
      "headers": {
        "CONTEXT7_API_KEY": "YOUR_CONTEXT7_KEY"
      }
    }
  }
}
```

要验证 key 是否激活，请要求代理使用 `use context7` 查找最近发布的库，然后检查你的 Context7 仪表板以了解使用量的增加。

#### 为 websearch 添加 API key（Exa）

默认情况下，websearch 使用 Exa。从 [exa.ai](https://exa.ai/) 获取 key，然后将其添加到 `oh-my-openagent.json`：

```json
{
  "mcp": {
    "websearch": {
      "type": "remote",
      "url": "https://mcp.exa.ai/mcp?tools=web_search_exa",
      "enabled": true,
      "headers": {
        "x-api-key": "YOUR_EXA_KEY"
      }
    }
  }
}
```

要验证 key 是否激活，请要求代理运行网页搜索，并在成功搜索后检查 [dashboard.exa.ai](https://dashboard.exa.ai/) 以了解使用量的增加。

#### 将 websearch 从 Exa 切换到 Tavily

Tavily 是一个替代网页搜索提供方，每月提供 1,000 次免费搜索。从 [app.tavily.com](https://app.tavily.com/) 获取 key，然后将 `websearch` 指向 Tavily：

```json
{
  "mcp": {
    "websearch": {
      "type": "remote",
      "url": "https://mcp.tavily.com/mcp/",
      "enabled": true,
      "headers": {
        "Authorization": "Bearer YOUR_TAVILY_KEY"
      }
    }
  }
}
```

要验证，请从代理运行网页搜索，并检查你的 [Tavily 仪表板](https://app.tavily.com/) 以了解使用量的增加。要切换回 Exa，请用上面的 Exa 定义替换此块。

### 禁用特定 MCP 服务器

<tabs>
<template #Disable-from-UI>

1. 点击右上角的状态图标，然后选择 **MCP**。
2. 关闭特定 MCP 服务器。

</template>
<template #Disable-via-config-file>

1. 编辑 `~/.config/opencode/oh-my-openagent.json` 并添加 `disabled_mcps` 字段。例如，要禁用 websearch：

   ```json
   {
     "disabled_mcps": ["websearch"]
   }
   ```

   允许的值：`"websearch"`、`"context7"`、`"grep_app"`。当你禁用 OMO（`OPENCODE_OMO=false`）时，所有三个 MCP 服务器都会随之停止。

2. 重启 OpenCode 以应用更改。

   当你刷新 OpenCode UI 时，websearch MCP 服务器不再出现，只列出其他两个 MCP 服务器。

   ![Disable an MCP server in oh-my-openagent.json](/images/manual/use-cases/opencode-omo-disable-mcp.png#bordered){width=50%}

</template>
</tabs>

## 使用自托管的 Context7

OMO 的内置 `context7` 服务器在 `https://mcp.context7.com/mcp` 查询公共 Context7 云。要使用你自己的实例，请在 Olares 上安装 Context7 并在 OpenCode 中注册它。有关步骤，请参阅 [将 Context7 连接到 OpenCode](context7.md#opencode)。

`opencode.json` 中的条目优先于 OMO 的内置 MCP 服务器。如果你已经在 `opencode.json` 中有自托管的 Context7 条目，启用 OMO 时会发生什么取决于条目名称：

- 如果名称不是 `context7`，你的自托管实例和 OMO 的内置 `context7` 服务器都会无冲突地加载。
- 如果名称是 `context7`，它会与 OMO 的内置服务器冲突。在 `opencode.json` 中重命名你的条目（例如，改为 `context7-local`），或添加 `"disabled_mcps": ["context7"]` 到 `~/.config/opencode/oh-my-openagent.json` 以禁用 OMO 的内置服务器。

## 高级配置

### 根据你的订阅重新生成 `oh-my-openagent.json`

预安装的配置启用了所有提供方。如果你只订阅了其中一些，请要求代理重新生成最佳配置。

使用 `--claude`、`--openai`、`--gemini` 和 `--copilot` 标志（每个设置为 `yes` 或 `no`）来匹配你的实际订阅。例如，如果你订阅了 Claude 和 Copilot 但没有订阅 OpenAI 或 Gemini，请在聊天窗口中输入：

```text
Run: npx oh-my-opencode install --no-tui --claude=yes --openai=no --gemini=no --copilot=yes
```

根据你实际拥有的提供方调整 `yes`/`no`。这会更新 `~/.config/opencode/oh-my-openagent.json`。

:::info 版本要求
此命令需要 OpenCode 版本 1.4.0 或更高。它会重写所有代理的配置，但只有子代理模型从文件中生效。主代理始终使用你在 OpenCode UI 中选择的模型。
:::

### 手动覆盖代理的模型

要为代理设置特定模型，请编辑 `~/.config/opencode/oh-my-openagent.json` 并更新 `model` 字段：

```json
{
  "agents": {
    "sisyphus": { "model": "kimi-for-coding/k2p5" },
    "oracle": { "model": "openai/gpt-5.4", "variant": "high" }
  }
}
```

:::warning
Explore 和 Librarian 是使用工具的代理。为它们分配高能力模型会增加 token 使用量而不会改善结果。
:::

## 常见问题

### `ultrawork` 无响应或抛出错误

要求 OpenCode 运行诊断：

```text
Run: npx oh-my-opencode doctor
```

通常这意味着没有提供方已通过身份验证。

### `oh-my-openagent.json` 中的设置与 UI 中选择的模型有什么关系？

UI 选择控制主代理（你与之聊天的代理）。`oh-my-openagent.json` 控制子代理，它们是通过 `delegate_task` 由主代理分派的后台任务。两者是独立的。

### 我可以同时使用多个提供方吗？

可以，而且推荐这样做。你配置的提供方越多，UI 选择器中可用的模型就越多，子代理委托可用的回退选项也越多。你仍然需要手动为主代理选择模型。

### `use context7`、`use grep_app` 或 `use websearch` 不起作用

确认 OMO 已启用。Settings 中的 MCP 标签页应在每个服务器旁边显示绿色圆点。如果服务器是灰色或断开连接的，远程端点可能暂时不可用。稍后再试。

### 首次启动时插件安装失败

检查 `init-packages` 容器的 Pod 日志并搜索 `oh-my-opencode`。如果 npm 下载超时，请确认你的 Olares 可以访问互联网。

### 重启后我的身份验证数据会保留吗？

会。身份验证存储在 `~/.config/opencode/` 下，重启后自动重新加载。

### Olares 升级会覆盖我的配置吗？

不会。`oh-my-openagent.json` 仅在文件不存在时写入。全局 `opencode.json` 只管理 Olares 拥有的 `plugin` 和 `instructions` 条目。你添加的任何其他条目都会被保留。技能文件（`web-preview`、`system-admin`）在每次启动时强制更新到最新版本。

### 我会知道模型何时自动切换吗？

会。当 `runtime_fallback` 切换到另一个模型时，会出现一个 toast 通知，显示当前正在使用的模型。

## 了解更多

- [将 OpenCode 设置为你的 AI 编码代理](opencode.md)：安装 OpenCode 并将其连接到 Ollama。
- [使用技能和插件扩展 OpenCode](opencode-extensions.md)：通过技能和插件添加功能。
- [使用 Context7 将 AI 编码助手连接到最新文档](context7.md#opencode)：在 OpenCode 中将 Context7 注册为远程 MCP 服务器。
- [OMO 概述](https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/overview.md)：OMO 架构和代理的官方介绍。
- [Agent-to-model 匹配参考](https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/agent-model-matching.md)：每个代理和任务类别的推荐模型。
