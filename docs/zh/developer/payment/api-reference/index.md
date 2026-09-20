---
description: Olares Payment API 总览——官方客户端、HMAC 认证、响应包络、错误码,以及按领域分组的方法地图。
head:
  - - meta
    - name: keywords
      content: Olares Payment API, HMAC 认证, 响应包络, 错误码, MerchantClient
---

# API 参考

Olares Payment API 是网关上的一组 `POST` 端点。**根地址:**

```
https://www.olares.com/payment/api
```

每个方法都是 `POST {根地址}/<方法名>`。下面的 SDK 只是对此的封装——任何语言都可以不走 SDK,按[认证](#认证)的规范自行签名后直接 HTTP 接入。

方法按领域分组:

| 领域 | 内容 |
|---|---|
| [支付](./payments) | `createPayment`、`getPayment`、`listPayments`、BuyerRef 三档、Payment 对象、幂等 |
| [退款](./refunds) | `createRefund`、`getRefund`、`cancelRefund`、`reissueRefundLink`、执行链接、退款生命周期 |
| [收款](./receiving) | 通道、收款钱包及其链上流水、收款方式配置、支持的链 |
| [账户与连通](./account) | `getAccount`、`ping`、`verifyTransaction` |

Webhook 是本地动作——用 `constructEvent` 验签,见 [Webhook](../webhooks)。

## 官方客户端

| 语言 | 包 | 安装 |
|---|---|---|
| TypeScript (Node.js 18+) | `@olares/payment-sdk` | `npm install @olares/payment-sdk` |
| Go | `github.com/Above-Os/olares-payment/packages/payment-sdk-go` | `go get github.com/Above-Os/olares-payment/packages/payment-sdk-go` |
| 任意语言(裸 HTTP) | — | 不需要 SDK——按[认证](#认证)签名即可 |

SDK 开箱即指向生产网关 `https://www.olares.com/payment`——无需任何配置。

## 认证

认证调用携带四个 HMAC 请求头:

| 请求头 | 取值 |
|---|---|
| `x-olares-payment-key` | 你的 API key,`pk_live_…` |
| `x-olares-payment-timestamp` | Unix 时间,**毫秒** |
| `x-olares-payment-nonce` | 每次请求的随机 nonce |
| `x-olares-payment-signature` | `hex(HMAC-SHA256(apiSecret, canonical))` |

规范化字符串(canonical string),含字面换行:

```
canonical = "{method}\n{body}\n{timestamp}\n{nonce}"
   method  = createPayment
   body    = the exact raw request body string
```

`method` 是短方法名(`createPayment`、`getPayment` …)。请求 URL 是 `POST /api/{method}`。不要把 `/api/` 或主机名写进签名。`timestamp` 与 `nonce` 必须与请求头逐字节一致。`apiSecret` 是完整的 `sk_live_…` 字符串,不去前缀。

超过 5 分钟的请求会被拒绝(`1502 TIMESTAMP_EXPIRED`)。SDK 会为你完成所有这些计算;只有在自己实现新客户端时才需要手写。

### 公开端点

`POST /api/ping` 无需认证。`POST /api/getPayment` 接受 **HMAC** 或请求体中的 `client_secret` **之一**(即买家收银台使用的配对方式——见 [getPayment](./payments#getpayment))。

## 响应包络

每个响应都是统一的包络:

```json
// success
{ "code": 0, "payload": { "…": "…" } }
// failure
{ "code": 1203, "message": "payment not found" }
```

Wire 层字段为 `snake_case`。TS SDK 在边界处将其转换为 `camelCase`(例如 `intent_id` → `paymentId`);Go SDK 使用生成的 protobuf 类型。

## 错误码

SDK 调用抛出 `PaymentError`(TS)/ 返回 `*paymentsdk.Error`(Go),含 `code`、`httpStatus` 与 `message`。

| 区间 | 含义 |
|---|---|
| `1000–1899` | 网关业务错误:请求已到达网关。按 `code` 处理。 |
| `1900–1949` | 网关**退款**错误,由四个退款方法抛出。见[退款 → 错误码](./refunds#错误码)。 |
| `1950–1999` | SDK 本地传输错误,由 SDK 抛出,网关不会返回。查询类可直接重试;创建支付、创建退款等写操作请带 `Idempotency-Key` 重试——超时可能发生在网关已处理之后,幂等键可避免重复建单。 |

这条分界线值得直接写进代码:**`code < 1950` 表示请求到过网关**,`code >= 1950` 表示请求根本没发出去。

| 错误码 | 常量 | 含义 |
|---|---|---|
| 1000 | `INTERNAL_ERROR` | 服务端内部错误 |
| 1100 | `INVALID_ARGUMENT` | 参数错误 |
| 1104 | `INVALID_RETURN_URL` | `returnUrl` 不是绝对 http(s) URL |
| 1203 | `PAYMENT_NOT_FOUND` | 支付单不存在 |
| 1300 | `PERMISSION_DENIED` | 该 API key 的权限被拒绝 |
| 1301 | `UNAUTHENTICATED` | 需要认证 |
| 1500 | `SIGNATURE_MISMATCH` | HMAC 或 webhook 签名不匹配 |
| 1501 | `INVALID_API_KEY` | 未知或已吊销的 API key |
| 1502 | `TIMESTAMP_EXPIRED` | 时间戳超出 5 分钟窗口 |
| 1901 | `REFUND_PRECONDITION_FAILED` | 支付单尚不可退 |
| 1902 | `REFUND_AMOUNT_EXCEEDED` | 退款额超出可退余额 |
| 1903 | `REFUND_STATE_CONFLICT` | 当前退款状态不允许该操作 |
| 1904 | `REFUND_EXECUTION_INVALID` | 执行凭据失效或已过期 |
| 1905 | `REFUND_ROUTE_UNSUPPORTED` | 这笔支付无法退款 |
| 1906 | `REFUND_ALREADY_OPEN` | 该支付单已有未结清退款 |
| 1907 | `REFUND_CANCEL_UNAVAILABLE` | 暂时不能取消(静默期未满) |
| 1951 | `SDK_TIMEOUT` | 客户端超时 |
| 1952 | `SDK_NETWORK_ERROR` | 网络故障或非 JSON 响应 |
| 1953 | `SDK_RPC_ERROR` | 直连 RPC 核实失败 |

::: warning 老集成注意
退款上线之前,SDK 本地码曾是 `1901` / `1902` / `1903`——这三个号如今属于退款段。仍按老号分支的代码会把 `1902` 读成二义:既可能是「退款超出可退余额」(绝不该重试),也可能是「网络失败」(可以重试)。请把分支迁到 `1951` / `1952` / `1953`。
:::

## 客户端配置

```ts
const client = new MerchantClient({
  apiKey: 'pk_live_…',
  apiSecret: 'sk_live_…',
});
```

| 选项 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `apiKey` | `string` | — | 商户 API key(`pk_live_…`)。HMAC 方法必填。 |
| `apiSecret` | `string` | — | HMAC 私钥(`sk_live_…`)。仅限服务端使用。 |
| `baseUrl` | `string` | `https://www.olares.com/payment` | 网关根地址,不含 `/api`。默认即生产环境,无需修改。 |
| `timeoutMs` | `number` | `30000` | 单次请求超时,也用于直连 RPC 核实。 |
| `logger` | `SdkLogger` | `console` | `(level, msg, ctx?) => void`;传 `() => {}` 可静音。 |
| `rpcUrl` | `string` | — | 设置后,`verifyTransaction` 直接查询该 EVM RPC,而不经过网关。 |
