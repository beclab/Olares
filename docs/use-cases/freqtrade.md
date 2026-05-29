---
outline: [2, 3]
description: Run automated crypto trading strategies on Olares with Freqtrade. Backtest strategies against historical data, paper trade in dry-run mode, and go live with the spot or future bot.
head:
  - - meta
    - name: keywords
      content: Olares, Freqtrade, crypto trading, trading bot, backtest, strategy, OKX, hyperliquid, self-hosted
app_version: "1.0.4"
doc_version: "1.1"
doc_updated: "2026-05-29"
---

# Backtest and run crypto trading strategies with Freqtrade

Freqtrade is an open-source crypto trading bot for backtesting, paper trading, and live trading. On Olares, it runs as a self-hosted app with separate bots for backtesting, spot trading, and futures trading.

This guide starts with a safe backtest workflow using the bundled demo strategy. Paper and live trading are covered later for users who are ready to edit configuration files and manage exchange credentials.

:::warning Trading involves risk
This software is for educational purposes only. Do not risk money which you are afraid to lose. USE THE SOFTWARE AT YOUR OWN RISK. THE AUTHORS, OLARES, AND ALL AFFILIATES ASSUME NO RESPONSIBILITY FOR YOUR TRADING RESULTS.
:::

## Learning objectives

In this guide, you will learn how to:
- Add Freqtrade bots to FreqUI.
- Run and review a backtest.
- Manage strategies, exchanges, and trading configuration.

## How Freqtrade works on Olares

A single Freqtrade installation on Olares includes one FreqUI front end and three independent bots:

- `backtest`: Backtests strategies against historical data.
- `spot`: Runs spot market paper or live trading.
- `future`: Runs futures market paper or live trading.

Each bot has its own folder under `Application/Data/freqtrade/`, such as `freqtrade-backtest`, `freqtrade-spot`, or `freqtrade-future`.

## Install Freqtrade

1. Open Market and search for "Freqtrade".

   ![Freqtrade in Market](/images/manual/use-cases/freqtrade.png#bordered){width=95%}

2. Click **Get**, then **Install**, and wait for installation to complete.

## Add bots to FreqUI

After installation, open Freqtrade from Launchpad. You need to add at least one bot before you can use FreqUI. Start with the backtest bot so you can run a safe test without placing real orders.

### Obtain the Freqtrade endpoint

1. Open Settings and go to **Applications** > **Freqtrade**.

2. Under **Entrances**, click **Freqtrade**.

3. On the **Set up endpoint** page, copy the value next to **Endpoint**.

   For example:

   ```text
   https://71e1dd45.hzfysttq.olares.com
   ```

If you use a cloned Freqtrade app, repeat these steps for the cloned app and use its own endpoint.

### Add the backtest bot

When adding a bot to FreqUI, append the bot's API path to the endpoint you copied from Settings.

| Bot | API URL format |
|:--|:--|
| backtest | `<Freqtrade endpoint>/api-backtest` |
| spot | `<Freqtrade endpoint>/api-spot` |
| future | `<Freqtrade endpoint>/api-future` |

For example, if your endpoint is `https://71e1dd45.hzfysttq.olares.com`, the backtest bot API URL is `https://71e1dd45.hzfysttq.olares.com/api-backtest`.

By default, all three bots use `admin` / `admin` as the FreqUI login username and password.

1. Click **Login** in the upper-right corner.

2. On the bot login page, fill in the fields:

   - **Bot Name**: `backtest`
   - **API URL**: Enter the Freqtrade endpoint from Settings with `/api-backtest` appended.
   - **Username**: `admin`
   - **Password**: `admin`

   ![Bot login prompt](/images/manual/use-cases/freqtrade-bot-login.png#bordered){width=95%}

3. Click **Submit**.

After the bot is added, select it from **Available bots**. The backtest pages are now available in FreqUI.

To add the other bots, click **Add new bot** under **Available bots**. Repeat the steps above with the matching API path (`/api-spot` or `/api-future`).

![Add new bot](/images/manual/use-cases/freqtrade-add-new-bot.png#bordered){width=95%}

### Switch between bots

After you add multiple bots, use either method to switch or manage them:

- Open the bot selector in the upper-right corner, then select the bot you want to use.

![Bot selector](/images/manual/use-cases/freqtrade-bot-list.png#bordered){width=95%}

- Click **Freqtrade UI** in the upper-left corner to open the **Available bots** page. From there, select an existing bot or click **Add new bot**.

![Available bot](/images/manual/use-cases/freqtrade-available-bot.png#bordered){width=95%}
  
## Run a backtest

Use the backtest bot to evaluate a strategy against historical data.

### Download historical data

The default backtest configuration uses OKX, but FreqUI defaults to Binance on the data download page. You need to switch the exchange before downloading.

1. Click the backtest bot, then click **Download Data** in the toolbar at the top.

2. Expand **Advanced Options** and check **Custom Exchange**.

3. Select **OKX**, set the data type to **spot**, then click **Start Download**.

Wait until the progress bar reaches 100%. The downloaded data is then available for backtesting.

![Download historical data](/images/manual/use-cases/freqtrade-download-data.png#bordered)

### Start the backtest

1. Click **Backtest** in the top toolbar.

2. Select the strategy name and timeframe, then click **Start backtest**.

   ![Run backtest](/images/manual/use-cases/freqtrade-run-backtest.png#bordered)

3. Track progress in the bottom-right corner. When the progress indicator disappears, the backtest is complete.

### View results

After the backtest finishes, load the result before switching between result views.

1. Open the **Load Results** tab.

   ![Load results](/images/manual/use-cases/freqtrade-load-result.png#bordered)

2. Find the backtest run you want to inspect.

3. Click <i class="material-symbols-outlined">arrow_right_alt</i> to load the run.

:::tip Refresh if the arrow does not respond
If clicking the arrow icon does not respond, refresh the FreqUI page, return to **Load Results**, and try again.
:::

FreqUI provides several result views:

| View | What to check |
|:--|:--|
| **Analyze result** | Overall performance summary, including profit, trade count, win <br>rate, and drawdown. Start here to decide whether the result is <br>worth deeper inspection. |
| **Compare result** | Differences between loaded backtest runs. This is most useful<br> after you run the same strategy with different time ranges,<br> timeframes, trading pairs, or strategy parameters. |
| **Visualize summary** | Profit and loss per trade together with the strategy equity curve.<br> Use it to spot large losses, uneven returns, or long flat sections. |
| **Visualize result** | Entry and exit signals on a kline chart. Use it to check whether the<br> strategy enters and exits at reasonable points. |

### Show strategy indicators on the chart

To show strategy indicators in **Visualize result**, add a `plot_config` section to the Python strategy file in the bot's `strategies/` folder, such as `Strategy001.py`:

```python
plot_config = {
    "main_plot": {
        "ema20":  {"color": "#f39c12"},
        "ema50":  {"color": "#3498db"},
        "ema100": {"color": "#9b59b6"},
        "ha_open":  {"color": "#2ecc71"},
        "ha_close": {"color": "#e74c3c"}
    }
}
```

To add indicators to the chart:

1. Click <i class="material-symbols-outlined">settings</i> on the right to open **Plot Configurator**.

2. Click **From strategy** to load the indicators.

3. Select all indicators and click **Add new indicator**.

   ![Plot Configurator](/images/manual/use-cases/freqtrade-plot-configurator.png#bordered)

The chart now shows the indicators alongside entry and exit signals.

## Manage bot configuration

Each bot keeps its own configuration and strategy files under `Application/Data/freqtrade/`:

| Bot | Folder | Default market |
|:--|:--|:--|
| backtest | `freqtrade-backtest` | OKX spot |
| spot | `freqtrade-spot` | Hyperliquid spot |
| future | `freqtrade-future` | Hyperliquid futures |

In each bot folder:

- Edit `config/config.json` to change the exchange, trading pairs, strategy name, credentials, or API server settings.
- Upload custom strategy files to `strategies/`. The bundled `Strategy001.py` is from the official [freqtrade-strategies](https://github.com/freqtrade/freqtrade-strategies) repo and is for demonstration only.
- For the backtest bot, downloaded historical market data is stored in `data/`.
- Running bots may generate `tradesv3.*` database files for trade records.

:::tip Change the default credentials
To change the bot login username and password, edit the `api_server` section in `config/config.json` and restart Freqtrade.
:::

## Switch the exchange

Change the `config/config.json` file for the target bot, then restart the target pod to apply the changes.

The following example changes the exchange for the backtest bot.

1. Follow the [official guide](https://www.freqtrade.io/en/stable/exchanges/#exchange-configuration) to edit the exchange section in `config/config.json` for the bot.

2. Restart the target Freqtrade pod to apply the new config.

   a. Open Control Hub, go to **Browse**, then select the freqtrade project.

   b. Under **Deployments**, select **freqtrade-backtest**.

   c. Click **Restart** in the right panel.

   d. Enter the pod name and click **Confirm**.

3. Return to the data download page, download data for the new exchange, then re-run the backtest.

## Add a custom strategy

You can add a custom strategy to the target bot. Take the backtest bot as an example.

1. Upload your strategy file (for example, `DoubleSmaTrend.py`) into `Application/Data/freqtrade/freqtrade-backtest/strategies/`.

2. Refresh the **Run Backtest** page. The new strategy appears in the strategy dropdown.

   ![Select new strategy](/images/manual/use-cases/freqtrade-select-new-strategy.png#bordered)

3. Select the new strategy and start the backtest.

## Run paper or live trading

Use the `spot` bot for spot trading or the `future` bot for futures trading. Both ship with `dry_run` set to `true`, which records signals without placing real orders.

:::warning Test before live trading
Live trading requires exchange credentials or wallet credentials supported by Freqtrade. Before setting `dry_run` to `false`, verify your strategy in backtest and dry-run mode. Double-check exchange credentials, pair whitelist, and order types.
:::

### Configure config.json

Edit `Application/Data/freqtrade/freqtrade-spot/config/config.json` (or `freqtrade-future/config/config.json`) and update the key fields:

| Field | What to set |
|:--|:--|
| `dry_run` | Set to `true` for paper trading.<br> Set to `false` to place real orders on the exchange. |
| `strategy` | Enter the strategy name to run. <br>When the bot starts, Freqtrade scans Python files in the<br> `strategies/` folder and loads the strategy class with the<br> matching name. |
| `exchange.name` | Enter the exchange name supported by Freqtrade. |
| `exchange.walletAddress` and `exchange.privateKey` | Set these for decentralized exchanges that require wallet<br> credentials. |
| `exchange.key` and `exchange.secret` | Set these for centralized exchanges that require API<br> credentials. |
| `exchange.pair_whitelist` | List the trading pairs the bot can trade. |
| `timeframe` | Set the trading timeframe, such as `5m` or `1h`. |
| `entry_pricing`, `exit_pricing`, `order_types` | Configure order pricing and order placement behavior. |

### Apply the new configuration

Restart the target pod from Control Hub for the new config to take effect.

If the **Trade** page shows **Open Trades**, you need to close them before reloading the config:

1. On the **Multi Pane**, click <i class="material-symbols-outlined">tab_close</i> to close all open trades.

2. After **Open Trades** is empty, click <i class="material-symbols-outlined">refresh</i> to reload `config.json` and the strategy.

## Compare backtest with live data

The backtest bot can pull live market data from an exchange and overlay strategy signals on real-time charts. This lets you sanity-check a strategy against current market conditions.

1. Select the backtest bot.

2. Click **Chart** in the top-left.

3. Set the exchange, strategy, timeframe and other parameters.

4. Click the refresh icon to load the chart.

## FAQ

### Can I run multiple trading strategies at the same time?

Each Freqtrade bot container runs one strategy at a time. To run additional strategies in parallel, clone the Freqtrade app:

1. Open Market.

2. In My Olares, click the Freqtrade dropdown menu and select **Clone**.

3. Enter a new app name and confirm.

Each clone is independent of the original app. Cloned data is stored in a separate `freqtrade+<random 6 characters>` directory.

![Cloned app directory](/images/manual/use-cases/freqtrade-cloned-directory.png#bordered)

## Learn more

- [Freqtrade documentation](https://www.freqtrade.io/en/stable/): Official Freqtrade docs.
- [Freqtrade strategies repo](https://github.com/freqtrade/freqtrade-strategies): Sample strategies you can adapt for your own use.
