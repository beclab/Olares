---
description: 介绍可验证凭证（VC）的定义、相关角色与六步验证流程，以及 VC 与 DID 的协同工作方式。
---

# 可验证凭证（Verifiable Credential）

在实体（无论是现实还是数字世界中的有形或无形对象）通过 **DID** 标识自身后，可以针对该实体的多项 “声明” 提供佐证文件。用于证明这些声明的文件即 **可验证凭证（VC）**。VC 是 W3C 已正式推荐的标准，与 DID 协同，能够在无需第三方信任的情况下实现安全交互。

举例而言：  
> Alice 拥有一项“学历”声明；她的毕业证书即为验证该声明的凭证。该证书由其就读的大学签发，可通过密码学方式安全验证。当 Alice 求职时，用人单位可在面试中验证此凭证。

## VC 相关角色

以下示例展示了如何将一份毕业证书从普通凭证转化为 VC：

![Verifiable Credential Process](/images/manual/concepts/vc-diploma.jpeg)

1. **Issuer（签发者）**  
   > 对主体（Subject）做出声明并签发 VC 的实体。  
   示例中，Issuer 是 Alice 就读的大学。

2. **Holder（持有者）**  
   > 持有一个或多个 VC，并向 Verifier 提交证明（Proof）的实体。  
   此处，Holder 是 Alice。她通过钱包应用（此例为 TermiPass）存储与管理 VC。

3. **Verifier（验证者）**  
   > 通过 Presentation Definition 指定所需证明，并验证提交材料的实体。  
   示例中，Verifier 是面试 Alice 的公司。

## 六步验证流程

1. **Issuer 注册 DID**：Issuer 将其信息注册到 DID Registry。  
2. **Holder 发起请求**：Holder 向 Issuer 发送凭证签发请求。  
3. **Issuer 签发 VC**：Issuer 向 Holder 颁发含“学历”声明的 VC。  
4. **Holder 存储 VC**：Holder 将 VC 安全存储于 TermiPass。  
5. **提交 VP**：Holder 生成包含 VC 的 **Verifiable Presentation (VP)** 并发送给 Verifier。  
6. **Verifier 验签**：Verifier 通过 DID Registry 校验 VC 与 VP 的签名，确认声明有效。  

::: tip 提示
VC 本身并不存储在区块链上。
:::
