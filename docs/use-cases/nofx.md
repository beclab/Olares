---
outline: deep
description: Run NOFX, an open-source autonomous AI trading agent, on Olares. Fund an AI wallet, connect an exchange, configure a strategy, and let the agent trade.
head:
  - - meta
    - name: keywords
      content: Olares, NOFX, AI trading, autonomous agent, crypto, Hyperliquid, self-hosted
app_version: "1.0.4"
doc_version: "1.0"
doc_updated: "2026-04-30"
---

# Set up an autonomous AI trading agent with NOFX

NOFX is an open-source autonomous AI trading agent. Unlike traditional AI tools that require you to configure models, manage API keys, and wire up data sources, NOFX senses the market, picks a model, and pulls the data on its own. You set the strategy, and the agent handles the rest.

:::warning Disclaimer
NOFX is an open-source AI trading aid. It does not provide investment advice or any guarantee of returns.

Digital asset trading carries high risk, and market volatility can lead to partial or total loss of funds. The strategies, parameters, and examples in this guide are for technical demonstration only and do not represent trading advice. Ensure you fully understand the risks before trading, and bear the full consequences of your trading decisions.

The NOFX project developers and contributors are not responsible for any direct or indirect losses resulting from the use of this project.
:::

## Learning objectives

In this guide, you will learn how to:
- Install NOFX on Olares.
- Create an account and fund your AI wallet.
- Connect an exchange account for trading.
- Create a trading strategy and configure your trading bot.
- Start the trading bot and monitor its performance.

## Prerequisites

- An active exchange account funded with USDC on the corresponding chain. For example, Hyperliquid requires USDC on the Arbitrum chain.

## Install NOFX

1. Open Market and search for "NOFX".

   ![NOFX in Market](/images/manual/use-cases/nofx.png#bordered)

2. Click **Get**, and then click **Install**. Wait for installation to finish.

## Sign up and fund the AI wallet

NOFX uses an internal AI wallet to pay for model calls through the [x402 protocol](https://www.x402.org/). You fund this wallet with USDC on the Base chain. NOFX automatically pays per call when the agent invokes a model.

1. Open NOFX from the Launchpad. On first launch, register with your email and password.
   
   ![NOFX sign-up screen](/images/manual/use-cases/nofx-signup.png#bordered)

2. After you sign up, NOFX generates a wallet address and the corresponding private key. Store the private key in a secure location and never share it with anyone.

   ![NOFX wallet generated](/images/manual/use-cases/nofx-wallet.png#bordered)

3. Click **I saved it, continue**.
4. On the **Config** page, click **OPEN AI WALLET**.
   
   ![Open AI Wallet](/images/manual/use-cases/nofx-open-ai-wallet.png#bordered)

5. In the **Edit AI Model** window, scroll down to the **Deposit USDC (Base Chain)** section, and then deposit USDC into the wallet. Start with about 10 USDC to test the flow before depositing the amount you plan to use for live strategies.

   :::warning Use USDC on the Base chain
   Only USDC on the Base chain is recognized. If your USDC is on another chain, use the wallet's swap feature to bridge it to Base first.
   :::

   ![Deposit screen](/images/manual/use-cases/nofx-deposit.png#bordered){width=70%}

6. After the deposit transaction completes, refresh the page to update the balance. If the balance does not update, click **Test Connection** to check the connectivity.

   Once funded, your balance appears in the **AI MODELS** section.
   
   ![AI Models with wallet](/images/manual/use-cases/nofx-ai-models.png#bordered)

7. Click **OPEN AI WALLET** again, select an AI model to use, scroll down to the bottom of the window, and then click **Start Trading**.

   ![Select model](/images/manual/use-cases/nofx-select-model.png#bordered){width=70%}

## Connect an exchange account

To allow NOFX to execute trades on your behalf securely, you must connect it to your exchange. This is a two-part process: first, you create a dedicated API wallet on the exchange to generate access keys, and then you add those keys to NOFX. 

NOFX supports multiple exchanges, which all share the same configuration process. The following steps use Hyperliquid as the example.

### Create a Hyperliquid API wallet

An API wallet acts as a secure bridge, allowing NOFX to trade without ever exposing your main account credentials.

:::info Prerequisite
Before you create an API wallet, ensure you have an active Hyperliquid account funded with USDC on the Arbitrum chain. If you are new to Hyperliquid, follow the [Hyperliquid onboarding guide](https://hyperliquid.gitbook.io/hyperliquid-docs/onboarding/how-to-start-trading) to get started.
:::

1. Open the Hyperliquid web interface at https://app.hyperliquid.xyz, and then log in to your account.
2. On the top menu bar, go to **More** > **API**.
3. Create a new API wallet.

   ![Hyperliquid API page](/images/manual/use-cases/nofx-hyperliquid-api-page.png#bordered)
   
4. Authorize the API wallet by signing the prompt with your main Hyperliquid wallet.
5. Copy the API wallet address and private key, and store them in a secure location. You need these credentials for the next step.

### Add Hyperliquid to NOFX

Now that you have your API credentials, you add them to NOFX so the agent can actively communicate with your exchange.

1. In NOFX, go to the **Config** page.
2. On the **Add Exchange** card, click **Configure**.
3. On the **Select Exchange** screen, choose **Hyperliquid**.
   
   ![Select Hyperliquid](/images/manual/use-cases/nofx-add-exchange.png#bordered){width=70%}

4. On the **Configure** screen, fill in the following fields:

    - **Account Name**: Enter a memorable name for this exchange account, such as `Main Account`.
    - **Agent Private Key**: Enter the private key of the API wallet you created on Hyperliquid.
    - **Main Wallet Address**: Enter the address of your main Hyperliquid trading account that holds the funds. The Agent wallet signs transactions, while the main wallet keeps the balance and remains unexposed.

   ![Configure Hyperliquid](/images/manual/use-cases/nofx-configure-hyperliquid.png#bordered){width=70%}

5. Click **Save Configuration**. Once finished, Hyperliquid appears in the **EXCHANGES** section.

   ![Hyperliquid configured](/images/manual/use-cases/nofx-hyperliquid-configured.png#bordered){width=90%}

## Create a trading strategy

1. On the **Pick Strategy** card, click **Open strategy** to open the strategy editor.
   
   ![Open strategy](/images/manual/use-cases/nofx-open-strategy.png#bordered)

2. Select a default strategy from the left panel, select **Static List** as the coin source, and add **BTC** and **ETH** as the trading pairs.
   
   ![Strategy config](/images/manual/use-cases/nofx-strategy-config.png#bordered)

3. To fine-tune the strategy, configure the following panels:
    - **Indicators**: Choose the technical indicators the strategy uses.
    - **Risk control**: Set risk management rules.
    - **Prompt Editor**: Edit the prompt the AI uses for decision-making.
    - **Extra Prompt**: Enter an extra prompt appended to the System Prompt for a personalized trading style.

4. Click **Save**, and then click **Activate** to enable the strategy.

## Configure your trading bot

A Trader represents your automated trading bot. Creating a Trader combines your chosen AI model, exchange connection, and trading strategy into a single runnable trading bot.

1. On the **Create Trader** card, click **Create now**.

   ![Create trader](/images/manual/use-cases/nofx-create-trader.png#bordered)

2. Enter the trader name, and select the AI model and exchange you configured earlier.
   
   ![Select trader components](/images/manual/use-cases/nofx-config-trader.png#bordered){width=70%}

3. Select the trading strategy you configured earlier.
4. Specify the trading parameters.
5. Click **Create Trader**.

## Start and monitor your trading bot

1. In the **Current Traders** section, find the new Trader you just created.

   ![Trader created](/images/manual/use-cases/nofx-trader-created.png#bordered)

2. Start the Trader by clicking **Start**.
3. Go to the **Dashboard** page to watch the strategy run in real time.
   
   ![Dashboard](/images/manual/use-cases/nofx-dashboard.png#bordered)

4. To stop the Trader, return to the **Config** page, and then click **Stop** in the same section.

## FAQs

### What is the difference between the AI wallet and the API wallet?

NOFX requires two different wallets for distinct purposes:

* **AI wallet**: Lives inside NOFX and pays for the AI model's data analysis and decision-making process. You fund this wallet with USDC on the Base chain.
* **API wallet**: Lives on your exchange, such as Hyperliquid. It does not hold funds. Instead, it securely grants NOFX permission to execute trades using the funds in your main exchange account.

### How to switch models?

1. On the **Config** page, click **+ MODELS_CONFIG**.

   ![Switch models](/images/manual/use-cases/nofx-models-config.png#bordered)

2. Click **Other API Providers**, and then select another provider. For example, **Qwen**.
3. Enter the API key. The base URL and model name are optional.

   ![Configure models](/images/manual/use-cases/nofx-models-config-api.png#bordered){width=70%}

4. Click **Save Configuration**. API keys are encrypted before storage. After saving, test the connection to verify it works.

### Can I use a local model?

The current version does not support local models.

### Model didn't output structured JSON decision

Some models cannot reliably produce decisions in JSON format. To work around this issue, try one of the following methods:
- Switch to a different model.
- Adjust the model parameters.
- Edit the prompt below the model settings to instruct the model to output valid JSON format.
