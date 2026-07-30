---
outline: [2, 4]
title: 使用 LobeHub 构建本地 AI 智能体
description: 在 Olares 上安装 LobeHub 并接入本地模型，构建支持知识库、技能和多模态输入的自托管 AI 助手。
head:
  - - meta
    - name: keywords
      content: Olares, LobeHub, LobeChat, self-hosted lobechat, AI agent, lobechat on olares
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/lobechat.md)。
:::

# 使用 LobeHub 构建你的本地 AI 智能体

LobeHub（前身为 LobeChat）是一个开源平台，用于构建安全、自托管的 AI 智能体和聊天体验。它可以连接到你的本地模型，支持文件处理和知识库，并允许你创建具有自定义技能的专用智能体。

本指南涵盖 LobeHub 的安装、配置和实际使用，以创建你的个性化 AI 智能体。

:::tip 关于产品名称
LobeHub 是官方平台名称，但应用目前在 Olares Market 中列为 "LobeChat"。本指南中同时使用两个名称，以精确匹配你在屏幕上看到的内容。Market 将在未来更新中反映新的 LobeHub 品牌。
:::

## 学习目标

- 在 Olares 上安装 LobeHub 并连接到本地模型。
- 使用 Lobe AI 完成日常任务。
- 使用 Agent Builder 或自定义设置创建专用智能体。

## 前提条件

开始前，你需要以下模型：

| 模型类型 | 模型 | 获取方式 |
| :--- | :--- | :--- |
| 聊天 | Qwen3.6-27B (llama.cpp) | 从应用市场安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 LobeHub

1. 从 Olares Market 搜索 "LobeChat"。

   ![从 Market 搜索 LobeChat](/images/manual/use-cases/find-lobechat2.png#bordered)

2. 点击 **获取**，然后点击 **安装**。等待安装完成。

## 登录 LobeHub

1. 从 Launchpad 打开 **LobeChat**。
2. 输入你的电子邮件地址，然后按照页面提示创建 LobeHub 账户并登录。

   ![LobeHub 主页](/images/manual/use-cases/lobehub-start.png#bordered)

## 配置连接

将 LobeHub 连接到你的本地模型，以使聊天界面正常工作。

### 获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### 在 LobeHub 中配置连接

1. 从左侧边栏，前往**设置** > **AI 服务提供商** > **OpenAI**。

   ![在 LobeHub 中配置模型连接](/images/manual/use-cases/lobehub-config-model.png#bordered)

2. 配置以下设置：

   - **API 密钥**：输入任意占位文本，例如 `local`。
   - **API 代理 URL**：输入你从 Model Console 复制的 **Base URL**。例如，`https://e46e044d.laresprime.olares.com/v1`。
   - **使用 Responses API 规范**：确保此选项已禁用。
   - **使用客户端请求模式**：确保此选项已禁用。

      :::tip
      运行本地模型时，不要启用**使用客户端请求模式**选项。此模式专为远程 API 调用设计，可能会导致连接错误。
      :::

3. 在**模型列表**部分，点击**获取模型**以拉取支持的模型列表。模型名称 `unsloth/Qwen3.6-27B-GGUF:Q4_K_M` 会出现在列表中。

   ![获取模型列表并启用模型](/images/manual/use-cases/lobehub-fetch-enable-model1.png#bordered)

4. 点击 <i class="material-symbols-outlined">toggle_off</i> 启用它。
5. 在**连接检查**部分，从列表中选择你刚刚启用的模型，然后点击**检查**以验证连接。如果模型较大，加载可能需要更长时间。

   按钮变为**检查通过**，表示连接已建立。

   ![连接检查成功](/images/manual/use-cases/lobehub-checkpass2.png#bordered)  

6. 点击左上角的 Home 图标返回 LobeHub 主页。

   ![返回主页](/images/manual/use-cases/lobehub-return-home.png#bordered){width=50%}

## 使用 Lobe AI

Lobe AI 是 LobeHub 的官方默认智能体。它旨在帮助你完成各种任务，无需复杂设置，例如软件开发、学习支持、创意写作、数据分析和日常个人任务。

如果 Lobe AI 不能满足你的特定工作流需求，你可以构建自己的专用智能体。有关更多信息，请参阅[创建智能体](#创建智能体)。

1. 从左侧边栏，点击 **Lobe AI**。

   ![点击 Lobe AI](/images/manual/use-cases/lobe-ai.png#bordered)

2. 在聊天窗口中，点击模型选择器并选择刚才配置的本地语言模型。
3. 像与任何标准对话 AI 一样聊天。

## 创建智能体

使用对话式 Agent Builder 或通过手动从头开始配置设置，创建你自己的专用智能体。

LobeHub 允许你通过利用各种语言模型并将它们与技能相结合，创建处理特定任务的专用助手。
- **灵活的模型切换**：你可以在同一次聊天中即时切换语言模型以获得最佳结果。例如，如果你对某个回答不满意，可以从列表中选择不同的模型以利用它们的独特优势。
- **技能扩展**：你还可以安装额外的技能来扩展和增强智能体的能力。
   要安装技能，请确保你选择一个兼容 Function Calling 的模型。寻找模型名称旁边的 <i class="material-symbols-outlined">brick</i>，它表示该模型支持函数调用。

### 使用 Agent Builder 创建

Agent Builder 是 LobeHub 内置的助手，帮助你通过对话创建专用智能体。描述你的需求，它会自动生成完整的智能体配置，包括角色设置、系统提示和技能。

1. 在主页上，点击聊天框下方的**创建智能体**。

   ![创建智能体按钮](/images/manual/use-cases/lobehub-create-agent1.png#bordered)

2. 在聊天框中，描述你希望智能体处理的具体任务。例如，

   ```text
   我需要一个智能体来审查我的日常工作事项并总结它们。总结应重点关注任务的整体目的，
   并突出具体的行动项。
   ```
3. 选择本地语言模型。
4. 按 **Enter**。新智能体的资料页面打开，你可以看到 Agent Builder 开始自动配置你的智能体。

   ![Agent Builder](/images/manual/use-cases/lobehub-agent-builder1.png#bordered)

5. 使用右下角的聊天界面与 Agent Builder 交互。当你提供更多细节或完善需求时，Agent Builder 会自动起草并更新配置。
6. 创建完成后，点击**开始对话**以使用该智能体。
7. 在聊天中提供你的文本，然后你可以获得精炼的结果。例如，

   ```text
   - 修复登录页面的 bug 405
   - 与设计讨论新仪表板
   - 在邮件中回答客户关于账单的问题。
   - 审查 pr112，截止日期明天上午 11:00
   ```
   你获得输出：

   ![Agent Builder 示例输出](/images/manual/use-cases/agent-builder-example1.png#bordered)

8. 如果你对智能体的表现满意，请将其固定以便快速访问：

   a. 返回主页。

   b. 将鼠标悬停在左侧边栏的智能体上，点击 <i class="material-symbols-outlined">more_horiz</i>，然后点击**固定**。

### 创建自定义智能体

如果你有特定需求并更喜欢完全手动配置智能体，请创建自定义智能体。

自定义智能体提供最高级别的个性化。你可以设置智能体的头像、名称、AI 模型、技能和提示，以创建一个独特的 AI 智能体。

1. 在主页上，点击左上角的机器人图标，然后选择**创建智能体**。

   ![创建自定义智能体](/images/manual/use-cases/lobehub-create-custom-agent.png#bordered){width=50%}

   **智能体资料**页面打开。

   ![自定义智能体资料](/images/manual/use-cases/lobehub-custom-agent-profile1.png#bordered)
2. 点击默认的机器人头像，为你的智能体选择一个新图标。
3. 输入智能体名称。例如，`SEO 文案撰写`。
4. 选择配置好的本地语言模型。
5. 点击 **+ 添加技能**为智能体配备额外的工具。例如，选择**网页浏览**以收集 SEO 数据。
6. 通过填写结构化 markdown 模板来定义角色和行为，以精确定义智能体的运作方式。例如，

   ```text
   #### 目标
   根据用户提供的主题撰写 SEO 优化的博客文章。
   #### 技能
   - 关键词研究、部署和密度优化
   - 引人入胜的标题生成
   - Markdown 格式
   #### 工作流
   1. 询问用户主题。
   2. 建议目标关键词、H1 标题和最佳元描述。
   3. 生成针对 Google 精选摘要优化的结构化大纲。
   4. 生成待批准的结构化大纲。
   5. 大纲获批后撰写完整的博客文章。
   #### 约束
   - 使用简单语言，避免技术术语。
   - 关注用户价值而不是罗列产品功能。
   - 避免使用被动语态。
   - 用第二人称 "你" 面向用户
   ```
7. 点击**开始对话**以使用它。例如，输入以下请求：

   ```text
   我想为 "本地 AI 替代方案" 排名
   ```
8. 审查提案和输出，然后与其迭代，直到你对结果满意。

   ![自定义智能体结果示例](/images/manual/use-cases/lobehub-seo-sample1.png#bordered)

9. 如果你对智能体的表现满意，请将其固定以便快速访问：

   a. 返回主页。

   b. 将鼠标悬停在左侧边栏的智能体上，点击 <i class="material-symbols-outlined">more_horiz</i>，然后点击**固定**。

## 常见问题

### 为什么连接检查失败？

如果你遇到 `Error requesting service` 错误，请按以下步骤排查并重试：

   ![连接错误](/images/manual/use-cases/lobehub-connection-error.png#bordered)

1. 检查 Model Console，确认 **Model** 的状态显示为 **READY**，且 **Engine** 的状态显示为 **RUNNING**。
2. 确保提供商设置页面上的**使用客户端请求模式**选项已禁用。

   ![禁用使用客户端请求模式选项](/images/manual/use-cases/lobehub-disable-client-request-mode3.png#bordered)
