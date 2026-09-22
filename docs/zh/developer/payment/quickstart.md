---
description: 五步完成第一笔稳定币收款——从 LarePass 登录开始:创建 API 密钥、创建支付单、发送收银台链接、确认履约。
head:
  - - meta
    - name: keywords
      content: Olares Payment 快速开始, 创建支付, 收银台 URL, API 密钥, HMAC, LarePass 登录
---

# 快速开始

五步把稳定币支付集成到你自己的服务。你需要一个 [Olares ID](/zh/manual/get-started/create-olares-id),并在手机上装好 LarePass 应用;还需要一个服务端运行时(Node.js 18+)。

::: tip 想先看能跑的代码?
[demo 商店示例仓库](https://github.com/beclab/olares-payment-developer-example)就是本篇的可运行版本——克隆、填入你的密钥,就能看到一笔支付到账。
:::

## 第 1 步:用 LarePass 登录

在浏览器打开 [https://www.olares.com/payment/dashboard](https://www.olares.com/payment/dashboard/)(商户后台),用 LarePass 扫描二维码。

![登录 Olares Payment](/images/payment/dashboard-login.png#bordered)

首次登录时,Olares Payment 会创建你的商户账户并**预配置收款地址**:你的 LarePass 钱包 EVM 地址会被设置到每一条支持的链上。你可以立即开始收款——无表单、无需手动填写地址。

![收款就绪](/images/payment/dashboard-home.png#bordered)

## 第 2 步:创建 API 密钥

在商户后台进入 **Checkouts → Advanced settings → API keys**,为密钥命名(例如 `Store backend`),然后点击 **New key**。

![API 密钥与 webhook](/images/payment/dashboard-advanced-settings.png#bordered)

你会得到一对 `pk_live_…` / `sk_live_…`:

- `pk_live_…` 标识你的集成。
- `sk_live_…` 用于签名请求(HMAC)。**只放在你的服务器上——绝不要带进浏览器、移动应用或公开仓库。**

## 第 3 步:安装 SDK

<Tabs>
<template #TypeScript>

```bash
npm install @olares/payment-sdk
```

```ts
import { MerchantClient } from '@olares/payment-sdk';

const client = new MerchantClient({
  apiKey: process.env.PAYMENT_API_KEY!,     // pk_live_…
  apiSecret: process.env.PAYMENT_API_SECRET!, // sk_live_…
});
// Defaults to the production gateway https://www.olares.com/payment — no baseUrl needed.
```

</template>
<template #cURL>

每个端点都是普通的 `POST /api/<method>`,带 HMAC 请求头——任何语言都可以直接调用:

```bash
# See "API reference → Authentication" for how to compute the signature.
curl -X POST https://www.olares.com/payment/api/ping \
  -H 'content-type: application/json' \
  -d '{}'
# {"code":0,"payload":{"server_time":"2026-09-07T08:30:00Z"}}
```

</template>
</Tabs>

## 第 4 步:创建支付单

`createPayment` 一步拿到托管收银台。把链接发给买家——币种、网络和钱包连接都由收银台处理。

<Tabs>
<template #TypeScript>

```ts
const { paymentId, checkoutUrl } = await client.createPayment({
  amountCents: 499,              // $4.99, priced in USD
  metadata: { order_id: 'ord_001' },
  returnUrl: 'https://your-shop.example/orders/ord_001', // optional redirect after payment
});
// → checkoutUrl: https://www.olares.com/payment/?intent_id=pi_…&client_secret=…
```

</template>
<template #cURL>

```bash
curl -X POST https://www.olares.com/payment/api/createPayment \
  -H 'content-type: application/json' \
  -H "x-olares-payment-key: $PAYMENT_API_KEY" \
  -H "x-olares-payment-timestamp: $(date +%s%3N)" \
  -H "x-olares-payment-nonce: $(uuidgen)" \
  -H "x-olares-payment-signature: $SIGNATURE" \
  -d '{"amount_cents":499,"metadata":{"order_id":"ord_001"}}'
```

</template>
</Tabs>

::: details 安全重试:幂等键
传入稳定的幂等键(`{ idempotencyKey: 'order:ord_001' }` 或 `Idempotency-Key` 请求头)。重放的请求会返回首次调用的结果,而不会重复创建支付单。
:::

你的买家会看到托管收银台,并用任意 EVM 钱包付款——无需 Olares 账户:

![托管收银台](/images/payment/checkout-page.png#bordered)

## 第 5 步:确认并履约

当 `getPayment` 返回 `paid: true` 时,支付单即可履约——状态为 `SUCCEEDED` **且**带有链上交易哈希。只依据这个信号履约。

<Tabs>
<template #TypeScript>

```ts
const result = await client.getPayment(paymentId);
if (result.paid) {
  const { txHash, payAmount, payCurrency, chain } = result.credential;
  // fulfill the order
}
```

</template>
<template #cURL>

```bash
curl -X POST https://www.olares.com/payment/api/getPayment \
  -H 'content-type: application/json' \
  -H "x-olares-payment-key: $PAYMENT_API_KEY" \
  -H "x-olares-payment-timestamp: $(date +%s%3N)" \
  -H "x-olares-payment-nonce: $(uuidgen)" \
  -H "x-olares-payment-signature: $SIGNATURE" \
  -d '{"intent_id":"pi_…"}'
```

</template>
</Tabs>

除了轮询,你也可以注册 [webhook](./webhooks),在收到 `payment.succeeded` 时履约。未支付的支付单会过期(默认 30 分钟)并收敛为 `CANCELED`——没有取消 API。

## 接下来

- [API 参考](./api-reference/)——全部方法、参数和错误码
- [退款](./api-reference/refunds)——把一笔支付沿原路退回买家
- [Webhook](./webhooks)——签名、重试,以及在商户后台测试与调试
- [用 AI 智能体构建](./ai-agents)——把本文档交给智能体,让它为你生成一个 demo 商店
- [demo 商店示例仓库](https://github.com/beclab/olares-payment-developer-example)——一个可以直接克隆的最小集成示例
