---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/open-design
outline: [2, 3]
description: 在 Olares 上使用 Open Design，通过本地或云端 AI 模型，将提示词转化为 HTML 原型、落地页和幻灯片。
head:
  - - meta
    - name: keywords
      content: Olares, Open Design, AI 设计工作室, AI 原型, 落地页生成, 演示文稿生成, OpenAI 兼容, 本地大模型, Qwen3.8, 自托管
app_version: "0.22.1"
doc_version: "1.0"
doc_updated: "2026-09-23"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/open-design.md)为准。
:::

# 使用 Open Design 创建设计文件

<VersionRouteSelect />

Open Design 是一款开源 AI 设计工作室，可以将自然语言需求转化为可直接使用的设计文件。它通过内置的 OpenCode 调用你配置的模型，生成可预览的 HTML 原型、落地页、仪表盘、线框图和幻灯片。

在 Olares 上，你可以通过 Router 将 Open Design 连接到本地模型，也可以使用云端模型的 API Key。每个项目集中保存提示词、参考素材、预览和生成文件，方便你持续调整并导出最终设计。

:::warning 当前输出限制
Olares 上的 Open Design 0.22.1 暂不支持通过文本直接生成独立图片。请用它创建页面、原型、仪表盘或幻灯片。
:::

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Open Design。
- 通过 OpenAI 兼容 API 连接本地模型。
- 创建、调整和导出设计项目。
- 排查常见的模型连接和生成问题。

## 前提条件

开始前，你需要：

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- 以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 对话 | Qwen3.8-27B (llama.cpp) | 从 Market 安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 Open Design

1. 打开 Market，搜索 "Open Design"。
2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 连接本地模型

### 获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

本指南使用 **OpenAI-Compatible** API 格式，将 Open Design 连接到 Qwen3.8-27B (llama.cpp)：

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### 在 Open Design 中添加模型

1. 从启动台打开 Open Design。
2. 打开 **Settings**。在 **Models & providers** 页面选择 **API provider**。
3. 在 **Provider preset** 中选择 **Custom provider**。
4. 配置提供方：

   - **Base URL**：粘贴从 Router 复制的 Base URL，包括末尾的 `/v1`。
   - **API Key**：输入任意非空值，例如 `olares`。同一 Olares 集群中的应用调用本地模型时，不需要真实密钥。
   - **Model**：填写 `default-chat`。
   - **Max tokens（可选）**：输入 `65536`。如需使用更小的上下文窗口，可改为 `32768`。

   <!-- ![在 Open Design 中配置 API 提供方](/images/manual/use-cases/open-design-api-provider.png#bordered) -->

5. 点击 **Test** 检查连接。显示绿色的 **Connected** 状态表示连接成功。修改会自动保存。

:::tip 先做小规模测试
开始制作多页幻灯片或落地页之前，先让 Open Design 生成一张封面幻灯片。预览成功后，说明 Base URL、模型名称和 token 上限可以正常配合使用。
:::

## 创建设计项目

1. 在首页选择项目类型，例如 **Prototype**、**Slide deck**、**Document** 或 **Website clone**。
2. 选择符合任务需求的 Skill 预设，例如 **Blog Post**。也可以打开提示词下方的示例，将它作为起点。
3. 根据需要选择 **Design system** 和 **Working directory**。
4. 输入具体需求。建议说明目标受众、内容、视觉风格和输出格式。例如：

   ```text
   为产品发布会制作一张封面幻灯片。
   使用深色背景、醒目的几何字体和青绿色点缀。
   产品名称为“Northstar”，副标题为“Find your next move.”。
   ```

   <!-- ![选择项目类型和设计系统](/images/manual/use-cases/open-design-create-project.png#bordered) -->

5. 确认已选择目标模型，然后点击箭头开始生成。
6. 等待工作区打开并显示预览。

   <!-- ![Open Design 工作区和预览](/images/manual/use-cases/open-design-workspace-preview.png#bordered) -->

7. 在对话中继续调整结果。你可以拖入参考图片，或输入 `@` 添加项目文件，然后说明要修改的部分。
8. 完成后打开导出菜单，根据项目类型下载 ZIP、PDF 或 PPTX 文件。

## 可选：连接云端模型

你也可以使用云端模型。打开**设置**，在**模型与提供商**页面选择 **API 提供方**，然后选择提供商预设，或使用匹配的协议、Base URL、API Key 和模型配置自定义提供方。

| 提供商 | 协议 | Base URL |
|:-------|:-----|:---------|
| OpenAI | OpenAI | `https://api.openai.com/v1` |
| OpenRouter | OpenAI | `https://openrouter.ai/api/v1` |
| DeepSeek | OpenAI | `https://api.deepseek.com` |
| Anthropic | Anthropic | `https://api.anthropic.com` |
| 通义 Qwen | OpenAI | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| 硅基流动 | OpenAI | `https://api.siliconflow.cn/v1` |
| Kimi | OpenAI | `https://api.moonshot.cn/v1` |

云端模型的调用费用由所选提供商收取。请使用该提供商的完整模型名称和有效 API Key。

## 故障排查

### 连接测试失败或返回 404

从 Router 的 **How to call this model** 窗口重新复制 **Apps in Olares** 下的 Base URL，保留 `/v1`。在 **LLM** 页面确认默认聊天模型显示 **Callable**。

### 设计尚未生成完成便停止

将**最大 tokens（可选）**设置为 `65536` 或 `32768`。多页幻灯片需要的输出 token 比单页或封面幻灯片更多。

### Open Design 一直没有响应

返回模型应用的 Model Console。确认**模型**显示**就绪**，且**引擎**显示**运行中**后再继续。

### Open Design 使用了错误的模型

使用 `default-chat`，并在 Router 的 **Default models** 页面检查它指向的模型。如需固定使用某个模型，请改填 Router 中显示的完整模型名称。

### Open Design 无法生成独立图片

这是 Olares 上 Open Design 0.22.1 的当前限制。请改用页面、幻灯片、仪表盘、线框图或交互式原型对应的 Skill。

## 了解更多

- [使用引擎基座应用托管本地大语言模型](llm-base-apps.md)：使用引擎基座应用部署本地模型，并通过 Router 连接。
- [将 AI 应用连接到模型服务](/zh/manual/best-practices/connect-ai-apps.md)：了解模型名称、API 格式和 Base URL。
- [Open Design 官网](https://open-design.ai/)：了解上游项目及其功能。
