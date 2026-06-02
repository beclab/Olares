---
outline: [2, 3]
description: 详细指导如何为 Olares 设置自定义域名，包括域名验证、组织创建、成员管理和 Olares ID 激活的完整流程。
---

# 为 Olares 设置自定义域名

默认情况下，在 LarePass 中创建账户时，系统会分配一个 `olares.cn` 域名的 Olares ID。这意味着你通过类似 `desktop.{your-username}.olares.cn` 的 URL 访问 Olares 服务。如果你希望使用自己的域名，可以按照本教程完成配置。

:::warning 先准备一台未激活的 Olares 设备
请先创建自定义域名 Olares ID，再用它激活 Olares。要完成本教程，你需要一台全新设备、已恢复出厂设置的设备，或一台尚未激活过 Olares 的设备。

如果设备已经使用 `olares.cn` 域名的 Olares ID 激活，无法直接改用自定义域名。请先[将 Olares 恢复出厂设置](../larepass/manage-olares.md#将-olares-恢复出厂设置)，再重新按照本教程操作。
:::

## 学习目标
通过本教程，你将学习：
- 在 Olares Space 中添加并验证自定义域名
- 在 LarePass 中创建组织和域名下的 Olares ID
- 在尚未激活的设备上使用自定义域名安装并激活 Olares
- 在 Olares Space 中为组织添加成员

## 自定义域名工作原理

在 Olares 中，自定义域名通过组织进行管理。自定义域名流程分为两部分：

1. 在账户仍处于 DID 阶段时，创建自定义域名下的 Olares ID。
2. 用该自定义域名 Olares ID 激活一台全新或已恢复出厂设置的 Olares 设备。

无论你是个人用户还是代表公司，都需要先创建一个组织。所需操作取决于你的角色：

| 步骤                                                          | 组织管理员 | 组织成员 |
|---------------------------------------------------------------|:--------:|:------:|
| 准备一台全新、已恢复出厂设置或尚未激活过的 Olares 设备           | ✅       | ✅     |
| 在 LarePass 中创建 DID                                         | ✅       | ✅     |
| 将自定义域名添加到 Olares Space                                   | ✅       |        |
| 为域名创建组织，并以管理员身份<br>在 LarePass 中创建 Olares ID        | ✅       |        |
| 在 Olares Space 中为组织添加新用户                                | ✅       |        |
| 加入组织，并在 LarePass 中创建 Olares ID                          |          | ✅     |
| 使用自定义域名 Olares ID 安装并激活 Olares                         | ✅       | ✅     |

如果你要加入已有组织，可以直接跳到[以成员身份加入组织](#以成员身份加入组织)。

## 前提条件

请确保你已准备好：
- 从域名注册商购买了一个有效的域名。
- 在手机上安装了 LarePass 应用。你将使用 LarePass 登录 Olares Space，并将自定义域名与 Olares ID 关联。
- 一台全新、已恢复出厂设置或尚未激活过的 Olares 设备。如果要复用已激活过的硬件，请先[将 Olares 恢复出厂设置](../larepass/manage-olares.md#将-olares-恢复出厂设置)，再继续操作。

## 第 1 步：创建 DID

<!--@include: ../../reusables/custom-domain.md#custom-domain-create-did-->

## 第 2 步：添加域名

以下步骤以 `space.n1.monster` 为例。

1. 在浏览器中打开 [Olares Space](https://space.olares.com/)，使用 LarePass 扫码登录。

   ![LarePass 扫码登录](/images/manual/tutorials/scan-qr-code1.png)

<!--@include: ../../reusables/custom-domain.md#custom-domain-add-domain-steps-->

## 第 3 步：创建组织

在 LarePass 中，为你的域名创建组织并获得管理员权限的 Olares ID。

:::warning 继续前请检查账户和设备
请使用仍处于 DID 阶段的账户创建组织，并确认准备用来安装 Olares 的设备尚未激活。如果要复用已激活过的设备，请先停下，将 Olares 恢复出厂设置。
:::

<!--@include: ../../reusables/custom-domain.md#custom-domain-create-organization-->

6. 点击**下一步**，进入 Olares 激活页面。
   ![在 LarePass 中发现 Olares](/images/manual/tutorials/custom-domain-discover-olares.png#bordered)

## 第 4 步：安装并激活 Olares

<!--@include: ../../reusables/custom-domain.md#custom-domain-install-and-activate-olares-->

## 第 5 步：添加成员

作为管理员，在 Olares Space 中为组织添加成员。

<!--@include: ../../reusables/custom-domain.md#custom-domain-add-user-->

   成员将使用这些凭据在 LarePass 中创建自己的 Olares ID。

7. （可选）如果成员将使用你现有的 Olares 实例而非单独安装设备，你还需要在 Olares 上创建该用户并分配资源。详见[管理团队](../olares/settings/manage-team.md)。

## 以成员身份加入组织

在 LarePass 应用中，点击**创建账户**开始账户创建流程。

<!--@include: ../../reusables/custom-domain.md#custom-domain-join-organization-->

自定义域名的 Olares ID 创建完成。下一步是激活 Olares：

- **在单独的设备上安装**：在全新、已恢复出厂设置或尚未激活过的设备上按照[第 4 步](#第-4-步：安装并激活-olares)操作。
- **在与管理员相同的设备上**：从管理员处获取激活向导地址和一次性密码，然后在 LarePass 中扫描向导二维码完成激活。详见[激活 Olares](../get-started/join-olares.md)。


## 了解更多

- [Olares 账户](../../developer/concepts/account.md)：DID、Olares ID 和组织的工作原理。
- [安装 Olares](../get-started/install-olares.md)：不同平台和环境的安装方式。
- [管理团队](../olares/settings/manage-team.md)：在 Olares 实例中创建和管理用户账户。
