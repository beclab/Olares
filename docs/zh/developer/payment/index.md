---
description: 用 Olares Payment 接受稳定币支付——用 LarePass 登录、获取 API 密钥、集成一个 API。资金在链上直接结算到你自己的钱包。
head:
  - - meta
    - name: keywords
      content: Olares Payment, 稳定币支付, USDC, USDT, 支付 API, 商户后台, LarePass, webhook
---

# Olares Payment

Olares Payment 让你把稳定币支付集成到自己的服务——接受来自任意 EVM 钱包的 USDC 和 USDT 付款,资金**在链上直接结算到你自己的钱包**。网关从不托管你的资金。

::: tip 想直接跑起来再看文档?
克隆 [demo 商店示例仓库](https://github.com/beclab/olares-payment-developer-example)——一个完整的最小集成,几分钟就能跑通;再回来对照本文档理解每一步。
:::

## 全流程一览

![从登录到收款全流程:LarePass 登录、收款地址就绪、创建 API 密钥、集成 API、买家任意钱包付款](/images/payment/payment-flow-zh.svg#bordered)

- **商户侧 = LarePass。** 只需扫一次二维码。商户账户自动创建,你的 LarePass 钱包 EVM 地址会在每条支持的链上预配置为收款地址。
- **买家侧 = 任意钱包。** 买家在托管收银台用 MetaMask、OKX 或任何其他 EIP-1193 浏览器钱包付款,无需 Olares 账户。

## 为什么商户选择它

| 特性 | 含义 |
|---|---|
| **自托管结算** | 买家直接向你在链上的收款地址转账。网关只核实转账,从不托管资金。 |
| **零配置入驻** | 首次用 LarePass 登录即创建账户并配好全部 7 条链——无表单、无需手动填写地址。 |
| **一个 API** | 创建支付单、拿到托管收银台、通过 webhook 或轮询确认。这就是全部集成工作。 |
| **全球化稳定币** | Optimism、Base、BNB Smart Chain、Ethereum、Arbitrum One、Polygon 和 Avalanche 上的 USDC 与 USDT。 |

## 核心概念

| 概念 | 说明 |
|---|---|
| **支付单(Payment)** | 一笔可支付的订单,以 `paymentId`(wire 层为 `intent_id`)标识。通过 API 创建,在托管收银台完成支付。 |
| **收银台** | 创建时返回的托管支付页面(代码中的 `checkoutUrl` 字段)。把链接发给买家——币种、网络和钱包连接都由它处理。 |
| **状态(Status)** | 支付单沿 `REQUIRES_PAYMENT_METHOD → PROCESSING → SUCCEEDED` 流转,过期(默认 TTL 30 分钟)则终止为 `CANCELED`。链上交易失败会发出 `payment.failed` 事件,支付单回落为等待下一次支付尝试。 |
| **退款(Refund)** | 沿付款来时那条路把钱退回去——从你的收款钱包出,回到当初付款的地址。可部分可全额;网关负责链上核实,资金与付款一样保持自托管。 |
| **退款执行链接** | `createRefund` 交还的不是 calldata,而是这条链接:短时效、只对应一笔退款,由操作员打开并用原收款钱包签名完成转账。 |
| **API 密钥** | 在商户后台创建的一对 `pk_live_…` / `sk_live_…`。私钥用于签名请求(HMAC),必须只留在你的服务器上。 |
| **Webhook** | 你在商户后台注册的 HTTPS 端点。网关向其 POST 签名事件,如 `payment.succeeded`。 |
| **Webhook 签名密钥** | 用于验证 webhook 签名的 `whsec_…`,独立于你的 API 密钥。 |

## 接下来

<div class="cta">
  <a href="./quickstart">
    <h3>快速开始 →</h3>
    <p>五步之内,从 LarePass 登录到收到第一笔付款。</p>
  </a>
</div>

- [API 参考](./api-reference/)——每个方法、参数和错误码
- [退款](./api-reference/refunds)——把一笔支付沿原路退回买家
- [Webhook](./webhooks)——签名事件、重试与验证
- [用 AI 智能体构建](./ai-agents)——一个 URL,让智能体生成可用的 demo
- [demo 商店示例仓库](https://github.com/beclab/olares-payment-developer-example)——可以直接克隆的最小集成示例
- [用户手册](/zh/manual/payment/)——产品角度的介绍,不需要技术背景
