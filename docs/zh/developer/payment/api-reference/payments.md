---
description: Olares Payment API 支付方法——createPayment、getPayment、listPayments、BuyerRef 三档、Payment 对象与幂等。
head:
  - - meta
    - name: keywords
      content: createPayment, getPayment, listPayments, BuyerRef, 幂等, Olares Payment
---

# 支付

创建、查询、列出支付单。认证与错误码见 [API 总览](./index)。

## createPayment

`POST https://www.olares.com/payment/api/createPayment`

```ts
createPayment(
  params: CreatePaymentRequest,
  opts?: { idempotencyKey?: string },
): Promise<CreatePaymentResult>
```

收款账户由密钥推断。一次调用即拿到托管收银台。

### CreatePaymentRequest

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `amountCents` | `number` | 是 | 价格,单位为分;`1000` = $10.00。 |
| `currency` | `'usd'` | 否 | 计价币种;默认 `usd`。 |
| `returnUrl` | `string` | 否 | 支付完成后跳转的绝对 http(s) URL。不传则停留在成功页。 |
| `metadata` | `Record<string, unknown>` | 否 | 透传对象,在查询、列表和 webhook 中原样回显——用于对账(例如你的 `order_id`)。 |
| `buyer` | [`BuyerRef`](#buyerref) | 否 | 买家信息披露。不传即为匿名支付。 |

### BuyerRef

三个互斥档位。混用字段或只披露一半会被 `1100` 拒绝。

| 档位 | 字段 | 网关行为 |
|---|---|---|
| 缺省 | — | 匿名。不创建身份,不建账户。 |
| `{ kind: 'external' }` | `ref`(1–128 字符);可选 `display.name` / `display.avatarUrl` | 你对买家的自有标识。不做验证;快照到支付单上,供商户后台对账。 |
| `{ kind: 'olares' }` | `olaresId` **和** `did`,均为必填 | 经 DID gate 解析;确定买家的 customer 账户。 |

`ref` 是**你**的用户标识,不是 Olares 账户。`display` 仅用于展示;只校验格式。

```ts
await client.createPayment({ amountCents: 500 });

await client.createPayment({
  amountCents: 500,
  buyer: { kind: 'external', ref: 'user_8817', display: { name: 'Ada' } },
});

await client.createPayment({
  amountCents: 500,
  buyer: { kind: 'olares', olaresId: 'alice.olares.com', did: 'did:olares:0x…' },
});
```

### CreatePaymentResult

| 字段 | 说明 |
|---|---|
| `paymentId` | 支付单 ID,用于后续查询与对账。 |
| `checkoutUrl` | 这笔支付的托管收银台(携带 `intent_id` 与 `client_secret` 的链接)。把它发给买家。 |

支付单会过期(默认 30 分钟,网关侧配置项 `INTENT_TTL_SECS`)。过期的收银台无法再支付。

### 幂等

```ts
await client.createPayment(params, { idempotencyKey: 'order:ord_123' });
```

网关按 `(scope, account, key)` 去重。同一 key 重放相同请求体会返回首次结果;同一 key 配不同请求体则报 `400`。幂等键应从你的业务身份派生,并在重试之间复用——SDK 从不自动生成。

---

## getPayment

`POST https://www.olares.com/payment/api/getPayment`

```ts
getPayment(paymentId: string): Promise<PaymentResult>
```

未支付的支付单**不是**错误。当且仅当状态为 `PAYMENT_STATUS_SUCCEEDED` **且**最近一次尝试带有 `txHash` 时,`paid` 才为 `true`。成功状态但没有 `txHash` 时返回 `paid: false`——不要履约。

```ts
const result = await client.getPayment(paymentId);
if (result.paid) {
  const { txHash, payAmount, payCurrency, chain } = result.credential;
} else {
  // result.payment.status
}
```

该端点为双认证:商户用 HMAC,买家收银台页面用 `client_secret`。无密钥构造的 SDK 客户端可用 `getPayment(paymentId, { clientSecret })` 查询。

支付成功的单子还会带上 `refundSummary`——这笔订单的退款账本(已收、已退、仍可退)。由于同一个对象也会下发给买家的收银台会话,其中的 `refunds` 列表**只含已成功的退款**。见[退款](./refunds#在支付单上读退款)。

<TryIt endpoint="getPayment" />

---

## listPayments

`POST https://www.olares.com/payment/api/listPayments`

```ts
listPayments(params?: ListPaymentsRequest): Promise<ListPaymentsResponse>
```

| 参数 | 说明 |
|---|---|
| `status` | 精确状态过滤,例如 `PAYMENT_STATUS_SUCCEEDED`。 |
| `metadata` | JSONB 包含匹配(`metadata @> filter`)。 |
| `cursor` | 上一页返回的 `nextCursor`。 |
| `limit` | 默认 20,最大 200。 |

列表项从不包含 `clientSecret`。

```ts
const { items, hasMore, nextCursor } = await client.listPayments({
  status: 'PAYMENT_STATUS_SUCCEEDED',
  metadata: { order_id: 'ord_123' },
  limit: 20,
});
```

---

## Payment 对象

由 `getPayment` / `listPayments` 返回(TS 命名;wire 层使用 `snake_case`):

| 字段 | 说明 |
|---|---|
| `paymentId` | 支付单 ID |
| `merchantAccountId` | 收款账户 |
| `buyer` | 创建时的买家快照;匿名时为 `null` |
| `amountCents` / `currency` | 计价 |
| `status` | `PAYMENT_STATUS_*` |
| `settlementCurrency` / `settlementAmount` | 链上结算币种与金额 |
| `metadata` | 创建时传入的透传对象 |
| `clientSecret` | 仅当对象随创建响应返回时出现;查询与列表结果永不包含 |
| `latestAttempt` | 最近一次尝试,含 `txHash` |
| `refundSummary` | [退款账本](./refunds#在支付单上读退款)。仅 `getPayment` 且订单已成功时非空;`listPayments` 条目恒为 `null` |
| `expiresAt` / `canceledAt` / `paidAt` / `createdAt` / `updatedAt` | RFC 3339 时间戳 |

`PaymentResult` 在 `paid: true` 时额外携带 `credential`——`txHash`、`payAmount`、`payCurrency`、`chain`、`chainType`、`networkId`——与 `payment.succeeded` webhook 的 credential 形状相同。
