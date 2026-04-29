---
outline: deep
description: Run NOFX, an open-source autonomous AI trading agent, on Olares. Fund an AI wallet, connect an exchange, configure a strategy, and let the agent trade.
head:
  - - meta
    - name: keywords
      content: Olares, NOFX, AI trading, autonomous agent, crypto, Hyperliquid, x402, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-29"
---

# Run NOFX as your autonomous AI trading agent

NOFX is an open-source autonomous AI trading agent. Unlike traditional AI tools that require you to configure models, manage API keys, and wire up data sources, NOFX senses the market, picks a model, and pulls the data on its own. You set the strategy, and the agent handles the rest.

This guide uses Hyperliquid as the example exchange.

:::warning Disclaimer
NOFX is an open-source AI trading aid. It does not provide investment advice or any guarantee of returns.

Digital asset trading carries high risk, and market volatility can lead to partial or total loss of funds. The strategies, parameters, and examples in this guide are for technical demonstration only and do not represent trading advice. Make sure you fully understand the risks before trading, and bear the full consequences of your trading decisions.

The NOFX project developers and contributors are not responsible for any direct or indirect losses resulting from the use of this project.
:::

## Learning objectives

In this guide, you will learn how to:
- Install NOFX on Olares and fund the AI wallet.
- Connect an exchange for trading.
- Configure a trading strategy and create a Trader.
- Start the Trader and monitor its performance.

## Prerequisites

- An Olares device with admin privileges to install apps from Market.
- An exchange account with USDC funded on the chain the exchange uses. Hyperliquid uses USDC on Arbitrum.

## Install NOFX

1. Open Market and search for "NOFX".
   <!-- ![NOFX in Market](/images/manual/use-cases/nofx.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Sign up and fund the AI wallet

NOFX uses an internal AI wallet to pay for model calls through the [x402 protocol](https://www.x402.org/). You fund this wallet with USDC on the Base chain, and NOFX automatically pays per call when the agent invokes a model.

1. Open NOFX from Launchpad. On first launch, register with your email and password.
   <!-- ![NOFX sign-up screen](/images/manual/use-cases/nofx-signup.png#bordered) -->

2. After you sign up, NOFX generates a wallet address and the corresponding private key. Store the private key in a secure location and never share it with anyone.
   <!-- ![NOFX wallet generated](/images/manual/use-cases/nofx-wallet.png#bordered) -->

3. Click **Open AI Wallet** to open the wallet panel.
   <!-- ![Open AI Wallet](/images/manual/use-cases/nofx-open-ai-wallet.png#bordered) -->

4. Deposit USDC into the wallet. Start with about 10 USDC to test the flow before depositing the amount you plan to use for live strategies.

   :::warning Use USDC on the Base chain
   Only USDC on the Base chain is recognized. If your USDC is on another chain, use the wallet's swap feature to bridge it to Base first.
   :::

   <!-- ![Deposit screen](/images/manual/use-cases/nofx-deposit.png#bordered) -->

5. After the deposit transaction completes, refresh the page to update the balance. If the balance does not update, click **Test Connection** to check the connectivity.

   Once funded, your wallet appears in the **AI Models** section.
   <!-- ![AI Models with wallet](/images/manual/use-cases/nofx-ai-models.png#bordered) -->

6. In **AI Models**, click **Open AI Wallet**, choose a model to use, and click **Start Trading**.
   <!-- ![Select model](/images/manual/use-cases/nofx-select-model.png#bordered) -->

## Configure an exchange

NOFX supports multiple exchanges, all configured the same way. The steps below show how to add Hyperliquid.

### Create a Hyperliquid API Wallet

:::info Hyperliquid runs on Arbitrum
Hyperliquid trades USDC on the Arbitrum chain. Make sure your Hyperliquid account has USDC deposited on Arbitrum before creating an API Wallet.
:::

1. If you don't have a Hyperliquid account yet, follow the [Hyperliquid onboarding guide](https://hyperliquid.gitbook.io/hyperliquid-docs/onboarding/how-to-start-trading) to create a wallet and deposit USDC.
   <!-- ![Hyperliquid deposit](/images/manual/use-cases/nofx-hyperliquid-deposit.png#bordered) -->

2. After you log in, open <https://app.hyperliquid.xyz/API>.
   <!-- ![Hyperliquid API page](/images/manual/use-cases/nofx-hyperliquid-api-page.png#bordered) -->

3. Create a new API Wallet from the API page.
   <!-- ![Create API wallet](/images/manual/use-cases/nofx-hyperliquid-create-api-wallet.png#bordered) -->

4. Authorize the API Wallet by signing the prompt with your main Hyperliquid wallet.
   <!-- ![Authorize API wallet](/images/manual/use-cases/nofx-hyperliquid-authorize-api-wallet.png#bordered) -->

5. Copy the API Wallet address and private key, and store them in a secure location. You'll paste them into NOFX next.
   <!-- ![Saved API wallet credentials](/images/manual/use-cases/nofx-hyperliquid-api-credentials.png#bordered) -->

### Add Hyperliquid to NOFX

1. In NOFX, open the exchange configuration page.
   <!-- ![Exchange config](/images/manual/use-cases/nofx-exchange-config.png#bordered) -->

2. Click **Add Exchange**, then select **Hyperliquid**.
   <!-- ![Select Hyperliquid](/images/manual/use-cases/nofx-add-exchange.png#bordered) -->

3. On the **Configure** screen, fill in the following fields:
    - **Account Name**: A memorable name for this exchange account, such as `Main Account`.
    - **Agent Private Key**: The private key of the API Wallet you created on Hyperliquid.
    - **Main Wallet Address**: The address of your main Hyperliquid trading account that holds the funds. The Agent Wallet signs transactions, while the main wallet keeps the balance and is never exposed.

   <!-- ![Configure Hyperliquid](/images/manual/use-cases/nofx-configure-hyperliquid.png#bordered) -->

4. Click **Save Configuration**.

## Configure a trading strategy

1. Click **Open Strategy** to open the strategy editor.
   <!-- ![Open strategy](/images/manual/use-cases/nofx-open-strategy.png#bordered) -->

2. Select a default strategy from the left panel. Set **Coin Source** to **Static**, and add **BTC** and **ETH** as the trading pairs.
   <!-- ![Strategy config](/images/manual/use-cases/nofx-strategy-config.png#bordered) -->

3. In the panels below, fine-tune the strategy:
    - **Parameters**: Adjust the strategy's parameters.
    - **Indicators**: Choose the technical indicators the strategy uses.
    - **Risk control**: Set risk management rules.
    - **Prompt**: Edit the prompt the AI uses for decision-making.

4. Click **Save**, then **Activate** to enable the strategy.

## Create a Trader

A Trader combines an AI model, an exchange, and a strategy into a runnable trading bot.

1. On the **Config** page, click **Create Trader**.

2. Select the AI model, exchange, and strategy you configured earlier.
   <!-- ![Select trader components](/images/manual/use-cases/nofx-create-trader.png#bordered) -->

3. Set the trading parameters and save the Trader.
   <!-- ![Trader parameters](/images/manual/use-cases/nofx-trader-parameters.png#bordered) -->

## Start and monitor the Trader

1. On the **Config** page, start the Trader you just created.
   <!-- ![Start trader](/images/manual/use-cases/nofx-start-trader.png#bordered) -->

2. Switch to the **Dashboard** page to watch the strategy run in real time.
   <!-- ![Dashboard](/images/manual/use-cases/nofx-dashboard.png#bordered) -->

To stop the Trader, return to the **Config** page and stop it from the same controls.

## FAQs

### How do I switch models?

On the **Config** page, go to the **AI Models** section.
<!-- ![AI Models section on Config page](/images/manual/use-cases/nofx-config-ai-models.png#bordered) -->

For each model:

1. Get an API key from the provider. The UI links to the provider's API key page.
2. Enter the API key. Optionally customize the base URL and model name.
   <!-- ![Enter API key for the model](/images/manual/use-cases/nofx-configure-model-api-key.png#bordered) -->

3. Save the configuration.

API keys are encrypted before storage. After saving, test the connection to verify it works.

### Can I use a local model?

The current version does not support local models.

### The model didn't output a structured JSON decision

Some models cannot reliably produce decisions in JSON format. To work around this, try one of the following:
- Switch to a different model.
- Adjust the model parameters.
- Edit the prompt below the model settings to nudge the output toward valid JSON.

<!-- ![JSON decision error](/images/manual/use-cases/nofx-json-decision-error.png#bordered) -->
