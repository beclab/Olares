---
description: Olares Payment webhook——注册端点、验证签名、处理 payment.succeeded,并用免费公网 URL 测试投递。
head:
  - - meta
    - name: keywords
      content: Olares Payment webhook, payment.succeeded, webhook 签名, constructEvent, 重试, whsec
---

# Webhook

Webhook 把支付生命周期事件推送到你的服务器,让你无需轮询即可履约。每次投递都有签名;信任之前先验证。

## 工作原理

1. 你在商户后台注册一个 HTTPS 端点(**Checkouts → Advanced settings → Webhooks**)。
2. 支付单状态变化时,网关向你的端点 POST 一个已签名的 JSON 事件。
3. 你的服务器验证签名并返回 `2xx`。其他任何结果都会触发重试。

在商户后台新建的端点默认订阅 `payment.succeeded`、`refund.succeeded` 与 `refund.failed`。

::: warning 退款上线之前创建的端点
投递按端点订阅的事件过滤,而既有端点不会被自动补订阅——老端点会**悄无声息地收不到任何退款回调**。请打开 **Checkouts → Advanced settings → Webhooks**,给它补上这两个退款事件。
:::

![注册端点](/images/payment/dashboard-advanced-settings.png#bordered)

## 事件

| 事件 | 触发时机 | 额外字段 |
|---|---|---|
| `payment.succeeded` | 支付在链上确认 | `credential`(交易哈希、金额、币种、链)、`paid_at` |
| `payment.failed` | 买家提交的交易在链上失败;支付单保持打开,等待下一次尝试 | `tx_hash`、`fail_reason` |
| `payment.canceled` | 支付单过期未支付 | `cancellation_reason`(`expired`) |
| `refund.succeeded` | 退款转账通过链上核实 | `refund_id`、`amount`、`token_symbol`、`chain`、`network_id`、`tx_hash`、`succeeded_at` |
| `refund.failed` | 退款交易 revert 或与冻结路由不符——可重试,不是终局 | 同上,另加 `fail_reason` 与 `failed_at` |
| `endpoint.test` | 由商户后台 **Test** 按钮发送,用于验证你的接收端——返回 200 并忽略即可 | 仅示例载荷 |

支付类事件携带 `event_type`、`intent_id`、`merchant_account_id`、`buyer`,以及你的 `metadata`。其中 `buyer` 是创建支付时提供的买家身份快照——没提供(匿名单)就是 `null`。退款事件有[自己的载荷形状](#退款事件)。

```json
{
  "event_type": "payment.succeeded",
  "intent_id": "pi_1a07b0148515b9a95bc",
  "merchant_account_id": "acct_x8y9…",
  "buyer": {
    "kind": "olares",
    "olares_id": "buyer.olares.cn",
    "did": "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH"
  },
  "metadata": { "order_id": "ord_001" },
  "credential": {
    "tx_hash": "0x9f3c…",
    "pay_amount": "4990000",
    "pay_currency": "USDC",
    "chain": "optimism",
    "chain_type": "CHAIN_TYPE_EVM",
    "network_id": "10"
  },
  "paid_at": "2026-09-08T15:20:00Z"
}
```

凭 `credential.tx_hash` 存在与否履约——与 `getPayment` 的 `paid` 判定是同一规则。

## 退款事件

`refund.succeeded` 与 `refund.failed` **不用**上面那套载荷,它们携带的是退款形状的报文。因此请先读 `event_type`,再决定按什么解析:

```json
{
  "event_type": "refund.succeeded",
  "refund_id": "re_7f2c9a1b34",
  "intent_id": "pi_1a07b0148515b9a95bc",
  "merchant_account_id": "acct_x8y9…",
  "amount": "600000",
  "token_symbol": "USDC",
  "chain": "optimism",
  "network_id": 10,
  "tx_hash": "0x4d1e…",
  "reason": "customer downgraded the plan",
  "buyer": {
    "kind": "olares",
    "olares_id": "buyer.olares.cn",
    "did": "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH"
  },
  "succeeded_at": "2026-09-08T15:20:00Z"
}
```

三件必须做对的事:

| 规则 | 原因 |
|---|---|
| **解析前先按 `event_type` 分支** | 把退款事件当成支付回调去解,会静默丢掉 `refund_id`、`amount`、`token_symbol`——正是你入账所需的字段。 |
| **退款事件没有 `metadata`** | 退款回调协议里没有这个字段。请用 `intent_id`(TS 里是 `paymentId`)回查你自己的订单。 |
| **`refund.failed` 不是终态** | 金额仍被占住,操作员可以从同一条执行链接重签一笔修正后的交易。不要因此释放占额或把订单改回未退款。 |

`amount` 是原 token 的最小单位,不是分——与 `createRefund` 入参同一单位。`reason` 是你自己填的内部备注原样回传,买家从未见过它。

完整生命周期见[退款](./api-reference/refunds)。

## 验证签名

每次投递带两个请求头:

- `x-olares-payment-webhook-timestamp` —— Unix 时间,毫秒
- `x-olares-payment-webhook-signature` —— `hex(HMAC-SHA256(whsec_…, "{timestamp}\n{rawBody}"))`

签名密钥(`whsec_…`)在注册时签发,可随时在商户后台查看。它独立于你的 API 密钥。

验证规则:

- 被签名的内容是 `{timestamp}\n{rawBody}`——**精确的原始请求字节**。验证通过后再解析 JSON。
- 拒绝超过 5 分钟的时间戳(防重放)。
- 在 Express 中,webhook 路由使用 `express.raw`——先挂 `express.json()` 会破坏原始请求体。

<Tabs>
<template #TypeScript>

```ts
import { webhooks } from '@olares/payment-sdk';

app.post('/webhook', express.raw({ type: 'application/json' }), (req, res) => {
  try {
    const event = webhooks.constructEvent(req, process.env.PAYMENT_WEBHOOK_SECRET!);
    switch (event.type) {
      case 'payment.succeeded':
        // event.paymentId, event.credential.txHash, event.metadata
        break;
      case 'payment.failed':
        break;
      case 'payment.canceled':
        break;
      case 'refund.succeeded':
        // event.refundId、event.paymentId、event.amount、event.txHash——这里没有 metadata
        break;
      case 'refund.failed':
        // event.failReason;退款仍可重试,此刻先别回滚任何东西
        break;
    }
    res.status(200).send('ok');
  } catch {
    res.status(400).send('bad signature');
  }
});
```

如果你的框架已经缓冲了请求体,也可以传 `{ body, headers }`。验证失败会抛出 `PaymentError`——绝不要将本次投递视为有效。

</template>
<template #Go>

```go
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    e, err := paymentsdk.ConstructEvent(r, webhookSecret) // SDK reads the raw body itself
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    switch e.Type {
    case "payment.succeeded":
        // e.PaymentID / e.Buyer / e.Metadata / e.PaidAt
    case "refund.succeeded", "refund.failed":
        // e.Refund 里是退款事实:RefundId、Amount、TokenSymbol、Chain。
        // e.PaymentID / e.TxHash / e.FailReason 也会回填到顶层。
    }
    w.WriteHeader(http.StatusOK)
}
```

如果请求体已被消费(队列重放、急切中间件),使用 `ConstructEventFromParts(rawBody, r.Header, secret)`。

</template>
</Tabs>

## 重试

你的端点必须返回 `2xx`。其他任何情况——非 2xx、超时(60 秒)、不可达——都计为一次失败尝试,并按以下计划重试:

| 第几次尝试 | 1 | +1 | +2 | +3 | +4 | +5 | +6 | +7 |
|---|---|---|---|---|---|---|---|---|
| 失败后的延迟 | — | 1 分钟 | 5 分钟 | 30 分钟 | 2 小时 | 12 小时 | 24 小时 | 24 小时 |

8 次尝试后投递进入死信(`failed`)。事件处理必须**幂等**——重试意味着同一事件可能多次到达。

## 在商户后台测试与调试

- **Test**——向你的端点发送一笔真实签名的 `endpoint.test` 投递。这是验证接收端的最快方式。你的处理器应对它返回 200(未知事件类型安全忽略即可)。
- **Delivery log**——每次投递的记录,含状态、尝试次数与最后一次错误。
- **Replay**——端点修复后,重发失败的投递。

## 用免费公网 URL 做本地开发

网关必须够得到你的服务器,所以本地开发需要一个公网 HTTPS 地址。按场景选——全部免费、无需注册:

**场景 A:我只想看看 webhook 投递的内容长什么样。** 打开 [webhook.site](https://webhook.site)——它直接给你一个现成的公网 URL。把这个 URL 填进商户后台当 webhook 地址,之后每笔投递的内容都会实时显示在那个网页上。全程不需要你自己的服务器。

**场景 B:我要本地运行的 demo 程序真正收到并处理 webhook。** 用隧道把本地服务映射成公网地址,再把这个地址登记到商户后台:

| 工具 | 一条命令 | 说明 |
|---|---|---|
| **localtunnel**(推荐) | `npx localtunnel --port 3000` | 有 npm 就零安装、免注册。输出就一行:`your url is: https://….loca.lt` |
| **Cloudflare Quick Tunnel** | `cloudflared tunnel --url http://localhost:3000` | 免费、免账号。URL 在输出的 "Visit it at" 框里;网络不稳老掉线就加 `--protocol http2`。也可走 npm:`npx cloudflared tunnel --url …` |

免费隧道的域名是随机的,每次重启都会变——变了就到商户后台改 webhook 地址。

::: warning 切勿用于生产
这些免费 URL 仅供开发使用。生产环境的 webhook 请指向你自己的稳定 HTTPS 端点。
:::
