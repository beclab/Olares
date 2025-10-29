---
description: 了解数字身份钱包 LarePass 的核心功能，包括 Olares ID（去中心化身份证明）与 VC 的管理及与 Olares 的无缝交互。
---

# 身份钱包应用（Identity Wallet App）

**数字身份钱包** 是自主主权身份 (SSI) 生态的重要工具，可让用户无需中心化机构即可管理去中心化标识符（DID）、凭证 (VC) 以及与数字服务的交互。

**LarePass** 是 Olares 的官方身份钱包应用，为基于 Olares ID 的自托管去中心化操作系统提供安全、便捷的身份管理与访问能力。

## 管理 Olares ID

创建 Olares 账户时，系统会先生成 DID 并绑定至新的 Olares ID。LarePass 提供以下功能来简化 Olares ID 的管理：

![Olares ID management](/images/manual/concepts/create-terminus-name.png)

- **Olares ID 创建**  
  - 快速创建（无需 VC 绑定）  
  - 高级创建（绑定 VC，现支持 Gmail Issuer Service）  
- **备份 / 导入 Olares ID**  
  使用助记词快速备份与恢复账户  
- **多身份管理**  
  同时管理多个 DID / Olares ID  

详细说明参见 [使用 LarePass 管理账户](../../manual/larepass/create-account.md)。

## 管理可验证凭证（VC）

LarePass 支持绑定至 Olares ID 的 VC 全流程操作，包括签名、验证、发现与呈现：

![VC management](/images/manual/concepts/vc-management.png)

- **签发与存储**：安全保存来自 Issuer 的 VC  
- **验证与呈现**：按需向 Verifier 提交 Verifiable Presentation  
- **发现与检索**：便捷查看各身份对应的 VC 列表  

::: tip 说明
以上为 LarePass 与身份钱包相关的核心功能。更多使用方法请参阅 [LarePass 文档](../../manual/larepass/)。
:::
