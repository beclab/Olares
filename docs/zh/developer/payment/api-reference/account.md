---
description: Olares Payment API 账户与连通方法——getAccount、ping 与 verifyTransaction。
head:
  - - meta
    - name: keywords
      content: getAccount, ping, verifyTransaction, Olares Payment API
---

# 账户与连通

密钥身份、账户查询、健康检查与独立的链上核实。认证与错误码见 [API 总览](./index)。

## getAccount

`POST https://www.olares.com/payment/api/getAccount`

```ts
const account = await client.getAccount();
// { accountId, did, olaresId, status }
```

你的 API 密钥绑定的商户账户。

---

## ping

`POST https://www.olares.com/payment/api/ping`

```ts
const { serverTime } = await client.ping();
```

公开健康检查,返回网关时间。注意 SDK 仍会发送签名请求头,因此构造客户端时需要密钥。

<TryIt endpoint="ping" />

---

## verifyTransaction

`POST https://www.olares.com/payment/api/verifyTx`

日常履约使用 `getPayment` 与 webhook。`verifyTransaction` 按哈希独立读取 EVM 回执:

```ts
const v = await client.verifyTransaction(txHash, 'optimism', 'mainnet');
if (v.confirmed) {
  // v.blockNumber, v.status === 'success'
}
```

未配置 `rpcUrl` 时,调用经网关代理(`POST /api/verifyTx`);配置 `rpcUrl` 后,SDK 直接查询该 EVM RPC——适合不想采信网关结论的场景。仅限 EVM 链;`chain` / `network` 仅在走网关路径时需要。
