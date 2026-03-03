---
search: false
---
<!-- 可复用的 LarePass VPN 内容。按行号范围引用。
     步骤（无标题）：Step 1 7-16，Step 2 18-41，Step 3 42-49。
     常见问题：50-75 -->

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

如果之前能正常启用 VPN 但现在失效，可能需要重置系统扩展。

:::info
不同 macOS 版本下界面可能略有差异。
:::

1. 打开**系统设置**，搜索“扩展”，选择**登录项与扩展**。
2. 滚动到**网络扩展**，点击信息图标 (ⓘ) 查看已加载的扩展。
3. 找到 LarePass，点击三点 (...)，选择**删除扩展**。
4. 确认卸载。
5. 重启 Mac，在 LarePass 桌面客户端中重新启用 VPN。

### 为什么在 Windows 上无法启用 LarePass VPN？

第三方杀毒软件可能误将 LarePass 桌面客户端标记为可疑，导致无法启动 VPN 服务。

首次打开 LarePass 时如果杀毒软件有提示，选择允许应用继续运行。

如果 VPN 仍然无法启用：
1. 打开安全软件，查看是否拦截了 LarePass。
2. 将 LarePass 主程序加入杀毒软件的白名单或排除项。
3. 重启 LarePass 并再次启用 VPN。
