---
outline: [2, 3]
description: 在 Olares 上使用 Open Design，通过本地或云端 AI 模型，将提示词转化为 HTML 原型、落地页和幻灯片。
head:
  - - meta
    - name: keywords
      content: Olares, Open Design, AI 设计工作室, AI 原型, 落地页生成, 演示文稿生成, OpenAI 兼容, 本地大模型, Qwen3.6, 自托管
app_version: "0.22.1"
doc_version: "1.0"
doc_updated: "2026-09-18"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/open-design.md)为准。
:::

# 使用 Open Design 创建设计文件

Open Design 是一款开源 AI 设计工作室，可以将自然语言需求转化为可直接使用的设计文件。它通过内置的 OpenCode 调用你配置的模型，生成可预览的 HTML 原型、落地页、仪表盘、线框图和幻灯片。

在 Olares 上，你可以通过 Model Console 将 Open Design 连接到本地模型，也可以使用云端模型的 API Key。每个项目集中保存提示词、参考素材、预览和生成文件，方便你持续调整并导出最终设计。

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

- Olares 1.12.6 或更高版本。
- 已从 Market 安装 Qwen3.6-27B (llama.cpp)，并确认模型在 Model Console 中已就绪。本指南使用该模型作为本地模型。

如需部署其他本地模型，请参阅[使用引擎基座应用托管本地大语言模型](llm-base-apps.md)。

## 安装 Open Design

1. 打开 Market，搜索 "Open Design"。

   <!-- ![Market 中的 Open Design](/images/manual/use-cases/open-design.png#bordered) -->

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 连接本地模型

### 获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

本指南使用 **OpenAI-Compatible** API 格式，将 Open Design 连接到 Qwen3.6-27B (llama.cpp)：

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### 在 Open Design 中添加模型

1. 从启动台打开 Open Design。
2. 打开**设置**。在**模型与提供商**页面选择 **API 提供方**。
3. 在**提供商预设**中选择**自定义提供方**，然后将 **API 协议**设置为 **OpenAI**。
4. 配置提供方：

   - **Base URL**：原样粘贴从 Model Console 复制的 Base URL，包括末尾的 `/v1`。
   - **API Key**：输入任意非空值，例如 `olares`。同一 Olares 集群中的应用调用本地模型时，不需要真实密钥。
   - **模型**：输入从 Model Console 复制的完整 Model name。
   - **最大 tokens（可选）**：输入 `65536`。如需使用更小的上下文窗口，可改为 `32768`。

   <!-- ![在 Open Design 中配置 API 提供方](/images/manual/use-cases/open-design-api-provider.png#bordered) -->

5. 点击**保存**，然后点击**测试**。
6. Open Design 显示连接成功后再继续。

:::tip 先做小规模测试
开始制作多页幻灯片或落地页之前，先让 Open Design 生成一张封面幻灯片。预览成功后，说明 Base URL、模型名称和 token 上限可以正常配合使用。
:::

## 创建设计项目

1. 在 Open Design 中新建项目。
2. 根据目标产物选择 Skill，例如幻灯片、落地页或原型。
3. 选择设计系统。首次使用时，可以直接采用默认设计系统。

   <!-- ![选择 Skill 和设计系统](/images/manual/use-cases/open-design-skill-design-system.png#bordered) -->

4. 输入具体需求。建议说明目标受众、内容、视觉风格和输出格式。例如：

   ```text
   为产品发布会制作一张封面幻灯片。
   使用深色背景、醒目的几何字体和青绿色点缀。
   产品名称为“Northstar”，副标题为“Find your next move.”。
   ```

5. 发送提示词，等待右侧生成预览。

   <!-- ![Open Design 工作区和预览](/images/manual/use-cases/open-design-workspace-preview.png#bordered) -->

6. 在对话中继续调整结果。你可以拖入参考图片，或输入 `@` 添加项目文件，然后说明要修改的部分。
7. 完成后打开导出菜单，根据项目类型下载 ZIP、PDF 或 PPTX 文件。

   <!-- ![从 Open Design 导出项目](/images/manual/use-cases/open-design-export.png#bordered) -->

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

重新复制 Model Console 中的 Base URL，不要修改路径。本指南使用本地模型时，需要在 Model Console 中选择 **OpenAI-Compatible**，并原样使用显示的 URL，包括 `/v1`。

### 设计尚未生成完成便停止

将**最大 tokens（可选）**设置为 `65536` 或 `32768`。多页幻灯片需要的输出 token 比单页或封面幻灯片更多。

### Open Design 一直没有响应

返回模型应用的 Model Console。确认**模型**显示**就绪**，且**引擎**显示**运行中**后再继续。

### Open Design 使用了错误的模型

比较 Open Design 中的**模型**与 Model Console 中的 **Model name**。两者必须完全一致。

### Open Design 无法生成独立图片

这是 Olares 上 Open Design 0.22.1 的当前限制。请改用页面、幻灯片、仪表盘、线框图或交互式原型对应的 Skill。

## 了解更多

- [使用引擎基座应用托管本地大语言模型](llm-base-apps.md)：通过 Model Console 部署和管理模型。
- [将 AI 应用连接到模型服务](/zh/manual/best-practices/connect-ai-apps.md)：了解模型名称、API 格式和 Base URL。
- [Open Design 官网](https://open-design.ai/)：了解上游项目及其功能。
