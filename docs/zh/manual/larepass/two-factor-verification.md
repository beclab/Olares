---
description: 在 Vault 中设置和使用双因素身份验证。生成安全的 2FA 代码，增强账户安全性，实现与主流验证器的兼容。
---
# 生成双因素身份验证（2FA）代码

双因素身份验证（2FA）在登录时需要你的密码和身份验证代码。这些代码由基于时间的一次性密码（TOTP）生成，自动刷新。与 Google Authenticator 或 Microsoft Authenticator 类似，Vault 可以为你的在线账户生成安全的 2FA 代码。

本文档将展示如何在 Vault 中生成双因素身份验证（2FA）代码。

## 准备目标服务

1. 登录你希望启用 2FA 的网站或应用（例如 GitHub 或 OpenAI）。
2. 转到安全设置页面，启用基于身份验证应用的双因素身份验证。

   ![Enable GitHub 2FA](/images/manual/olares/2fa-github.png#bordered)

3. 保存提供的密钥或二维码以备后续使用。

:::tip 注意
如果服务提供恢复代码，请安全存储。这些代码在你无法访问 Vault 时对账户恢复至关重要。
:::

## 在 Vault 中创建身份验证器

:::tip 提示
访问 [官方页面](https://olares.cn/larepass) 获取下载选项。
:::

<tabs>
<template #Olares、LarePass-桌面端和浏览器扩展>

1. 在 Vault 中，右上角点击 **<i class="material-symbols-outlined">add</i>添加**。
2. 选择**验证器**作为项目类型，并点击**创建**。
3. 填写必填字段：
    - 项目名称：输入服务的描述性名称，例如 `GitHub`。
    - 一次性密码：粘贴上一步提供的密钥。
4. 点击**保存**。

</template>

<template #LarePass-移动端>

1. 在你的设备上打开 LarePass，并进入应用的 **Vault** 页面。
2. 右上角点击 **<i class="material-symbols-outlined">add</i>添加**。
3. 选择**验证器**作为项目类型，并点击**创建**。
4. 填写必填字段：
    - 项目名称：输入服务的描述性名称，例如 `GitHub`。
    - 一次性密码：点击文本字段中的 <i class="material-symbols-outlined">qr_code</i> 以扫描二维码。
5. 点击**保存**。

</template>
</tabs>

保存后，新的身份验证器将立即开始生成代码。

## 使用你的 2FA 生成器

1. 使用你的用户名和密码登录网站。
2. 当系统提示输入验证码时，打开 Vault 查看当前的 6 位验证码。
3. 输入验证码完成登录。