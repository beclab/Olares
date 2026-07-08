---
outline: [2, 3]
description: 在 Olares 上运行 TradingAgents，通过多个 AI 智能体模拟专业金融交易公司。配置本地模型、运行市场分析并生成交易策略。
head:
  - - meta
    - name: keywords
      content: Olares, TradingAgents, AI trading, multi-agent, local LLM, Ollama, market analysis, financial research
app_version: "1.0.4"
doc_version: "1.0"
doc_updated: "2026-05-27"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/tradingagents.md)为准。
:::

# 使用 TradingAgents 分析金融市场

TradingAgents 是一个多智能体金融交易框架，模拟真实交易公司的运作方式。它部署专门的 AI 智能体来评估市场状况、辩论策略并提供交易决策。这些智能体包括基本面分析师、情绪专家、技术分析师、交易员和风险经理。

:::warning 免责声明
TradingAgents 是一个开源的 AI 交易和市场分析辅助工具。它不提供经过认证的金融投资建议或任何收益保证。

Olares 提供运行 TradingAgents 的平台，但不运营、认可或控制 TradingAgents 软件或通过它进行的任何交易活动。Olares 对软件的分析结果不承担任何责任。

金融市场具有高风险，市场波动可能导致资金部分或全部损失。本指南中的策略、参数和示例仅用于技术演示，不代表交易建议。在交易前请确保您充分了解风险，并对基于这些 AI 生成报告的交易决策承担全部后果。

TradingAgents 项目开发者、贡献者和 Olares 对因使用本项目而导致的任何直接或间接损失不承担责任。
:::

## 学习目标

在本指南中，您将学习如何：
- 在 Olares 上安装 TradingAgents。
- 连接本地 AI 模型为交易智能体提供动力。
- 通过交互式终端运行多智能体市场分析。
- 访问和查看生成的分析报告。

## 前提条件

确保您已在 Olares 上运行本地 AI 模型，使用以下方法之一：
- **Ollama 应用**：一个托管多个模型的应用。确保已[安装 Ollama](ollama.md) 并至少下载了一个模型，例如 `llama3.1:8b`。
- **单模型应用**：将特定模型作为独立应用运行。确保已从 Market 安装模型应用且模型已完全下载，例如 **Qwen3.5 27B Q4_K_M (Ollama)**。

## 安装 TradingAgents

1. 打开 **Market**，搜索 "TradingAgents"。

   ![TradingAgents in Market](/images/manual/use-cases/trading-agents.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。

## 连接本地模型

要为 AI 智能体提供动力，请将 TradingAgents 连接到本地模型。

### 获取模型名称和端点

获取您已准备的本地模型的模型名称和共享端点 URL。选择与您设置匹配的选项卡。

<Tabs>
<template #Ollama>

一个应用，在单个共享端点后面托管多个模型。

1. 从 Launchpad 打开 Ollama 应用。
2. 通过运行以下命令检查已安装的模型：

    ```bash
    ollama list
    ```
3. 复制并保存 **NAME** 列中显示的模型名称。例如，`qwen3.5:27b`。
4. 打开 Settings，然后进入 **Applications** > **Ollama** > **Shared entrances** > **Ollama API**。

   ![Get shared endpoint](/images/manual/use-cases/ollama-shared.png#bordered){width=70%}

5. 记下共享端点 URL。例如，`http://d54536a50.shared.olares.com`。
</template>
<template #Single-model-apps>

每个应用打包一个特定模型并暴露其自己的共享端点。以下示例使用 **Qwen3.5 27B Q4_K_M (Ollama)**。

1. 从 Launchpad 打开 Qwen3.5 27B Q4_K_M (Ollama) 应用，然后记下显示的模型名称。例如，`qwen3.5:27b-q4_K_M`。

   ![Get model name in app](/images/manual/use-cases/qwen3.5-27b-model-name.png#bordered){width=70%}

2. 打开 Settings，然后进入 **Applications** > **Qwen3.5 27B Q4_K_M (Ollama)**。
3. 在 **Shared entrances** 中，点击模型名称查看其端点 URL。

   ![Get shared endpoint](/images/one/qwen3.5-27b-shared-entrance.png#bordered){width=70%}

4. 记下共享端点 URL。例如，`http://94a553e00.shared.olares.com`。

:::tip 为什么不使用模型页面上显示的 URL？
模型应用页面上显示的 URL 是用户特定的，依赖于基于浏览器的前端调用。如果您的设备和 Olares 不在同一本地网络中，这些调用可能会触发 Olares 登录，并且您可能会遇到跨域限制（CORS）。为避免这些问题，请使用共享端点 URL。
:::
</template>
</Tabs>

### 配置 TradingAgents

1. 打开 Settings，然后进入 **Applications** > **TradingAgents** > **Manage environment variables**。
2. 点击 `OLLAMA_BASE_URL` 旁边的 <i class="material-symbols-outlined">edit_square</i>。

   ![Configure environment variables](/images/manual/use-cases/tradingagents-env-vars.png#bordered){width=70%}

3. 输入您之前记下的共享端点 URL，然后在末尾追加 `/v1`。例如，`http://d54536a50.shared.olares.com/v1`。
4. 点击 **Confirm**，然后点击 **Apply**。

   :::tip 连接云模型
   要使用云模型而非本地模型，请编辑相应的 API key 变量，例如 `OPENAI_API_KEY` 或 `ANTHROPIC_API_KEY`。输入您的密钥，点击 **Confirm**，然后点击 **Apply**。
   :::

5. （可选）要设置 UI 中未列出的高级环境变量，请打开 **Files**，进入 **Data** > **tradingagents**，然后打开 `.env` 文件直接添加您的自定义配置。

   ![Edit environment file](/images/manual/use-cases/tradingagents-env-file.png#bordered)

## 运行市场分析

配置好本地模型后，使用应用的交互式命令行界面（CLI）开始市场分析会话。

1. 从 Launchpad 打开 TradingAgents。
2. 在终端窗口中，运行以下命令启动框架：

   ```bash
   tradingagents
   ```

3. 按照屏幕提示逐步配置分析参数：
   
   a. **Ticker Symbol**：输入您想要分析的资产的精确股票代码。例如，`SPY`。
   
   b. **Analysis Date**：以 `YYYY-MM-DD` 格式输入分析的目标日期。例如，`2026-05-20`。
   
   c. **Output Language**：选择生成报告的语言。
   
   d. **Analysts Team**：选择要纳入研究的特定 AI 分析师，例如市场、情绪、新闻和基本面分析师。
   
      :::tip 从核心分析师开始
      新闻和情绪分析师依赖实时外部数据流，如果数据源暂时不可用，可能会失败或返回不准确的结果。对于首次测试运行，请选择市场和基本面分析师以熟悉工作流程。
      :::
   
   e. **Research Depth**：选择智能体研究和辩论策略的深入程度。
   
      :::tip 从浅层深度开始
      更深入的研究需要显著更多的处理时间。对于初始运行，请选择最低深度以了解工作流程，然后再开始全面分析。
      :::
      
   f. **LLM Provider**：选择 **Ollama**。
   
   g. **Thinking Agents**：为 **Quick-Thinking LLM Engine** 和 **Deep-Thinking LLM Engine** 分配特定模型。如果您的本地模型名称未列出，请选择 **Custom model ID**，然后输入精确的模型名称。

   ![Configure CLI parameters](/images/manual/use-cases/tradingagents-cli-parameters.png#bordered)

## 查看分析报告

提交配置后，智能体将进行研究、辩论并做出决策。

1. 直接在终端窗口中监控实时进度、工具使用和内部智能体辩论。

   ![Analysis progress](/images/manual/use-cases/tradingagents-analysis-progress1.png#bordered)

2. 分析完成后，终端会显示执行摘要并提示您保存综合报告。

   ![Analysis complete](/images/manual/use-cases/tradingagents-analysis-complete.png#bordered)

3. 保存报告：

   a. 输入 `y`，然后按 **Enter**。

   b. 再次按 **Enter** 确认默认目录。

4. 要在屏幕上显示完整报告，输入 `y`，然后按 **Enter**。
5. 稍后查看详细报告：

   a. 打开 **Files**，然后进入 **Data** > **tradingagents** > **reports**。

   ![Analysis detail reports in Files](/images/manual/use-cases/tradingagents-reports.png#bordered)

   b. 打开特定的分析文件夹，访问每个分析师团队生成的 markdown 文件和最终的投资组合决策。

   c. 下载 markdown 文件并在支持 markdown 查看的编辑器中打开，以查看详细分析。
