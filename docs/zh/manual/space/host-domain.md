---
outline: [2, 3]
description: 在 Olares Space 中设置自定义域名，包括域名验证和 DNS 配置。
---

# 设置自定义域名

无论你是希望团队成员使用公司域名的组织用户，还是想使用个人域名，Olares Space 都支持为 Olares 系统配置自定义域名。

:::tip 第一次设置自定义域名？
如需从域名设置到 Olares 安装的完整流程，请参阅[为 Olares 设置自定义域名](../best-practices/set-custom-domain.md)。
:::

## 前提条件

请确保：
- 账户处于 DID 阶段。只有 DID 阶段的账户才能关联自定义域名。具体操作见[创建 DID](../larepass/create-org-account.md#创建-did)。
- 已使用 DID 登录 Olares Space。参见[登录和管理账户](manage-accounts.md)。
- 拥有一个通过域名注册商注册的域名，该域名未在 Olares Space 中关联到其他账户。
- 在手机上安装了 LarePass。
- 可以访问域名的 DNS 设置，用于配置 TXT 和 NS 记录。

## 添加域名

<!--@include: ../../reusables/custom-domain.md{34,73}-->

## 下一步

域名添加完成后，在 LarePass 中创建组织和 Olares ID。详见[使用自定义域名创建 Olares ID](../larepass/create-org-account.md)。
