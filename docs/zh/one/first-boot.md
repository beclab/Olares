---
description: 了解如何首次设置 Olares One，包括设置硬件、安装客户端应用、创建 Olares 账户、连接设备、安装和激活系统，以及登录 Olares。
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, 首次启动, 初始设置, 首次使用
---

# 首次启动

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/first-boot.md)。
:::

本指南将引导你完成 Olares One 的首次设置。

## 设置概览
- 你无需将显示器、键盘或鼠标连接到 Olares One。整个配置通过手机完成。
- LarePass 应用是你进行初始设置、认证和远程访问管理的主要工具。

## 前提条件
**硬件**
- Olares One 已连接电源。
- （推荐）一根以太网线将 Olares One 连接到路由器。

**网络**
- 可靠的互联网连接。
- 你的手机（iOS 或 Android）连接到同一网络。

## 步骤 1：开机并安装 LarePass
1. 打开 Olares One 电源。状态 LED 变为白色常亮，表示设备已开机。
2. 在 Apple App Store 或 Google Play Store 中搜索 "LarePass"。在你的移动设备上安装并打开应用。
3. 按照屏幕上的指示创建你的 Olares ID。这个唯一标识符在整个 Olares 生态系统中充当你的用户名。
   ![创建 Olares ID](/images/one/create-olares-id.png#bordered){width=90%}

## 步骤 2：连接 Olares One

当你的 ID 准备好后，你需要发现并关联你的 Olares One。

<tabs>
<template #通过有线局域网设置>

1. 确保你的 Olares One 通过以太网连接到路由器。
2. 在 LarePass 应用中，点击 **发现附近的 Olares**。
   ![发现附近的 Olares](/images/one/discover-nearby-olares.png#bordered){width=90%}

3. 从可用设备列表中选择你的 Olares One。
</template>

<template #通过 Wi-Fi 设置（蓝牙）>
如果无法使用有线连接，可以使用蓝牙配置 Wi-Fi 凭据。

1. 在 LarePass 应用中，点击 **发现附近的 Olares**。
2. 点击底部的 **蓝牙网络设置**。
3. 从蓝牙列表中选择你的设备，然后点击 **网络设置**。
4. 按照提示将 Olares One 连接到你的手机当前使用的 Wi-Fi 网络。
5. 连接完成后，返回主屏幕并再次点击 **发现附近的 Olares** 以找到你的设备。

</template>
</tabs>

## 步骤 3：安装并激活 Olares OS

1. 在 LarePass 应用中，在你刚刚找到的设备上点击 **立即安装**。
2. 安装完成后，点击 **立即激活** 以初始化系统。
3. 选择离你位置最近的反向代理节点，然后点击 **确认**。反向代理节点充当远程访问的安全网关。选择最近的节点可确保最快的连接速度和最佳稳定性。
4. 设置 Olares 的登录密码。
5. 复制或记下你的个人桌面 URL。你需要此 URL 来访问你的 Olares 服务。

    ![获取 URL](/images/one/obtain-url.png#bordered)

6. 点击 **知道了** 关闭提示。

## 步骤 4：登录 Olares Desktop
1. 在计算机上打开 Web 浏览器，访问你的桌面 URL。
2. 在登录页面，输入你的登录密码。
3. 系统会提示你完成双重验证。打开 LarePass 批准登录请求，或手动输入应用中显示的 6 位验证码。
   ::: info
   验证码具有时效性。请确保在过期前输入。
   :::

## 后续步骤
恭喜！你的 Olares One 已设置完成并激活。强烈建议完成以下步骤以保护你的账户并优化体验。

### 备份助记词
:::warning 安全警告
你有责任保护自己的安全。切勿分享此短语。如果丢失这 12 个单词，你将永久失去对 Olares ID 和存储在 Vault 中所有数据的访问权限。
:::
你的 Olares ID 由唯一的 12 个单词的助记词保护。如果你丢失手机或需要在新设备上登录，这个短语是恢复账户的唯一方法。

1. 打开 LarePass 应用，进入 **设置** > **安全**。
2. 点击 **助记词** 并验证你的身份。
3. 点击 **点击查看**。
4. 按提示输入本地密码。
5. 将这 12 个单词写到 **恢复表** 上，然后将恢复表存放在安全的离线位置。

### 重置 SSH 密码

<!--@include: ./reusables-reset-ssh.md#reset-ssh-upon-activation-->

有关如何 SSH 连接到 Olares One 的说明，请参阅[通过 SSH 连接 Olares One](access-terminal-ssh.md)。

### 安全访问 Olares 服务
为了安全远程访问而无需复杂的网络配置，建议启用 LarePass VPN。

请参阅[使用 LarePass VPN 安全访问 Olares 服务](access-olares-via-vpn.md)。

### 探索
Olares OS 预装了系统应用。你还可以浏览 **Market** 下载最适合你需求的额外应用。

你可以继续浏览本文档以发现更多使用场景和高级配置。

## 常见问题

<!--@include: ../manual/help/olares.md#faq-why-olares-id-->

<!--@include: ../manual/help/olares.md#custom-domain-->

详细信息，请参阅[为你的 Olares 设置自定义域名](/manual/best-practices/set-custom-domain.md)。

## 资源
- [使用本地网络访问 Olares](access-olares-via-vpn.md)
- [Olares ID](../developer/concepts/olares-id.md)
- [备份助记词](../manual/larepass/back-up-mnemonics.md)
