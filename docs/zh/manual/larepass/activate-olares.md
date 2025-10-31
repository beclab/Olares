---
description: 了解如何首次激活 Olares、在重新安装后重新激活，以及使用 LarePass 移动端完成安全的双因素登录。
---

# 激活 Olares

Olares 通过 **Olares ID** 与 **LarePass 移动应用**提供安全且流畅的身份验证体验。本文介绍如何激活 Olares，并在登录时使用 LarePass 完成双因素验证。

:::warning 管理员网络要求
为避免激活失败，管理员用户在激活时需确保手机和 Olares 设备处于同一网络。
对于成员用户，则没有此限制。
:::

## 通过一键脚本安装后激活

如果你通过一键安装脚本完成[Olares 安装和初始配置](../get-started/install-olares.md#安装-olares-1)后，可使用以下步骤激活 Olares 实例：

![激活](/images/manual/larepass/activate-olares.png#bordered)

1. 打开 LarePass。  
2. 点击**扫码**，扫描安装向导中的二维码。  
3. 按照 LarePass 指引重置 Olares 登录密码。  

激活成功后，LarePass 将返回主页，安装向导将跳转至登录页。

## 通过 ISO 安装后激活

如果你通过 ISO 安装 Olares，或使用预装了 ISO 的 Olares 硬件，请按以下步骤激活：

<!--@include: ../get-started/install-and-activate-olares.md{9,23}-->

:::tip 无法发现设备？
如果手机无法连接到 Olares 的现有网络，LarePass 将无法发现待激活设备。此时，可通过[蓝牙配网](manage-olares.md#蓝牙配网)功能，将 Olares 连接至手机所在 Wi-Fi 网络，再重复此激活流程。
:::

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
2. 点击**确认**完成登录。  

### 手动输入验证码
![OTP](/images/manual/larepass/otp-larepass.jpg#bordered)

1. 在安装向导页面选择 **使用 LarePass 生成的一次性密码验证**。  
2. 在手机上打开 LarePass，进入**设置**。  
3. 在**我的 Olares** 卡片里，点击身份验证器，生成一次性验证码。  
4. 返回安装向导页面，输入验证码完成登录。  

:::tip 提示
验证码具有时效性，请在过期前输入；若已过期，请重新生成。
:::

验证成功后，你将自动跳转至 Olares 桌面。
