---
description: Olares 的 BEC 架构概述，阐述分布式节点实现数据存储和安全机制。包含 Snowinning Protocol、Olares OS 和 LarePass 三大核心组件。
---
# Olares 架构

Olares 通过区块链-边缘-客户端（BEC）架构提供了新一代去中心化互联网框架。BEC 通过在不同平台间合理分配信息，实现了数据存储的去中心化和安全性提升。

![BEC](/images/overview/snowinning/network-topology.jpeg)

- **区块链层**：通过智能合约在区块链上存储高价值数据，包括去中心化身份标识（DID）和交易信息。这确保了数据的透明度、不可篡改性和可发现性。
- **边缘层**：作为用户的去中心化网络节点，在私有边缘服务器上托管个人数据，如文档、聊天记录和照片。数据始终保持在边缘层用户的掌控之中，保障隐私和本地数据主权。
- **客户端**：身份钱包应用，让用户可以安全地管理身份并与自托管系统进行交互，同时保持对数字凭证的所有权和隐私控制。

## Olares 核心组件

对应 BEC 架构，Olares 包含以下核心组件：

- [**Olares ID**](olares-id.md)：一个整合了去中心化身份标识（DID）、可验证凭证（VC）和信誉数据的去中心化身份与信誉系统。通过实现去中心化环境中的透明和可验证交互来增强信任。
- [**Olares OS**](https://github.com/beclab/Olares)：专为边缘设备设计的完整自托管操作系统。用户可以托管和管理自己的数据和应用，将个人边缘设备转变为强大的主权云系统。
- [**LarePass**](https://olares.cn/larepass)：安全统一的界面软件，连接用户和其 Olares 系统。提供身份管理、远程访问、设备管理和数据存储等核心功能，确保与 Olares 的无缝交互。

## 延伸阅读

- [主权网络](self-sovereign-network.md)