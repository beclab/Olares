---
outline: [2, 3]
description: 在 Olares 上部署 Penpot 作为自托管设计工作空间，然后通过 MCP 连接 Cursor，以检查和修改活跃的 Penpot 设计文件。
head:
  - - meta
    - name: keywords
      content: Olares, Penpot, MCP, Model Context Protocol, Cursor, design collaboration, prototype, self-hosted design tool
app_version: "1.0.15"
doc_version: "1.0"
doc_updated: "2026-05-22"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/penpot.md)为准。
:::

# 通过 MCP 使用 Cursor 检查和编辑 Penpot 文件

Penpot 是一款开源的、基于网页的设计和原型工具，支持 UI 设计、交互原型、组件系统以及开发者交付。它使用 CSS、SVG 和 HTML 等开放标准，使其成为设计文件与前端实现之间的实用桥梁。

在 Olares 上，你可以将 Penpot 作为自托管设计工作空间运行，并通过 Penpot MCP 将其连接到 Cursor。本指南将带你完成一个完整的工作流程：打开 Penpot 文件、让 Cursor 读取其结构、请求 Cursor 添加一个卡片式组件，并在 Penpot 中检查结果。

## 学习目标

在本指南中，你将学习如何：

- 从 Market 安装 Penpot。
- 从 Olares Settings 获取 Penpot MCP 端点。
- 将活跃的 Penpot 文件连接到 MCP 服务器。
- 配置 Cursor 以读取已连接的 Penpot 文件。
- 使用 Cursor 检查画板并添加新的设计元素。
- 在 Penpot 中审查更改，并通过后续提示进行优化。

## 前提条件

- 你的电脑上已安装 Cursor。
- 在 Penpot 中已创建或导入一个 Penpot 文件。

## 安装 Penpot

1. 打开 Market 并搜索 "Penpot"。

   ![Penpot](/images/manual/use-cases/penpot.png#bordered)

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 准备 Penpot 工作流程

本指南使用一个简单的任务：请求 Cursor 检查一个 Penpot 文件，并在一个画板中添加一个卡片式组件。

1. 从 Launchpad 打开 Penpot。

2. 创建一个新的 Penpot 文件，或打开一个已有文件。

3. 选择你希望 Cursor 处理的页面和画板。

:::tip 保持文件打开
Cursor 只能读取当前活跃且已连接的 Penpot 文件。在整个工作流程中，请保持此浏览器标签页打开。
:::

## 通过 MCP 连接 Penpot 和 Cursor

从 Olares 获取你的 Penpot MCP 端点，然后将它们添加到 Penpot 和 Cursor 中。

### 获取 MCP 端点

1. 打开 Settings，然后进入 **Applications** > **Penpot**。

   ![Penpot application endpoints](/images/manual/use-cases/penpot-entrances.png#bordered)

2. 进入目标入口，找到 **Penpot MCP Plugin** 和 **Penpot MCP HTTP** 的端点 URL。

   - **Penpot MCP Plugin**：使用此端点安装 Penpot 插件。

      ![Copy Penpot MCP Plugin endpoint](/images/manual/use-cases/lp-penpotmcpplugin-endpoint.png#bordered)

   - **Penpot MCP HTTP**：使用此端点连接 Cursor。

      ![Copy Penpot MCP HTTP endpoint](/images/manual/use-cases/lp-penpotmcphttp-endpoint.png#bordered)

3. 保持此页面打开，或复制两个 URL。你将在接下来的步骤中使用它们。

### 在 Penpot 中安装并连接 MCP 插件

在目标文件已在 Penpot 中打开的情况下：

1. 在 Penpot 编辑器中，点击 <i class="material-symbols-outlined">more_vert</i> 打开主菜单，然后选择 **Plugins** > **Plugin manager**。

   ![Open Penpot plugin manager](/images/manual/use-cases/penpot-plugin-manager.png#bordered)

2. 在你的 MCP Plugin 端点后面追加 `/manifest.json`。按以下格式输入 URL，然后点击 **Install**：

   ```text
   <your-mcp-plugin-endpoint>/manifest.json
   ```

   例如：
   ```text
   https://2550d96f1.laresprime.olares.com/manifest.json
   ```

   ![Add Penpot MCP plugin](/images/manual/use-cases/penpot-add-plugin.png#bordered){width=60%}

3. 查看权限提示，然后点击 **Allow**。

4. 该插件现在会出现在 **INSTALLED PLUGINS** 部分。

   ![Penpot MCP plugin installed](/images/manual/use-cases/penpot-plugin-installed.png#bordered){width=60%}

5. 在 Plugin manager 中，点击 MCP 插件旁边的 **Open**。

6. 在插件面板中，点击 **CONNECT TO MCP SERVER**。

   ![Connect Penpot plugin to MCP server](/images/manual/use-cases/penpot-plugin-connect.png#bordered){width=95%}

7. 等待状态变为 **Connected to MCP server**。

   ![Penpot plugin connected](/images/manual/use-cases/penpot-plugin-connected.png#bordered){width=40%}

### 将 Cursor 配置为 MCP 客户端

1. 在你的电脑上打开 Cursor。

2. 进入 **Cursor** > **Settings** > **Tools & MCPs**，然后点击 **Add Custom MCP**。

   ![Add Custom MCP in Cursor](/images/manual/use-cases/penpot-cursor-add-mcp.png#bordered)

3. 在你的 MCP HTTP 端点后面追加 `/mcp`。在 `~/.cursor/mcp.json` 中，添加以下配置：

   ```json
   {
     "mcpServers": {
       "penpot": {
         "url": "<your-mcp-http-endpoint>/mcp"
       }
     }
   }
   ```

   ![Configure Penpot MCP in Cursor](/images/manual/use-cases/penpot-cursor-mcp-config.png#bordered)

   :::warning 检查 JSON 语法
   确保你完全按照上述格式复制，包括所有引号 `"` 和大括号 `{}`。JSON 无效会导致 Cursor 无法加载 MCP 服务器。
   :::

4. 保存文件。在 macOS 上，按 `Cmd + S`。在 Windows 上，按 `Ctrl + S`。

5. 在 **Tools & MCPs** 中，启用 **penpot** 旁边的开关。如果没有出现，请重启 Cursor 并重新打开 **Tools & MCPs**。

   ![Enable Penpot MCP in Cursor](/images/manual/use-cases/penpot-cursor-mcp-enabled.png#bordered)

## 使用 Cursor 编辑 Penpot 文件

### 检查文件结构

首先，请求 Cursor 读取设计结构。

1. 保持你的 Penpot 文件打开，并确保插件状态为 **Connected to MCP server**。

2. 在 Cursor 中，开启一个新对话并提问：

   ```text
   List all frames in the current Penpot file.
   ```

   ![Cursor lists Penpot frames](/images/manual/use-cases/penpot-cursor-list-frames.png#bordered)

:::tip 一次处理一个画板
如果文件包含多个画板，请在下一个提示中指定目标画板。这样可以使更改更集中，也更容易在 Penpot 中审查。
:::

### 添加卡片式组件

在 Cursor 读取文件后，请它进行具体的设计更改。以下示例会在选定的画板中添加一个可复用的卡片式组件。

1. 在 Cursor 中，发送类似这样的提示。将 `Home` 替换为你想要修改的画板：

   ```text
   In the current Penpot file, add a card-style component to the Home frame.
   The card should include a title, a short description, and one primary button.
   Match the existing spacing, colors, and typography as closely as possible.
   Name the main group "Feature card" and explain what you changed.
   ```

2. 如果 Cursor 要求你从多个选项中选择，请选择最符合你布局的选项。

3. 等待 Cursor 完成工具调用。

4. 返回 Penpot 并检查活跃页面。新的设计元素应该会出现在目标画板中。

   ![Penpot file updated by Cursor](/images/manual/use-cases/penpot-result-card.png#bordered)

5. 在 Penpot 中选择新元素，检查其图层名称、位置、文本和视觉样式。

### 审查并优化结果

将 Cursor 的首次编辑视为草稿。在 Penpot 中审查结果，然后向 Cursor 请求具体的调整。

使用如下提示：

```text
Move the Feature card 24 px below the hero heading and align it with the left edge of the content column.
```

```text
Make the button label shorter and adjust the card width so it matches the other content blocks.
```

```text
Rename the card layers so they are easy for developers to inspect.
```

当满足以下条件时，工作流程即告完成：

- Cursor 可以列出已连接 Penpot 文件中的画板。
- Cursor 可以解释它更改了哪个画板或图层。
- 新卡片出现在选定的 Penpot 画板中。
- 卡片使用清晰的图层名称，并与周围设计协调。

## 常见问题

### Cursor 无法看到我的 Penpot 文件

#### 原因

MCP 插件未连接、Penpot 浏览器标签页已关闭，或者活跃的是另一个文件或页面。

#### 解决方案

打开目标 Penpot 文件，打开 MCP 插件，点击 **Connect to server**，并等待状态显示为 **Connected to MCP server**。然后在 Cursor 中重试你的提示。

### Penpot MCP 连接是如何工作的？

Penpot MCP 将 Cursor 连接到你浏览器中当前打开的 Penpot 文件。

| 组件 | 使用者 | 功能 | 手动设置 |
|:----------|:--------|:-------------|:-------------|
| MCP Plugin | Penpot 文件 | 将活跃文件、页面、画板、图层、组件、样式和 token 暴露给 MCP 服务器。 | 在 Penpot 中添加。 |
| MCP HTTP | Cursor | 接收来自 Cursor 的 MCP 请求，并将其转发给已连接的 Penpot 文件。 | 在 Cursor 中配置。 |
| MCP WebSocket | 插件和 MCP 服务器 | 实时保持 Penpot 文件与 MCP 服务器的连接。 | 无需手动设置。 |

Cursor 只需要在 MCP HTTP 端点后面追加 `/mcp`。MCP WebSocket 连接在 Penpot 插件和 MCP 服务器之间内部使用。

在 Cursor 处理文件时，插件必须保持与 Penpot 的连接。

## 了解更多

- [Penpot Help Center](https://help.penpot.app/)：官方 Penpot 指南和产品文档。
- [Model Context Protocol](https://modelcontextprotocol.io/)：了解 MCP 如何将 AI 客户端连接到外部工具和数据源。
