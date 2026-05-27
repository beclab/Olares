---
outline: [2, 3]
description: Run TradingAgents on Olares to simulate a professional financial trading firm with multiple AI agents. Configure local models, run market analysis, and generate trading strategies.
head:
  - - meta
    - name: keywords
      content: Olares, TradingAgents, AI trading, multi-agent, local LLM, Ollama, market analysis, financial research
app_version: "1.0.4"
doc_version: "1.0"
doc_updated: "2026-05-27"
---

# Analyze financial markets with TradingAgents

TradingAgents is a multi-agent financial trading framework that simulates a real-world trading firm. It deploys specialized AI agents to assess market conditions, debate strategies, and provide trading decisions. These agents include fundamental analysts, sentiment experts, technical analysts, traders, and risk managers. 

:::warning Disclaimer
TradingAgents is an open-source AI trading and market analysis aid. It does not provide certified financial investment advice or any guarantee of returns.

Olares provides the platform to run TradingAgents but does not operate, endorse, or control the TradingAgents software or any trading activities conducted through it. Olares assumes no responsibility for the software's analysis outcomes.

Financial markets carry high risk, and market volatility can lead to partial or total loss of funds. The strategies, parameters, and examples in this guide are for technical demonstration only and do not represent trading advice. Ensure you fully understand the risks before trading, and bear the full consequences of your trading decisions based on these AI-generated reports.

The TradingAgents project developers, contributors, and Olares are not responsible for any direct or indirect losses resulting from the use of this project.
:::

## Learning objectives

In this guide, you will learn how to:
- Install TradingAgents on Olares.
- Connect local AI models to power your trading agents.
- Run a multi-agent market analysis via the interactive terminal.
- Access and review the generated analysis reports.

## Prerequisites

Ensure you have a local AI model running on Olares using one of the following methods:
- **Ollama application**: One app that hosts multiple models. Ensure [Ollama is installed](ollama.md) with at least one model downloaded, such as `llama3.1:8b`.
- **Single-model application**: Runs one specific model as a standalone application. Ensure a model app is installed from Market with the model fully downloaded, such as **Qwen3.5 27B Q4_K_M (Ollama)**.

## Install TradingAgents

1. Open **Market**, and search for "TradingAgents".

   ![TradingAgents in Market](/images/manual/use-cases/trading-agents.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Connect a local model

To power your AI agents, connect TradingAgents to a local model.

### Get model name and endpoint

Obtain the model name and shared endpoint URL for the local model you have prepared. Select the tab that matches your setup.

<Tabs>
<template #Ollama>

One app that hosts multiple models behind a single shared endpoint.

1. Open the Ollama app from the Launchpad.
2. Check the installed models by running the following command:

    ```bash
    ollama list
    ```
3. Copy and save your model name exactly as shown in the **NAME** column. For example, `qwen3.5:27b`.
4. Open Settings, and then go to **Applications** > **Ollama** > **Shared entrances** > **Ollama API**.

   ![Get shared endpoint](/images/manual/use-cases/ollama-shared.png#bordered){width=70%}

5. Note down the shared endpoint URL. For example, `http://d54536a50.shared.olares.com`.
</template>
<template #Single-model-apps>

Each app packages one specific model and exposes its own shared endpoint. The following example uses **Qwen3.5 27B Q4_K_M (Ollama)**.

1. Open the Qwen3.5 27B Q4_K_M (Ollama) app from the Launchpad, and then note down the model name exactly as shown. For example, `qwen3.5:27b-q4_K_M`.

   ![Get model name in app](/images/manual/use-cases/qwen3.5-27b-model-name.png#bordered){width=70%}

2. Open Settings, and then go to **Applications** > **Qwen3.5 27B Q4_K_M (Ollama)**.
3. In **Shared entrances**, click the model name to view its endpoint URL.

   ![Get shared endpoint](/images/one/qwen3.5-27b-shared-entrance.png#bordered){width=70%}

4. Note down the shared endpoint URL. For example, `http://94a553e00.shared.olares.com`.

:::tip Why not use the URL shown on the model page?
The URL shown on the model app page is user-specific and relies on browser-based frontend calls. If your device and Olares are not on the same local network, those calls might trigger Olares sign-in and you might encounter cross-origin restrictions (CORS). To avoid these issues, use the shared endpoint URL.
:::
</template>
</Tabs>

### Configure TradingAgents

1. Open Settings, and then go to **Applications** > **TradingAgents** > **Manage environment variables**.
2. Click <i class="material-symbols-outlined">edit_square</i> next to `OLLAMA_BASE_URL`.

   ![Configure environment variables](/images/manual/use-cases/tradingagents-env-vars.png#bordered){width=70%}

3. Enter the shared endpoint URL you noted down earlier, then append `/v1` to the end. For example, `http://d54536a50.shared.olares.com/v1`.
4. Click **Confirm**, and then click **Apply**.

   :::tip Connect a cloud model
   To use a cloud model instead of a local model, edit the corresponding API key variable, such as `OPENAI_API_KEY` or `ANTHROPIC_API_KEY`. Enter your key, click **Confirm**, and then click **Apply**.
   :::

5. (Optional) To set advanced environment variables that are not listed in the UI, open **Files**, go to **Data** > **tradingagents**, and then open the `.env` file to add your custom configurations directly.

   ![Edit environment file](/images/manual/use-cases/tradingagents-env-file.png#bordered)

## Run a market analysis

With your local model configured, start a market analysis session using the app's interactive command-line interface (CLI).

1. Open TradingAgents from the Launchpad.
2. In the terminal window, run the following command to start the framework:

   ```bash
   tradingagents
   ```

3. Follow the on-screen prompts to configure your analysis parameters step by step:
   
   a. **Ticker Symbol**: Enter the exact ticker code for the asset you want to analyze. For example, `SPY`.
   
   b. **Analysis Date**: Enter the target date for the analysis in the `YYYY-MM-DD` format. For example, `2026-05-20`.
   
   c. **Output Language**: Select the language for the generated reports.
   
   d. **Analysts Team**: Select the specific AI analysts to include in the research, such as market, sentiment, news, and fundamental analysts.
   
      :::tip Start with core analysts
      News and sentiment analysts rely on live external data streams, which might fail or return inaccurate results if the sources are temporarily unavailable. For your first test run, select the market and fundamental analysts to familiarize yourself with the workflow.
      :::
   
   e. **Research Depth**: Choose how thoroughly the agents should research and debate the strategy.
   
      :::tip Start with the shallow depth
      Deeper research requires significantly more time to process. Select the lowest depth for your initial run to understand the workflow before starting a comprehensive analysis.
      :::
      
   f. **LLM Provider**: Select **Ollama**.
   
   g. **Thinking Agents**: Assign specific models for **Quick-Thinking LLM Engine** and **Deep-Thinking LLM Engine**. If your local model name is not listed, select **Custom model ID**, and then enter the exact model name.

   ![Configure CLI parameters](/images/manual/use-cases/tradingagents-cli-parameters.png#bordered)

## Review analysis reports

After you submit the configuration, the agents research, debate, and make decisions.

1. Monitor the live progress, tool usage, and internal agent debates directly in the terminal window.

   ![Analysis progress](/images/manual/use-cases/tradingagents-analysis-progress1.png#bordered)

2. When the analysis finishes, the terminal displays an executive summary and prompts you to save the comprehensive report.

   ![Analysis complete](/images/manual/use-cases/tradingagents-analysis-complete.png#bordered)

3. To save the report:

   a. Type `y`, and then press **Enter**.

   b. Press **Enter** again to confirm the default directory.

4. To display the full report on screen, type `y`, and then press **Enter**.
5. To view the detailed reports later:

   a. Open **Files**, and then go to **Data** > **tradingagents** > **reports**.

   ![Analysis detail reports in Files](/images/manual/use-cases/tradingagents-reports.png#bordered)

   b. Open the specific analysis folder to access the markdown files generated by each analyst team and the final portfolio decision.

   c. Download the markdown files and open them in an editor that supports markdown viewing to see the detailed analysis.