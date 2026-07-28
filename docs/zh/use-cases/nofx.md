---
outline: deep
description: 在 Olares 上运行 NOFX，一个开源的自主 AI 交易智能体。为 AI 钱包充值，连接交易所，配置策略，并让智能体进行交易。
head:
  - - meta
    - name: keywords
      content: Olares, NOFX, AI trading, autonomous agent, crypto, Hyperliquid, self-hosted
app_version: "1.0.5"
doc_version: "1.1"
doc_updated: "2026-05-09"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/nofx.md)为准。
:::

# 使用 NOFX 设置自主 AI 交易智能体

NOFX 是一个开源的自主 AI 交易智能体。与传统需要你配置模型、管理 API 密钥和连接数据源的 AI 工具不同，NOFX 感知市场、选择模型，并自行拉取数据。你设置策略，智能体处理其余部分。

:::warning 免责声明
NOFX 是一个开源的 AI 交易辅助工具。它不提供投资建议或任何回报保证。

Olares 提供运行 NOFX 的平台，但不运营、认可或控制 NOFX 软件或通过它进行的任何交易活动。Olares 不是任何交易的一方，也不对软件的功能、安全性或交易结果承担任何责任。

数字资产交易具有高风险，市场波动可能导致资金部分或全部损失。本指南中的策略、参数和示例仅供技术演示，不代表交易建议。在交易前确保你充分理解风险，并承担交易决策的全部后果。

NOFX 项目开发者、贡献者和 Olares 不对使用本项目造成的任何直接或间接损失负责。
:::

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 上安装 NOFX。
- 创建账户并为你的 AI 钱包充值。
- 连接交易所账户进行交易。
- 创建交易策略并配置你的交易机器人。
- 启动交易机器人并监控其性能。

## 前提条件

- 一个活跃的交易所账户，在相应链上充有 USDC。例如，Hyperliquid 需要 Arbitrum 链上的 USDC。

## 安装 NOFX

1. 打开 Market 并搜索 "NOFX"。

   ![Market 中的 NOFX](/images/manual/use-cases/nofx.png#bordered)

2. 点击 **获取**，然后点击 **安装**。等待安装完成。

## 注册并为 AI 钱包充值

NOFX 使用内部 AI 钱包通过 [x402 协议](https://www.x402.org/) 支付模型调用费用。你用 Base 链上的 USDC 为这个钱包充值。当智能体调用模型时，NOFX 自动按调用付费。

1. 从 Launchpad 打开 NOFX。首次启动时，使用你的电子邮件和密码注册。

   ![NOFX 注册屏幕](/images/manual/use-cases/nofx-signup.png#bordered)

2. 注册后，NOFX 生成一个钱包地址和相应的私钥。将私钥存放在安全位置，切勿与任何人分享。

   ![NOFX 钱包已生成](/images/manual/use-cases/nofx-wallet.png#bordered)

3. 点击 **我已保存，继续**。
4. 在 **配置** 页面上，点击 **打开 AI 钱包**。

   ![打开 AI 钱包](/images/manual/use-cases/nofx-open-ai-wallet.png#bordered)

5. 在 **编辑 AI 模型** 窗口中，向下滚动到 **存入 USDC (Base 链)** 部分，然后将 USDC 存入钱包。先存入约 10 USDC 来测试流程，然后再存入你计划用于实盘策略的金额。

   :::warning 使用 Base 链上的 USDC
   只识别 Base 链上的 USDC。如果你的 USDC 在另一条链上，请使用钱包的兑换功能先将其桥接到 Base。
   :::

   ![存款屏幕](/images/manual/use-cases/nofx-deposit.png#bordered){width=70%}

6. 存款交易完成后，刷新页面以更新余额。如果余额没有更新，点击 **测试连接** 检查连接性。

   充值后，你的余额将出现在 **AI 模型** 部分。

   ![带钱包的 AI 模型](/images/manual/use-cases/nofx-ai-models.png#bordered)

7. 再次点击 **打开 AI 钱包**，选择一个要使用的 AI 模型，向下滚动到窗口底部，然后点击 **开始交易**。

   ![选择模型](/images/manual/use-cases/nofx-select-model.png#bordered){width=70%}

## 连接交易所账户

为了允许 NOFX 代表你安全地执行交易，你必须将其连接到你的交易所。这是一个两部分过程：首先，你在交易所创建一个专用的 API 钱包以生成访问密钥，然后你将这些密钥添加到 NOFX。

NOFX 支持多个交易所，它们都共享相同的配置过程。以下步骤使用 Hyperliquid 作为示例。

### 创建 Hyperliquid API 钱包

API 钱包充当安全的桥梁，允许 NOFX 交易而无需暴露你的主账户凭据。

:::info 前提条件
在创建 API 钱包之前，请确保你有一个活跃的 Hyperliquid 账户，在 Arbitrum 链上至少充有 10 USDC。Hyperliquid 要求此最低余额才能解锁创建 API 钱包的能力。

如果你是 Hyperliquid 的新用户，请按照 [Hyperliquid 入门指南](https://hyperliquid.gitbook.io/hyperliquid-docs/onboarding/how-to-start-trading) 开始使用。
:::

1. 在 https://app.hyperliquid.xyz 打开 Hyperliquid Web 界面，然后登录你的账户。
2. 在顶部菜单栏上，前往 **更多** > **API**。
3. 创建一个新的 API 钱包。

   ![Hyperliquid API 页面](/images/manual/use-cases/nofx-hyperliquid-api-page.png#bordered)

4. 通过使用你的主 Hyperliquid 钱包签名提示来授权 API 钱包。
5. 复制 API 钱包地址和私钥，并将它们存放在安全位置。下一步需要这些凭据。

### 将 Hyperliquid 添加到 NOFX

现在你已经有了 API 凭据，将它们添加到 NOFX 中，以便智能体可以主动与你的交易所通信。

1. 在 NOFX 中，前往 **配置** 页面。
2. 在 **添加交易所** 卡片上，点击 **配置**。
3. 在 **选择交易所** 屏幕上，选择 **Hyperliquid**。

   ![选择 Hyperliquid](/images/manual/use-cases/nofx-add-exchange.png#bordered){width=70%}

4. 在 **配置** 屏幕上，填写以下字段：

   - **账户名称**：输入此交易所账户的易记名称，例如 `主账户`。
   - **智能体私钥**：输入你在 Hyperliquid 上创建的 API 钱包的私钥。
   - **主钱包地址**：输入持有资金的主 Hyperliquid 交易账户的地址。智能体钱包签署交易，而主钱包保持余额且不被暴露。

   ![配置 Hyperliquid](/images/manual/use-cases/nofx-configure-hyperliquid.png#bordered){width=70%}

5. 点击 **保存配置**。完成后，Hyperliquid 将出现在 **交易所** 部分。

   ![Hyperliquid 已配置](/images/manual/use-cases/nofx-hyperliquid-configured.png#bordered){width=90%}

## 创建交易策略

1. 在 **选择策略** 卡片上，点击 **打开策略** 以打开策略编辑器。

   ![打开策略](/images/manual/use-cases/nofx-open-strategy.png#bordered)

2. 从左侧面板选择一个默认策略，选择 **静态列表** 作为币源，并添加 **BTC** 和 **ETH** 作为交易对。

   ![策略配置](/images/manual/use-cases/nofx-strategy-config.png#bordered)

3. 要微调策略，请配置以下面板：
   - **指标**：选择策略使用的技术指标。
   - **风险控制**：设置风险管理规则。
   - **提示编辑器**：编辑 AI 用于决策的提示。
   - **额外提示**：输入附加到系统提示的额外提示，以实现个性化的交易风格。

4. 点击 **保存**，然后点击 **激活** 以启用策略。

## 配置你的交易机器人

Trader 代表你的自动交易机器人。创建 Trader 将你选择的 AI 模型、交易所连接和交易策略组合成一个可运行的交易机器人。

1. 在 **创建 Trader** 卡片上，点击 **立即创建**。

   ![创建 Trader](/images/manual/use-cases/nofx-create-trader.png#bordered)

2. 输入 Trader 名称，并选择你之前配置的 AI 模型和交易所。

   ![选择 Trader 组件](/images/manual/use-cases/nofx-config-trader.png#bordered){width=70%}

3. 选择你之前配置的交易策略。
4. 指定交易参数。
5. 点击 **创建 Trader**。

## 启动和监控你的交易机器人

1. 在 **当前 Trader** 部分，找到你刚刚创建的新 Trader。

   ![Trader 已创建](/images/manual/use-cases/nofx-trader-created.png#bordered)

2. 点击 **启动** 启动 Trader。
3. 前往 **仪表板** 页面实时观看策略运行。

   ![仪表板](/images/manual/use-cases/nofx-dashboard.png#bordered)

4. 要停止 Trader，返回 **配置** 页面，然后在同一部分点击 **停止**。

## 常见问题

### AI 钱包和 API 钱包有什么区别？

NOFX 需要两个不同的钱包用于不同的目的：

* **AI 钱包**：存在于 NOFX 内部，支付 AI 模型的数据分析和决策过程费用。你用 Base 链上的 USDC 为这个钱包充值。
* **API 钱包**：存在于你的交易所上，例如 Hyperliquid。它不持有资金。相反，它安全地授予 NOFX 使用主交易所账户中的资金执行交易的权限。

### 如何切换模型？

:::info
如果你配置自己的 API 密钥，NOFX 会直接通过你的提供商路由请求。这意味着 NOFX 不会从你的内部 AI 钱包中扣除模型调用费用。
:::

1. 在 **配置** 页面上，点击 **+ 模型配置**。

   ![切换模型](/images/manual/use-cases/nofx-models-config.png#bordered)

2. 点击 **其他 API 提供商**，然后选择另一个提供商。例如，**千问**。
3. 输入 API 密钥。基础地址和模型名称是可选的。

   ![配置模型](/images/manual/use-cases/nofx-models-config-api.png#bordered){width=70%}

4. 点击 **保存配置**。API 密钥在存储前会被加密。保存后，测试连接以验证其是否正常工作。

### 我可以使用本地模型吗？

是的，你可以使用本地模型，但请确保它满足以下要求：
- 它必须支持 OpenAI 兼容的 API 调用。
- 它需要强大的指令遵循能力、足够的上下文窗口和快速的推理速度。否则，模型可能无法输出有效的交易指令。

要配置本地模型：
1. 在 **配置** 页面上，点击 **+ 模型配置**。
2. 点击 **其他 API 提供商**，然后选择 **OpenAI**。
3. 在 **API 密钥** 字段中，输入任意文本字符串。
4. 在 **基础地址** 字段中，输入你的本地模型端点 URL。确保 URL 以 `/v1` 结尾。

   Olares 提供两种提供本地模型的方式。对于任一方式，获取共享入口 URL：

   <Tabs>
   <template #Ollama>

   一个应用托管多个模型，位于单个共享端点后面。

   a. 打开 **设置**，然后前往 **应用** > **Ollama**。

   b. 在 **共享入口** 中，选择 **Ollama API** 以查看共享端点 URL。

      ![设置中的 Ollama 共享入口](/images/manual/use-cases/ollama-shared.png#bordered){width=80%}

   c. 复制共享端点。例如，`http://d54536a50.shared.olares.com`。

   d. 在此端点 URL 后附加 `/v1`，即 `http://d54536a50.shared.olares.com/v1`。
   </template>
   <template #单模型应用>

   每个应用打包一个特定模型并暴露其自己的共享端点。以 **Qwen3.5 9B Q4_K_M (Ollama)** 为例。

   a. 打开 **设置**，然后前往 **应用** > **Qwen3.5 9B Q4_K_M (Ollama)**。

   b. 在 **共享入口** 中，选择 **Qwen3.5 9B Q4_K_M** 以查看端点 URL。

      ![Qwen3.5 9B 共享入口](/images/manual/use-cases/anythingllm-qwen359b-shared-entrance.png#bordered){width=80%}

   c. 复制共享端点 URL。例如，`http://bd5355000.shared.olares.com`。

   d. 在此端点 URL 后附加 `/v1`，即 `http://bd5355000.shared.olares.com/v1`。
   </template>
   </Tabs>

5. 点击 **保存配置**。

### 模型未输出结构化 JSON 决策

某些模型无法可靠地以 JSON 格式生成决策。要解决此问题，请尝试以下方法之一：
- 切换到不同的模型。
- 调整模型参数。
- 编辑模型设置下方的提示，指示模型输出有效的 JSON 格式。

### 上下文截止时间超出

以下错误表示 AI 模型响应时间过长：

`Failed to get AI decision: AI API call failed: failed to send request: Post "https://.../v1/chat/completions": context deadline exceeded`

NOFX 目前强制执行严格的 120 秒超时限制。

如果你持续遇到此问题，请尝试以下方法来减少模型的响应时间：
- 切换到更快的 AI 模型。
- 降低模型的推理或思考参数以生成更快的响应。
- 在设置中启用 `KEEP_ALIVE` 设置以保持模型加载。

### 开仓金额太小

以下错误发生是因为 NOFX 对合约强制执行严格的最低交易规模：

`Failed to get AI decision: failed to parse AI response: decision validation failed: decision #1 validation failed: ETHUSDT opening amount too small (12.00 USDT), must be ≥60.00 USDT`

要解决此问题，请调整策略的交易规模以满足以下最低要求：
- BTC 和 ETH 合约：每笔交易至少分配 60.00 USDT。
- 其他代币：每笔交易至少分配 12.00 USDT。
