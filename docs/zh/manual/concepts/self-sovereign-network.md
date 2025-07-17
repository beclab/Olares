---
description: 介绍 Olares 基于 DID 与区块链的自主主权网络及 BEC 架构，实现去信任的信息与价值交换。
---

# 自主主权网络（Self-Sovereign Network）

借助 [**DID**](did.md) 与 [**区块链注册表**](registry.md)，Olares 构建了基于 **Blockchain-Edge-Client（BEC）** 架构的点对点自主主权网络，消除了对第三方信任的依赖，使任意两方均可直接进行信息与价值的去信任交换。

## BEC 架构

BEC 通过在最合适的位置分布存储数据，实现彻底去中心化，由三大支柱组成：

- **Edge（边缘）**  
  用户将个人数据（文档、聊天记录、照片等）存储在私有边缘服务器上。用户与他人或服务的所有交互都通过该服务器完成。  
  > Olares：开源自托管操作系统，运行在本地边缘设备，即此组件的具体实现。

- **Blockchain（区块链）**  
  高价值数据（如 DID 和交易）存储在链上，确保透明、安全与可发现性。  
  > Olares ID 通过智能合约将 DID Registry 部署于 EVM 兼容链（如 Optimism）。

- **Client（客户端）**  
  身份钱包应用，私钥保存在移动设备，用户完全掌控。  
  > LarePass：Olares 的全功能客户端，即此组件的具体实现。

## 通过 BEC 进行去信任信息交换

以下示例展示 Alice 向 Bob 发送消息的流程。二者已在区块链上注册 DID。

![BEC 拓扑](/images/manual/concepts/network-topology.jpeg)

1. **终端 → Edge**：Alice 在终端设备发送消息至其 Edge 服务器。  
2. **链上定位**：Alice 的 Edge 通过区块链这一去中心化 **DNS** 定位 Bob 的 Edge。  
3. **Edge → Edge**：消息从 Alice 的 Edge 转发至 Bob 的 Edge。  
4. **签名验证**：Bob 的 Edge 以区块链为 **CA**，验证 Alice 的加密签名。  
5. **Edge → 终端**：验证通过后，消息安全地转发至 Bob 的终端设备。  

## 通过 Otmoic 协议进行去信任价值交换

**Otmoic** 是建立在 **Olares ID** 之上的去信任自动价值交换协议，目标是在 **公用品** 场景中提供公平价格与透明交易。

![Otmoic RFQ 流程](/images/manual/concepts/rfq.jpeg)

核心特性：

| 功能                     | 说明                                                                                     |
|--------------------------|------------------------------------------------------------------------------------------|
| 链上声誉机制             | 对交易者与流动性提供者建立声誉体系，解决 **Free-Mint** 问题                             |
| VC-驱动的 KYC            | 通过可验证凭证完成身份验证而不损害去中心化                                               |
| 基于 RFQ 的价格发现      | 高效的 **Request-for-Quote** 模式进行价格撮合                                           |
| 原子交换（Atomic Swap）  | 支持链上无信任原子交换交易                                                               |
| 自动做市                 | 流动性提供者可在 **Olares** 上安装应用参与自动做市                                       |
