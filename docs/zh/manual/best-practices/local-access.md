---
outline: [2,3]
description: 比较通过局域网直连 Olares 的不同方式。
head:
  - - meta
    - name: keywords
      content: Olares, 本地访问, LarePass 专用网络, hosts 映射, 本地 DNS, hosts 文件, .local 域名
---
# 在局域网内访问 Olares 服务

Olares 服务通常使用标准 `olares.com` 地址，可从本地或远程网络访问。当电脑或其他设备与 Olares 位于同一局域网时，可以选择本地路径，避免连接经过公网反向代理。

让流量留在局域网内可以降低延迟、提高传输速度，并在互联网不可用时继续访问服务。具体选择哪种方式，取决于你是否会切换网络、是否需要继续使用标准地址、是否要让多台设备统一生效，以及应用是否需要独立的局域网 IP。

## 选择本地访问方式

| 使用需求 | 推荐方式 | 地址 | 生效范围 |
|:---------|:---------|:-----|:---------|
| 经常在本地和远程网络之间切换 | [LarePass 专用网络](#使用-larepass-专用网络) | 标准 `olares.com` 地址 | 当前设备 |
| 在 Windows 或 macOS 上通过局域网直连 | [LarePass hosts 映射](#使用-larepass-配置-hosts-映射) | `olares.com` 或 `olares.local` | 当前电脑 |
| 不使用 LarePass 桌面端进行本地访问 | [`.local` 地址](#不使用-larepass-时使用-local-地址) | `olares.local` | 当前设备 |
| 为多台设备配置本地解析 | [本地 DNS](#配置本地-dns) | 标准 `olares.com` 地址 | 整个局域网 |
| 为支持的应用分配独立局域网 IP | [Overlay gateway](#通过-overlay-gateway-访问应用) | 应用专用 IP 地址 | 整个局域网 |

:::warning 避免绕行公网
在同一局域网内，如果使用标准 `olares.com` 地址，但没有启用专用网络、匹配的 hosts 条目或本地 DNS，连接可能会经过公网反向代理。服务仍可正常加载，但速度较慢，不建议将此路径用于本地访问。
:::

## 使用 LarePass 专用网络

如果经常切换网络，请使用此方式。LarePass 会自动选择连接类型。当设备与 Olares 位于同一局域网时，会切换到**内网**连接进行本地直连。

<!--@include: ../../reusables/larepass-vpn.md#vpn-setup-notes-->

<!--@include: ../../reusables/larepass-vpn.md#enable-larepass-vpn-->

<!--@include: ../../reusables/larepass-vpn.md#check-vpn-status-->

## 使用 LarePass 配置 hosts 映射

如果需要从 Windows 或 macOS 电脑通过局域网直连，同时不运行 LarePass 专用网络，请使用此方式。

<!--@include: ../../reusables/local-domain.md#local-domain-overview-->

<!--@include: ../../reusables/local-domain.md#larepass-local-domains-->

## 不使用 LarePass 时使用 .local 地址

当设备与 Olares 位于同一局域网，并且不想配置 LarePass 桌面端时，可以使用此方式。

### 多级域名

<!--@include: ../../reusables/local-domain.md#local-domain-url-format-->

在 macOS 和 iOS 上，本地服务发现功能无需额外配置即可解析多级 `.local` 主机名。在 Windows 上，请使用 [LarePass hosts 映射](#使用-larepass-配置-hosts-映射)。

### 单级域名

单级 `.local` 主机名适用于所有操作系统，但仅支持社区应用。Desktop 和文件管理器等 Olares 系统应用不支持此格式。

```text
http://<entrance_id>-<username>-olares.local
```

## 配置本地 DNS

如果希望多台设备继续使用标准 `olares.com` 地址，并将其解析到 Olares 的局域网 IP，请配置本地 DNS。此配置通常对整个局域网生效，无需在每台设备上安装 LarePass。

:::info
如果使用 `.local` 地址，请跳过此方式。`.local` 使用本地名称解析，不依赖 `olares.com` DNS 记录。
:::

### 查找 Olares 的局域网 IP

<tabs>
<template #使用-LarePass-手机端>

1. 确保手机与 Olares 位于同一网络。
2. 打开 LarePass，前往**设置** > **系统**。

   ![在 LarePass 中打开系统设置](/images/zh/manual/get-started/larepass-system.png#bordered)
3. 点击 Olares 设备卡片。

   ![打开 Olares 设备卡片](/images/zh/manual/get-started/larepass-device-card.png#bordered)
4. 在**网络**中找到**内网 IP**。

   ![查找内网 IP](/images/zh/manual/get-started/larepass-network.png#bordered)

</template>
<template #使用控制面板>

1. 在控制面板中打开**终端**，然后选择 **Olares**。

   ![在控制面板中打开 Olares 终端](/images/zh/manual/get-started/find-internal-ip-from-controlhub.png#bordered)
2. 运行 `ifconfig`。
3. 找到当前有线或 Wi-Fi 接口的 `inet` 地址。该地址通常属于 `192.168.x.x` 等私有地址范围。

</template>
</tabs>

### 配置 DNS 服务器

可以在一台电脑上配置 DNS，也可以在路由器上统一配置。

<tabs>
<template #一台电脑>

具体步骤取决于操作系统。以 macOS 为例：

1. 打开 Apple 菜单，进入**系统设置**。
2. 选择 **Wi-Fi**，然后点击已连接网络的**详细信息**。
3. 选择 **DNS**。
4. 在 **DNS 服务器**下添加 Olares 的局域网 IP，并将其移到列表顶部。
5. 在其下方保留原有 DNS，或添加 `1.1.1.1` 等公共 DNS 作为备用。
6. 点击**好**。

</template>
<template #所有设备>

1. 登录路由器管理页面。
2. 打开 DHCP 或 DNS 设置。
3. 将**首选 DNS** 设置为 Olares 的局域网 IP。
4. 将原来的首选 DNS 或 `1.1.1.1` 等公共 DNS 设置为**备用 DNS**。
5. 保存设置并重新连接客户端设备，使其获取更新后的 DNS 配置。

</template>
</tabs>

配置完成后，像往常一样打开标准 `olares.com` 地址。在同一网络内，这些地址应解析到 Olares 的局域网 IP。

:::tip
可以从 Olares 应用市场安装 AdGuard Home，通过图形界面监控流量并管理 DNS 映射。
:::

:::info 检查主机名解析
通过 LarePass 配置 hosts 映射或配置本地 DNS 后，在客户端电脑上解析一个标准服务主机名：

```bash
ping desktop.<username>.olares.com
```

返回的地址应与 Olares 的局域网 IP 一致。常见的私有局域网地址以 `192.168`、`10` 或 `172.16`–`172.31` 开头。
:::

## 通过 Overlay gateway 访问应用

某些应用需要在局域网中显示为独立设备，以支持设备发现、投屏、多人游戏连接或其他不使用 Olares 网页地址的协议。对于支持的应用，Overlay gateway 会通过虚拟网络接口为应用分配独立的局域网 IP。

此方式与 LarePass 专用网络、hosts 映射和本地 DNS 相互独立。访问时请使用应用显示的 IP 地址，而不是 `olares.com` 或 `olares.local` 地址。

Overlay gateway 要求 Olares 运行在使用有线网络连接的原生 Linux 主机上。有关可用性、权限和设置步骤，请参阅[管理应用的 Overlay gateway](../olares/settings/overlay-gateway.md)。

## 故障排除

### LarePass 管理的地址无法再通过本地网络打开

Olares 的局域网 IP 可能已经变化。请确保电脑与 Olares 位于同一局域网，关闭 **专用网络连接**，然后执行 LarePass 提供的 hosts 更新。

### LarePass 无法添加或更新 hosts 条目

请确保 LarePass 有权更新系统 hosts 文件。如需调整条目，请在 LarePass 的**更新 hosts 映射**弹窗中编辑，并保留以 `#` 开头的管理标记。

## 常见问题

<!--@include: ../../reusables/larepass-vpn.md#larepass-vpn-faq-->

<!--@include: ../../reusables/local-domain.md#larepass-local-domain-faq-->

<!--@include: ../../reusables/local-domain.md#local-domain-faq-->
