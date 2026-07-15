---
outline: [2, 3]
description: 了解如何在 Olares 上连接 AI 应用，重点讲解共享端点的用法，并以 Ollama 为例演示。
---

# 连接 AI 应用

:::warning
本页面由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../../manual/best-practices/connect-ai-apps.md)。
:::

Olares 上的很多 AI 应用都遵循同一种模式：一个应用通过 API 提供 AI 能力，另一个应用提供你日常使用的界面。理解了这种模式，你就能用同样的步骤，把几乎任意一对兼容的应用连接起来。

本教程讲解其中的核心概念，并以 Ollama 作为 AI 服务应用，带你走完几个实际示例。

## 学习目标

完成本教程后，你将能够：

- 区分 AI 服务应用和客户端应用。
- 配置认证级别，让应用之间顺畅通信。
- 判断什么时候用共享端点、什么时候用用户端点。
- 把 LobeHub（原 LobeChat）、n8n、Continue.dev 等常见客户端应用连接到 Ollama。

## 工作原理

把一个客户端应用连接到 AI 服务应用，通常分三步：

1. 在 Olares 设置中找到 AI 服务应用的 API 入口，把它的**认证级别**设为**内部**。
2. 复制该入口显示的端点地址。
3. 在客户端应用中，把这个端点填进模型或 API 配置页。如果连不上，参照[该用哪个端点](#which-endpoint-to-use)里的规则调整端点。

## 核心概念

### AI 服务应用与客户端应用

- **AI 服务应用**：充当后端引擎，通过 API 对外提供 AI 能力，通常作为服务在后台运行，自身没有聊天界面。例如 Ollama 和 ComfyUI Shared。
- **客户端应用**：面向用户的应用，提供你直接操作的聊天界面，但要靠 AI 服务应用来生成回复。例如 LobeHub、Open WebUI 和 n8n。

### 认证级别

Olares 为应用入口提供以下访问级别：
- **内部（推荐）**：应用之间通信时不会弹出登录提示，同时支持通过本地网络或 LarePass VPN 访问。
- **公开**：对互联网上任何人开放，私有服务不建议使用。

### 前端调用与后端调用 {#frontend-calls-vs-backend-calls}

客户端应用向 AI 服务应用发送 API 请求，有两种方式：
- **后端调用（强烈推荐）**：由客户端应用的服务端进程直接向 AI 服务应用发请求。把服务应用的 API 设为“内部”后，这类请求会跳过认证，是最稳定的连接方式。
- **前端调用**：请求直接从浏览器发出，不经过服务端转发，通常更快。但即使设成了“内部”权限，这类请求仍可能触发 Olares 登录提示，或被跨域（CORS）限制拦截，导致连接失败。

### 端点

端点（endpoint）就是访问应用入口的 URL。当 AI 服务应用开放一个 API 入口时，你通常会看到两种端点：

| 类型 | 格式 | 说明 |
|------|------|------|
| 用户端点 | `https://{route-ID}.{OlaresID}.olares.com` | 用于前端调用，或通过 VPN 从外部访问。 |
| 共享端点 | `http://{route-ID}.shared.olares.com` | 用于后端调用。系统内全局可达，应用间通信非常可靠。 |

### 该用哪个端点 {#which-endpoint-to-use}

:::tip
本教程演示的是通过 `olares.com` 域名连接。如果客户端设备与 Olares 在同一局域网内，用 `.local` 地址的方法完全相同。
:::

1. 先试共享端点（`http://{route-ID}.shared.olares.com`）。

   共享端点专为应用间的直接 API 访问设计，不需要用户凭证，通常是最可靠的选择。
2. 不行再用用户端点（`https://{route-ID}.{OlaresID}.olares.com`）。

   如果共享端点不可用，或者客户端应用是从浏览器（而非自己的服务端）发请求，就改用用户端点。把它的**认证级别**设为**内部**（推荐），这样既能免登录访问，又不会公开暴露到互联网。
3. 需要时补上后缀。

   很多客户端应用要求基础 URL 以 `/v1`（OpenAI 兼容 API）或 `/api`（其他格式）结尾。如果连不上，试着加上对应后缀，例如 `http://{route-ID}.shared.olares.com/v1`。两种端点都适用。
4. 用占位 API key。

   如果客户端应用要求填 API key，但服务本身并不需要，随便填个占位文字（比如 `ollama`）满足必填项即可。

## 示例

### 将 Ollama 连接到 LobeHub

本示例中，Ollama 作为 AI 服务应用，LobeHub 作为客户端应用。

示例使用 `qwen2.5:1.5b` 模型，开始前请确认你已经下载好它。

1. 在 Olares 上打开设置，进入 **应用** > **Ollama**。
2. 在**共享入口**中，选择 **Ollama API**。
   ![Ollama 共享入口](/images/manual/use-cases/obtain-ollama-hosturl2.png#bordered)

3. 复制共享端点 URL。
4. 打开 LobeHub，进入 **设置** > **AI 服务商** > **Ollama**。
5. 在 **接口代理地址** 字段中，粘贴刚才复制的共享端点。
   :::warning
   如果你用的是本地 Ollama 模型，不要开启 **客户端请求模式**。这个开关会让应用改用[前端调用](#frontend-calls-vs-backend-calls)，而在使用本地 AI 服务应用时，前端调用经常会触发登录提示或导致连接失败。
   :::
   ![填入共享端点](/images/manual/tutorials/api-lobechat-enter-url.png#bordered)
6. 验证连接：

   a. 点击 **获取模型列表**，你在 Ollama 里下载过的模型会出现在列表中。
   ![获取模型列表](/images/manual/tutorials/api-lobechat-fetch-models.png#bordered)

   b. 启用你想用的模型，例如 **Qwen2.5 1.5B**。

   c. 在 **连通性检查** 处，从下拉菜单选择 **qwen2.5:1.5b**，然后点击 **检查**。

      出现 **检查通过** 时，连接就建立好了。
      ![检查通过](/images/manual/tutorials/api-lobechat-check-passed.png#bordered)

### 将 Ollama 连接到 n8n

n8n 是从浏览器（而非自己的服务端）发请求的，所以需要用户端点。把认证级别设为**内部**，访问时就不会弹出登录提示。

:::tip 网络要求
请确保你的设备与 Olares 处于同一局域网，或已在 LarePass 中开启 VPN，否则连接无法成功。
:::

1. 在 Olares 上打开设置，进入 **应用** > **Ollama**。
2. 在**入口**下，点击 **Ollama API**。
3. 把**认证级别**设为**内部**。
4. 在**端点配置**下，复制 **端点** 旁显示的端点 URL。
5. 在 n8n 中新建一个 Ollama 凭证：

   a. 在 n8n 左侧导航栏选择 **+** > **Credential**。

   b. 在 **Add new credential** 对话框中，从下拉菜单选择 **Ollama**，然后点击 **Continue**。

   c. 粘贴刚才复制的 Ollama 端点 URL。

   d. 点击 **Save**，n8n 会自动测试连接。

      出现 **Connection tested successfully** 时，连接就建立好了。
      ![n8n 已连接 Ollama](/images/manual/tutorials/api-n8n-connected.png#bordered)

### 将 Ollama 连接到 Continue.dev（Olares 之外）

你可以把本地 IDE 连接到运行在 Olares 上的 Ollama，让 AI 辅助和代码补全都跑在你自己的硬件上，而不是第三方云端。

示例使用 `llama3.1:8b`、`qwen2.5-coder:7b` 和 `qwen2.5-coder:1.5b`，开始前请确认这些模型都已下载。

1. 在 Olares 上打开设置，进入 **应用** > **Ollama**。
2. 在**入口**下，点击 **Ollama API**。
3. 把**认证级别**设为**内部**。
4. 在**端点配置**下，复制 **端点** 旁显示的端点 URL。
5. 在本地 IDE（例如 IntelliJ IDEA）中，打开 Continue 面板。
6. 在 Continue 中配置模型以使用 Ollama：

   a. 点击 **Local Config** 打开 **Configs** 菜单，再点击 **Local Config** 旁的设置图标。
   ![打开 local config](/images/manual/tutorials/api-continue-local-config.png#bordered){width=45%}

   b. 在打开的 `config.yaml` 中，用刚才复制的 Ollama 端点更新各个模型条目：
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
7. 在 LarePass 桌面客户端上开启 VPN。
   ![在桌面端开启 LarePass VPN](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

8. 在 Continue 聊天面板中输入一个 prompt 测试连接，例如：
   ```plain
   Write a hello world python script
   ```
   ![输入 prompt](/images/manual/tutorials/api-continue-prompt.png#bordered){width=45%}

   Continue 会把请求转发到你 Olares 上的 Ollama 并返回结果。开启 LarePass VPN 后，你的 IDE 就能像在同一私有网络里一样访问 Ollama 端点。
   ![结果](/images/manual/tutorials/api-continue-hello-world.png#bordered){width=45%}

## 延伸阅读

- [应用](../../developer/concepts/application.md)
- [网络](../../developer/concepts/network.md)
- [管理应用入口](../olares/settings/manage-entrance.md)
- [使用场景](../../use-cases/index.md)
