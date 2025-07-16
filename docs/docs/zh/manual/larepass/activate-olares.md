---
description: 了解如何首次激活 Olares、在重新安装后重新激活，以及使用 LarePass 移动端完成安全的双因素登录。
---

# 激活 Olares

Olares 通过 **Olares ID** 与 **LarePass 移动应用**提供安全且流畅的身份验证体验。本文介绍如何激活 Olares，并在登录时使用 LarePass 完成双因素验证。

:::warning 管理员网络要求
为避免激活失败，管理员用户需确保手机和 Olares 设备连接到同一网络。
:::

## 首次激活

完成 [Olares 安装向导](../get-started/install-olares.md#安装-olares-1)后，可在 **LarePass** 中使用 Olares ID 激活实例。

:::tip Note
新成员用户无需本地安装 Olares 即可通过激活向导激活账户。
详细请参阅[创建成员](../olares/settings/manage-team.md#创建新成员)。
:::

![激活](/images/manual/larepass/activate-olares.png#bordered)

1. 打开 LarePass。  
2. 点击 **扫码**，扫描安装向导中的二维码。  
3. 按照 LarePass 指引重置 Olares 登录密码。  

激活成功后，LarePass 将返回主页，安装向导将跳转至登录页。

## 使用同一 Olares ID 重新激活

如果重新安装了 Olares，仍然想用原有 Olares ID 重新激活：

1. 在手机上打开 LarePass，看到红色提示 “未发现运行中的Olares”。  
2. 点击**了解更多** > **重新激活**，进入扫码界面。  
3. 点击**扫码**，扫描安装向导中的二维码以激活新实例。  

## 使用 LarePass 进行双因素验证

登录 Olares 时，需要完成双因素验证。你可以在 LarePass 中直接确认，或手动输入 6 位验证码。


### 在 LarePass 中确认登录
![2FA](/images/manual/larepass/second-confirmation.png#bordered)

1. 在手机上打开登录通知。  
2. 点击 **确认** 完成登录。  

### 手动输入验证码
![OTP](/images/manual/larepass/otp-larepass.jpg#bordered)

1. 在安装向导页面选择 **使用 LarePass 生成的一次性密码验证**。  
2. 在手机上打开 LarePass，进入**设置**。  
3. 点击顶部的身份验证器，生成一次性验证码。  
4. 返回安装向导页面，输入验证码完成登录。  

:::tip 提示
验证码具有时效性，请在过期前输入；若已过期，请重新生成。
:::

验证成功后，你将自动跳转至 Olares 桌面。
