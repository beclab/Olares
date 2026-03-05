---
search: false
---
<!-- 可复用的 LarePass VPN 内容。按行号范围引用。
     步骤（无标题）：Step 1 7-16，Step 2 18-41，Step 3 42-49。
     常见问题：50-57 -->

要使用安全 VPN 连接，必须在用来访问 Olares 的设备上安装 LarePass 客户端。

- **移动端**：使用在创建 Olares ID 时安装的 LarePass 应用。
- **桌面端**：下载并安装 LarePass 桌面客户端。

1. 访问 <AppLinkCN />。
2. 下载与当前操作系统匹配的版本。
3. 安装应用并使用 Olares ID 登录。

安装完成后，在设备上直接启用 VPN。

:::tip 始终启用 VPN 以进行远程访问
保持 LarePass VPN 开启。它会自动优先选择最快可用路由，无需手动切换即可获得最佳速度。
:::
:::info iOS 和 macOS 设置
首次在 iOS 或 macOS 上启用时，系统可能会提示添加 VPN 配置。选择允许以完成设置。
:::

<tabs>
<template #使用-LarePass-移动端>

1. 打开 LarePass 应用，进入**设置**。
2. 在**我的 Olares** 卡片中，打开 VPN 开关。

   ![移动端开启 LarePass VPN](/images/zh/manual/get-started/larepass-vpn-mobile.png#bordered)
</template>
<template #使用-LarePass-桌面端>

1. 打开 LarePass 应用，点击左上角头像打开用户菜单。
2. 打开**专用网络连接**开关。

   ![桌面端开启 LarePass VPN](/images/zh/manual/get-started/larepass-vpn-desktop.png#bordered)
</template>
</tabs>

启用后，在 LarePass 中查看状态指示以确认连接类型：

| 状态       | 描述                                             |
|:-----------|:--------------------------------------------------|
| **内网**   | 通过本地局域网 IP 直连，速度最快。                 |
| **P2P**    | 设备间直接加密隧道，速度较快。                    |
| **DERP**   | 经安全中继服务器路由，作为备用。                  |

### 为什么在 Mac 上无法再启用 LarePass VPN？

如果网络扩展或 VPN 配置未完整设置，或网络扩展出现卡死、损坏，macOS 会阻止 LarePass 建立 VPN 隧道。参考 [LarePass VPN 无法使用](/zh/manual/help/ts-larepass-vpn-not-working)，重置扩展并恢复 VPN。

### 为什么在 Windows 上无法启用 LarePass VPN？

第三方杀毒或安全软件可能误将 LarePass 标记为可疑程序，导致 VPN 服务无法启动。参考 [LarePass VPN 无法使用](/zh/manual/help/ts-larepass-vpn-not-working) 解决该问题。
