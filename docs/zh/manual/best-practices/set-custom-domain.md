---
outline: [2, 3]
description: 详细指导如何为 Olares 设置自定义域名，包括域名验证、组织创建、成员管理和 Olares ID 激活的完整流程。
---

# 为 Olares 设置自定义域名

默认情况下，在 LarePass 中创建账户时，系统会分配一个 `olares.cn` 域名的 Olares ID。这意味着你通过类似 `desktop.{your-username}.olares.cn` 的 URL 访问 Olares 服务。如果你希望使用自己的域名，可以按照本教程完成配置。

## 学习目标
通过本教程，你将学习：
- 在 Olares Space 中添加并验证自定义域名
- 在 LarePass 中创建组织和域名下的 Olares ID
- 使用自定义域名安装并激活 Olares
- 在 Olares Space 中为组织添加成员

## 自定义域名工作原理

在 Olares 中，自定义域名通过组织进行管理。无论你是个人用户还是代表公司，都需要先创建一个组织。所需操作取决于你的角色：

| 步骤                                                          | 组织管理员 | 组织成员 |
|---------------------------------------------------------------|:--------:|:------:|
| 在 LarePass 中创建 DID                                         | ✅       | ✅     |
| 将自定义域名添加到 Olares Space                                   | ✅       |        |
| 为域名创建组织，并以管理员身份<br>在 LarePass 中创建 Olares ID        | ✅       |        |
| 在 Olares Space 中为组织添加新用户                                | ✅       |        |
| 加入组织，并在 LarePass 中创建 Olares ID                          |          | ✅     |
| 安装 Olares                                                    | ✅       | ✅     |

如果你要加入已有组织，可以直接跳到[以成员身份加入组织](#以成员身份加入组织)。

## 前提条件

请确保你已准备好：
- 从域名注册商购买了一个有效的域名。
- 在手机上安装了 LarePass 应用。你将使用 LarePass 登录 Olares Space，并将自定义域名与 Olares ID 关联。

:::info
如果你之前已在设备上安装并激活过 Olares，并且想在同一硬件上使用自定义域名，请先[恢复出厂设置](../larepass/manage-olares.md#将-olares-恢复出厂设置)，然后[创建新账户](../larepass/create-account.md)，再按照本教程操作。
:::

## 第 1 步：创建 DID

<!--@include: ../../reusables/custom-domain.md{21,31}-->

## 第 2 步：添加域名

以下步骤以 `space.n1.monster` 为例。

1. 在浏览器中打开 [Olares Space](https://space.olares.com/)，使用 LarePass 扫码登录。

   ![LarePass 扫码登录](/images/manual/tutorials/scan-qr-code1.png)

<!--@include: ../../reusables/custom-domain.md{36,73}-->

## 第 3 步：创建组织

在 LarePass 中，为你的域名创建组织并获得管理员权限的 Olares ID。

<!--@include: ../../reusables/custom-domain.md{76,107}-->

## 第 4 步：安装并激活 Olares

现在可以使用你的 Olares ID 安装并激活 Olares 了。

安装步骤与标准流程类似。以下以 Linux 为例，其他系统请参阅[安装指南](../get-started/install-olares.md)。
<!--@include: ../get-started/install-and-activate-olares.md{5,7}-->

1. 在要安装 Olares 的设备上打开终端，运行以下命令：

   ```bash
   export PREINSTALL=1 &&
   curl -sSfL https://cn.olares.sh | bash -
   ```
   此命令仅执行预安装阶段，不会进入完整安装流程。

<!--@include: ../get-started/install-and-activate-olares.md{9,10}-->

<!--@include: ../get-started/install-and-activate-olares.md{11,12}-->

<!--@include: ../get-started/install-and-activate-olares.md{19,19}-->

激活完成后，LarePass 将显示带有自定义域名的 Olares 桌面地址，如 `https://desktop.alex.space.n1.monster`。

## 第 5 步：添加成员

作为管理员，在 Olares Space 中为组织添加成员。

<!--@include: ../../reusables/custom-domain.md{109,123}-->

   成员将使用这些凭据在 LarePass 中创建自己的 Olares ID。

7. （可选）如果成员将使用你现有的 Olares 实例而非单独安装设备，你还需要在 Olares 上创建该用户并分配资源。详见[管理团队](../olares/settings/manage-team.md)。

## 以成员身份加入组织

在 LarePass 应用中，点击**创建账户**开始账户创建流程。

<!--@include: ../../reusables/custom-domain.md{130,141}-->

自定义域名的 Olares ID 创建完成。下一步是激活 Olares：

- **在单独的设备上安装**：按照[第 4 步](#第-4-步：安装并激活-olares)操作。
- **在与管理员相同的设备上**：从管理员处获取激活向导地址和一次性密码，然后在 LarePass 中扫描向导二维码完成激活。详见[激活 Olares](../get-started/join-olares.md)。


## 了解更多

- [Olares 账户](../../developer/concepts/account.md)：DID、Olares ID 和组织的工作原理。
- [安装 Olares](../get-started/install-olares.md)：不同平台和环境的安装方式。
- [管理团队](../olares/settings/manage-team.md)：在 Olares 实例中创建和管理用户账户。
