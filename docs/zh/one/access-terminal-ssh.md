---
outline: [2, 3]
description: 了解如何通过 SSH 访问 Olares One 主机终端以进行命令行操作。
head:
  - - meta
    - name: keywords
      content: Olares One, 终端, SSH
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/access-terminal-ssh.md)为准。
:::

# 通过网络访问 Olares One 终端（SSH）

安全外壳（SSH）通过网络建立加密会话，让你可以从自己的电脑上对 Olares One 执行命令行操作。

你可以通过本地网络连接，或者使用 LarePass VPN 从不同位置安全连接。

## 前提条件

**硬件**
- Olares One 已完成设置并连接到网络。
- 一台用于访问终端的客户端设备，如电脑。
- 一台安装了 LarePass 应用的移动设备。

**经验**
- 熟悉基本的终端命令和命令行界面（CLI）。

## 通过本地网络连接

如果你的设备和 Olares One 处于同一本地网络，请按照以下步骤操作。

### 步骤 1：获取 Olares One 的本地 IP 地址

1. 在你的移动设备上打开 LarePass 应用，然后进入 **Settings** > **System**。

   ![点击 System 卡片](/images/manual/get-started/larepass-system.png#bordered)

2. 点击 Olares One 设备卡片。
3. 向下滚动到 **Network** 部分，然后记下 **Intranet IP**。

   :::tip 通过 Control Hub 查看
   你可以在 Control Hub 终端中使用 `ifconfig` 命令查看 IP。

   查找你的活动接口，通常是 `enp3s0`（有线）或 `wlo1`（无线）。IP 地址显示在 `inet` 之后。
   :::

### 步骤 2：从 Vault 获取登录密码

:::info 与 Olares Desktop 密码不同
此密码用于通过 SSH 登录 Olares One 主机系统。它与你在浏览器中登录 Olares Desktop 时使用的密码不同。
:::

<!--@include: ./reusables-reset-ssh.md#reset-ssh-upon-activation-->

### 步骤 3：通过 SSH 连接

Olares One 的默认用户名是 `olares`。

1. 在你的电脑上打开终端。
2. 运行以下命令，将 `<local_ip_address>` 替换为你记下的内网 IP：
   ```bash
   ssh olares@<local_ip_address>
   ```
3. 按提示输入步骤 2 中获取的密码。

## 高级：从不同网络远程连接

如果你的设备与 Olares One 不在同一本地网络，请使用 LarePass VPN 建立安全隧道来连接 Olares One，而无需将其暴露在互联网上。LarePass VPN 使用 [Tailscale](https://tailscale.com/)，这是一种网状 VPN，为每个连接的设备分配一个 `100.64.0.0/10` 范围内的稳定 IP 地址（Tailscale IP），以实现它们之间的直接加密通信。

### 步骤 1：通过 VPN 启用 SSH

1. 打开你的 Olares desktop，然后进入 **Settings** > **VPN**。
2. 打开 **Allow SSH via VPN**。
3. 在你的电脑上，打开 LarePass 桌面客户端。
4. 点击左上角的头像，然后打开 **VPN connection**。
    ![在桌面上启用 LarePass VPN](/images/one/ssh-enable-vpn.png#bordered)

### 步骤 2：查找 Tailscale IP 地址

1. 在你的 Olares desktop 上，进入 **Settings** > **VPN** > **View VPN connection status**。
2. 找到 **olares**，然后点击它以展开连接详情。
3. 找到以 `100.64` 开头的 IP 地址，然后记下它。
    
    ![在 VPN 设置中查看 Tailscale IP 地址](/images/one/ssh-remote-ip.png#bordered){width=80%}

### 步骤 3：从 Vault 获取登录密码

<!--@include: ./reusables-reset-ssh.md#reset-ssh-upon-activation-->

### 步骤 4：通过 SSH 连接

Olares One 的默认用户名是 `olares`。

1. 在你的电脑上打开终端。
2. 运行以下命令，将 `<tailscale_ip_address>` 替换为你记下的 Tailscale IP 地址：
   ```bash
   ssh olares@<tailscale_ip_address>
   ```
   :::info
   启用 SSH via VPN 后，首次连接速度较慢，因为 VPN 路由正在应用。请稍等片刻以完成连接。
   :::
3. 按提示输入步骤 3 中获取的密码。

:::tip 改用本地 IP 地址连接
如果在 **Settings** > **VPN** 中启用了 **Subnet routes**，Olares One 本地网络上的所有设备都将通过 VPN 可达。这样即使从不同网络访问，你也可以使用本地 IP 地址（`192.168.x.x`）而不是 Tailscale IP（`100.64.x.x`）进行 SSH 连接。
:::

## 重置 SSH 密码
<!--@include: ./reusables-reset-ssh.md#reset-ssh-in-settings-->
