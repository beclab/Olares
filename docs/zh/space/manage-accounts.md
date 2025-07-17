---
description: 介绍 Olares Space 的账号操作方法，通过 LarePass 扫码实现 DID 和 Olares ID 登录，支持多账号导入和切换，可随时退出登录切换身份。
---
# 管理 Olares Space 账号

本指南将介绍 Olares Space 的常见账号操作，包括登录、多账号管理和退出登录。

## 登录 Olares Space

Olares Space 使用去中心化身份（DID）或 Olares ID 进行身份验证。请确保你已在 LarePass 中获取了相应的凭据。

1. 在 LarePass 应用中，选择要用于登录的 DID 或 Olares ID。
2. 在浏览器中打开 [https://space.olares.com/](https://space.olares.com/)。
3. 使用 LarePass 扫描二维码。

::: tip DID 与 Olares ID 的差异
根据使用 DID 还是 Olares ID 登录，Olares Space 提供的功能和服务会有所不同。
- **使用 DID 登录**：由于账号未关联域名，你可以设置自己的域名。但在激活 Olares 设备前，必须先将 DID 绑定到 Olares ID。
- **使用 Olares ID 登录**：只要该名称尚未被其他 Olares 设备使用，你就可以创建 Olares。但因为 Olares ID 已对应唯一域名，所以无法使用自定义域名。
  :::

## 退出登录

退出账号有以下方式：

1. 点击右上角的头像。
2. 选择**退出登录**。

或者：

1. 从菜单中选择**切换账号**。
2. 点击任意列出账号旁边的 <i class="material-symbols-outlined">logout</i> 图标。

## 管理多个账号

每个 Olares ID 只能关联一个 Olares。通过 Olares Space 的多账号管理功能，你可以轻松切换账号，方便管理多个 Olares ID 和实例。

添加账号的步骤：

1. 点击右上角的头像。
2. 在弹出菜单中选择**导入账号**。
3. 打开 LarePass，扫描二维码登录。

添加多个账号后，可以通过菜单中的**切换账号**选项进行切换。如果账号已退出登录，系统会跳转到二维码登录页面。