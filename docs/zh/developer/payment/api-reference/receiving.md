---
description: Olares Payment API 收款配置——通道、收款钱包、链上流水、收款方式配置与支持链目录。
head:
  - - meta
    - name: keywords
      content: listChannels, listReceiveWalletTransactions, upsertOnchainPmc, listSupportedChains, 收款钱包
---

# 收款

收款地址在首次登录时已预配置——这些方法用于回读、列出链上进账,或变更配置。认证与错误码见 [API 总览](./index)。

## listChannels

`POST https://www.olares.com/payment/api/listChannels`

```ts
const { channels } = await client.listChannels();
```

你账户的收款通道。链上通道携带各自的收款钱包。

---

## listReceiveWalletTransactions

`POST https://www.olares.com/payment/api/listReceiveWalletTransactions`

```ts
const { items, hasMore, nextCursor } = await client.listReceiveWalletTransactions({
  address: '0x…', // optional; must belong to your account's receive wallets
  limit: 50,
});
```

收款钱包的链上进账,按区块号降序排列。支付单历史在 [listPayments](./payments#listpayments)——本方法是钱包级账本,包括未经收银台的直接转账。

---

## listPaymentMethodConfigs

`POST https://www.olares.com/payment/api/listPaymentMethodConfigs`

你账户当前的收款方式配置(`onchain` 通道及其各链钱包与代币白名单)。

---

## upsertOnchainPmc

`POST https://www.olares.com/payment/api/upsertOnchainPmc`

覆盖链上收款配置:逐链设置收款钱包与代币白名单。某链的 `tokens` 为空列表即关闭该链收款。

```ts
await client.upsertOnchainPmc({
  chains: [
    { chainId: '10', receiveWallet: '0x…', tokens: ['USDC', 'USDT'] },
  ],
});
```

收款地址在首次登录时已预配置——仅在需要变更时使用本方法。

---

## listSupportedChains

`POST https://www.olares.com/payment/api/listSupportedChains`

官方链与代币目录:Optimism、Base、BNB Smart Chain、Ethereum、Arbitrum One、Polygon、Avalanche,各链的 USDC/USDT。只读。
