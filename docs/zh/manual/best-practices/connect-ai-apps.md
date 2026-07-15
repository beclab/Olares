---
outline: [2, 3]
description: 学习如何在 Olares 上使用共享端点连接 AI 应用，并以 Ollama 作为实际示例。
---

:::warning
本文档由 AI 翻译生成，仅供参考。如有歧义，请参阅[英文原文](/manual/best-practices/connect-ai-apps.md)。
:::

# 连接 AI 应用

Olares 上的许多 AI 应用都遵循相同的模式：一个应用通过 API 提供 AI 能力，另一个应用则提供你每天使用的界面。理解这个模式后，你就可以将几乎任何兼容的应用组合连接起来。

本教程介绍核心概念，并通过使用 Ollama 作为 AI 服务应用的实际示例带你完成连接。

## 学习目标

完成本教程后，你将能够：

- 区分 AI 服务应用与客户端应用。
- 配置认证级别以实现应用之间的无缝通信。
- 理解何时使用共享端点或用户端点。
- 将常见的客户端应用（例如 LobeHub（前身为 LobeChat）、n8n 和 Continue.dev）连接到 Ollama。

## 工作原理

将客户端应用连接到 AI 服务应用通常涉及三个步骤：

1. 在 Olares **设置**中，找到 AI 服务应用的 API 入口，并将其**认证级别**设置为 **Internal**。
2. 复制该入口显示的端点。
3. 在客户端应用中，将此端点粘贴到模型或 API 配置页面。如果连接失败，请按照[使用哪种端点](#使用哪种端点)中的规则进行调整。

## 核心概念

### AI 服务应用与客户端应用

- **AI 服务应用**：作为后端引擎。它们通过 API 提供 AI 能力，通常以服务形式运行，没有自己的聊天界面。例如 Ollama 和 ComfyUI Shared。
- **客户端应用**：作为用户-facing 应用。它们提供你直接交互的聊天界面，但依赖 AI 服务应用来生成回复。例如 LobeHub、Open WebUI 和 n8n。

### 认证级别

Olares 为应用入口提供以下访问级别：

- **Internal（推荐）**：允许应用之间无需登录提示即可通信，也允许通过本地网络或 LarePass VPN 访问。
- **Public**：向互联网上的任何人开放。不建议用于私有服务。

### 前端调用与后端调用

客户端应用通过以下方式之一向 AI 服务应用发送 API 请求：

- **后端调用（强烈推荐）**：客户端应用的服务器进程直接向 AI 服务应用发起请求。将服务应用的 API 设置为 "Internal" 后，这些调用可以绕过认证，是最稳定的连接方式。
- **前端调用**：请求直接从你的浏览器发送。这种方式避免了服务器端转发，通常更快。然而，即使拥有 "Internal" 权限，这些调用也可能触发 Olares 登录提示或被跨域（CORS）限制阻止，导致连接失败。

### 端点

端点是访问应用入口的 URL。当 AI 服务应用暴露 API 入口时，你通常会看到两种类型的端点：

| 类型 | 格式 | 说明 |
|------|--------|-------------|
| 用户端点 | `https://{route-ID}.{OlaresID}.olares.com` | 前端调用或通过 VPN 的外部访问。 |
| 共享端点 | `http://{route-ID}.shared.olares.com` | 后端调用。系统级访问，应用间通信非常可靠。 |

### 使用哪种端点

:::tip
本教程介绍使用 `olares.com` 域名的连接方式。如果你的客户端设备与 Olares 位于同一本地网络，也可以使用 `.local` 地址进行同样的操作。
:::

1. 首先尝试共享端点（`http://{route-ID}.shared.olares.com`）。

   共享端点专为直接的应用间 API 访问而设计。它们不需要用户凭证，通常是最可靠的选择。
2. 如果共享端点不可用，或者客户端应用从浏览器而不是自己的服务器发送请求，请回退到用户端点（`https://{route-ID}.{OlaresID}.olares.com`）。

   将其**认证级别**设置为 **Internal**（推荐），这样无需登录提示即可访问，但不会公开暴露。
3. 根据需要添加后缀。

   许多客户端应用期望 OpenAI 兼容 API 的基础 URL 以 `/v1` 结尾，或其他格式以 `/api` 结尾。如果连接失败，请尝试添加合适的后缀。例如：`http://{route-ID}.shared.olares.com/v1`。这适用于两种端点类型。
4. 使用占位 API key。

   如果客户端应用需要 API key 但服务并不使用，请输入任意占位文本（例如 `ollama`）以满足必填字段的要求。

## 示例

### 将 Ollama 连接到 LobeHub

在本示例中，Ollama 作为 AI 服务应用，LobeHub 是客户端应用。

本示例使用 `qwen2.5:1.5b` 作为模型。开始之前请确保已下载该模型。

1. 在 Olares 上，打开**设置**，然后进入 **应用** > **Ollama**。
2. 在**共享入口**中，选择 **Ollama API**。
   ![Ollama 共享入口](/images/manual/use-cases/obtain-ollama-hosturl2.png#bordered)
   
3. 复制共享端点 URL。
4. 打开 LobeHub，然后进入 **Settings** > **AI Service Provider** > **Ollama**。
5. 在 **Interface proxy address** 字段中，粘贴你复制的共享端点。
   :::warning
   如果你使用的是本地 Ollama 模型，请不要启用 **Use Client Request Mode**。该设置会将应用切换为使用[前端调用](#前端调用与后端调用)，在使用本地 AI 服务应用时经常会触发登录提示或连接失败。
   :::
   ![输入共享端点](/images/manual/tutorials/api-lobechat-enter-url.png#bordered)
6. 验证连接：

   a. 点击 **Fetch models**。你在 Ollama 中下载的模型将出现在列表中。
   ![获取模型](/images/manual/tutorials/api-lobechat-fetch-models.png#bordered)

   b. 启用你想要使用的模型。例如，启用 **Qwen2.5 1.5B**。

   c. 对于 **Connectivity Check**，从下拉列表中选择 **qwen2.5:1.5b**，然后点击 **Check**。

      当显示 **Check Passed** 时，连接即建立成功。
      ![检查通过](/images/manual/tutorials/api-lobechat-check-passed.png#bordered)

### 将 Ollama 连接到 n8n

n8n 从浏览器而不是其服务器发起请求，因此需要使用用户端点。将认证级别配置为 **Internal**，以便无需登录提示即可访问。

:::tip 网络要求
确保你的设备与 Olares 位于同一本地网络，或在 LarePass 中启用了 VPN，连接才能正常工作。
:::

1. 在 Olares 上，打开**设置**，然后进入 **Application** > **Ollama**。
2. 在 **Entrances** 下，点击 **Ollama API**。
3. 将 **Authentication level** 设置为 **Internal**。
4. 在 **Endpoint settings** 下，复制 **Endpoint** 旁显示的端点 URL。
5. 在 n8n 中创建新的 Ollama 凭证：

   a. 在 n8n 中，点击左侧导航栏的 **+** > **Credential**。

   b. 在 **Add new credential** 对话框中，从下拉列表选择 **Ollama**，然后点击 **Continue**。

   c. 粘贴你复制的 Ollama 端点 URL。

   d. 点击 **Save**。n8n 会自动测试连接。

      当显示 **Connection tested successfully** 时，连接即建立成功。   
      ![n8n Ollama 已连接](/images/manual/tutorials/api-n8n-connected.png#bordered)

### 将 Ollama 连接到 Continue.dev（Olares 外部）

你可以将本地 IDE 连接到运行在 Olares 系统上的 Ollama，这样 AI 辅助和代码补全就由你自己的硬件驱动，而不是第三方云。

本示例使用 `llama3.1:8b`、`qwen2.5-coder:7b` 和 `qwen2.5-coder:1.5b`。开始之前请确保已下载这些模型。

1. 在 Olares 上，打开**设置**，然后进入 **Application** > **Ollama**。
2. 在 **Entrances** 下，点击 **Ollama API**。
3. 将 **Authentication level** 设置为 **Internal**。
4. 在 **Endpoint settings** 下，复制 **Endpoint** 旁显示的端点 URL。
5. 在你的本地 IDE（例如 IntelliJ IDEA）中，打开 Continue 面板。
6. 在 Continue 中配置使用 Ollama 的模型：

   a. 点击 **Local Config** 打开 **Configs** 菜单，然后点击 **Local Config** 旁的设置图标。
   ![打开本地配置](/images/manual/tutorials/api-continue-local-config.png#bordered){width=45%}

   b. 在打开的 `config.yaml` 中，使用你复制的 Ollama 端点更新模型条目：
   ```yaml
   name: Local Config
   version: 1.0.0
   schema: v1
   models:
   - name: Llama3.1-8B
      provider: ollama
      model: llama3.1:8b
      apiBase: https://<your-ollama-endpoint>
      roles:
         - chat
   - name: Qwen2.5-Coder 7B
      provider: ollama
      model: qwen2.5-coder:7b
      apiBase: https://<your-ollama-endpoint>
      roles:
         - edit
         - embed
         - rerank
   - name: Qwen2.5-Coder 1.5B
      provider: ollama
      model: qwen2.5-coder:1.5b
      apiBase: https://<your-ollama-endpoint>
      roles:
         - autocomplete
         - apply
   ```
7. 在 LarePass 桌面客户端上启用 VPN。
   ![在桌面端启用 LarePass VPN](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

8. 在 Continue 聊天面板中，输入提示词测试连接。例如：
   ```plain
   Write a hello world python script
   ```
   ![输入提示词](/images/manual/tutorials/api-continue-prompt.png#bordered){width=45%}

   Continue 会将请求路由到你 Olares 系统上的 Ollama 并返回结果。启用 LarePass VPN 后，你的 IDE 可以像访问同一私有网络一样访问 Ollama 端点。
   ![结果](/images/manual/tutorials/api-continue-hello-world.png#bordered){width=45%}

## 了解更多

- [应用](/zh/developer/concepts/application.md)
- [网络](/zh/developer/concepts/network.md)
- [管理应用入口](/zh/manual/olares/settings/manage-entrance.md)
- [使用场景](/zh/use-cases/index.md)
