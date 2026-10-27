---
outline: [2, 3]
description: 将 Vault 用作身份验证应用，生成基于时间的一次性密码（TOTP）验证码，提升在线服务的账户安全。
head:
  - - meta
    - name: keywords
      content: Olares, LarePass, Vault, 双因素认证, 2FA, TOTP, 验证码
---
# 生成双因素身份验证（2FA）代码

双因素身份验证（2FA）可以为账户安全增加一层保护。输入密码后，服务会要求你提供验证码。

Vault 可以作为身份验证应用使用。它会自动生成基于时间的一次性密码（TOTP）验证码，功能与 Google Authenticator 或 Microsoft Authenticator 类似。

本文档将介绍如何将 Vault 设为验证方式，并为 GitHub 或 OpenAI 等网站生成 2FA 验证码。

## 前提条件

- 你已经在设备上安装了 LarePass。访问[官方页面](https://olares.cn/larepass)下载。
- 你可以访问目标在线服务的安全设置页面。

## 准备目标服务

1. 登录你希望启用 2FA 的网站或应用，例如 GitHub 或 OpenAI。
2. 转到安全设置页面，选择使用身份验证应用来启用双因素身份验证。

   ![启用 GitHub 2FA](/images/manual/olares/2fa-github.png#bordered)

3. 保存服务提供的密钥或二维码，下一节会用到。

:::tip 注意
如果服务提供了恢复代码，请妥善保存。当你无法访问 Vault 时，需要这些代码来恢复账户。
:::

## 在 Vault 中创建身份验证器

<tabs>
<template #Olares、LarePass-桌面端>

1. 打开 **Vault**。
2. 在右上角点击 **<i class="material-symbols-outlined">add</i> 添加**。
3. 选择**验证器**作为项目类型，然后点击**创建**。
4. 填写必填字段：
    - **项目名称**：输入可识别该服务的名称，例如 `GitHub`。
    - **一次性密码**：粘贴上一步保存的密钥。
5. 点击**保存**。

</template>

<template #LarePass-移动端>

1. 在你的设备上打开 LarePass，然后进入 **Vault** 页面。
2. 在右上角点击 **<i class="material-symbols-outlined">add</i> 添加**。
3. 选择**验证器**作为项目类型，然后点击**创建**。
4. 填写必填字段：
    - **项目名称**：输入可识别该服务的名称，例如 `GitHub`。
    - **一次性密码**：点击文本字段中的 <i class="material-symbols-outlined">qr_code</i> 扫描二维码。
5. 点击**保存**。

</template>
</tabs>

保存后，Vault 会立即开始为该账户生成验证码。

## 使用 2FA 验证码

1. 使用你的用户名和密码登录网站。
2. 当网站要求输入验证码时，打开 Vault 查看当前的 6 位验证码。
3. 输入验证码完成登录。
