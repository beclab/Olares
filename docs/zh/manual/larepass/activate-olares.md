---
outline: [2, 3]
description: 了解如何首次激活 Olares、在重新安装后重新激活，以及使用 LarePass 移动端完成安全的双因素登录。
---

# 激活与登录 Olares

Olares 通过 **Olares ID** 与 **LarePass 移动应用**提供安全且流畅的身份验证体验。本文介绍如何激活 Olares，并在登录时使用 LarePass 完成双因素验证。

:::warning 管理员网络要求
为避免激活失败，管理员用户在激活时需确保手机和 Olares 设备处于同一网络。
对于成员用户，则没有此限制。
:::

## 通过一键脚本安装后激活

如果你通过一键安装脚本完成[Olares 安装和初始配置](../get-started/install-olares.md)后，可使用以下步骤激活 Olares 实例：

![激活](/images/manual/larepass/activate-olares.png#bordered)

1. 打开 LarePass。  
2. 点击**扫码**，扫描安装向导中的二维码。  
3. 按照 LarePass 指引重置 Olares 登录密码。  

激活成功后，LarePass 将返回主页，安装向导将跳转至登录页。

## 通过 ISO 安装后激活

如果你通过 ISO 安装 Olares，或使用预装了 ISO 的 Olares 硬件，请按以下步骤激活：

<!--@include: ../get-started/install-and-activate-olares.md{9,23}-->

### 通过蓝牙激活

如果 LarePass 找不到你的 Olares 设备，可以使用蓝牙激活。这通常发生在 Olares 没有连接有线网络，或者你的手机和 Olares 处于不同网络的情况下。
通过蓝牙，你可以将 Olares 直接连接到你手机当前的 Wi-Fi 网络，以便继续操作。
![蓝牙配网](/images/zh/manual/larepass/bluetooth-network.png#bordered)

1. 在**未发现 Olares** 提示页面底部，点击**蓝牙配网**选项。LarePass 将使用手机蓝牙扫描附近的 Olares 设备。
2. 设备显示后，点击**配置网络**。
3. 选择手机当前连接的 Wi-Fi 网络。如果该网络有密码保护，请输入密码并点击**确认**。
4. Olares 将开始切换网络。完成后你会看到成功消息。此时，如返回到**蓝牙配网**页面，你将看到 Olares 的 IP 地址已更改为与你手机 Wi-Fi 相同的网络。
5. 返回到设备扫描页面，点击**发现附近的 Olares**，找到你的设备并继续激活。

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
