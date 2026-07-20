# 官方 Tagger

## RootTagger

目前我们定义了以下标签，并使用 `RootTagger` 作为它们的 Tagger：

-   **rsaPubKey**: TName 的 RSA 公钥
    -   类型: `bytes`
    -   访问权限: operator、TName 的 owner (所有者)，以及父 TName 的 owner。
-   **dnsARecord**: TName 的 IP 地址
    -   类型: `bytes4`
    -   访问权限: operator、TName 的 owner，以及父 TName 的 owner。
-   **latestDID**: TName 最新的 DID (添加此标签是因为元数据是不可变的)
    -   类型: `string`
    -   访问权限: operator、TName 的 owner，以及父 TName 的 owner。
-   **authAddresses**: 由 TName 的 owner 控制的地址
    -   类型: `tuple(uint8,address)[]`
    -   访问权限: 任何持有 TName 的 owner 和被添加地址的 EIP-712 签名的人。

## AppStoreReputation

此外还有一个特殊的 Tagger `AppStoreReputation`，用于在 TName `app.myterminus.com` 中定义的 `ratings` 标签，其类型为 `tuple(string,uint8)[]`。它为 Terminus OS 中应用的评分提供链上存储。每个应用都有一个对应的子 TName `<appVersion>.<appId>.app.myterminus.com`，任何拥有 TName 的人都可以为这些应用提交评分。
```