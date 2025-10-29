---
description: 了解 Olares ID 的定义、结构及其用途，涵盖个人/组织/实体域名类型，并阐释与去中心化标识符 (DID) 的关系。
---

# Olares ID

本文介绍 Olares ID——Olares 生态中的身份与信任基础。

## 什么是 Olares ID？

**Olares ID** 是一种去中心化身份与信用系统，用于实现信息和价值的安全、无需信任的交换。它是你在 Olares 生态中的数字身份，使个人和组织无需依赖第三方即可自主管理身份。

Olares ID 具备以下特点：

- **唯一性**：类似电子邮件地址，例如 `alice123@olares.com`
- **易读易记**：人性化命名，便于分享
- **与 DID 绑定**：具备加密安全与可验证性

## 为什么需要 Olares ID？

Olares ID 带来无缝访问、增强安全与个性化体验：

- **便捷访问**：系统自动为应用配置子域名与访问策略，可随时随地通过域名访问。
- **免费 HTTPS 证书**：Olares 域名自带 TLS 证书，保障加密通信。
- **个性化且易记**：ID 与域名简单易记，展示独特在线身份。

例如，若你的 Olares ID 为 `alice123@olares.com`，系统自动分配以下地址：

- `alice123.olares.com`：个人主页
- `desktop.alice123.olares.com`：访问 Olares 桌面
- `files.alice123.olares.com`：访问文件管理器应用

## Olares ID 结构

格式与邮箱类似，由两部分组成：

- **本地名（前缀）**
- **域名（后缀）**

如 `alice123@olares.com` 中，`alice123` 为本地名，`olares.com` 为域名。本地名在域内唯一，保证整体唯一性。

### 域名类型

Olares 提供三类域名：

| 类型       | 说明                                       |
|------------|--------------------------------------------|
| 个人域名   | 供个人使用，目前提供默认域名 `olares.com` |
| 组织域名   | 供企业/组织使用，管理员可在成员离职时回收 |
| 实体域名   | 供应用或其他无法归类为个人/组织的实体使用 |

### 个人 Olares ID 创建方式

- **[快速创建](../../manual/larepass/create-account.md#快速创建)**：选择一个可用的本地名立即生成 ID。
- **[高级创建](../../manual/larepass/create-account.md#高级创建)**：使用可验证凭证 (VC) 将现有可信身份（如邮箱）绑定至 Olares ID。
    - 通过 OAuth 验证
    - 将社交身份与 Olares DID 建立加密关联

## Olares ID 与 DID 的关系

**DID**（去中心化标识符）是无需中心化机构即可验证的唯一标识。但 DID 难以记忆，日常使用不便。

Olares ID 采用类似邮箱的可读格式，使 DID 更易用，同时保持其安全性。用户创建 Olares 账户时，系统同时生成并绑定 DID。详情见 [Account 生命周期](./account#understand-the-stage-of-account)。

## 深入阅读

- [创建 Olares ID](../../manual/larepass/create-account.md)
- [去中心化标识符 (DID)](did.md)
- [Gmail issuer 服务](/zh/developer/contribute/olares-id/verifiable-credential/olares.md#gmail-issuer-service)
