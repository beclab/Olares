---
outline: [2, 3]
description: Run automated crypto trading strategies on Olares with Freqtrade. Backtest strategies against historical data, paper trade in dry-run mode, and go live with spot or futures bots.
head:
  - - meta
    - name: keywords
      content: Olares, Freqtrade, crypto trading, trading bot, backtest, strategy, OKX, hyperliquid, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-05-16"
---

# Run automated crypto trading with Freqtrade

Freqtrade is an open-source crypto trading bot written in Python. It supports backtesting, paper trading, and live trading on major centralized and decentralized exchanges, and lets you develop custom strategies in Python. Running Freqtrade on Olares gives you a self-hosted trading platform that keeps your strategies, configs, and trade history on your own device.

:::warning Trading involves risk
This software is for educational purposes only. Do not risk money which you are afraid to lose. USE THE SOFTWARE AT YOUR OWN RISK. THE AUTHORS, OLARES, AND ALL AFFILIATES ASSUME NO RESPONSIBILITY FOR YOUR TRADING RESULTS.
:::

## Learning objectives

In this guide, you will learn how to:
- Install Freqtrade and add the bundled bots to the FreqUI.
- Edit bot configurations and upload custom strategies.
- Download historical data and run backtests.
- Switch exchanges and reload configurations.
- Run paper trading or live trading with the spot or futures bot.
- Clone the app to run multiple strategies in parallel.

## How Freqtrade works on Olares

A single Freqtrade installation on Olares includes three independent bots, each running its own configuration, strategy, and data:

| Bot      | Purpose                          | API path        | Default credentials | Data directory                  |
|:---------|:---------------------------------|:----------------|:--------------------|:--------------------------------|
| backtest | Backtest strategies against historical data | `/api-backtest` | admin / admin       | `Application/Data/freqtrade/freqtrade-backtest` |
| spot     | Trade spot markets               | `/api-spot`     | admin / admin       | `Application/Data/freqtrade/freqtrade-spot`     |
| future   | Trade futures markets            | `/api-future`   | admin / admin       | `Application/Data/freqtrade/freqtrade-future`   |

All three bots share the same FreqUI front end. You add each bot to the UI by pointing it to the matching API path.

## Install Freqtrade

1. Open Market and search for "Freqtrade".

   <!-- ![Freqtrade in Market](/images/manual/use-cases/freqtrade.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Add bots to FreqUI

After installation, open Freqtrade from Launchpad. Clicking any tab in FreqUI brings up the bot login prompt. You need to add at least one bot before you can use FreqUI.

<!-- ![Bot login prompt](/images/manual/use-cases/freqtrade-bot-login.png#bordered) -->

To add the backtest bot:

1. In the bot login prompt, fill in the fields:

   - **Bot Name**: Any descriptive name (e.g., `backtest`).
   - **API URL**: Your Freqtrade app endpoint followed by `/api-backtest`. For example, `https://71e1dd45.hzfysttq.olares.com/api-backtest`. You can copy the endpoint from your browser address bar after opening Freqtrade.
   - **Username**: `admin`
   - **Password**: `admin`

   <!-- ![Add bot dialog](/images/manual/use-cases/freqtrade-add-bot-dialog.png#bordered) -->

2. Click **Add** to save the bot.

To add the other bots, click the icon in the top-left corner to expand the bot list, then click **Add New Bot**. Repeat the steps above with the matching API path (`/api-spot` or `/api-future`).

<!-- ![Bot list](/images/manual/use-cases/freqtrade-bot-list.png#bordered) -->

## Configure a bot

All bot configurations are stored under `Application/Data/freqtrade/`. Each bot has its own subfolder containing:

- `config.json`: Exchange, trading pairs, strategy name, and credentials.
- `strategies/`: Strategy files. The bundled `Strategy001.py` is from the official [freqtrade-strategies](https://github.com/freqtrade/freqtrade-strategies) repo and is for demonstration only. Upload your own files here.
- `tradesv3.*`: Stores trade history.

<!-- ![Bot data directory](/images/manual/use-cases/freqtrade-data-directory.png#bordered) -->

The default configurations use:

- **backtest**: OKX (spot)
- **spot**: hyperliquid
- **future**: hyperliquid

To change the exchange, edit the relevant fields in `config.json`. For full configuration reference, see the [Freqtrade documentation](https://www.freqtrade.io/en/stable/).

:::tip Change the default credentials
To change the bot login username and password, edit the `api_server` section in `config.json` and restart Freqtrade.
:::

## Run a backtest

Use the backtest bot to evaluate a strategy against historical data.

### Download historical data

The default backtest configuration uses OKX, but FreqUI defaults to Binance on the data download page. You need to switch the exchange before downloading.

1. In the backtest bot, open the data download page.

2. Expand **Advanced Options** and check **Custom Exchange**.

3. Select **OKX** and set the data type to **spot**.

4. Start the download and wait for the progress bar to reach 100%.

   <!-- ![Download historical data](/images/manual/use-cases/freqtrade-download-data.png#bordered) -->

### Start the backtest

1. Open the **Run Backtest** page.

2. Select the strategy name and timeframe, then click **Start backtest**.

   <!-- ![Run backtest](/images/manual/use-cases/freqtrade-run-backtest.png#bordered) -->

3. Track progress in the bottom-right corner. When the progress indicator disappears, the backtest is complete.

   <!-- ![Backtest progress](/images/manual/use-cases/freqtrade-backtest-progress.png#bordered) -->

### View results

After the backtest finishes, click **Load backtest result**. Each tab shows a different view of the run.

#### Analyze result

Displays the overall backtest summary, including total return, win rate, and drawdown.

<!-- ![Analyze result](/images/manual/use-cases/freqtrade-analyze-result.png#bordered) -->

#### Compare result

Compares performance across strategies. With a single strategy loaded, you see the strategy compared against itself.

<!-- ![Compare result](/images/manual/use-cases/freqtrade-compare-result.png#bordered) -->

#### Visualize summary

Shows profit and loss per trade alongside the strategy equity curve.

<!-- ![Visualize summary](/images/manual/use-cases/freqtrade-visualize-summary.png#bordered) -->

#### Visualize result

Plots entry and exit signals on a kline chart. If your strategy exports plot indicators, you can overlay them on the chart.

To export indicators, add a `plot_config` section to your strategy:

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

To overlay indicators in the chart:

1. Click the config button on the right to open **Plot Configurator**.

2. Click **From strategy** to load the exported indicators.

3. Select all indicators and click **Add new indicator**.

   <!-- ![Plot Configurator](/images/manual/use-cases/freqtrade-plot-configurator.png#bordered) -->

The chart now shows the indicators alongside entry and exit signals.

<!-- ![Kline chart with indicators](/images/manual/use-cases/freqtrade-kline-indicators.png#bordered) -->

## Switch the exchange

To run a backtest on a different exchange:

1. Edit the exchange name in `config.json` for the bot.

2. In Olares Settings, restart the Freqtrade pods to apply the new config:

   a. Open Settings, then go to **Applications** > **Freqtrade**.

   b. Click **Restart** in the right panel.

   c. Enter the pod name and click **Confirm**.

   <!-- ![Restart pod](/images/manual/use-cases/freqtrade-restart-pod.png#bordered) -->

3. Return to the data download page, download data for the new exchange, then re-run the backtest.

## Add a custom strategy

1. Drag your strategy file (for example, `DoubleSmaTrend.py`) into `Application/Data/freqtrade/freqtrade-backtest/strategies/`.

   <!-- ![Upload strategy file](/images/manual/use-cases/freqtrade-upload-strategy.png#bordered) -->

2. Refresh the **Run Backtest** page. The new strategy appears in the strategy dropdown.

3. Select the new strategy and start the backtest.

   <!-- ![Select new strategy](/images/manual/use-cases/freqtrade-select-new-strategy.png#bordered) -->

## Run paper or live trading

Use the spot or future bot for paper trading or live trading. Both ship with `dry_run` set to `true`, which records signals without placing real orders.

### Configure config.json

Edit `Application/Data/freqtrade/freqtrade-spot/config.json` (or `freqtrade-future/config.json`) and update the key fields:

- **dry_run**: Set to `true` for paper trading. Set to `false` to place real orders on the exchange.
- **strategy**: Name of the strategy class in the `strategies/` folder. Freqtrade scans the folder for a file matching this name.
- **exchange**:
  - `name`: Exchange name.
  - For decentralized exchanges, set `walletAddress` and `privateKey`.
  - For centralized exchanges, set `key` and `secret`.
  - `pair_whitelist`: Trading pairs.
- **timeframe**: Trading timeframe (e.g., `5m`, `1h`).
- **entry_pricing**, **exit_pricing**, **order_types**: Order placement options.

### Apply the new configuration

Restart the app for the new config to take effect.

If the **Trade** page shows **Open Trades**, you need to close them before reloading the config:

1. On the **Multi Pane**, click **Force exit all** (rightmost button) to close all open trades.

2. After **Open Trades** is empty, click **Reload Config** (second from the right) to reload `config.json` and the strategy.

   <!-- ![Force exit and reload config](/images/manual/use-cases/freqtrade-reload-config.png#bordered) -->

:::warning Verify before going live
Before setting `dry_run` to `false`, verify your strategy in backtest and dry-run mode. Double-check exchange credentials, pair whitelist, and order types.
:::

## Compare backtest with live data

The backtest bot can pull live market data from an exchange and overlay strategy signals on real-time charts. This lets you sanity-check a strategy against current market conditions.

1. Select the backtest bot.

2. Click **Chart** in the top-left.

3. Select the exchange, strategy, and timeframe.

4. Click the refresh icon to load the chart.

   <!-- ![Chart comparison](/images/manual/use-cases/freqtrade-chart-comparison.png#bordered) -->

## Run multiple strategies in parallel

Each Freqtrade installation runs one strategy per bot. To run additional strategies at the same time, clone the app.

1. In **My Olares**, click the Freqtrade dropdown menu and select the clone option.

2. Enter a new app name and confirm.

   <!-- ![Clone Freqtrade](/images/manual/use-cases/freqtrade-clone-menu.png#bordered) -->

   <!-- ![Clone dialog](/images/manual/use-cases/freqtrade-clone-dialog.png#bordered) -->

Each clone is independent of the original app. Cloned data is stored in a separate `freqtrade+<random>` directory.

<!-- ![Cloned app directory](/images/manual/use-cases/freqtrade-cloned-directory.png#bordered) -->

## FAQ

### Can I run multiple trading strategies at the same time?

Each Freqtrade bot container runs one strategy at a time. To run multiple strategies in parallel, clone the Freqtrade app from **My Olares**. Each clone keeps its own configs, strategies, and data, so the instances do not interfere with each other.

### Where do I find the API URL for each bot?

The API URL is your Freqtrade app endpoint plus the bot's API path. Open Freqtrade from Launchpad, then copy the URL from the browser address bar and append `/api-backtest`, `/api-spot`, or `/api-future`.

### How do I change the default `admin` / `admin` credentials?

Edit the `api_server` section in the bot's `config.json` to set new credentials, then restart Freqtrade.

## Learn more

- [Freqtrade documentation](https://www.freqtrade.io/en/stable/): Official Freqtrade docs.
- [Freqtrade strategies repo](https://github.com/freqtrade/freqtrade-strategies): Sample strategies you can adapt for your own use.
