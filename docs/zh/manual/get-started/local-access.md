---
outline: [2,3]
description: 了解如何通过局域网 (LAN) 直接访问 Olares 应用和服务，以获得最快速度、隐私保护和离线可靠性。
---
# 内网访问 Olares 服务
虽然远程访问非常方便，但通过内网直接访问设备具有显著优势：
- **极致性能：** 无需经过互联网路由，即可高速传输文件。
- **增强隐私：** 将流量严格限制在家庭网络内。
- **离线独立性：** 即使互联网服务提供商断网，也能访问数据和应用。

本文档介绍了建立本地连接的几种方法：
- [启用 LarePass VPN（推荐）](#方法-1启用-larepass-vpn)<br/>这是最简单的方案，自动建立最快连接，无需手动配置。
- [使用 `.local` 域名](#方法-2使用-local-域名)<br/>此方法无需安装，但需根据操作系统使用特定的 URL 格式。
- [配置本地 DNS（进阶）](#方法-3配置本地-dns)<br/>此方法通过更新路由器或单台电脑的 DNS 设置，支持使用标准 URL 进行本地访问。
- [修改 Host 文件（备选）](#方法-4修改-hosts-文件)<br/>此方法在单台电脑上手动将标准 URL 映射到本地 IP，确保在无外网连接时也能访问。

## 方法 1：启用 LarePass VPN
无论是在设备旁还是在旅途中，连接的最稳健方式都是使用 LarePass VPN。它能智能检测你是否处于同一网络，并切换到直接的**内网**模式以获得最大速度。

:::tip 始终启用 VPN 以进行远程访问
保持 LarePass VPN 启用。它会自动优先选择最快的可用路由，确保你无需手动切换即可获得最佳速度。
:::
:::info iOS 和 macOS 设置
在 iOS 或 macOS 上首次启用该功能时，系统可能会提示你添加 VPN 配置文件。允许此操作以完成设置。
:::

直接在你当前用于访问 Olares 的设备上启用 LarePass VPN。

<tabs>
<template #使用-LarePass-移动端>

1. 打开 LarePass 应用，进入**设置**。
2. 在**我的 Olares** 卡片中，打开 VPN 开关。

   ![移动端开启 LarePass VPN](/images/zh/manual/get-started/larepass-vpn-mobile.png#bordered)
</template>
<template #使用-LarePass-桌面端>

1. 打开 LarePass 应用，点击左上角的头像打开用户菜单。
2. 打开**专用网络连接**开关。

   ![桌面端开启 LarePass VPN](/images/zh/manual/get-started/larepass-vpn-desktop.png#bordered)
</template>
</tabs>

启用后，可以查看 LarePass 中的网络状态以确认是否正在使用本地连接：

| 状态         | 描述                         |
|:-----------|:---------------------------|
| **内网**     | 通过本地局域网 IP 直连。速度最快。        |
| **P2P**    | 设备间的直接加密隧道。连接速度快。          |
| **DERP**   | 通过安全中继服务器路由，仅在无法直连时作为备用方案。 |

## 方法 2：使用 `.local` 域名
如果不想使用 VPN，可以使用 `.local` 域名访问服务。根据兼容性需求，有两种域名格式可用。

### 单级域名（所有操作系统适用）
:::warning 仅支持社区应用
Desktop 和文件管理器等 Olares 系统应用不支持此 URL 格式，因此无法正确加载。
:::
此格式通过连字符（`-`）连接入口 ID 和用户名来使用单级主机名。
- **默认 URL**:
   ```plain
   https://<entrance_id>.<username>.olares.cn
   ```
- **本地访问 URL**:
   ```plain
   http://<entrance_id>-<username>-olares.local
   ```

### 多级域名 (macOS 和 iOS 适用)
Apple 设备支持通过 [Bonjour](https://developer.apple.com/bonjour/)（零配置网络）进行本地服务发现，因此能够在 macOS 和 iOS 上解析 `.local` 下的多级域名。这使得本地 URL 格式可以与默认的远程访问地址保持结构一致。

- **默认 URL**:
   ```plain
   https://<entrance_id>.<username>.olares.cn
   ```
- **本地访问 URL**:
   ```plain
   http://<entrance_id>.<username>.olares.local
   ```
![多级域名](/images/manual/get-started/multilevel-local-domain-mac.png#bordered)

## 方法 3：配置本地 DNS
为了获得无缝体验（即标准 URL 自动解析为你的本地 IP 地址），你可以配置网络 DNS。此配置确保网络上所有设备的访问一致，无需单独设置客户端。

### 查找 Olares 设备的内网 IP
要配置 DNS，首先需要找到 Olares 设备的内网 IP。
<tabs>
<template #通过-LarePass-手机端查看>

如果你的手机和 Olares 设备处于同一网络：
1. 打开 LarePass 应用，点击**设置** > **系统**，进入 **Olares 管理**页面。
   ![点击 System 卡片](/images/zh/manual/get-started/larepass-system.png#bordered)

2. 点击设备卡片。
   ![点击设备卡片](/images/zh/manual/get-started/larepass-device-card.png#bordered)

3. 并向下滚动到**网络**部分。可以在此处找到**内网 IP**。
   ![找到网络部分](/images/zh/manual/get-started/larepass-network.png#bordered)

</template>
<template #通过-Olares-终端查看>

控制面板提供了内置终端，允许你直接从浏览器运行系统命令，无需外部 SSH 客户端。
1. 打开控制面板，在左侧导航栏的**终端**下选择 **Olares**。
   ![从控制面板查找内网 IP](/images/zh/manual/get-started/find-internal-ip-from-controlhub.png#bordered)

2. 在终端输入 `ifconfig` 并按下 **Enter** 确认。
3. 寻找你的活动连接，通常命名为 `enp3s0`（有线）或 `wlo1`（无线）。IP 地址位于 `inet` 之后。

   输出示例：
   ```bash
    enp3s0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
    inet 192.168.50.116  netmask 255.255.255.0  broadcast 192.168.50.255
    inet6 fe80::4194:4045:c35e:7b32  prefixlen 64  scopeid 0x20<link>
    ether d8:43:ae:54:ce:fc  txqueuelen 1000  (Ethernet)
    RX packets 80655321  bytes 71481515308 (71.4 GB)
    RX errors 0  dropped 136  overruns 0  frame 0
    TX packets 51867817  bytes 15924740708 (15.9 GB)
    TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
   ```
在此示例中，`192.168.50.116` 即为内网 IP。
</template>
</tabs>

### 配置 DNS
确定内网 IP 地址后，现在必须配置 DNS 设置以正确路由流量。你可以将此配置应用于单台电脑以供个人访问，或者更新路由器以实现网络上所有设备的无缝本地解析。
<tabs>
<template #设置单台设备>

更新特定电脑上的 DNS 设置。以 macOS 为例：
1. 打开 Apple 菜单，进入**系统设置**。
2. 选择 **Wi-Fi**，然后点击已连接网络的**详细信息**。
3. 选择 **DNS** 并更新服务器列表：

   a. 点击 **DNS 服务器**下的 **+** 按钮，添加 Olares 设备的内网 IP（例如 `192.168.x.x`）。

   b. 确保 Olares IP 列在顶部。在其下方添加你原来的 DNS（或 `1.1.1.1`）作为备用。<br/>这确保了如果 Olares 设备关机，路由器会自动切换到辅助 DNS，保持互联网连接正常。

4. 点击**好**保存更改。

</template>

<template #设置所有设备>

更新路由器上的 DNS，将更改应用于网络中的所有设备。

1.  登录路由器的管理面板。
2.  导航至 **DHCP / DNS 设置**。
3.  将**首选 DNS** 设置为 Olares 设备的内网 IP（例如 `192.168.x.x`）。
4.  将**备用 DNS** 设置为当前的首选 DNS（或公共提供商，如 `1.1.1.1`）。<br/>这确保了如果 Olares 设备关机，路由器会自动切换到辅助 DNS，保持互联网连接正常。
5.  保存并重新连接设备以刷新 DNS 缓存。
</template>
</tabs>

配置完成后，可以使用标准公网地址和本地地址访问 Olares。
:::tip
你可以从 Olares 应用市场安装 AdGuard Home，以图形化方式监控流量并管理 DNS 映射。
:::
## 方法 4：修改 hosts 文件
如果无法更改路由器设置且需要在特定电脑上立即离线访问，可以在 hosts 文件中手动映射域名。

1.  找到 hosts 文件：
    - **Windows:** `C:\Windows\System32\drivers\etc\hosts`
    - **macOS/Linux:** `/etc/hosts`
2.  使用文本编辑器打开文件（需要管理员权限）。
3.  添加映射行：
    ```plain
    # 替换为内网 IP 和用户名
    # Olares 应用
    192.168.31.208  desktop.<username>.olares.cn
    192.168.31.208  auth.<username>.olares.cn
    192.168.31.208  files.<username>.olares.cn
    192.168.31.208  market.<username>.olares.cn
    192.168.31.208  settings.<username>.olares.cn
    192.168.31.208  dashboard.<username>.olares.cn
    192.168.31.208  control-hub.<username>.olares.cn
    192.168.31.208  profile.<username>.olares.cn
    192.168.31.208  vault.<username>.olares.cn
    # 根据需要添加其他社区应用
    192.168.31.208  <entrance_id>.<username>.olares.cn
    ```
这允许你在无需互联网连接的情况下本地访问 Olares。
## 常见问题 (FAQs)
### 为什么在 Mac 上无法再启用 LarePass VPN？
如果之前成功启用过 VPN，但现在停止工作，可能需要重置系统扩展。
:::info
根据 macOS 版本不同，界面可能略有差异。
:::
1. 打开**系统设置**，搜索 "Extension"（扩展），选择**登录项与扩展**。
2. 滚动到**网络扩展** 部分，点击信息图标 (ⓘ) 查看已加载的扩展。
3. 找到 LarePass，点击三个点 (...)，选择**删除扩展**。
4. 确认卸载。
5. 重启 Mac 并在 LarePass 桌面客户端中重新启用 VPN。

### 为什么我在 Windows 上无法启用 LarePass VPN？
第三方杀毒软件可能会错误地将 LarePass 桌面客户端标记为可疑，从而阻止其启动 VPN 服务。

如果在首次打开 LarePass 时收到杀毒软件提示，请允许应用程序继续运行。

如果 VPN 仍然无法启用：
1. 打开安全软件，检查 LarePass 是否被拦截。
2. 将 LarePass 主程序添加到杀毒软件的白名单或排除项中。
3. 重启 LarePass 并启用 VPN。

### 为什么 `.local` 域名在 Chrome 中无法使用 (macOS) ？
如果 macOS 未授予局域网访问权限，Chrome 可能会无法访问本地 URL。
要启用访问权限：
1. 打开 Apple 菜单，进入**系统设置**。
2. 进入 **隐私与安全性** > **局域网**。
3. 在列表中找到 Google Chrome 和 Google Chrome Helper，并开启开关。
   ![启用局域网权限](/images/manual/larepass/mac-chrome-local-access.png#bordered){width=400}

4. 重启 Chrome 并再次尝试访问本地 URL。

### 为什么在 Chrome 上使用 `.local` 域名时，应用无法在 iFrame 中加载 (macOS) ？
使用本地域名时，Chrome 可能会默认使用 HTTPS，你可能会看到“连接不安全”的警告。
![本地地址错误](/images/manual/get-started/incorrect-local-address.png#bordered)

要解决此问题，在 URL 开头显式添加 HTTP 协议头 (`http://`)。

