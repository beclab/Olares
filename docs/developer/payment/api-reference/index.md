---
description: Olares Payment API overview — official clients, HMAC authentication, response envelope, error codes, and the method map by domain.
head:
  - - meta
    - name: keywords
      content: Olares Payment API, HMAC authentication, response envelope, error codes, MerchantClient
---

# API reference

The Olares Payment API is a set of `POST` endpoints on the gateway. **Base URL:**

```
https://www.olares.com/payment/api
```

Every method is `POST {base}/<method>`. The SDKs below are wrappers around exactly this — you can equally integrate over plain HTTP from any language by signing requests yourself (see [Authentication](#authentication)).

Methods are organized by domain:

| Domain | What it covers |
|---|---|
| [Payments](./payments) | `createPayment`, `getPayment`, `listPayments`, the `BuyerRef` tiers, the Payment object, idempotency |
| [Refunds](./refunds) | `createRefund`, `getRefund`, `cancelRefund`, `reissueRefundLink`, the execution link, the refund lifecycle |
| [Receiving](./receiving) | Channels, receive wallets and their on-chain transactions, payment method configuration, supported chains |
| [Account & connectivity](./account) | `getAccount`, `ping`, `verifyTransaction` |

Webhooks are a local concern — signature verification with `constructEvent` lives in [Webhooks](../webhooks).

## Official clients

| Language | Package | Install |
|---|---|---|
| TypeScript (Node.js 18+) | `@olares/payment-sdk` | `npm install @olares/payment-sdk` |
| Go | `github.com/Above-Os/olares-payment/packages/payment-sdk-go` | `go get github.com/Above-Os/olares-payment/packages/payment-sdk-go` |
| Any (raw HTTP) | — | No SDK needed — sign per [Authentication](#authentication) |

The SDK points at the production gateway `https://www.olares.com/payment` out of the box — nothing to configure.

## Authentication

Authenticated calls carry four HMAC headers:

| Header | Value |
|---|---|
| `x-olares-payment-key` | Your API key, `pk_live_…` |
| `x-olares-payment-timestamp` | Unix time in **milliseconds** |
| `x-olares-payment-nonce` | Random nonce per request |
| `x-olares-payment-signature` | `hex(HMAC-SHA256(apiSecret, canonical))` |

The canonical string, with literal newlines:

```
canonical = "{method}\n{body}\n{timestamp}\n{nonce}"
   method  = createPayment
   body    = the exact raw request body string
```

`method` is the short name (`createPayment`, `getPayment`, …). The request URL is `POST /api/{method}`. Do not put `/api/` or the host in the signature. `timestamp` and `nonce` must match the headers byte-for-byte. `apiSecret` is the full `sk_live_…` string; nothing is stripped.

A request more than 5 minutes old is rejected (`1502 TIMESTAMP_EXPIRED`). The SDKs compute all of this for you; implement it yourself only when writing a new client.

### Public endpoints

`POST /api/ping` needs no authentication. `POST /api/getPayment` accepts **either** HMAC **or** a `client_secret` in the body — the pairing a buyer's checkout uses (see [getPayment](./payments#getpayment)).

## Response envelope

Every response is a single envelope:

```json
// success
{ "code": 0, "payload": { "…": "…" } }
// failure
{ "code": 1203, "message": "payment not found" }
```

Wire keys are `snake_case`. The TS SDK translates them to `camelCase` at the boundary (e.g. `intent_id` → `paymentId`); the Go SDK uses generated protobuf types.

## Errors

SDK calls throw `PaymentError` (TS) / return `*paymentsdk.Error` (Go) with `code`, `httpStatus`, and `message`.

| Range | Meaning |
|---|---|
| `1000–1899` | Gateway business errors — the request reached the gateway. Handle by `code`. |
| `1900–1949` | Gateway **refund** errors, raised by the four refund methods. See [Refunds → Errors](./refunds#errors). |
| `1950–1999` | SDK-local transport errors, raised by the SDK and never returned by the gateway. Reads can be retried directly; for writes such as creating a payment or a refund, retry with an `Idempotency-Key` — a timeout can happen after the gateway has already processed the request, and the key prevents duplicates. |

The dividing line is worth coding against: **`code < 1950` means the request reached the gateway**, while `code >= 1950` means it never got there.

| Code | Constant | Meaning |
|---|---|---|
| 1000 | `INTERNAL_ERROR` | Server-side internal error |
| 1100 | `INVALID_ARGUMENT` | Bad parameters |
| 1104 | `INVALID_RETURN_URL` | `returnUrl` is not an absolute http(s) URL |
| 1203 | `PAYMENT_NOT_FOUND` | Payment does not exist |
| 1300 | `PERMISSION_DENIED` | Capability denied for this key |
| 1301 | `UNAUTHENTICATED` | Authentication required |
| 1500 | `SIGNATURE_MISMATCH` | HMAC or webhook signature mismatch |
| 1501 | `INVALID_API_KEY` | Unknown or revoked key |
| 1502 | `TIMESTAMP_EXPIRED` | Timestamp outside the 5-minute window |
| 1901 | `REFUND_PRECONDITION_FAILED` | Payment is not refundable yet |
| 1902 | `REFUND_AMOUNT_EXCEEDED` | Refund amount exceeds the refundable balance |
| 1903 | `REFUND_STATE_CONFLICT` | The refund is not in a state that allows this action |
| 1904 | `REFUND_EXECUTION_INVALID` | Execution credential invalid or expired |
| 1905 | `REFUND_ROUTE_UNSUPPORTED` | This payment cannot be refunded |
| 1906 | `REFUND_ALREADY_OPEN` | Another refund is already open on this payment |
| 1907 | `REFUND_CANCEL_UNAVAILABLE` | Cancellation is unavailable right now (quiet period) |
| 1951 | `SDK_TIMEOUT` | Client-side timeout |
| 1952 | `SDK_NETWORK_ERROR` | Network failure or non-JSON response |
| 1953 | `SDK_RPC_ERROR` | Direct-RPC verification failed |

::: warning Older integrations
Before refunds shipped, the SDK-local codes were `1901` / `1902` / `1903` — numbers that now belong to refunds. Code still branching on them reads `1902` ambiguously: "refund exceeds the balance" (never retry) or "network failure" (safe to retry). Move that branching to `1951` / `1952` / `1953`.
:::

## Client configuration

```ts
const client = new MerchantClient({
  apiKey: 'pk_live_…',
  apiSecret: 'sk_live_…',
});
```

| Option | Type | Default | Description |
|---|---|---|---|
| `apiKey` | `string` | — | Merchant API key (`pk_live_…`). Required for HMAC methods. |
| `apiSecret` | `string` | — | HMAC secret (`sk_live_…`). Server-side only. |
| `baseUrl` | `string` | `https://www.olares.com/payment` | Gateway root, without `/api`. Already the production default — leave it alone. |
| `timeoutMs` | `number` | `30000` | Per-request timeout, also used for direct-RPC verification. |
| `logger` | `SdkLogger` | `console` | `(level, msg, ctx?) => void`; pass `() => {}` to silence. |
| `rpcUrl` | `string` | — | When set, `verifyTransaction` queries this EVM RPC directly instead of the gateway. |
