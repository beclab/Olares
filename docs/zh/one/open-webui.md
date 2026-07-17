---
noindex: true
outline: [2, 3]
description: 了解如何在 Olares One 上设置 Open WebUI，使用 Ollama 与本地 LLM 聊天。
head:
  - - meta
    - name: keywords
      content: Open WebUI, Ollama, 本地 LLM, 聊天机器人, AI
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/open-webui.md)。
:::

# 使用 Open WebUI 与本地 LLM 聊天 <Badge type="tip" text="20 min" />

Open WebUI 提供了一个直观的界面，用于管理大语言模型（LLM），支持 Ollama 和 OpenAI 兼容的 API。

本指南将带你完成在 Olares One 上安装 Open WebUI 和 Qwen3.5 27B、连接模型以及开始第一次聊天的全过程。完成后，你将拥有一个私有的本地聊天机器人，随时可供日常使用。

## 学习目标

- 在 Olares One 上使用 Open WebUI 运行本地 LLM。
- 让 Qwen3.5 27B Q4_K_M 模型可供其他应用使用。

## 前提条件

**硬件** <br>
- Olares One 连接到稳定的网络。
- 至少 20 GB 的可用磁盘空间用于下载模型。
- 足够的 GPU VRAM 和系统内存来运行 LLM。

**用户权限**
- 管理员权限，用于从 Market 安装共享应用和管理 GPU 资源。

## 步骤 1：安装 Open WebUI

1. 打开 Market，搜索 "Open WebUI"。
   ![安装 Open WebUI](/images/one/open-webui.png#bordered)

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 步骤 2：安装 Qwen3.5 27B 并获取共享端点

1. 在 Market 中搜索并安装 "Qwen3.5 27B Q4_K_M (Ollama)"。等待安装完成。
   ![安装 Qwen3.5 27B](/images/one/qwen3.5-27b.png#bordered)

2. 安装完成后，点击 **Open** 查看模型下载进度。
   ![正在下载 Qwen3.5 27B](/images/one/qwen3.5-27b-downloading.png#bordered)
   :::tip
   模型文件约为 17 GB。下载时间取决于你的网络速度。
   :::

3. 当你看到以下界面时，模型已准备好使用。
   ![Qwen3.5 27B 已下载](/images/one/qwen3.5-27b-downloaded.png#bordered)

4. 要让 Open WebUI 访问此模型，你需要获取其共享端点 URL。

   a. 打开 Olares Settings，然后导航到 **Application** > **Qwen3.5 27B Q4_K_M (Ollama)**。

   b. 在 **Shared entrances** 中，选择 **Qwen3.5 27B Q4_K_M** 查看端点 URL。
   ![Qwen3.5 27B 共享入口](/images/manual/use-cases/deerflow2-shared-entrance.png#bordered)

   c. 复制共享端点。例如：
      ```plain
      http://94a553e00.shared.olares.com
      ```
   你在后续步骤中需要用到此 URL。

   :::tip 为什么使用共享端点 URL？
   模型应用页面上显示的 URL 是用户特定的。如果你的设备和 Olares One 不在同一本地网络中，前端调用可能会触发 Olares 登录，并且你可能会遇到跨域限制（CORS）。为避免这些问题，请使用共享端点 URL。
   :::

## 步骤 3：创建 Open WebUI 管理员账户

1. 打开 Open WebUI 应用。
2. 在欢迎页面，点击 **Get started**。
3. 输入你的姓名、邮箱和密码来创建账户。
   ![创建账户](/images/one/open-webui-create-account.png#bordered)
   :::info
   你的所有数据，包括登录信息，都存储在 Olares One 本地。
   :::
   :::tip 第一个账户是管理员
   在 Open WebUI 上创建的第一个账户具有管理员权限，让你完全控制用户管理和系统设置。
   :::

## 步骤 4：配置连接

1. 点击左下角的 **profile icon**，选择 **Admin Panel**。
2. 进入 **Settings** > **Connections**。
   :::info
   默认情况下，本地 Ollama API 已预先配置，并在 **Manage Ollama API connections** 下可见。
   :::
3. 点击 <span class="material-symbols-outlined">add</span> 打开 Add Connection 对话框。
4. 在 **URL** 字段中，粘贴你在步骤 2 中复制的共享端点 URL，然后点击 **Save**。Open WebUI 会自动验证连接。当你看到 "Ollama API settings updated" 时，连接已建立。
   ![连接已建立](/images/one/open-webui-connection-established.png#bordered)

## 步骤 5：与本地 LLM 聊天

1. 在聊天主页面，确认模型下拉菜单中已选择 **qwen3.5:27b-q4_K_M**。
   ![与本地 LLM 聊天](/images/one/open-webui-qwen3.5-27b.png#bordered)

2. 在文本框中输入你的提示词，按 **Enter** 开始聊天。
   ![与本地 LLM 聊天](/images/one/open-webui-chat1.png#bordered)

## 故障排查

### Qwen3.5 27B 卡在 "Waiting for Ollama" 或 "Needs attention" 状态

如果 Qwen3.5 27B 应用在这些状态停留超过几分钟，首先检查 **Settings** > **GPU** 中的 GPU 模式：

- 如果你处于 **Memory slicing** 模式，请确保你已关联 Qwen3.5 27B 应用并为其分配了足够的 VRAM。
- 如果你处于 **App exclusive** 模式，请确保具有完整 GPU 访问权限的应用是 Qwen3.5 27B。

## 资源

- [Open WebUI 文档中心](https://docs.openwebui.com/getting-started/)
- [切换 GPU 模式](gpu.md)
- [更多 Open WebUI 功能](../use-cases/openwebui.md)
