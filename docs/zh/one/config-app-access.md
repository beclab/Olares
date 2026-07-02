---
outline: [2, 3]
description: 配置认证级别和模型，以控制用户如何访问 Olares One 上的应用。
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/config-app-access.md)为准。
:::

# 配置应用访问权限 <Badge text="5 min"/>

Olares 上的每个应用都有一个入口，用于控制用户访问方式。你可以为每个入口配置访问策略，以匹配应用的敏感度和共享需求。

## 开始之前

访问策略有两个设置项：

- **认证级别**：定义何时需要认证。
- **认证模型**：定义用户如何认证。

下表展示了可用的组合及其访问行为：

| 认证级别 | 认证模型 | 访问行为 |
| --- | --- | --- |
| **Private** | **System**、**One factor**、**Two factor** | 所有用户在访问应用前都必须认证。 |
| **Internal** | **System**、**One factor**、**Two factor** | 使用 LarePass VPN 的用户跳过认证。其他访问方式需要认证。 |
| **Public** | **None** | 任何人无需登录即可访问应用。 |

## 设置访问策略

1. 前往 **Settings** > **Applications**。
2. 选择目标应用。
3. 在 **Entrances** 部分，点击要配置的入口。

   ![设置入口](/images/one/settings-entrance.png#bordered){width=80%}

4. 在 **Access policy** 下，选择 **Authentication level**：
   - **Private**：仅允许认证后访问。
   - **Internal**：VPN 访问无需登录。
   - **Public**：无需登录即可访问。

5. 选择 **Authentication model**：
   - **System**：遵循系统范围的认证设置。
   - **One factor** 或 **Two factor**：为该应用单独应用认证。
   - **None**：无需认证。

   :::warning
   **None** 仅在 **Authentication level** 设置为 **Public** 时可用。这意味着任何拥有 URL 的人都可以无需登录访问该应用。
   :::

6. 点击 **Confirm** 保存更改。

## 资源
- [Entrance 概念](/developer/concepts/network.md#entrance)：了解更多技术背景。
- [自定义应用域名](../../manual/olares/settings/custom-app-domain.md#custom-domain-name)：了解如何使用自己的域名访问应用。
