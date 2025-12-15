---
outline: [2, 3]
description: 本文介绍如何在任意环境安全访问 Olares。
---
# 通过 LarePass VPN 随时随地访问 Olares
Olares 设备上运行着 Vault 和 Ollama 等供个人或内部使用的关键应用。出于安全考量，这些应用只能通过[私有或内部入口](../../developer/concepts/network.md#内部入口) 访问。

为了获得最流畅的连接体验，建议你始终开启 LarePass VPN。开启后，LarePass 会利用 Tailscale 建立安全连接，并根据你所处的位置智能选择最快线路：

- **居家**： 通过局域网 (LAN) 直连，传输速度最快。
- **外出**： 与设备建立点对点 (P2P) 加密隧道直连。

若未开启 VPN，流量将经由 Cloudflare 或 FRP 等公网进行转发，速度可能会受到影响。

## 在 LarePass 中启用专用网络
:::info iOS 和 macOS 设置
在 iOS 或 macOS 上首次启用该功能时，系统可能会提示你添加 VPN 配置文件。允许此操作以完成设置。
:::

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

## 了解连接状态
LarePass 会展示设备到 Olares 的连接状态，帮助你判断或诊断网络情况。

![连接状态](/images/manual/larepass/connection-status.jpg)

| 状态          | 描述                                                   |
|--------------|--------------------------------------------------------|
| Internet     | 通过公网连接到 Olares                                  |
| Intranet     | 通过局域网连接到 Olares                                |
| FRP          | 通过 FRP 连接到 Olares                                 |
| DERP         | 通过专用网络使用 DERP 中继服务器连接到 Olares             |
| P2P          | 通过专用网络使用点对点 (P2P) 连接到 Olares                 |
| Offline mode | 当前离线，无法连接到 Olares                             |

::: info
若在外网环境使用专用网络访问私有入口时，连接状态显示 "DERP"，说明专用网络无法直接通过 P2P 连接 Olares，需要借助 Tailscale 中继服务器，这可能影响连接质量。如长期如此，请联系 Olares 支持。
:::

## 故障排查
出现连接问题时，LarePass 会显示诊断信息。常见提示及处理办法如下：

![异常消息](/images/zh/manual/larepass/abnormal-state.jpg)

| 状态提示               | 可能原因与解决方案                                                                                                                                                                                                                                                                                   |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 检测到网络问题，请检查本地网络设置。 | **本地网络异常**<br>1. 等待自动重连，系统会在网络恢复后同步数据。<br>2. 若问题持续，检查本地网络设置。                                                                                                                                                                                              |
| 需要专用网络连接至 Olares。  | **未开启专用网络**<br>点击通知横幅并按提示开启专用网络。                                                                                                                                                                                                                                                     |
| 需要重新登录 Olares。     | **会话过期或认证问题**<br>点击通知横幅并按提示重新登录。                                                                                                                                                                                                                                             |
| 需要重新连接到 Olares。    | **连接中断或超时**<br>点击通知横幅并按提示重新登录，登录后 Vault 数据将与服务器同步合并。                                                                                                                                                                                                           |
| 无活跃 Olares。        | **临时网络问题或 Olares 正在重启/关机**<br>等待自动恢复，通常会很快解决。<br>**Olares 实例已不存在**<br>1. 点击通知横幅并按提示重新激活 Olares、启用离线模式或忽略通知。<br>2. 若问题持续，请联系 Olares 管理员。                                                                         |
