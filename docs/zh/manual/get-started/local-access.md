---
outline: [2,3]
description: 了解如何使用 LarePass VPN 安全访问 Olares 服务。
---
# 使用 LarePass VPN 安全访问 Olares 服务

通常，你会使用类似 `https://desktop.<username>.olares.cn` 的网址在浏览器中访问 Olares。这种方式允许你在任何设备上随时随地访问服务。

虽然该地址可从任意网络访问，但为了获得更安全、高效的连接，建议启用 LarePass VPN。客户端会自动检测网络环境并智能切换连接模式：
- **居家**：建立直连内网通道，大幅提升局域网文件传输速度。
- **远程**：切换至加密安全隧道，保障远程访问的数据安全。

## 下载 LarePass
要使用安全 VPN 连接，需先在当前设备上安装 LarePass 客户端。
- **移动端**：使用在创建 Olares ID 过程中安装的 LarePass 移动端。
- **桌面端**：下载并安装 LarePass 桌面客户端。

1. 访问 <AppLinkCN />。
2. 下载对应操作系统的版本。

## 启用 LarePass VPN
安装完成后，在当前设备上启用 VPN。

:::tip 始终启用 VPN 以进行远程访问
建议保持 LarePass VPN 开启。系统会自动优选最佳路由，让你无需手动切换即可确保持续获得最快速度。
:::
:::info iOS 和 macOS 设置
首次在 iOS 或 macOS 上开启此功能时，系统可能会弹窗请求添加 VPN 配置。允许此操作以完成设置。
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

## 确认连接状态
开启后，查看 LarePass 中的状态标签来确认当前的连接模式：

| 状态         | 描述                         |
|:-----------|:---------------------------|
| **内网**     | 通过本地局域网 IP 直连。速度最快。        |
| **P2P**    | 设备间的直接加密隧道。连接速度快。          |
| **DERP**   | 通过安全中继服务器路由，仅在无法直连时作为备用方案。 |

## 常见问题
### 为什么在 Mac 上无法再启用 LarePass VPN？
如果之前成功启用过 VPN，但现在停止工作，可能需要重置系统扩展。
:::info
根据 macOS 版本不同，界面可能略有差异。
:::
1. 打开**系统设置**，搜索`扩展`，选择**登录项与扩展**。
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

## 了解更多
- [本地访问 Olares 服务](../best-practices/local-access.md)：了解所有 Olares 本地连接方式。
- [网络](../../../zh/developer/concepts/network.md)：了解 Olares 中的应用的各类入口。