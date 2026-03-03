---
outline: [2,3]
description: Learn the different methods to access Olares services locally for improved speed and offline capability.
---
# 本地访问 Olares 服务

Olares 的设计初衷是让你随时随地都能无缝访问自己的服务。

然而，通过内网直接访问设备具有显著优势：
- **极致性能**：在本地网络上传输文件时不受互联网延迟和瓶颈的影响。
- **隐私保护**：将流量严格限制在家庭网络内以增加安全性。
- **离线可用**：即使互联网服务不可用，也可以访问数据和应用。

## 目标
通过本教程，你将学会如何：

- 使用 LarePass VPN 建立安全、高速的本地连接。
- 使用 `.local` 域名访问 Olares 服务。
- 配置本地 DNS，使全网设备可通过标准 URL 进行本地访问。
- 手动修改 Hosts 文件，确保特定计算机在无外网环境下也能访问。

## 选择连接方式
有四种建立本地连接的方法：

* **[方法 1：启用 LarePass VPN](#方法-1-启用-larepass-vpn)**<br/>
  利用 LarePass VPN 自动检测本地网络并优化连接速度，无需更改任何设置。
* **[方法 2：使用 `.local` 域名](#方法-2-使用-local-域名)**<br/>
  通过特定的本地 URL 格式访问设备。无需安装任何软件。
* **[方法 3：配置本地 DNS](#方法-3-配置本地-dns)**<br/>
  更新路由器或计算机的 DNS 设置，将标准 Olares URL 映射到本地 IP 地址。
* **[方法 4：修改 Hosts 文件](#方法-4-修改-hosts-文件)**<br/>
  在单台计算机上手动将标准 Olares URL 映射到本地 IP。

## 方法 1：启用 LarePass VPN
LarePass VPN 旨在兼顾连接安全与性能优化。启用后，LarePass 会自动检测设备是否处于同一网络，并切换至**内网**模式。

<!--@include: ../../reusables/larepass-vpn.md{19,24}-->

在用来访问 Olares 的当前设备上直接启用 LarePass VPN。

<!--@include: ../../reusables/larepass-vpn.md{26,50}-->

## 方法 2：使用 `.local` 域名

如果不希望安装额外应用，可使用 `.local` 域名访问服务。根据操作系统不同，有两种域名格式。

### 单级域名（所有操作系统适用）
:::warning 仅支持社区应用
Desktop 和文件管理器等 Olares 系统应用不支持此 URL 格式，因此无法正确加载。
:::
此格式通过连字符（`-`）连接入口 ID 和用户名来使用单级主机名。

**标准 URL**
```plain
https://<entrance_id>.<username>.olares.cn
```
**本地访问 URL**
```plain
http://<entrance_id>-<username>-olares.local
```

### 多级域名

The multi-level format below matches the structure of your standard Olares URL. Use it as shown.

<!--@include: ../../reusables/local-domain.md{7,23}-->

![多级域名](/images/manual/get-started/multilevel-local-domain-mac.png#bordered)

#### macOS and iOS
Apple devices support local service discovery via [Bonjour](https://developer.apple.com/bonjour/) (zero‑configuration networking), which can resolve multi‑label domains under `.local` on macOS and iOS.

Therefore, no extra setup is needed. You can directly use local URL in your browser.

#### Windows

<!--@include: ../../reusables/local-domain.md{26,40}-->

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

1. 找到 hosts 文件：
   - **Windows**：`C:\Windows\System32\drivers\etc\hosts`
   - **macOS/Linux**：`/etc/hosts`
2. 使用文本编辑器打开文件（需要管理员权限）。
3. 添加映射行：
    ```plain
    # 替换为实际内网 IP 和用户名
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
4. 保存文件以应用更改。这能确保即使在断网情况下，你也能进行本地访问。

你可以通过检查 URL 加载速度或使用终端来验证更改。例如：
```bash
ping desktop.<username>.olares.cn
```
如果返回的 IP 地址以 `192.168` 开头，即表示配置成功。


## 常见问题

<!--@include: ../../reusables/larepass-vpn.md{50,75}-->

<!--@include: ../../reusables/local-domain.md{42,75}-->