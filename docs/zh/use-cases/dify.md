---
description: Dify 在 Olares 上的部署教程，帮助你快速构建和管理 AI 应用，打造专属的智能服务生态系统。
---
# Dify 定制 AI 助手
Dify 是一个 AI 应用开发平台。它是 Olares 集成的关键开源项目之一，帮助你构建和管理 AI 应用，同时确保数据完全由自己掌控。
此外，你也可以在 Dify 中接入个人知识库文档，让 AI 应用更懂你。

## 开始之前
要使用本地 AI 模型，请确保你的环境中已配置以下内容：
- Olares 环境中已安装并运行 [Ollama](ollama.md)。
- 已安装 [Open WebUI](openwebui.md)，并下载了你偏好的语言模型。
- 访问设备已安装 LarePass 并已[启用专用网络](../manual/larepass/private-network.md)。
  :::tip 提示
  建议使用 `gemma2` 或 `qwen` 等轻量但功能强大的模型，可在速度和性能间取得良好平衡。
  :::

## 安装 Dify
:::info
从 Olares 1.11.6 开始，如果已安装 "Dify For Cluster" 或 "Dify"，需先卸载这些版本。
:::

1. 从应用市场中安装 “Dify 共享版”。
2. 从桌面打开 Dify。请确保管理员已安装 Dify 共享版。

## 创建 AI 助手应用

1. 打开 Dify，在**工作室**选项卡下，点击**创建空白应用**创建一个 AI 助手应用。这里我们创建一个名为 “Ashia” 的 Agent。
   ![创建应用](/images/zh/manual/use-cases/dify-create-app.png#bordered)

2. 右侧点击**去设置**，进入模型供应商配置页面。你可以选择远程模型或本地托管模型。
   ![应用初始页面](/images/zh/manual/use-cases/dify-app-init.png#bordered)

## 添加 Ollama 作为模型提供商

1. 进入**设置** > **应用** > **Ollama** > **入口**，设置 Ollama 的认证级别为“内部”。该设置允许其他应用在本地网络环境下可无需认证即可访问 Ollama 服务。
   
   ![Ollama entrance](/images/zh/manual/use-cases/dify-ollama-entrance.png#bordered)

2. 在 Dify 的 模型提供者商配置页面，选择 Ollama 作为模型提供者商，并进行以下配置：
    - **模型名称**：填写模型名称，例如：`gemma2`。
    - **基础 URL**：填入 Ollama 本地地址: `https://39975b9a1.{username}.olares.cn`。将 `{username}` 替换为 Olares 管理员的用户名。例如：`https://39975b9a1.marvin123.olares.com`。
   
    ![配置 Ollama](/images/zh/manual/use-cases/dify-add-gemma2.png#bordered){width=70%}

      :::info 提示
      其他必填字段可以保留默认值。
      :::
3. 点击**保存**。

## 配置 Ashia

1. 切换至 Dify 的**工作室**选项，并进入 **Ashia** 应用。
2. 从右侧模型列表中选择已配置好的 Gemma2 本地模型。

   ![选择模型](/images/zh/manual/use-cases/dify-select-model.png#bordered)
3. 点击**发布**。现在可以在**调试与预览**窗口试着和 Gemma2 聊天了。

   ![聊天](/images/zh/manual/use-cases/dify-chat-with-ashia.png#bordered)

## 设置本地知识库
1. 在 Dify 中，进入**知识库**选项卡。
2. 找到你的默认知识库。Dify 会监听你 Olares ID 名下的 `/Documents` 文件夹，作为其默认知识库。
    ![默认知识库](/images/zh/manual/use-cases/dify-default-knowledge-base.png#bordered)
3. 进入 `/Documents` 文件夹，添加文档至知识库。
   ![添加文档](/images/zh/manual/use-cases/dify-add-kb-file.png#bordered)
4. 在 Ashia 的编排页面中，点击 **<i class="material-symbols-outlined">add</i>添加**，选择创建的知识库，为 Ashia 添加上下文支持。
   ![添加知识库](/images/zh/manual/use-cases/dify-add-knowledge-base.png#bordered)
5. 点击**发布**。现在有了知识库的帮助，你可以试着向助手问一个专业问题：
   ![知识库聊天](/images/zh/manual/use-cases/dify-chat-kb.png#bordered)