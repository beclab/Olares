---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/dify
description: 在 Olares 上自托管 Dify，构建本地 AI 应用与助手。部署 Dify、通过 Router 接入模型，并添加个人知识库，实现私有的 RAG 工作流。
head:
  - - meta
    - name: keywords
      content: Olares, Dify, dify self hosted, dify vs n8n, dify ollama, dify on olares
---
# Dify 定制 AI 助手

<VersionRouteSelect />
Dify 是一个 AI 应用开发平台。它是 Olares 集成的关键开源项目之一，帮助你构建和管理 AI 应用，同时确保数据完全由自己掌控。
此外，你也可以在 Dify 中接入个人知识库文档，让 AI 应用更懂你。

## 开始之前

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- 从应用市场安装 Qwen3.8-27B (llama.cpp)。

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

## 通过 Router 添加聊天模型

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

1. 在 Dify 中打开 **Settings** > **Model Provider**，安装并选择 **OpenAI-API-compatible** 提供商插件。
2. 添加模型，**Model Name** 填写 `default-chat`，模型类型选择聊天或 **LLM**。
3. **Base URL** 填写从 Router 复制的地址，保留 `/v1`。API 密钥允许留空时留空，必填时填写 `olares`。
4. 保存配置。

## 配置 Ashia

1. 切换至 Dify 的**工作室**选项，并进入 **Ashia** 应用。
2. 从右侧模型列表中选择已配置好的 Gemma2 本地模型。


   <!--
   TODO: 素材清单 05，待补 Router 截图：dify-model-router.png；替换下方旧图后再取消注释。
   ![选择模型](/images/zh/manual/use-cases/dify-select-model.png#bordered)
   -->

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
