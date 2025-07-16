---
description: 介绍在区块链上注册 DID 的机制，涵盖智能合约实现、去中心化优势与当前部署方案。
---

# 区块链上的 DID Registry

当生成 DID 后，用户需在 **DID Registry** 中注册其所有权。W3C 标准并未规定具体实现方式，**Olares ID** 采用 [智能合约](/zh/developer/contribute/olares-id/contract/contract.md#smart-contract.md) 完成注册，具有以下优势：

- **去中心化**：无需依赖中心化机构或服务商  
- **抗审查**：结构对审查与干预具有韧性  
- **全网可发现**：其他用户可轻松检索到已注册的 DID  

发行方可将关键的 DID 元数据上传到 Registry，包括：

- DID 本身  
- [Olares ID](olares-id.md)  
- RSA 公钥  
- 其他相关信息  

DID 元数据体积小、更新频率低，适合使用主流区块链技术进行存储与检索，系统开销低且效率高。  

目前，Olares ID 的智能合约部署在以太坊二层网络 **[Optimistic Rollups](https://optimism.io/)** 上。

::: info
值得一提的是，Web5 开发团队已从 ION Registry 迁移至 **DHT（分布式哈希表）网络**。与 Layer2 区块链相比，DHT 拥有更高去中心化程度，但在高节点数量下保持效率、稳定性与数据完整性更为复杂。
:::
