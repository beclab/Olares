---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/opencode
outline: deep
description: 在 Olares 上设置 OpenCode 以运行 AI 编码代理。将其连接到本地模型或 OpenAI，并使用自然语言编写、测试和管理代码。
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, AI coding agent, Qwen3.8, OpenAI, ChatGPT, self-hosted, code generation, TUI
app_version: "1.0.11"
doc_version: "1.3"
doc_updated: "2026-09-23"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/opencode.md)为准。
:::

# 将 OpenCode 设置为你的 AI 编码代理

<VersionRouteSelect />

OpenCode 是一个 AI 驱动的编码代理，允许你通过自然语言编写、测试和管理代码。它支持多个 AI 提供方，并可以从聊天界面运行 shell 命令、创建文件和安装开发环境。

在 Olares 上，你可以通过两种方式使用 OpenCode：

- **浏览器**：在 Olares 上将 OpenCode 安装为应用，并通过浏览器访问它。
- **本地 CLI**：在你的计算机上安装 OpenCode，并将其连接到 Olares 上的模型，以获得原生终端体验。

## 学习目标

在本教程结束时，你将学习如何：
- 在 Olares 上安装 OpenCode，并将其连接到本地模型或 OpenAI。
- 通过聊天界面或基于终端的 UI（TUI）创建项目和运行编码任务。
- 通过 LarePass VPN 从你的本地计算机使用 OpenCode，或从 VS Code 内部使用。
- 编辑 OpenCode 配置文件以管理提供方、模型和工具。

## 前提条件

开始前，你需要：

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- 从 Market 安装应用的管理员权限。
- 其他条件取决于 OpenCode 的使用方式：

   | 使用方式 | 所需条件 |
   | :--- | :--- |
   | 在浏览器中使用本地模型 | 从 Market 安装 Qwen3.8-27B (llama.cpp) |
   | 在浏览器中使用 OpenAI | ChatGPT Plus/Pro 账号或 OpenAI API key |
   | 在计算机上使用 OpenCode CLI | Olares 上已配置 Router 默认聊天模型，并已创建 Router API 密钥。 |

## 在浏览器中运行 OpenCode

此选项在你的 Olares 设备上将 OpenCode 安装为应用。你通过浏览器访问它。

### 安装 OpenCode

1. 打开 Market 并搜索 "OpenCode"。
   ![Install OpenCode](/images/manual/use-cases/opencode.png#bordered)

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

安装后，你将在 Launchpad 上看到两个图标：
- OpenCode：OpenCode 的主界面。
- OpenCode Terminal：OpenCode 的命令行终端。如果你更喜欢在 TUI 中工作，请使用此选项。

### 获取模型连接信息

:::tip 计划使用 oh-my-openagent？
对于 [oh-my-openagent](opencode-omo.md) 下的多本地模型设置，请使用单一模型应用。每个模型都有自己的端点，这允许你将其注册为单独的提供方，并调整每个模型的设置（如并发限制）。
:::

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

本示例中，OpenCode 通过 OpenAI-compatible API 格式连接 Qwen3.8-27B：

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### 连接到自定义提供方

在 OpenCode 中将 Qwen3.8-27B 添加为自定义提供方。

1. 在 OpenCode 中，点击左下角的 <i class="material-symbols-outlined">settings</i>。
   ![Open OpenCode settings](/images/manual/use-cases/opencode-settings.png#bordered)

2. 选择 **Providers**，然后向下滚动并选择 **Custom Provider** 旁边的 **Connect**。
   ![Select custom provider](/images/manual/use-cases/opencode-custom-provider.png#bordered)

3. 输入以下详细信息：
   - **Provider ID**：输入唯一标识符，例如 `olares`。
   - **Display name**：输入提供方列表中显示的名称，例如 `Router chat`。
   - **Base URL**：粘贴从 Router 复制的 Base URL，并按显示内容原样使用。
   - **Models**：
     - **Model ID**：填写 `default-chat`。
     - **Display Name**：输入该模型的显示名称，例如 `Router chat`。

4. 点击 **Submit** 保存配置。你新添加的提供方将出现在提供方列表中。

   <!--
   TODO: 素材清单 10，待补 Router 截图：opencode-provider-router.png；替换下方旧图后再取消注释。
   ![Provider list](/images/manual/use-cases/opencode-provider-list.png#bordered)
   -->


### 连接到 OpenAI

OpenCode 支持两种连接 OpenAI 的实用方式：

- **浏览器登录**：推荐用于 ChatGPT Plus 或 Pro 账户。
- **API key**：简单直接。即使你也有 ChatGPT 订阅，OpenAI 也会根据 token 使用量单独计费 API 使用。

:::warning 不要使用无头登录
无头登录方法容易出现 OpenAI 安全检查导致的已知故障。失败后，登录流可能会在短时间内无法恢复。请改用浏览器登录或 API key。
:::

<tabs>
<template #Browser-sign-in>

当你希望 OpenCode 通过你的 ChatGPT Plus 或 Pro 账户连接时，请使用此方法。

因为 OpenCode 在 Olares 容器内运行，以 `localhost:1455` 开头的浏览器回调 URL 无法直接从你的浏览器访问 OpenCode。启动浏览器流程，然后将最终的回调 URL 手动中继回 OpenCode。

:::warning 保持授权会话活动
- 在 5 分钟内完成步骤。如果会话过期，请重新开始。
- 在连接成功之前，保持 OpenCode 中的 **Connect OpenAI** 对话框打开。
- 每个回调 URL 只能使用一次。如果尝试失败，请关闭对话框并重复流程以获取新的 URL。
:::

1. 在 OpenCode 中，点击 <i class="material-symbols-outlined">settings</i> > **Providers**，然后向下滚动并选择 **OpenAI** 旁边的 **Connect**。
   ![Add the OpenAI provider](/images/manual/use-cases/opencode-openai-provider.png#bordered)

2. 对于登录方法，选择 **ChatGPT Pro/Plus (browser)**。
   ![Select browser sign-in for OpenAI](/images/manual/use-cases/opencode-openai-browser-signin.png#bordered){width=70%}

3. 在授权对话框中，点击授权链接以在浏览器中打开它。
   ![Open the OpenAI authorization link](/images/manual/use-cases/opencode-openai-authorization-link.png#bordered){width=70%}

4. 使用你的 ChatGPT 账户登录。

   登录后，浏览器会重定向到以 `http://localhost:1455` 开头的 URL，并可能显示页面无法访问。这是预期的。不要关闭页面。

5. 从浏览器地址栏复制完整的 URL，从 `http://localhost:1455` 到 URL 的末尾。
   ![Open the OpenAI authorization link](/images/manual/use-cases/opencode-openai-auth-url.png#bordered)

6. 从 Launchpad 打开 OpenCode Terminal。

7. 使用你复制的完整 URL 运行 `curl`：

   ```bash
   curl '<paste-the-full-url-you-copied>'
   ```
   :::warning
   运行 `curl` 时，将复制的 URL 用单引号（`''`）包裹。否则，shell 会将 `&` 视为特殊字符并截断 URL。
   :::

8. 等待终端返回 `Authorization Successful`。OpenCode 中的等待对话框应自动更改为已连接。
   ![OpenAI authorization succeeds in OpenCode Terminal](/images/manual/use-cases/opencode-openai-terminal-success.png#bordered)

9. 刷新 OpenCode 页面以加载新的 OpenAI 模型。
   ![OpenAI provider connected](/images/manual/use-cases/opencode-openai-connected.png#bordered){width=70%}
</template>

<template #API-key>

当你希望直接连接 OpenAI API 时，请使用此方法。它通常比浏览器登录更容易设置，但 OpenAI 会根据 token 单独计费 API 使用。

1. 从你的 OpenAI 账户创建或复制 OpenAI API key。
2. 在 OpenCode 中，点击 <i class="material-symbols-outlined">settings</i> > **Providers**，然后向下滚动并选择 **OpenAI** 旁边的 **Connect**。
   ![Add the OpenAI provider](/images/manual/use-cases/opencode-openai-provider.png#bordered)

3. 对于登录方法，选择 **API key**。
   ![Select API key sign-in for OpenAI](/images/manual/use-cases/opencode-openai-api-key-signin.png#bordered){width=70%}

4. 粘贴你的 OpenAI API key 并点击 **Continue**。
   ![Enter the OpenAI API key](/images/manual/use-cases/opencode-openai-api-key.png#bordered){width=70%}

5. 刷新 OpenCode 页面以加载新的 OpenAI 模型。
   ![OpenAI provider connected](/images/manual/use-cases/opencode-openai-connected.png#bordered){width=70%}
</template>
</tabs>

### 创建项目

默认工作空间是 Files 中的 `Home/Code`。点击左侧导航栏中的 **+** 创建项目。

![Create a project](/images/manual/use-cases/opencode-create-project.png#bordered)

要处理多个项目，首先在 `Home/Code` 下创建子文件夹：

1. 打开 Files 并导航到 `Home/Code/`。
2. 为每个项目创建一个子文件夹。
   ![Create subfolders in Files](/images/manual/use-cases/opencode-create-subfolders-in-files.png#bordered)

3. 返回 OpenCode，点击 **+**，然后打开带有子文件夹名称的项目。
   ![Open project from subfolder](/images/manual/use-cases/opencode-create-more-projects.png#bordered)

### 开始编码

你现在可以通过聊天界面与 OpenCode 交互。

1. 选择一个项目以打开编码代理界面。
2. 在聊天框下方，选择 **Big Pickle** 打开模型选择器，然后从列表中选择 **Router chat**。
3. 用自然语言输入编码任务。
   ![Code generation](/images/manual/use-cases/opencode-code-generation.png#bordered)

4. 点击 <i class="material-symbols-outlined">folder_open</i> 打开文件浏览器并查看生成的代码。
   ![View files](/images/manual/use-cases/opencode-view-files.png#bordered)

5. 使用 `@` 在聊天窗口中提及文件，并要求 OpenCode 编辑它。
   ![Mention file in chat](/images/manual/use-cases/opencode-mention.png#bordered)

### 使用 TUI

OpenCode 还提供基于终端的 UI（TUI）。你可以通过两种方式启动它：

- **在 OpenCode UI 内部**：在 OpenCode UI 底部打开终端面板，类似于 VS Code 的集成终端。代理聊天在上方保持可见，而 TUI 在面板中运行。
- **从 OpenCode Terminal**：从 Launchpad 打开 OpenCode Terminal。它直接打开到命令行，没有聊天 UI。

<tabs>
<template #Inside-the-OpenCode-UI>

1. 在 OpenCode 中，点击右上角的 <i class="material-symbols-outlined">terminal_2</i> 在窗口底部打开终端面板。
   ![Open the terminal panel in OpenCode](/images/manual/use-cases/opencode-web-terminal.png#bordered)

2. 在终端面板中，运行 `opencode` 启动 TUI。
   ![Launch TUI in the OpenCode UI](/images/manual/use-cases/opencode-web-launch-tui.png#bordered)

3. 使用 `/models` 命令切换到 **Router chat**。

   <!--
   TODO: 素材清单 11，待补 Router 截图：opencode-web-tui-model-router.png；替换下方旧图后再取消注释。
   ![Select model in TUI](/images/manual/use-cases/opencode-web-tui-select-model.png#bordered)
   -->


4. 直接输入你的提示。例如，要求 OpenCode 改进代码。
   ![Chat in TUI](/images/manual/use-cases/opencode-web-tui-chat.png#bordered)

</template>
<template #From-OpenCode-Terminal>

1. 从 Launchpad 打开 OpenCode Terminal。

2. 运行 `opencode` 启动 TUI。
   ![Launch TUI in OpenCode Terminal](/images/manual/use-cases/opencode-terminal-launch-tui.png#bordered)

3. 使用 `/models` 命令切换到 **Router chat**。

   <!--
   TODO: 素材清单 12，待补 Router 截图：opencode-terminal-model-router.png；替换下方旧图后再取消注释。
   ![Select model in OpenCode Terminal](/images/manual/use-cases/opencode-terminal-select-model.png#bordered)
   -->


4. 直接输入你的提示。例如，使用 `@` 提及文件并向 OpenCode 询问它。
   ![Chat in OpenCode Terminal](/images/manual/use-cases/opencode-terminal-mention.png#bordered)

</template>
</tabs>

## 从你的计算机运行 OpenCode

此可选流程在你的计算机上安装 OpenCode CLI，并通过 API 密钥连接 Olares Router。

### 安装 OpenCode CLI

1. 使用官方安装程序安装 CLI：

   ```bash
   curl -fsSL https://opencode.ai/install | bash
   ```

2. （可选）如果你遇到 `No config file found for zsh` 错误，请通过运行以下命令将 export 行添加到你的 `~/.zshrc` 文件：

   ```bash
   echo 'export PATH="$HOME/.opencode/bin:$PATH"' >> ~/.zshrc
   ```

   错误消息示例：

   ```text
   No config file found for zsh. You may need to manually add to PATH:
   export PATH=/Users/{username}/.opencode/bin:$PATH
   ```

3. 重新加载你的 shell 配置：

   ```bash
   source ~/.zshrc
   ```

4. 验证安装：

   ```bash
   opencode --version
   ```

   显示版本，例如 `1.15.6`。

5. 运行以下命令初始化 OpenCode 并创建配置文件：

   ```bash
   opencode
   ```

   配置文件在 `~/.config/opencode/opencode.jsonc` 创建。

### 获取 Router 连接信息

1. 打开 Router，在 **Default models** 页面将 Qwen3.8-27B (llama.cpp) 设为默认聊天模型。然后在 **LLM** 页面找到该模型，点击 **View connection example**。
2. 在 **How to call this model** 窗口选择 **Devices in LAN** 或 **Remote**，复制 Base URL。
3. 在 **API keys** 页面创建密钥，供电脑上的 OpenCode 使用。使用 VPN 也需要此密钥。

### 配置连接

1. 在运行 OpenCode 的终端中，将 `OLARES_ROUTER_API_KEY` 环境变量设为刚创建的 Router 密钥。
2. 打开 `~/.config/opencode/opencode.jsonc`，添加以下配置。将 `<router-base-url>` 替换为从 Router 复制的完整地址，保留 `/v1`。

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "olares/default-chat",
  "provider": {
    "olares": {
      "name": "Olares Router",
      "npm": "@ai-sdk/openai-compatible",
      "options": {
        "baseURL": "<router-base-url>",
        "apiKey": "{env:OLARES_ROUTER_API_KEY}"
      },
      "models": {
        "default-chat": { "name": "Router chat" }
      }
    }
  }
}
```

3. 保存文件，重新启动 OpenCode。

### 启动 OpenCode TUI

1. 在你的终端中，运行 `opencode` 启动 TUI：
   ![Launch OpenCode TUI in terminal](/images/manual/use-cases/opencode-terminal-tui.png#bordered)

2. 使用 `/models` 命令切换到 **Router chat**。

   <!--
   TODO: 素材清单 13，待补 Router 截图：opencode-local-tui-model-router.png；替换下方旧图后再取消注释。
   ![Select model in terminal TUI](/images/manual/use-cases/opencode-terminal-tui-select-model.png#bordered)
   -->


3. 开始与你的自托管模型聊天。
   ![Chat in terminal TUI](/images/manual/use-cases/router-client-connect-opencode.png#bordered)

:::info 首次连接
首次连接可能需要更长时间才能建立。
:::

### 在 VS Code 中使用 OpenCode

要直接使用你的代码库，请在 VS Code 中打开终端并从项目目录运行 `opencode`。

![OpenCode in VS Code](/images/manual/use-cases/opencode-vscode.png#bordered)

## 编辑配置文件

OpenCode 将其配置存储在 JSON 文件中。你可以直接编辑此文件以管理提供方、模型和工具。

1. 打开 Files 并导航到 `Application/Data/opencode/.config/opencode/`。
2. 右键点击 `opencode.jsonc` 并选择 **Rename**。
   ![Rename config file](/images/manual/use-cases/opencode-rename-config-file.png#bordered)

3. 将文件重命名为 `config.json`，以便你可以在 Files 中直接编辑它。OpenCode 识别这两个扩展名。
4. 打开 `config.json` 并点击 <i class="material-symbols-outlined">edit_square</i> 编辑它。
   ![Edit config file](/images/manual/use-cases/opencode-edit-config-file.png#bordered)

5. 保存更改。
6. 从 **Settings** > **Applications** 重启 OpenCode 以应用更改。

## 了解更多

- [管理包](opencode-packages.md)：安装系统级和语言特定的包。
- [技能和插件](opencode-extensions.md)：通过技能和插件添加功能。
- [使用 Olares CLI 管理 Olares](opencode-olares-cli.md)：认证 Olares CLI，让 Agent 管理设备上的文件和应用。
- [使用 oh-my-openagent 编排多代理工作流](opencode-omo.md)：启用 OMO 以在 OpenCode 中运行多代理协作。
- [常见问题](opencode-issues.md)：已知问题的解决方案。
- [使用 Context7 将 AI 编码助手连接到最新文档](context7.md#opencode)：在 OpenCode 中将 Context7 注册为远程 MCP 服务器。
- [OpenCode 官方文档](https://opencode.ai/docs)
