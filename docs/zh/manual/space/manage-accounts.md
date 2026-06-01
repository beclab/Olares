---
description: 使用 LarePass 扫码登录 Olares Space，支持 DID 和 Olares ID 两种登录方式，以及多账户管理和切换。
---

# 登录和管理账户

本页介绍如何登录 Olares Space、管理多个账户以及退出登录。

## 登录 Olares Space

在浏览器中打开 [Olares Space](https://space.olares.com/)，使用 LarePass 扫描二维码登录。可用功能取决于你使用的账户类型。

### 使用 Olares ID 登录

使用 Olares ID 登录后，可以直接创建和管理 Olares。

1. 在 LarePass 应用中，选择要使用的 Olares ID。
2. 点击右上角的扫码图标，扫描 Olares Space 登录页面上的二维码。

### 使用 DID 登录

如果你还没有创建 Olares ID，可以先使用 DID 登录。使用 DID 登录后可以[设置自定义域名](host-domain.md)。

:::tip 第一次设置自定义域名？
从域名设置到 Olares 安装的完整流程，请参阅[为 Olares 设置自定义域名](../best-practices/set-custom-domain.md)。
:::

1. 在 LarePass 应用中，创建一个 DID 或从账户列表中选择已有的 DID。
   ![LarePass 账户列表中的 DID](/images/manual/tutorials/did-stage1.png)

2. 点击右上角的扫码图标，扫描 Olares Space 登录页面上的二维码。
   ![LarePass 扫码](/images/manual/tutorials/scan-qr-code1.png)

## 退出登录

退出账户的方法：

1. 点击右上角的头像。
2. 选择**退出登录**。

或者：

1. 从菜单中选择**切换账户**。
2. 点击任意账户旁边的 <i class="material-symbols-outlined">logout</i>。

## 管理多个账户

每个 Olares ID 只能关联一个 Olares 实例。通过 Olares Space 的多账户管理功能，你可以在不同账户之间切换，管理多个 Olares ID 和实例。

添加账户：

1. 点击右上角的头像。
2. 在弹出菜单中选择**导入账户**。
3. 打开 LarePass，扫描二维码登录。

添加多个账户后，通过菜单中的**切换账户**进行切换。如果账户已退出登录，会跳转到二维码登录页面。
